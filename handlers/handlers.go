package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/utils"

	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

const help = `
Welcome to Canary service API.

Usage examples:

- GET /sitemaps
	{}
	Returns a list of sitemaps.

- POST /sitemaps
	{
		"name": "Some name",
		"url": "http://someurl.com",
		"concurrency": 10,
		"username": "Username when sitemap requires basic auth",
		"password": "Password when sitemap requires basic auth",
	}
	Adds a new sitemap.

- PATCH /sitemaps/1
	{
		"name": "Some name",
		"url": "http://someurl.com",
		"concurrency": 10,
		"username": "Username when sitemap requires basic auth",
		"password": "Password when sitemap requires basic auth",
	}
	Updates a sitemap.

- DELETE /sitemaps/1
	{}
	Updates a sitemap.

- GET /crawlings
	{
		"sitemap_id": 1,
		"processed": "true",
		"limit": 10,
		"offset": 0,
	}
	Returns a list of crawlings.

- POST /crawlings
	{
		"sitemap_id": 1
	}
	Creates a new crawling and starts it.

- DELETE /crawlings/1
	{}
	Deletes a crawling and cancels it.

- GET /page_results
	{
		"crawling_id": 1,
		"status": 500,
		"url": "some-url-substring"
		"limit": 10,
		"offset": 0,
	}
	Returns a list of page results.
`

// Handler - a function returning router compatible handler
type Handler func(db *storm.DB) httprouter.Handle

// NewRouter - defines all the routes and corresponding handlers
func NewRouter(db *storm.DB) *httprouter.Router {
	router := httprouter.New()

	router.GET("/", Ping)

	// Sitemaps
	router.GET("/sitemaps", GetSitemaps(db))
	router.POST("/sitemaps", CreateSitemap(db))
	router.PATCH("/sitemaps/:sitemapId", UpdateSitemap(db))
	router.DELETE("/sitemaps/:sitemapId", DeleteSitemap(db))

	// Crawlings
	router.GET("/crawlings", GetCrawlings(db))
	router.POST("/crawlings", CreateCrawling(db))
	router.DELETE("/crawlings/:crawlingId", DeleteCrawling(db))

	// Page results
	router.GET("/page_results", GetPageResults(db))

	return router
}

// Ping - return 200 status and some usage instructions
func Ping(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rw.Write([]byte(help))
}

// GetSitemaps - returns a list of all sitemaps in the database via `GET /sitemaps`
func GetSitemaps(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		sitemaps := store.GetSitemaps(db)
		renderOK(rw, sitemaps)
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

// UpdateSitemap - updates a sitemap via `PATCH /sitemaps/sitemapId`
func UpdateSitemap(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id, _ := strconv.Atoi(p.ByName("sitemapId"))
		sitemap := &store.Sitemap{ID: id}
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(sitemap)
		err := store.UpdateSitemap(db, sitemap)

		if err != nil {
			renderNotFound(rw, err)
		} else {
			renderOK(rw, map[string]interface{}{
				"status":  "Updated",
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
		f := &store.CrawlingsFilter{}

		if err := decodeBody(r, f); err != nil {
			renderBadRequest(rw, err)
		}

		crawlings := f.Query(db)
		renderOK(rw, crawlings)
	}
}

// CreateCrawling - creates a new crawling via `POST /sitemaps/:sitemapId/crawlings`
// Creates PageResult's for each page returned from the sitemap.
// These results are going to be processed separately by background worker.
func CreateCrawling(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := &struct {
			SitemapID int `json:"sitemap_id"`
		}{}

		if err := decodeBody(r, data); err != nil {
			renderBadRequest(rw, err)
		}

		sitemap, err := store.GetSitemap(db, data.SitemapID)
		if err != nil {
			renderNotFound(rw, err)
			return
		}

		urls, err := utils.ParseSitemap(sitemap)
		if err != nil {
			renderBadRequest(rw, err)
			return
		}

		crawling, err := store.CreateCrawling(db, sitemap, urls)

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

// DeleteCrawling - deletes a crawling via `DELETE /sitemaps/:sitemapId/crawlings/crawlingId`
// If the crawling is running, this will stop it.
func DeleteCrawling(db *storm.DB) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		crawlingID, _ := strconv.Atoi(p.ByName("crawlingId"))
		err := store.DeleteCrawling(db, crawlingID)

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
		f := &store.PageResultsFilter{}

		if err := decodeBody(r, f); err != nil {
			renderBadRequest(rw, err)
		}

		pageResults := f.Query(db)
		renderOK(rw, pageResults)
	}
}

func decodeBody(r *http.Request, to interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, to)
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
