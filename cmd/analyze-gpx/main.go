package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
	"github.com/twpayne/go-gpx"

	"github.com/ray1729/gpx-utils/pkg/openname"
)

func main() {
	openNames := flag.String("opname", "", "Path to Ordnance Server Open Names zip archive")
	gpxFile := flag.String("gpx", "", "Path to GPX file")
	dirName := flag.String("dir", "", "Directory to scan for GPX files")
	flag.Parse()
	if *openNames == "" {
		log.Fatal("--opname is required")
	}
	if (*gpxFile == "" && *dirName == "") || (*gpxFile != "" && *dirName != "") {
		log.Fatal("exactly one of --dir or --gpx is required")
	}
	rt, err := buildIndex(*openNames)
	if err != nil {
		log.Fatal(err)
	}
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		log.Fatal(err)
	}
	if *gpxFile != "" {
		err = summarizeSingleFile(rt, trans, *gpxFile)
	} else {
		err = summarizeDirectory(rt, trans, *dirName)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func summarizeDirectory(rt *rtreego.Rtree, trans osgb.CoordinateTransformer, dirName string) error {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || path.Ext(f.Name()) != ".gpx" {
			continue
		}
		filename := path.Join(dirName, f.Name())
		log.Printf("Analyzing %s", filename)
		summary, err := summarizeGPXTrack(rt, trans, filename)
		if err != nil {
			return fmt.Errorf("error creating summary of GPX track %s: %v", filename, err)
		}
		outfile := filename[:len(filename)-4] + ".json"
		wc, err := os.OpenFile(outfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("error creating output file %s: %v", outfile, err)
		}
		err = writeSummary(summary, wc)
		if err != nil {
			wc.Close()
			return fmt.Errorf("error marshalling JSON to %s: %v", outfile, err)
		}
		if err = wc.Close(); err != nil {
			return fmt.Errorf("error closing file %s: %v", outfile, err)
		}
	}
	return nil
}

func summarizeSingleFile(rt *rtreego.Rtree, trans osgb.CoordinateTransformer, filename string) error {
	summary, err := summarizeGPXTrack(rt, trans, filename)
	if err != nil {
		return fmt.Errorf("error creating summary of GPX track %s: %v", filename, err)
	}
	if err = writeSummary(summary, os.Stdout); err != nil {
		return fmt.Errorf("error marshalling summary for %s: %v", filename, err)
	}
	return nil
}

func writeSummary(s *Summary, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(s); err != nil {
		return err
	}
	return nil
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
	return math.Sqrt(s) / 1000.0
}

type POI struct {
	Name     string
	Distance float64
}

type Summary struct {
	Name             string
	Time             time.Time
	Link             string
	Start            string
	Finish           string
	Distance         float64
	Ascent           float64
	PointsOfInterest []POI
}

func summarizeGPXTrack(rt *rtreego.Rtree, trans osgb.CoordinateTransformer, filename string) (*Summary, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	g, err := gpx.Read(r)
	if err != nil {
		return nil, err
	}
	var s Summary
	s.Name = g.Metadata.Name
	s.Time = g.Metadata.Time
	for _, l := range g.Metadata.Link {
		if strings.HasPrefix(l.HREF, "http") {
			s.Link = l.HREF
			break
		}
	}

	var prevPlace string
	var prevPoint rtreego.Point
	var prevHeight float64

	init := true
	for _, trk := range g.Trk {
		for _, seg := range trk.TrkSeg {
			for _, p := range seg.TrkPt {
				gpsCoord := osgb.NewETRS89Coord(p.Lon, p.Lat, p.Ele)
				ngCoord, err := trans.ToNationalGrid(gpsCoord)
				if err != nil {
					return nil, err
				}
				thisPoint := rtreego.Point{ngCoord.Easting, ngCoord.Northing}
				thisHeight := ngCoord.Height
				nn, _ := rt.NearestNeighbor(thisPoint).(*openname.Record)
				if init {
					s.Start = nn.Name
					prevPlace = nn.Name
					prevPoint = thisPoint
					prevHeight = thisHeight
					s.PointsOfInterest = append(s.PointsOfInterest, POI{nn.Name, 0.0})
					init = false
					continue
				}
				s.Distance += distance(thisPoint, prevPoint)
				if ascent := thisHeight - prevHeight; ascent > 0 {
					s.Ascent += ascent
				}
				if insideLoc(thisPoint, nn) && nn.Name != prevPlace {
					s.PointsOfInterest = append(s.PointsOfInterest, POI{nn.Name, s.Distance})
					prevPlace = nn.Name
				}
				prevPoint = thisPoint
				prevHeight = thisHeight
			}
		}
	}
	s.Finish = prevPlace
	return &s, nil
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
