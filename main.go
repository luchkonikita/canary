package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"

	"github.com/luchkonikita/canary/handlers"
	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/workers"
)

var (
	port = "4000"
)

func main() {
	db := store.NewDB("storage.db", false)
	defer db.Close()

	go workers.Start(db)

	log.Printf("Listening on the port: %s", port)
	router := handlers.NewRouter(db)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})
	handler := corsMiddleware.Handler(router)
	http.ListenAndServe(":"+port, handler)
}
