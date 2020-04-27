package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ray1729/gpx-utils/pkg/placenames"
)

func main() {
	log.SetFlags(0)
	maxPOI := flag.Int("max-poi", 0, "Maximum number of points of interest")
	minDist := flag.Float64("min-dist", 0, "Minimum distance between points of interest")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("Usage: %s [--max-poi=N] ANALYSIS.json", os.Args[0])
	}
	summary, err := readTrackSummary(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	poi := summary.PointsOfInterest
	if *minDist > 0 {
		poi = condenseMinDist(poi, *minDist)
	}
	if *maxPOI > 0 {
		poi = condenseMaxPoi(summary.PointsOfInterest, *maxPOI)
	}
	result := make([]string, len(poi))
	for i, x := range poi {
		result[i] = x.Name
	}
	fmt.Println(strings.Join(result, ", "))
}

func readTrackSummary(filename string) (*placenames.TrackSummary, error) {
	log.Println("Reading summary from", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var summary placenames.TrackSummary
	err = json.Unmarshal(data, &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

var locationTypePriority = map[string]int{
	"City":    5,
	"Town":    4,
	"Village": 3,
	"Hamlet":  2,
}

func condenseMinDist(xs []placenames.POI, minDist float64) []placenames.POI {
	log.Printf("Condensing by min distance %f", minDist)
	cont := true
	for cont {
		cont = false
		for i := 0; i < len(xs)-1; i++ {
			if xs[i+1].Distance-xs[i].Distance < minDist {
				p1 := locationTypePriority[xs[i].Type]
				p2 := locationTypePriority[xs[i+1].Type]
				if i == 0 || p2 < p1 {
					xs = deleteElement(xs, i+1)
					cont = true
					break
				}
			}
		}
	}
	return xs
}

func condenseMaxPoi(xs []placenames.POI, maxPoi int) []placenames.POI {
	log.Printf("Condensing %d to %d points", len(xs), maxPoi)
	for len(xs) > maxPoi {
		var minI int
		var minD float64
		for i := 0; i < len(xs)-1; i++ {
			d := xs[i+1].Distance - xs[i].Distance
			if i == 0 || d < minD {
				minI = i
				minD = d
			}
		}
		p1 := locationTypePriority[xs[minI].Type]
		p2 := locationTypePriority[xs[minI+1].Type]
		if minI == 0 || p2 < p1 {
			xs = deleteElement(xs, minI+1)
			continue
		}
		if minI == len(xs)-2 || p1 < p2 {
			xs = deleteElement(xs, minI)
			continue
		}
		// p1 == p2
		d1 := xs[minI].Distance - xs[minI-1].Distance
		d2 := xs[minI+1].Distance - xs[minI].Distance
		if d1 < d2 {
			xs = deleteElement(xs, minI)
		} else {
			xs = deleteElement(xs, minI+1)
		}
	}
	return xs
}

func deleteElement(xs []placenames.POI, i int) []placenames.POI {
	log.Printf("Deleting %s (%0.1f)", xs[i].Name, xs[i].Distance)
	return append(xs[0:i], xs[i+1:]...)
}
