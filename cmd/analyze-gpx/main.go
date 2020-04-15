package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
	"github.com/twpayne/go-gpx"

	"github.com/ray1729/gpx-utils/pkg/openname"
)

func main() {
	openNames := flag.String("open-names", "", "Path to Ordnance Server Open Names zip archive")
	gpxFile := flag.String("gpx", "", "Path to GPX file")
	flag.Parse()
	if *openNames == "" {
		log.Fatal("--open-names is required")
	}
	if *gpxFile == "" {
		log.Fatal("--gpx is required")
	}
	rt, err := buildIndex(*openNames)
	if err != nil {
		log.Fatal(err)
	}
	points, err := readGPX(*gpxFile)
	if err != nil {
		log.Fatal(err)
	}
	var dist float64
	var prevPlace string
	var prevPoint rtreego.Point
	for i, p := range points {
		nn := rt.NearestNeighbor(p)
		loc, _ := nn.(*openname.Record)
		if i == 0 {
			fmt.Printf("%0.2f %s\n", dist, loc.Name)
			prevPlace = loc.Name
			prevPoint = p
			continue
		}
		dist += distance(prevPoint, p)
		if insideLoc(p, loc) && loc.Name != prevPlace {
			fmt.Printf("%0.2f %s\n", dist/1000, loc.Name)
			prevPlace = loc.Name
		}
		prevPoint = p
	}
}

func insideLoc(p rtreego.Point, loc *openname.Record) bool {
	return p[0] >= loc.MbrXMin && p[0] <= loc.MbrXMax && p[1] >= loc.MbrYMin && p[1] <= loc.MbrYMax
}

func distance(p1, p2 rtreego.Point) float64 {
	if len(p1) != len(p2) {
		panic("Length mismatch")
	}
	var s float64
	for i := range p1 {
		d := p1[i] - p2[i]
		s += d * d
	}
	return math.Sqrt(s)
}

func readGPX(filename string) ([]rtreego.Point, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	g, err := gpx.Read(r)
	if err != nil {
		return nil, err
	}
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		return nil, err
	}
	var points []rtreego.Point
	for _, trk := range g.Trk {
		for _, seg := range trk.TrkSeg {
			for _, p := range seg.TrkPt {
				gpsCoord := osgb.NewETRS89Coord(p.Lon, p.Lat, p.Ele)
				ngCoord, err := trans.ToNationalGrid(gpsCoord)
				if err != nil {
					return nil, err
				}
				points = append(points, rtreego.Point{ngCoord.Easting, ngCoord.Northing})
			}
		}
	}
	return points, nil
}

func buildIndex(filename string) (*rtreego.Rtree, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	rt := rtreego.NewTree(2, 25, 50)
	for _, f := range r.File {
		if !(strings.HasPrefix(f.Name, "DATA/") && strings.HasSuffix(f.Name, ".csv")) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		s, err := openname.NewScanner(rc)
		if err != nil {
			log.Fatalf("Error reading %s: %v", f.Name, err)
		}
		for s.Scan() {
			r := s.Record()
			if r.Type == "populatedPlace" && r.MbrXMax != r.MbrXMin && r.MbrYMax != r.MbrYMin {
				rt.Insert(r)
			}
		}
		if err = s.Err(); err != nil {
			log.Fatalf("Error parsing %s: %v", f.Name, err)
		}
	}
	return rt, nil
}
