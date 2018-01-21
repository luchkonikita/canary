package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

// Handler - a function returning router compatible handler
type Handler func(db *storm.DB) httprouter.Handle

// NewRouter - defines all the routes and corresponding handlers
func NewRouter(db *storm.DB) *httprouter.Router {
	router := httprouter.New()

	router.GET("/", Ping)

	// Crawlings
	router.GET("/crawlings", GetCrawlingsHandler(db))
	router.POST("/crawlings", CreateCrawlingHandler(db))
	router.DELETE("/crawlings/:crawlingId", DeleteCrawlingHandler(db))

	// Page results
	router.GET("/page_results", GetPageResultsHandler(db))

	return router
}

// Ping - return 200 status and some usage instructions
func Ping(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderOK(rw, map[string]interface{}{
		"alive": true,
	})
}

// GetCrawlings - returns a list of crawlings with results via `GET /crawlings`
func GetCrawlingsHandler(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		crawlings := NewCrawlingsFilter(r).Query(db)
		renderOK(rw, crawlings)
	}
}

// CreateCrawling - creates a new crawling via `POST /crawlings`
// Creates PageResult's for each page returned from the sitemap.
// These results are going to be processed separately by background worker.
func CreateCrawlingHandler(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		crawling := &Crawling{}
		err := CreateCrawling(db, crawling, r)

		if err != nil {
			renderBadRequest(rw, err)
		} else {
			renderOK(rw, map[string]interface{}{
				"status":   "Created",
				"crawling": crawling,
			})
		}
	}
}

// DeleteCrawling - deletes a crawling via `DELETE /crawlings/crawlingId`
// If the crawling is running, this will stop it.
func DeleteCrawlingHandler(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		crawlingID, _ := strconv.Atoi(p.ByName("crawlingId"))
		err := DeleteCrawling(db, crawlingID)

		if err != nil {
			renderBadRequest(rw, err)
		} else {
			renderOK(rw, map[string]string{
				"status": "Deleted",
			})
		}
	}
}

// GetPageResults - loads page results for a crawling via `GET /page_results`
func GetPageResultsHandler(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		pageResults := NewPageResultsFilter(r).Query(db)
		renderOK(rw, pageResults)
	}
}

func renderOK(rw http.ResponseWriter, data interface{}) {
	setHeaders(rw)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(data)
}

func renderBadRequest(rw http.ResponseWriter, err error) {
	setHeaders(rw)
	rw.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(rw).Encode(map[string]string{
		"status": "Bad request",
		"error":  err.Error(),
	})
}

func setHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
}
