package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

func testRequest(router *mux.Router, verb, path string, params url.Values) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	if verb == "POST" || verb == "PUT" {
		req, _ := http.NewRequest(verb, path, strings.NewReader(params.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(rr, req)
	} else {
		req, _ := http.NewRequest(verb, path+"?"+params.Encode(), nil)
		router.ServeHTTP(rr, req)
	}

	return rr
}

// func dbContains(db *storm.DB, sType interface{}, count int) bool {
// 	dbCount, _ := db.Count(sType)
// 	return dbCount == count
// }

func TestGetCrawlings(t *testing.T) {
	app := mockApp()
	app.useRoutes()
	defer app.db.Close()

	app.db.Save(&crawling{URL: "http://example.com/sitemap-1.xml", Processed: false})
	app.db.Save(&crawling{URL: "http://example.com/sitemap-2.xml", Processed: true})

	res := testRequest(app.router, "GET", "/crawlings", url.Values{})

	assert.Contains(t, res.Body.String(), "\"id\":1")
	assert.Contains(t, res.Body.String(), "\"id\":2")
}

func TestCreateCrawling(t *testing.T) {
	app := mockApp()
	app.useRoutes()
	defer app.db.Close()

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

	res := testRequest(app.router, "POST", "/crawlings", url.Values{})
	assert.Equal(t, res.Code, 400)

	res = testRequest(app.router, "POST", "/crawlings", url.Values{"url": {"NO"}, "concurrency": {"1"}})
	assert.Equal(t, res.Code, 400)

	res = testRequest(app.router, "POST", "/crawlings", url.Values{"url": {sitemapServer.URL}, "concurrency": {"1"}, "headers.0.name": {"Token"}, "headers.0.value": {"CorrectToken"}})
	assert.Equal(t, res.Code, 201)

	var cr crawling
	app.db.One("URL", sitemapServer.URL, &cr)

	assert.Equal(t, cr.Processed, false)
	assert.Equal(t, cr.Headers[0].Name, "Token")
	assert.Equal(t, cr.Headers[0].Value, "CorrectToken")
	assert.Equal(t, cr.PageResults[0].URL, "http://google.com/maps")
	assert.Equal(t, cr.PageResults[1].URL, "http://google.com/docs")
}

func TestGetCrawling(t *testing.T) {
	app := mockApp()
	app.useRoutes()
	defer app.db.Close()

	app.db.Save(&crawling{URL: "http://example.com/sitemap.xml", Processed: false, CreatedAt: time.Now()})

	res := testRequest(app.router, "GET", "/crawlings/2", url.Values{})
	assert.Equal(t, res.Code, 404)

	res = testRequest(app.router, "GET", "/crawlings/1", url.Values{})
	assert.Equal(t, res.Code, 200)

	assert.Contains(t, res.Body.String(), "\"id\":1")
	assert.Contains(t, res.Body.String(), "\"processed\":false")
	assert.Contains(t, res.Body.String(), "\"url\":\"http://example.com/sitemap.xml\"")
}

func TestDeleteCrawling(t *testing.T) {
	app := mockApp()
	app.useRoutes()
	defer app.db.Close()

	app.db.Save(&crawling{URL: "http://example.com/sitemap.xml", Processed: false, CreatedAt: time.Now()})
	count, _ := app.db.Count(&crawling{})
	assert.Equal(t, count, 1)

	res := testRequest(app.router, "DELETE", "/crawlings/2", url.Values{})
	assert.Equal(t, res.Code, 404)

	count, _ = app.db.Count(&crawling{})
	assert.Equal(t, count, 1)

	res = testRequest(app.router, "DELETE", "/crawlings/1", url.Values{})
	assert.Equal(t, res.Code, 200)

	count, _ = app.db.Count(&crawling{})
	assert.Equal(t, count, 0)
}
