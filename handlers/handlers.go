package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/utils"

	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

// Handler - a function returning router compatible handler
type Handler func(db *storm.DB) httprouter.Handle

// GetSitemaps - returns a list of all sitemaps in the database via `GET /sitemaps`
func GetSitemaps(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		setHeaders(rw)
		encoder := json.NewEncoder(rw)

		var sitemaps []store.Sitemap
		db.All(&sitemaps)
		encoder.Encode(sitemaps)
	}
}

// CreateSitemap - creates a new sitemap via `POST /sitemaps`
func CreateSitemap(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		sitemap := &store.Sitemap{}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(sitemap)
		err := store.CreateSitemap(db, sitemap)

		if err != nil {
			renderUnprocessableEntity(rw, err)
		} else {
			renderOK(rw, map[string]interface{}{
				"status":  "Created",
				"sitemap": sitemap,
			})
		}
	}
}

// DeleteSitemap - deletes a sitemap via `DELETE /sitemaps/:id`
func DeleteSitemap(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id, _ := strconv.Atoi(p.ByName("sitemapId"))
		err := store.DeleteSitemap(db, id)

		if err != nil {
			renderNotFound(rw, err)
		} else {
			renderOK(rw, map[string]string{
				"status": "Deleted",
			})
		}
	}
}

// GetCrawlings - returns a list of crawlings with results via `GET /sitemaps/:sitemapId/crawlings`
func GetCrawlings(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sitemapID, _ := strconv.Atoi(p.ByName("sitemapId"))
		sitemap, err := store.GetSitemap(db, sitemapID)

		if err != nil {
			renderNotFound(rw, err)
			return
		}

		crawlings := store.GetCrawlings(db, sitemap)

		renderOK(rw, crawlings)
	}
}

// CreateCrawling - creates a new crawling via `POST /sitemaps/:sitemapId/crawlings`
// Creates PageResult's for each page returned from the sitemap.
// These results are going to be processed separately by background worker.
func CreateCrawling(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sitemapID, _ := strconv.Atoi(p.ByName("sitemapId"))
		sitemap, err := store.GetSitemap(db, sitemapID)
		if err != nil {
			renderNotFound(rw, err)
			return
		}

		urls, err := utils.ParseSitemap(sitemap)
		if err != nil {
			renderBadRequest(rw, err)
			return
		}

		err = store.CreateCrawling(db, sitemap, urls)

		if err != nil {
			renderBadRequest(rw, err)
		} else {
			renderOK(rw, map[string]string{
				"status": "Created",
			})
		}
	}
}

// DeleteCrawling - deletes a crawling via `DELETE /sitemaps/:sitemapId/crawlings/crawlingId`
// If the crawling is running, this will stop it.
func DeleteCrawling(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sitemapID, _ := strconv.Atoi(p.ByName("sitemapId"))
		crawlingID, _ := strconv.Atoi(p.ByName("crawlingId"))
		err := store.DeleteCrawling(db, sitemapID, crawlingID)

		if err != nil {
			renderBadRequest(rw, err)
		} else {
			renderOK(rw, map[string]string{
				"status": "Deleted",
			})
		}
	}
}

// GetPageResults - loads page results for a crawling via `GET /crawlings/:crawlingId/page_results`
func GetPageResults(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		crawlingID, _ := strconv.Atoi(p.ByName("crawlingId"))
		err := db.One("ID", crawlingID, &store.Crawling{})

		if err != nil {
			renderNotFound(rw, err)
		} else {
			var pageResults []store.PageResult
			db.Find("CrawlingID", crawlingID, &pageResults)
			renderOK(rw, pageResults)
		}
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

func renderUnprocessableEntity(rw http.ResponseWriter, err error) {
	setHeaders(rw)
	rw.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(rw).Encode(map[string]string{
		"status": "Unprocessable Entity",
		"error":  err.Error(),
	})
}

func renderNotFound(rw http.ResponseWriter, err error) {
	setHeaders(rw)
	rw.WriteHeader(http.StatusNotFound)
	json.NewEncoder(rw).Encode(map[string]string{
		"status": "Not Found",
		"error":  err.Error(),
	})
}

func setHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
}
