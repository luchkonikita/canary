package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asdine/storm/q"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func (app *application) useRoutes() {
	app.router.HandleFunc("/crawlings", app.handleCrawlingsGET()).Methods("GET")
	app.router.HandleFunc("/crawlings/{id}", app.handleCrawlingGET()).Methods("GET")
	app.router.HandleFunc("/crawlings", app.handleCrawlingsPOST()).Methods("POST")
	app.router.HandleFunc("/crawlings/{id}", app.handleCrawlingDELETE()).Methods("DELETE")
	app.router.HandleFunc("/", app.handleIndex()).Methods("GET")
}

func (app *application) serveAssets() {
	staticFilesBox := packr.NewBox("web/dist")
	app.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(staticFilesBox)))
}

func (app *application) handleIndex() http.HandlerFunc {
	staticFilesBox := packr.NewBox("web/dist")
	indexTemplate := staticFilesBox.String("index.html")
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html")
		fmt.Fprint(rw, indexTemplate)
	}
}

func (app *application) handleCrawlingsGET() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		crs := []crawling{}

		query := app.db.Select().OrderBy("CreatedAt").Reverse()
		query.Find(&crs)

		render(rw, http.StatusOK, serializeCrawlings(crs, serializerFlags{SkipNested: true}))
	}
}

func (app *application) handleCrawlingGET() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		crawlingID, _ := strconv.Atoi(vars["id"])
		cr := &crawling{}

		if err := app.db.One("ID", crawlingID, cr); err != nil {
			render(rw, http.StatusNotFound, err.Error())
			return
		}

		render(rw, http.StatusOK, cr.Serialize(serializerFlags{SkipNested: false}))
	}
}

func (app *application) handleCrawlingsPOST() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		cr := &crawling{}
		cr.CreatedAt = time.Now()
		decoder.Decode(cr, r.Form)

		if len(cr.URL) == 0 {
			render(rw, http.StatusBadRequest, "Crawling should have a valid URL")
			return
		}

		if cr.Concurrency == 0 {
			render(rw, http.StatusBadRequest, "Crawling should have a concurrency bigger than 0")
			return
		}

		query := app.db.Select(q.And(
			q.Eq("URL", cr.URL),
			q.Eq("Processed", false),
		))

		if query.First(&crawling{}) == nil {
			render(rw, http.StatusBadRequest, "Current crawling is in progress, cannot create another")
			return
		}

		urls, err := requestSitemap(cr.URL, cr.Headers)

		if err != nil {
			render(rw, http.StatusBadRequest, err.Error())
			return
		}

		for _, url := range urls {
			cr.PageResults = append(cr.PageResults, pageResult{
				URL: url,
			})
		}

		if err := app.db.Save(cr); err != nil {
			render(rw, http.StatusBadRequest, err.Error())
		} else {
			render(rw, http.StatusCreated, cr.Serialize(serializerFlags{SkipNested: false}))
		}
	}
}

func (app *application) handleCrawlingDELETE() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		crID, _ := strconv.Atoi(vars["id"])

		var cr crawling

		if err := app.db.One("ID", crID, &cr); err != nil {
			render(rw, http.StatusNotFound, err.Error())
			return
		}

		if err := app.db.DeleteStruct(&cr); err != nil {
			render(rw, http.StatusBadRequest, err.Error())
		} else {
			render(rw, http.StatusOK, nil)
		}
	}
}

func render(rw http.ResponseWriter, status int, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(data)
}
