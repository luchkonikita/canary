package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"

	"github.com/luchkonikita/canary/handlers"
	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/workers"
)

func main() {
	var dbFile = flag.String("db", "canary.db", "database file")
	var port = flag.String("port", "4000", "port to listen on")
	var username = flag.String("username", "", "username for basic auth (if needed)")
	var password = flag.String("password", "", "password for basic auth (if needed)")
	var origin = flag.String("origin", "http://localhost:8080", "origin to allow cross-origin requests")
	flag.Parse()

	db := store.NewDB(*dbFile, false)
	defer db.Close()

	go workers.Start(db)

	log.Printf("Listening on the port: %s", *port)
	router := handlers.NewRouter(db)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{*origin},
		AllowCredentials: true,
	})
	handler := corsMiddleware.Handler(router)
	handler = basicAuthMiddleware(handler, *username, *password)

	http.ListenAndServe(":"+*port, handler)
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
