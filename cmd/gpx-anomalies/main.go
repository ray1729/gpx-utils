package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/fofanov/go-osgb"
	"github.com/twpayne/go-gpx"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(0)
	app := &cli.App{
		Name:  "gpx-anomalies",
		Usage: "Find repeated points in a GPX track",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "gpx-file",
				Aliases:  []string{"g"},
				Usage:    "Name of GPX file to process",
				Required: true,
			},
			&cli.Float64Flag{
				Name:    "fuzz",
				Aliases: []string{"f"},
				Usage:   "Consider two points coincident if they are within FUZZ kilometres of each other",
				Value:   0.005,
			},
			&cli.Float64Flag{
				Name:    "min-distance",
				Aliases: []string{"min"},
				Usage:   "Only show repeats that appear at least MIN kilometers apart",
				Value:   0.1,
			},
			&cli.Float64Flag{
				Name:    "max-distance",
				Aliases: []string{"max"},
				Usage:   "Do not show repeats that appear more than MAX kilometers apart",
				Value:   5.0,
			},
		},
		Action: func(c *cli.Context) error {
			points, err := readGPXTrack(c.String("gpx-file"))
			if err != nil {
				log.Fatal(err)
			}
			findDuplicates(
				points,
				c.Float64("fuzz")*1000.0,
				c.Float64("min-distance")*1000.0,
				c.Float64("max-distance")*1000.0,
			)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func findDuplicates(points []RoutePoint, fuzz, minDist, maxDist float64) {
	for i := range points {
		p := points[i]
		for j := i + 1; j < len(points); j++ {
			q := points[j]
			if p.Distance == q.Distance {
				continue
			}
			d := euclideanDistance(p.Coordinate, q.Coordinate)
			D := q.Distance - p.Distance
			if d < fuzz && D > minDist && D < maxDist {
				fmt.Printf("Point (%0.f, %0.f) revisited at %0.2f km and %0.2f km\n",
					p.Coordinate.Easting, p.Coordinate.Northing, p.Distance/1000.0, q.Distance/1000.0)
			}
		}
	}
}

func euclideanDistance(p, q *osgb.OSGB36Coordinate) float64 {
	x := p.Easting - q.Easting
	y := p.Northing - q.Northing
	return math.Sqrt(x*x + y*y)
}

func readGPXTrack(filename string) ([]RoutePoint, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening %s for reading: %v", filename, err)
	}
	defer r.Close()
	g, err := gpx.Read(r)
	if err != nil {
		return nil, fmt.Errorf("error reading GPS track %s: %v", filename, err)
	}
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		return nil, fmt.Errorf("error constructing coordinate transformer: %v", err)
	}
	distance := 0.0
	var prevPoint *osgb.OSGB36Coordinate
	var points []RoutePoint
	for _, trk := range g.Trk {
		for _, seg := range trk.TrkSeg {
			for _, trkPt := range seg.TrkPt {
				gpsCoord := osgb.NewETRS89Coord(trkPt.Lon, trkPt.Lat, trkPt.Ele)
				p, err := trans.ToNationalGrid(gpsCoord)
				if err != nil {
					return nil, fmt.Errorf("error converting coordinates to National Grid: %v", err)
				}
				if prevPoint != nil {
					distance += euclideanDistance(prevPoint, p)
				}
				points = append(points, RoutePoint{
					Coordinate: p,
					Distance:   distance,
				})
				prevPoint = p
			}
		}
	}
	return points, nil
}

type RoutePoint struct {
	Coordinate *osgb.OSGB36Coordinate
	Distance   float64
}
