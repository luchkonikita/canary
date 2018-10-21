package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppProcessCrawling(t *testing.T) {
	app := mockApp()
	defer app.db.Close()

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

	cr := crawling{
		URL:         "whatever",
		Processed:   false,
		Concurrency: 3,
		Headers: []requestHeader{
			requestHeader{Name: "Token", Value: "CorrectToken"},
		},
		PageResults: []pageResult{
			pageResult{
				URL: server.URL,
			},
		},
	}

	err := app.db.Save(&cr)
	assert.Nil(t, err)

	app.processCrawling(cr)
	app.db.One("ID", 1, &cr)

	assert.Equal(t, cr.PageResults[0].Status, 200)
}
