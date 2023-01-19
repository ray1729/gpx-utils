package rwgps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dhconnelly/rtreego"

	"github.com/ray1729/gpx-utils/pkg/cafes"
	"github.com/ray1729/gpx-utils/pkg/placenames"
)

type RWGPSHandler struct {
	gs    *placenames.GPXSummarizer
	stops *cafes.Cache
}

func NewHandler() (*RWGPSHandler, error) {
	gs, err := placenames.NewGPXSummarizer()
	if err != nil {
		return nil, fmt.Errorf("error creating GPX summarizer: %v", err)
	}
	stops := cafes.New()
	return &RWGPSHandler{gs, stops}, nil
}

func (h *RWGPSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Error parsing route id '%s': %v", rawRouteId, err)
		http.Error(w, fmt.Sprintf("Invalid routeId: %s", rawRouteId), http.StatusBadRequest)
		return
	}
	var stopsIndex *rtreego.Rtree
	if stopsName != "" {
		var err error
		stopsIndex, err = h.stops.Get(stopsName)
		if err != nil {
			log.Println(err)
			if errors.Is(err, cafes.ErrInvalidStops) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	track, err := FetchTrack(routeId)
	if err != nil {
		log.Println(err.Error())
		switch err.(type) {
		case *ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *ErrNotPublic:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	summary, err := h.gs.SummarizeTrack(bytes.NewReader(track), stopsIndex)
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
