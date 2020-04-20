package rwgps

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ErrNotFound struct {
	RouteId int
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("RideWithGPS track %d not found", e.RouteId)
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
	if IsNotFound(data) {
		return nil, &ErrNotFound{routeId}
	}
	return data, nil
}

func IsNotFound(data []byte) bool {
	return bytes.HasPrefix(data, []byte("<!DOCTYPE html>")) && bytes.Contains(data, []byte("Error (404 not found)"))
}
