package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/luchkonikita/canary/store"
	ts "github.com/luchkonikita/canary/test_support"
)

func TestCrawlerWorkerWork(t *testing.T) {
	workTimeout = time.Millisecond

	// Mock database
	db := store.NewDB(ts.GetTestDBName(), true)
	defer db.Close()

	// Mock server
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Token") == "CorrectToken" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "Yay")
			} else {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, "Noo")
			}
		}),
	)
	defer server.Close()

	// Create entities
	var pageResults []store.PageResult

	sitemap := store.Sitemap{
		Name: "The name",
		URL:  "http://example.com/sitemap.xml",
	}

	err := db.Save(&sitemap)
	ts.Assert(t, err == nil, "Expected to create a sitemap")

	NewCrawlerWorker(db, sitemap).Work(3)

	db.All(&pageResults)
	ts.Assert(t, len(pageResults) == 0, "Expected worked to idle when no page results are pending")

	crawling := store.Crawling{
		SitemapID: 1,
		Headers: []store.CrawlingHeader{
			store.CrawlingHeader{Name: "Token", Value: "CorrectToken"},
		},
	}
	err = db.Save(&crawling)
	ts.Assert(t, err == nil, "Expected to create a crawling")

	pageResult := store.PageResult{
		CrawlingID: 1,
		URL:        server.URL,
	}
	err = db.Save(&pageResult)
	ts.Assert(t, err == nil, "Expected to create a page result")

	NewCrawlerWorker(db, sitemap).Work(3)

	db.All(&pageResults)
	ts.Assert(t, pageResults[0].Status == 200, "Expected to request the page and store the result")

	db.One("ID", 1, &crawling)
	ts.Assert(t, crawling.Processed, "Expected to mark crawling as processed")
}
