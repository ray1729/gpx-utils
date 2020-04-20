package cafes

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
)

const cyclingMapsCafesUrl = "https://cafes.cyclingmaps.net/data/cafes.json"

type CyclingMapsCafe struct {
	Name    string
	Website string
	Lat     float64
	Lng     float64
}

func BuildCyclingMapsIndex(r io.Reader) (*rtreego.Rtree, error) {
	var cafes []CyclingMapsCafe
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cafes)
	if err != nil {
		return nil, err
	}
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		return nil, err
	}
	stops := make([]rtreego.Spatial, 0, len(cafes))
	for _, c := range cafes {
		gpsCoord := osgb.NewETRS89Coord(c.Lng, c.Lat, 0)
		ngCoord, err := trans.ToNationalGrid(gpsCoord)
		if err != nil {
			log.Printf("Error translating coordinates %v: %v", gpsCoord, err)
			continue
		}
		stops = append(stops, &RefreshmentStop{
			Name:     c.Name,
			Url:      c.Website,
			Easting:  ngCoord.Easting,
			Northing: ngCoord.Northing,
		})
	}
	return rtreego.NewTree(2, 25, 50, stops...), nil
}

func FetchCyclingMapsIndex() (*rtreego.Rtree, error) {
	res, err := http.Get(cyclingMapsCafesUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %v", cyclingMapsCafesUrl, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching %s: %s", cyclingMapsCafesUrl, res.Status)
	}
	index, err := BuildCyclingMapsIndex(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error building cyclingmaps.net cafe stops index: %v", err)
	}
	return index, nil
}
