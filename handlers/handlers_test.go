package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/luchkonikita/canary/store"
	ts "github.com/luchkonikita/canary/test_support"

	"github.com/asdine/storm"

	"github.com/julienschmidt/httprouter"
)

func requestCase(testFn func(db *storm.DB, router *httprouter.Router)) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()
	router := NewRouter(db)
	testFn(db, router)
}

func request(router *httprouter.Router, verb, path string, params url.Values) *httptest.ResponseRecorder {
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
		res := request(router, "GET", "/", url.Values{}).Body.String()

		ts.Assert(t, ts.ContainsJSON(res, "alive", "true"), "Expected response to contain server status")
	})
}

func TestGetSitemaps(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&store.Sitemap{Name: "The name", URL: "http://example.com"})

		res := request(router, "GET", "/sitemaps", url.Values{}).Body.String()

		ts.Assert(t, ts.ContainsJSON(res, "Name", "\"The name\""), "Expected response to contain sitemap name")
		ts.Assert(t, ts.ContainsJSON(res, "URL", "\"http://example.com\""), "Expected response to contain sitemap URL")
	})
}

func TestCreateSitemap(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Database should contain 0 sitemaps")

		rr := request(router, "POST", "/sitemaps", url.Values{})

		ts.Assert(t, rr.Code == 422, "Expected to return an error")
		ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Sitemap should not be created with blank name")

		validParams := url.Values{"name": {"The Name"}, "url": {"http://example/sitemap.xml"}}

		rr = request(router, "POST", "/sitemaps", validParams)
		sitemaps := store.GetSitemaps(db)

		ts.Assert(t, len(sitemaps) == 1, "CreateSitemap should create 1 sitemap")
		ts.Assert(t, sitemaps[0].Name == "The Name", "Expected to create sitemap with name 'The name'")
		ts.Assert(t, sitemaps[0].URL == "http://example/sitemap.xml", "Expected to create sitemap with url 'http://example/sitemap.xml'")

		rr = request(router, "POST", "/sitemaps", validParams)
		ts.Assert(t, len(store.GetSitemaps(db)) == 1, "CreateSitemap should not duplicate a sitemap")
	})
}

func TestUpdateSitemap(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Database should contain 0 sitemaps")

		db.Save(&store.Sitemap{
			Name: "Old name",
			URL:  "http://old.com",
		})

		ts.Assert(t, len(store.GetSitemaps(db)) == 1, "Database should contain 1 sitemap")

		validParams := url.Values{"name": {"New name"}, "url": {"http://new.com"}}

		rr := request(router, "PATCH", "/sitemaps/2", validParams)
		ts.Assert(t, rr.Code == 404, "Expected to have 404 response")

		var sitemap store.Sitemap
		db.One("ID", 1, &sitemap)
		ts.Assert(t, sitemap.Name == "Old name", "Should not update sitemap")
		ts.Assert(t, sitemap.URL == "http://old.com", "Should not update sitemap")

		rr = request(router, "PATCH", "/sitemaps/1", validParams)
		db.One("ID", 1, &sitemap)
		ts.Assert(t, rr.Code == 200, "Expected to have 200 response")
		ts.Assert(t, sitemap.Name == "New name", "Should update sitemap")
		ts.Assert(t, sitemap.URL == "http://new.com", "Should update sitemap")
	})
}

func TestDeleteSitemap(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		ts.Assert(t, dbContains(db, &store.Sitemap{}, 0), "Database should contain 0 sitemaps")
		ts.Assert(t, dbContains(db, &store.Crawling{}, 0), "Database should contain 0 crawlings")
		ts.Assert(t, dbContains(db, &store.PageResult{}, 0), "Database should contain 0 page results")

		db.Save(&store.Sitemap{Name: "The name", URL: "http://example.com"})
		db.Save(&store.Crawling{SitemapID: 1})
		db.Save(&store.PageResult{CrawlingID: 1})

		ts.Assert(t, dbContains(db, &store.Sitemap{}, 1), "Database should contain 1 sitemap")
		ts.Assert(t, dbContains(db, &store.Crawling{}, 1), "Database should contain 1 crawling")
		ts.Assert(t, dbContains(db, &store.PageResult{}, 1), "Database should contain 1 page result")

		rr := request(router, "DELETE", "/sitemaps/2", url.Values{})
		ts.Assert(t, rr.Code == 404, "Expected to have 404 response")

		rr = request(router, "DELETE", "/sitemaps/1", url.Values{})
		ts.Assert(t, rr.Code == 200, "Expected to have 200 response")

		ts.Assert(t, dbContains(db, &store.Sitemap{}, 0), "Database should contain 0 sitemaps")
		ts.Assert(t, dbContains(db, &store.Crawling{}, 0), "Database should contain 0 crawlings")
		ts.Assert(t, dbContains(db, &store.PageResult{}, 0), "Database should contain 0 page results")
	})
}

