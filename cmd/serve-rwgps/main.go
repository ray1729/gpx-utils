package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ray1729/gpx-utils/pkg/rwgps"
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8000"
	}
	rwgpsHandler, err := rwgps.NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/rwgps", rwgpsHandler)
	log.Printf("Listening for requests on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
