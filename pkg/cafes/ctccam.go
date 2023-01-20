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

const defaultWaypointsUrl = "https://ctccambridge.org.uk/ctccambridge-waypoints.gpx"

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
			return nil, fmt.Errorf("error translating coordinates %v: %v", gpsCoord, err)
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

type config struct {
	WaypointsUrl string
}

type Option func(*config)

func WithWaypointsUrl(u string) Option {
	return func(c *config) {
		c.WaypointsUrl = u
	}
}

func FetchCtcCamIndex(opt ...Option) (*rtreego.Rtree, error) {
	c := config{WaypointsUrl: defaultWaypointsUrl}
	for _, f := range opt {
		f(&c)
	}
	log.Printf("Fetching %s", c.WaypointsUrl)
	res, err := http.Get(c.WaypointsUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %v", c.WaypointsUrl, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching %s: %s", c.WaypointsUrl, res.Status)
	}
	index, err := BuildCtcCamIndex(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error building CTC Cambridge stops index: %v", err)
	}
	log.Printf("Loaded %d CTC Cambridge stops", index.Size())
	return index, nil
}
