package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/asdine/storm"
	"github.com/gorilla/mux"

	"github.com/rs/cors"
)

type application struct {
	db     *storm.DB
	router *mux.Router
}

// Crawling - a single crawling action, with page results related to it
type crawling struct {
	ID          int             `schema:"-" storm:"id,increment"`
	CreatedAt   time.Time       `schema:"-" storm:"index"`
	URL         string          `schema:"url" storm:"index"`
	Processed   bool            `schema:"-" storm:"index"`
	Concurrency int             `schema:"concurrency"`
	Headers     []requestHeader `schema:"headers"`
	PageResults []pageResult
}

type requestHeader struct {
	Name  string `schema:"name"`
	Value string `schema:"value"`
}

// PageResult - an entity representing a particular requested page
type pageResult struct {
	URL    string `json:"url"`
	Status int    `json:"status"`
}

func main() {
	var dbFile = flag.String("db", "canary.db", "database file")
	var port = flag.String("port", "4000", "port to listen on")
	var username = flag.String("username", "", "username for basic auth (if needed)")
	var password = flag.String("password", "", "password for basic auth (if needed)")
	var origin = flag.String("origin", "http://localhost:4000", "origin to allow cross-origin requests")
	flag.Parse()

	app := newApplication(*dbFile)
	defer app.db.Close()

	app.useRoutes()
	app.serveAssets()

	go app.runWorkers()

	log.Printf("Listening on the port: %s", *port)

	handler := corsMiddleware(app.router, []string{*origin})
	handler = basicAuthMiddleware(handler, *username, *password)

	http.ListenAndServe(":"+*port, handler)
}

func newApplication(dbFilename string) *application {
	db, err := storm.Open(dbFilename)
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()
	app := &application{
		db:     db,
		router: router,
	}
	return app
}

func corsMiddleware(r *mux.Router, origins []string) http.Handler {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE"},
	})
	return corsMiddleware.Handler(r)
}

func basicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		u, p, _ := r.BasicAuth()

		if username == "" && password == "" {
			next.ServeHTTP(rw, r)
			return
		}

		if u == username && p == password {
			next.ServeHTTP(rw, r)
		} else {
			rw.WriteHeader(http.StatusForbidden)
			rw.Write([]byte("Forbidden"))
		}
	})
}
