package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ray1729/gpx-utils/pkg/placenames"
	"github.com/ray1729/gpx-utils/pkg/rwgps"
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8000"
	}
	gs, err := placenames.NewGPXSummarizer()
	if err != nil {
		log.Fatal(err)
	}
	gpxSummarizer = gs
	http.HandleFunc("/rwgps", rwgpsHandler)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

var gpxSummarizer *placenames.GPXSummarizer

func rwgpsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	x := q.Get("routeId")
	log.Printf("Handilng request for routeId=%s", x)
	if x == "" {
		http.Error(w, "routeId is required", http.StatusBadRequest)
		return
	}
	routeId, err := strconv.Atoi(x)
	if err != nil {
		log.Printf("Invalid route id: %s", x)
		http.Error(w, "Invalid route id", http.StatusBadRequest)
		return
	}
	track, err := rwgps.FetchTrack(routeId)
	if err != nil {
		log.Println(err.Error())
		switch err.(type) {
		case *rwgps.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	summary, err := gpxSummarizer.SummarizeTrack(bytes.NewReader(track))
	if err != nil {
		log.Printf("Error analyzing route %d: %v", routeId, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(summary)
	if err != nil {
		log.Printf("Error marshalling JSON for route %d: %v", routeId, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
