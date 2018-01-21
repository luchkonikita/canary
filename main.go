package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func main() {
	var dbFile = flag.String("db", "canary.db", "database file")
	var port = flag.String("port", "4000", "port to listen on")
	var username = flag.String("username", "", "username for basic auth (if needed)")
	var password = flag.String("password", "", "password for basic auth (if needed)")
	var origin = flag.String("origin", "http://localhost:8080", "origin to allow cross-origin requests")
	flag.Parse()

	db := NewDB(*dbFile, false)
	defer db.Close()

	go StartWorkers(db)

	log.Printf("Listening on the port: %s", *port)
	router := NewRouter(db)

	handler := corsMiddleware(router, []string{*origin})
	handler = basicAuthMiddleware(handler, *username, *password)

	http.ListenAndServe(":"+*port, handler)
}

func corsMiddleware(r *httprouter.Router, origins []string) http.Handler {
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
