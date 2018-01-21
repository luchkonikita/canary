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
	"github.com/stretchr/testify/assert"
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

		assert.Contains(t, res, "\"alive\":true")
	})
}

func TestGetCrawlings(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&Crawling{URL: "http://example.com/sitemap-1.xml", Processed: false})
		db.Save(&Crawling{URL: "http://example.com/sitemap-2.xml", Processed: true})

		res := testRequest(router, "GET", "/crawlings", url.Values{"url": {"http://example.com/sitemap-1.xml"}}).Body.String()
		assert.Contains(t, res, "\"ID\":1")
		assert.NotContains(t, res, "\"ID\":2")

		res = testRequest(router, "GET", "/crawlings", url.Values{"processed": {"true"}}).Body.String()
		assert.Contains(t, res, "\"ID\":2")
		assert.NotContains(t, res, "\"ID\":1")
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

		assert.True(t, dbContains(db, &Crawling{}, 0))
		assert.True(t, dbContains(db, &PageResult{}, 0))

		rr := testRequest(router, "POST", "/crawlings", url.Values{})
		assert.Equal(t, rr.Code, 400)

		rr = testRequest(router, "POST", "/crawlings", url.Values{"url": {"NO"}, "concurrency": {"1"}})
		assert.Equal(t, rr.Code, 400)

		rr = testRequest(router, "POST", "/crawlings", url.Values{"url": {sitemapServer.URL}, "concurrency": {"1"}, "headers.0.name": {"Token"}, "headers.0.value": {"CorrectToken"}})
		assert.Equal(t, rr.Code, 200)

		var crawling Crawling
		var pageResults []PageResult
		db.One("URL", sitemapServer.URL, &crawling)
		db.All(&pageResults)

		assert.Equal(t, crawling.Processed, false)
		assert.Equal(t, crawling.Headers[0].Name, "Token")
		assert.Equal(t, crawling.Headers[0].Value, "CorrectToken")
		assert.Equal(t, pageResults[0].CrawlingID, crawling.ID)
		assert.Equal(t, pageResults[1].CrawlingID, crawling.ID)
	})
}

func TestDeleteCrawling(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&Crawling{URL: "http://example.com/sitemap.xml", Processed: false, CreatedAt: time.Now()})
		assert.True(t, dbContains(db, &Crawling{}, 1))

		rr := testRequest(router, "DELETE", "/crawlings/2", url.Values{})
		assert.Equal(t, rr.Code, 400)
		assert.True(t, dbContains(db, &Crawling{}, 1))

		rr = testRequest(router, "DELETE", "/crawlings/1", url.Values{})
		assert.Equal(t, rr.Code, 200)
		assert.True(t, dbContains(db, &Crawling{}, 0))
	})
}

func TestGetPageResults(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&PageResult{CrawlingID: 1, URL: "/first", Status: 200})
		db.Save(&PageResult{CrawlingID: 2, URL: "/second", Status: 404})
		db.Save(&PageResult{CrawlingID: 3, URL: "/third", Status: 500})

		res := testRequest(router, "GET", "/page_results", url.Values{"crawling_id": {"1"}}).Body.String()
		assert.Contains(t, res, "\"ID\":1")
		assert.NotContains(t, res, "\"ID\":2")
		assert.NotContains(t, res, "\"ID\":3")

		res = testRequest(router, "GET", "/page_results", url.Values{"status": {"404"}}).Body.String()
		assert.NotContains(t, res, "\"ID\":1")
		assert.Contains(t, res, "\"ID\":2")
		assert.NotContains(t, res, "\"ID\":3")

		res = testRequest(router, "GET", "/page_results", url.Values{"url": {"third"}}).Body.String()
		assert.NotContains(t, res, "\"ID\":1")
		assert.NotContains(t, res, "\"ID\":2")
		assert.Contains(t, res, "\"ID\":3")
	})
}
