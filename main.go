package main

import (
	"log"
	"net/http"

	"github.com/luchkonikita/canary/router"
	"github.com/luchkonikita/canary/store"
	"github.com/luchkonikita/canary/workers"
)

var (
	port = "8080"
)

func main() {
	db := store.NewDB("storage.db", false)
	defer db.Close()

	go workers.Start(db)

	log.Printf("Listening on the port: %s", port)
	router := router.NewRouter(db)

	http.ListenAndServe(":"+port, router)
}