func TestCreateCrawling(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		sitemapServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, ts.SitemapXML)
			}),
		)
		defer sitemapServer.Close()

		rr := request(router, "POST", "/crawlings", url.Values{})
		ts.Assert(t, rr.Code == 400, "Expected to have 400 response")

		rr = request(router, "POST", "/crawlings", url.Values{"sitemap_id": {"1"}})
		ts.Assert(t, rr.Code == 400, "Expected to have 400 response")

		db.Save(&store.Sitemap{Name: "The name", URL: sitemapServer.URL})
		db.Save(&store.Sitemap{Name: "Another name", URL: "NOT WORKING URL"})

		ts.Assert(t, dbContains(db, &store.Crawling{}, 0), "Expected to DB to have 0 crawlings")
		ts.Assert(t, dbContains(db, &store.PageResult{}, 0), "Expected to DB to have 0 page results")

		rr = request(router, "POST", "/crawlings", url.Values{"sitemap_id": {"2"}})
		ts.Assert(t, rr.Code == 400, "Expected to have 400 response when URL is broken")

		rr = request(router, "POST", "/crawlings", url.Values{"sitemap_id": {"1"}, "headers.0.name": {"HeaderName"}, "headers.0.value": {"HeaderValue"}})
		ts.Assert(t, rr.Code == 200, "Expected to have 200 response when URL is fine")

		var crawling store.Crawling
		var pageResults []store.PageResult
		db.One("SitemapID", 1, &crawling)
		db.All(&pageResults)

		ts.Assert(t, crawling.Processed == false, "Expected to create a crawling with processed as false")
		ts.Assert(t, crawling.Headers[0].Name == "HeaderName" && crawling.Headers[0].Value == "HeaderValue", "Expected to save crawling header")
		ts.Assert(t, pageResults[0].CrawlingID == crawling.ID, "Expected to create first page result")
		ts.Assert(t, pageResults[1].CrawlingID == crawling.ID, "Expected to create second page result")
	})
}

func TestGetCrawlings(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&store.Crawling{SitemapID: 1, Processed: false})
		db.Save(&store.Crawling{SitemapID: 2, Processed: true})

		res := request(router, "GET", "/crawlings", url.Values{"sitemap_id": {"1"}}).Body.String()
		ts.Assert(t, ts.ContainsJSON(res, "ID", "1"), "Expected response to contain only first crawling")
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "2"), "Expected response to contain only first crawling")

		res = request(router, "GET", "/crawlings", url.Values{"processed": {"true"}}).Body.String()
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "1"), "Expected response to contain only second crawling")
		ts.Assert(t, ts.ContainsJSON(res, "ID", "2"), "Expected response to contain only second crawling")
	})
}

func TestDeleteCrawling(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&store.Crawling{SitemapID: 1, Processed: false, CreatedAt: time.Now()})
		ts.Assert(t, dbContains(db, &store.Crawling{}, 1), "Expected to DB to have 1 crawling")

		rr := request(router, "DELETE", "/crawlings/2", url.Values{})
		ts.Assert(t, rr.Code == 400, "Expected response to have 404 status")

		ts.Assert(t, dbContains(db, &store.Crawling{}, 1), "Expected to DB to have 1 crawling")

		rr = request(router, "DELETE", "/crawlings/1", url.Values{})
		ts.Assert(t, rr.Code == 200, "Expected response to have 200 status")

		ts.Assert(t, dbContains(db, &store.Crawling{}, 0), "Expected to delete a crawling")
	})
}

func TestGetPageResults(t *testing.T) {
	requestCase(func(db *storm.DB, router *httprouter.Router) {
		db.Save(&store.PageResult{CrawlingID: 1, URL: "/first", Status: 200})
		db.Save(&store.PageResult{CrawlingID: 2, URL: "/second", Status: 404})
		db.Save(&store.PageResult{CrawlingID: 3, URL: "/third", Status: 500})

		res := request(router, "GET", "/page_results", url.Values{"crawling_id": {"1"}}).Body.String()
		ts.Assert(t, ts.ContainsJSON(res, "ID", "1"), "Expected to return only first page result")
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "2"), "Expected to return only first page result")
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "3"), "Expected to return only first page result")

		res = request(router, "GET", "/page_results", url.Values{"status": {"404"}}).Body.String()
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "1"), "Expected to return only second page result")
		ts.Assert(t, ts.ContainsJSON(res, "ID", "2"), "Expected to return only second page result")
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "3"), "Expected to return only second page result")

		res = request(router, "GET", "/page_results", url.Values{"url": {"third"}}).Body.String()
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "1"), "Expected to return only third page result")
		ts.Assert(t, !ts.ContainsJSON(res, "ID", "2"), "Expected to return only third page result")
		ts.Assert(t, ts.ContainsJSON(res, "ID", "3"), "Expected to return only third page result")
	})
}
