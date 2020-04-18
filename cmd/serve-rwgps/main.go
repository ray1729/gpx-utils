package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ray1729/gpx-utils/pkg/placenames"
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
	if x == "" {
		http.Error(w, "routeId is required", http.StatusBadRequest)
		return
	}
	routeId, err := strconv.Atoi(x)
	if err != nil {
		http.Error(w, "invalid route id", http.StatusBadRequest)
		return
	}
	track, err := FetchTrack(routeId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	summary, err := gpxSummarizer.SummarizeTrack(bytes.NewReader(track))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := json.Marshal(summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func FetchTrack(routeId int) ([]byte, error) {
	url := fmt.Sprintf("https://ridewithgps.com/routes/%d.gpx?sub_format=track", routeId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %v", url, err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %v", url, err)
	}
	return data, nil
}
