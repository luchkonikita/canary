package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/test_support"
)

func TestNewRouter(t *testing.T) {
	db := store.NewDB(test_support.GetTestDBName(), true)
	defer db.Close()

	router := NewRouter(db)

	makeRequest := func(method string, path string) int {
		// Prepare records
		db.Save(&store.Sitemap{
			Name: "The name",
			URL:  "http://example.com",
		})
		db.Save(&store.Crawling{
			SitemapID: 1,
		})

		// Make request
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(""))
		router.ServeHTTP(recorder, request)

		// Clean database
		db.Drop(&store.Sitemap{})
		db.Drop(&store.Crawling{})

		// Return the code
		return recorder.Code
	}

	var routes = []struct {
		verb string
		path string
	}{
		{http.MethodGet, "/sitemaps"},
		{http.MethodPost, "/sitemaps"},
		{http.MethodDelete, "/sitemaps/1"},
		{http.MethodGet, "/sitemaps/1/crawlings"},
		{http.MethodPost, "/sitemaps/1/crawlings"},
		{http.MethodDelete, "/sitemaps/1/crawlings/1"},
		{http.MethodGet, "/crawlings/1/page_results"},
	}

	for _, route := range routes {
		if makeRequest(route.verb, route.path) == 404 {
			t.Errorf("Expected router to handle %s via %s", route.path, route.verb)
		}
	}
}
