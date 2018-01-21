package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

func requestCase(testFn func(db *storm.DB, router *httprouter.Router)) {
	db := NewDB("test_storage.db", true)
	defer db.Close()
	router := NewRouter(db)
	testFn(db, router)
}

func testRequest(router *httprouter.Router, verb, path string, params url.Values) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(verb, path+"?"+params.Encode(), nil)
	router.ServeHTTP(rr, req)
	return rr
}

func dbContains(db *storm.DB, sType interface{}, count int) bool {
	dbCount, _ := db.Count(sType)
	return dbCount == count
}

func TestPing(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		res := testRequest(router, "GET", "/", url.Values{}).Body.String()

		Assert(t, ContainsJSON(res, "alive", "true"), "Expected response to contain server status")
	})
}

func TestGetCrawlings(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&Crawling{URL: "http://example.com/sitemap-1.xml", Processed: false})
		db.Save(&Crawling{URL: "http://example.com/sitemap-2.xml", Processed: true})

		res := testRequest(router, "GET", "/crawlings", url.Values{"url": {"http://example.com/sitemap-1.xml"}}).Body.String()
		Assert(t, ContainsJSON(res, "ID", "1"), "Expected response to contain only first crawling")
		Assert(t, !ContainsJSON(res, "ID", "2"), "Expected response to contain only first crawling")

		res = testRequest(router, "GET", "/crawlings", url.Values{"processed": {"true"}}).Body.String()
		Assert(t, !ContainsJSON(res, "ID", "1"), "Expected response to contain only second crawling")
		Assert(t, ContainsJSON(res, "ID", "2"), "Expected response to contain only second crawling")
	})
}

func TestCreateCrawling(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		sitemapServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Token") == "CorrectToken" {
					fmt.Fprintln(w, TestSitemapXML)
				} else {
					fmt.Fprintln(w, "")
				}
			}),
		)
		defer sitemapServer.Close()

		Assert(t, dbContains(db, &Crawling{}, 0), "Expected to DB to have 0 crawlings")
		Assert(t, dbContains(db, &PageResult{}, 0), "Expected to DB to have 0 page results")

		rr := testRequest(router, "POST", "/crawlings", url.Values{})
		Assert(t, rr.Code == 400, "Expected to have 400 response")

		rr = testRequest(router, "POST", "/crawlings", url.Values{"url": {"NO"}, "concurrency": {"1"}})
		Assert(t, rr.Code == 400, "Expected to have 400 response when URL is broken")

		rr = testRequest(router, "POST", "/crawlings", url.Values{"url": {sitemapServer.URL}, "concurrency": {"1"}, "headers.0.name": {"Token"}, "headers.0.value": {"CorrectToken"}})
		Assert(t, rr.Code == 200, "Expected to have 200 response when URL is fine")

		var crawling Crawling
		var pageResults []PageResult
		db.One("URL", sitemapServer.URL, &crawling)
		db.All(&pageResults)

		Assert(t, crawling.Processed == false, "Expected to create a crawling with processed as false")
		Assert(t, crawling.Headers[0].Name == "Token" && crawling.Headers[0].Value == "CorrectToken", "Expected to save crawling header")
		Assert(t, pageResults[0].CrawlingID == crawling.ID, "Expected to create first page result")
		Assert(t, pageResults[1].CrawlingID == crawling.ID, "Expected to create second page result")
	})
}

func TestDeleteCrawling(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&Crawling{URL: "http://example.com/sitemap.xml", Processed: false, CreatedAt: time.Now()})
		Assert(t, dbContains(db, &Crawling{}, 1), "Expected to DB to have 1 crawling")

		rr := testRequest(router, "DELETE", "/crawlings/2", url.Values{})
		Assert(t, rr.Code == 400, "Expected response to have 404 status")

		Assert(t, dbContains(db, &Crawling{}, 1), "Expected to DB to have 1 crawling")

		rr = testRequest(router, "DELETE", "/crawlings/1", url.Values{})
		Assert(t, rr.Code == 200, "Expected response to have 200 status")

		Assert(t, dbContains(db, &Crawling{}, 0), "Expected to delete a crawling")
	})
}

func TestGetPageResults(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&PageResult{CrawlingID: 1, URL: "/first", Status: 200})
		db.Save(&PageResult{CrawlingID: 2, URL: "/second", Status: 404})
		db.Save(&PageResult{CrawlingID: 3, URL: "/third", Status: 500})

		res := testRequest(router, "GET", "/page_results", url.Values{"crawling_id": {"1"}}).Body.String()
		Assert(t, ContainsJSON(res, "ID", "1"), "Expected to return only first page result")
		Assert(t, !ContainsJSON(res, "ID", "2"), "Expected to return only first page result")
		Assert(t, !ContainsJSON(res, "ID", "3"), "Expected to return only first page result")

		res = testRequest(router, "GET", "/page_results", url.Values{"status": {"404"}}).Body.String()
		Assert(t, !ContainsJSON(res, "ID", "1"), "Expected to return only second page result")
		Assert(t, ContainsJSON(res, "ID", "2"), "Expected to return only second page result")
		Assert(t, !ContainsJSON(res, "ID", "3"), "Expected to return only second page result")

		res = testRequest(router, "GET", "/page_results", url.Values{"url": {"third"}}).Body.String()
		Assert(t, !ContainsJSON(res, "ID", "1"), "Expected to return only third page result")
		Assert(t, !ContainsJSON(res, "ID", "2"), "Expected to return only third page result")
		Assert(t, ContainsJSON(res, "ID", "3"), "Expected to return only third page result")
	})
}
