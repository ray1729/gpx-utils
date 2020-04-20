package rwgps

import (
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

type ErrNotPublic struct {
	RouteId int
}

func (e *ErrNotPublic) Error() string {
	return fmt.Sprintf("RideWithGPS track %d is not public", e.RouteId)
}

func FetchTrack(routeId int) ([]byte, error) {
	url := fmt.Sprintf("https://ridewithgps.com/routes/%d.gpx?sub_format=track", routeId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, &ErrNotFound{routeId}
		}
		if resp.StatusCode == http.StatusForbidden {
			return nil, &ErrNotPublic{routeId}
		}
		return nil, fmt.Errorf("error retrieving route %d: %s", routeId, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %v", url, err)
	}
	return data, nil
}
