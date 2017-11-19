package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luchkonikita/canary/store"
	ts "github.com/luchkonikita/canary/test_support"

	"github.com/asdine/storm"

	"github.com/julienschmidt/httprouter"
)

var emptyBody = map[string]string{}
var emptyParams = map[string]string{}

type paramsMap map[string]string
type curriedRequest func(params map[string]string) *httptest.ResponseRecorder

func curryRequest(handlerFn Handler, db *storm.DB) curriedRequest {
	return func(params map[string]string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := ts.NewTestRequest(emptyBody)
		routerParams := httprouter.Params{}
		for key, value := range params {
			routerParams = append(routerParams, httprouter.Param{Key: key, Value: value})
		}
		handlerFn(db)(recorder, request, routerParams)
		return recorder
	}
}

func TestGetSitemaps(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	doRequest := curryRequest(GetSitemaps, db)

	db.Save(&store.Sitemap{
		Name: "The name",
		URL:  "http://example.com",
	})

	if len(store.GetSitemaps(db)) != 1 {
		t.Error("Database should contain 1 sitemap")
	}

	responseBody := doRequest(emptyParams).Body.String()

	if !strings.Contains(responseBody, "\"Name\":\"The name\"") {
		t.Error("Expected response to contain sitemap name")
	}
	if !strings.Contains(responseBody, "\"URL\":\"http://example.com\"") {
		t.Error("Expected response to contain sitemap url")
	}
}

func TestCreateSitemap(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	doRequest := func(params map[string]string) {
		request := ts.NewTestRequest(params)
		CreateSitemap(db)(httptest.NewRecorder(), request, httprouter.Params{})
	}

	blankParams := map[string]string{}
	fullParams := map[string]string{
		"name": "The Name",
		"url":  "http://example/sitemap.xml",
	}

	ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Database should contain 0 sitemaps")

	doRequest(blankParams)
	ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Sitemap should not be created with blank name")

	doRequest(fullParams)
	sitemaps := store.GetSitemaps(db)
	ts.Assert(t, len(sitemaps) == 1, "CreateSitemap should create 1 sitemap")
	ts.Assert(t, sitemaps[0].Name == "The Name", "Expected to create sitemap with name 'The name'")
	ts.Assert(t, sitemaps[0].URL == "http://example/sitemap.xml", "Expected to create sitemap with url 'http://example/sitemap.xml'")

	doRequest(fullParams)
	ts.Assert(t, len(store.GetSitemaps(db)) == 1, "CreateSitemap should not duplicate a sitemap")
}

func TestDeleteSitemap(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	doRequest := curryRequest(DeleteSitemap, db)

	ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Database should contain 0 sitemaps")

	recorder := doRequest(paramsMap{"sitemapId": "1"})

	ts.Assert(t, recorder.Code == 404, "Expected to have 404 response")

	db.Save(&store.Sitemap{
		Name: "The name",
		URL:  "http://example.com",
	})

	ts.Assert(t, len(store.GetSitemaps(db)) == 1, "Database should contain 1 sitemap")

	recorder = doRequest(paramsMap{"sitemapId": "1"})

	ts.Assert(t, recorder.Code == 200, "Expected to have 200 response")
	ts.Assert(t, len(store.GetSitemaps(db)) == 0, "Database should contain 0 sitemaps")
}

func TestCreateCrawling(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	doRequest := curryRequest(CreateCrawling, db)

	recorder := doRequest(map[string]string{"sitemapId": "1"})
	ts.Assert(t, recorder.Code == 404, "Expected to have 404 response")

	server := ts.NewServer(ts.SitemapXML, 200)
	defer server.Close()

	db.Save(&store.Sitemap{
		Name: "The name",
		URL:  server.URL,
	})

	db.Save(&store.Sitemap{
		Name: "Another name",
		URL:  "NOT WORKING URL",
	})

	recorder = doRequest(paramsMap{"sitemapId": "2"})
	ts.Assert(t, recorder.Code == 400, "Expected to have 400 response when URL is broken")

	recorder = doRequest(paramsMap{"sitemapId": "1"})
	ts.Assert(t, recorder.Code == 200, "Expected to have 200 response when URL is fine")

	var crawling store.Crawling
	ts.Assert(t, db.One("SitemapID", 1, &crawling) == nil, "Expected to create crawling")

	var pageResults []store.PageResult
	db.All(&pageResults)

	ts.Assert(t, pageResults[0].CrawlingID == crawling.ID, "Expected to create first page result")
	ts.Assert(t, pageResults[1].CrawlingID == crawling.ID, "Expected to create second page result")
}

