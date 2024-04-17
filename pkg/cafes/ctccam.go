package cafes

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
)

const ctcCamWaypointsUrl = "https://ctccambridge.org.uk/ctccambridge-waypoints.gpx"

type Waypoint struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Name string  `xml:"name"`
	Url  string  `xml:"url"`
}

type Waypoints struct {
	Waypoints []Waypoint `xml:"wpt"`
}

func BuildCtcCamIndex(r io.Reader) (*rtreego.Rtree, error) {
	dec := xml.NewDecoder(r)
	var wpt Waypoints
	err := dec.Decode(&wpt)
	if err != nil {
		return nil, err
	}
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		return nil, err
	}
	stops := make([]rtreego.Spatial, len(wpt.Waypoints))
	for i, w := range wpt.Waypoints {
		gpsCoord := osgb.NewETRS89Coord(w.Lon, w.Lat, 0)
		ngCoord, err := trans.ToNationalGrid(gpsCoord)
		if err != nil {
			return nil, fmt.Errorf("Error translating coordinates %v: %v", gpsCoord, err)
		}
		stops[i] = &RefreshmentStop{
			Name:     w.Name,
			Url:      w.Url,
			Easting:  ngCoord.Easting,
			Northing: ngCoord.Northing,
		}
	}
	return rtreego.NewTree(2, 25, 50, stops...), nil
}

func FetchCtcCamIndex() (*rtreego.Rtree, error) {
	log.Printf("Fetching %s", ctcCamWaypointsUrl)
	req, err := http.NewRequest(http.MethodGet, ctcCamWaypointsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error constructing waypoints request: %v", err)
	}
	req.Header.Set("User-Agent", "gpx-utils")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %v", ctcCamWaypointsUrl, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching %s: %s", ctcCamWaypointsUrl, res.Status)
	}
	index, err := BuildCtcCamIndex(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error building CTC Cambridge stops index: %v", err)
	}
	log.Printf("Loaded %d CTC Cambridge stops", index.Size())
	return index, nil
}
