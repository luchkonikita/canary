package workers

import (
	"net/http"
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
	server := ts.NewServer("Yay", http.StatusGone)
	defer server.Close()

	// Create entities
	var pageResults []store.PageResult

	sitemap := store.Sitemap{
		Name:     "The name",
		URL:      "http://example.com/sitemap.xml",
		Username: "USER",
		Password: "PASSWORD",
	}

	err := db.Save(&sitemap)
	ts.Assert(t, err == nil, "Expected to create a sitemap")

	NewCrawlerWorker(db, sitemap).Work(3)

	db.All(&pageResults)
	ts.Assert(t, len(pageResults) == 0, "Expected worked to idle when no page results are pending")

	crawling := store.Crawling{
		SitemapID: 1,
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
	ts.Assert(t, pageResults[0].Status == 410, "Expected to request the page and store the result")

	db.One("ID", 1, &crawling)
	ts.Assert(t, crawling.Processed, "Expected to mark crawling as processed")
}
