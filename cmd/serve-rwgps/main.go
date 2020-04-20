package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dhconnelly/rtreego"

	"github.com/ray1729/gpx-utils/pkg/cafes"
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
	if err = loadStops(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/rwgps", rwgpsHandler)
	log.Printf("Listening for requests on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func loadStops() error {
	var err error
	log.Println("Fetching CTC Cambridge cafe stops")
	stops["ctccam"], err = cafes.FetchCtcCamIndex()
	if err != nil {
		return err
	}
	log.Printf("Loaded %d ctccam stops", stops["ctccam"].Size())
	log.Println("Fetching cyclingmaps.net cafe stops")
	stops["cyclingmapsnet"], err = cafes.FetchCyclingMapsIndex()
	if err != nil {
		return err
	}
	log.Printf("Loaded %d cyclingmapsnet stops", stops["cyclingmapsnet"].Size())
	return nil
}

var gpxSummarizer *placenames.GPXSummarizer
var stops = make(map[string]*rtreego.Rtree)

func rwgpsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rawRouteId := q.Get("routeId")
	stopsName := q.Get("stops")
	log.Printf("Handling request for routeId=%s stops=%s", rawRouteId, stopsName)
	if rawRouteId == "" {
		log.Printf("Missing routeId")
		http.Error(w, "routeId is required", http.StatusBadRequest)
		return
	}
	routeId, err := strconv.Atoi(rawRouteId)
	if err != nil {
		log.Println("Error parsing route id '%s': %v", rawRouteId, err)
		http.Error(w, fmt.Sprintf("Invalid routeId: %s", rawRouteId), http.StatusBadRequest)
		return
	}
	var stopsIndex *rtreego.Rtree
	if stopsName != "" {
		stopsIndex = stops[stopsName]
		if stopsIndex == nil {
			log.Printf("Invalid stops: %s", stopsName)
			http.Error(w, fmt.Sprintf("Invalid stops: %s", stopsName), http.StatusBadRequest)
			return
		}
	}
	track, err := rwgps.FetchTrack(routeId)
	if err != nil {
		log.Println(err.Error())
		switch err.(type) {
		case *rwgps.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *rwgps.ErrNotPublic:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	summary, err := gpxSummarizer.SummarizeTrack(bytes.NewReader(track), stopsIndex)
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