func TestGetCrawlings(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	sitemap := &store.Sitemap{
		Name: "The name",
		URL:  "http://example.com/sitemap.xml",
	}

	db.Save(sitemap)

	store.CreateCrawling(db, sitemap, []string{
		"http://example.com/first",
	})

	doRequest := curryRequest(GetCrawlings, db)

	responseBody := doRequest(paramsMap{"sitemapId": "2"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "error", "\"not found\""), "Expected to return an error for unexisting crawling")

	responseBody = doRequest(paramsMap{"sitemapId": "1"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "ID", "1"), "Expected response to contain crawlings IDs")
	ts.Assert(t, ts.ContainsJSON(responseBody, "SitemapID", "1"), "Expected response to contain sitemap ID")
	ts.Assert(t, ts.ContainsJSON(responseBody, "Processed", "false"), "Expected response to processed flag")
}

func TestDeleteCrawling(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	sitemap := &store.Sitemap{
		Name: "The name",
		URL:  "http://example.com/sitemap.xml",
	}

	db.Save(sitemap)

	store.CreateCrawling(db, sitemap, []string{
		"http://example.com/first",
	})

	doRequest := curryRequest(DeleteCrawling, db)

	ts.Assert(t, db.One("ID", 1, &store.Crawling{}) == nil, "Expected DB to have a crawling with ID 1")
	ts.Assert(t, db.One("ID", 1, &store.PageResult{}) == nil, "Expected DB to have a page result with ID 1")

	responseBody := doRequest(paramsMap{"sitemapId": "1", "crawlingId": "2"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "error", "\"not found\""), "Expected to return an error for unexisting crawling")

	responseBody = doRequest(paramsMap{"sitemapId": "2", "crawlingId": "1"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "error", "\"not found\""), "Expected to return an error for unexisting sitemap")

	responseBody = doRequest(paramsMap{"sitemapId": "1", "crawlingId": "1"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "status", "\"Deleted\""), "Expected to return a status")
	ts.Assert(t, db.One("ID", 1, &store.Crawling{}) != nil, "Expected to delete a crawling with ID 1")
	ts.Assert(t, db.One("ID", 1, &store.PageResult{}) != nil, "Expected to delete a page result with ID 1")
}

func TestGetPageResults(t *testing.T) {
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	// Create entities
	err := db.Save(&store.Sitemap{
		Name:     "The name",
		URL:      "http://example.com/sitemap.xml",
		Username: "USER",
		Password: "PASSWORD",
	})
	ts.Assert(t, err == nil, "Expected to create a sitemap")

	err = db.Save(&store.Crawling{
		SitemapID: 1,
	})
	ts.Assert(t, err == nil, "Expected to create a crawling")

	err = db.Save(&store.PageResult{
		CrawlingID: 1,
		URL:        "http://example.com/page",
	})
	ts.Assert(t, err == nil, "Expected to create a page result")

	// Curry request
	doRequest := curryRequest(GetPageResults, db)

	responseBody := doRequest(paramsMap{"crawlingId": "2"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "error", "\"not found\""), "Expected to return an error for unexisting crawling")

	responseBody = doRequest(paramsMap{"crawlingId": "1"}).Body.String()
	ts.Assert(t, ts.ContainsJSON(responseBody, "URL", "\"http://example.com/page\""), "Expected to response to contain page URL")
	ts.Assert(t, ts.ContainsJSON(responseBody, "Status", "0"), "Expected to response to contain page status code")
}
