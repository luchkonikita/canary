package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCrawlerWorkerWork(t *testing.T) {
	workTimeout = time.Millisecond

	// Mock database
	db := NewDB("test_storage.db", true)
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
	var pageResults []PageResult

	crawling := Crawling{
		URL:         server.URL,
		Processed:   false,
		Concurrency: 3,
		Headers: []RequestHeader{
			RequestHeader{Name: "Token", Value: "CorrectToken"},
		},
	}
	err := db.Save(&crawling)
	Assert(t, err == nil, "Expected to create a crawling")

	pageResult := PageResult{
		CrawlingID: 1,
		URL:        server.URL,
	}
	err = db.Save(&pageResult)
	Assert(t, err == nil, "Expected to create a page result")

	worker := &CrawlerWorker{
		db:       db,
		crawling: crawling,
	}

	worker.Work()

	db.All(&pageResults)
	Assert(t, pageResults[0].Status == 200, "Expected to request the page and store the result")

	db.One("ID", 1, &crawling)
	Assert(t, crawling.Processed, "Expected to mark crawling as processed")
}
