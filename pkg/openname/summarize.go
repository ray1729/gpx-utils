package openname

import (
	"io"
	"math"
	"strings"
	"time"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
	"github.com/twpayne/go-gpx"
)

type GPXSummarizer struct {
	rt    *rtreego.Rtree
	trans osgb.CoordinateTransformer
}

func NewGPXSummarizer(rt *rtreego.Rtree) (*GPXSummarizer, error) {
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		if err != nil {
			return nil, err
		}
	}
	return &GPXSummarizer{rt: rt, trans: trans}, nil
}

func insideLoc(p rtreego.Point, loc *Record) bool {
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

type TrackSummary struct {
	Name             string
	Time             time.Time
	Link             string
	Start            string
	Finish           string
	Distance         float64
	Ascent           float64
	PointsOfInterest []POI
}

func (gs *GPXSummarizer) SummarizeTrack(r io.Reader) (*TrackSummary, error) {
	g, err := gpx.Read(r)
	if err != nil {
		return nil, err
	}
	var s TrackSummary
	s.Name = g.Metadata.Name
	s.Time = g.Metadata.Time
	for _, l := range g.Metadata.Link {
		if strings.HasPrefix(l.HREF, "http") {
			s.Link = l.HREF
			break
		}
	}

	var prevPlace string
	var prevPlacePoint rtreego.Point
	var prevPoint rtreego.Point
	var prevHeight float64

	init := true
	for _, trk := range g.Trk {
		for _, seg := range trk.TrkSeg {
			for _, p := range seg.TrkPt {
				gpsCoord := osgb.NewETRS89Coord(p.Lon, p.Lat, p.Ele)
				ngCoord, err := gs.trans.ToNationalGrid(gpsCoord)
				if err != nil {
					return nil, err
				}
				thisPoint := rtreego.Point{ngCoord.Easting, ngCoord.Northing}
				thisHeight := ngCoord.Height
				nn, _ := gs.rt.NearestNeighbor(thisPoint).(*Record)
				if init {
					s.Start = nn.Name
					prevPlace = nn.Name
					prevPlacePoint = thisPoint
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
				if insideLoc(thisPoint, nn) && nn.Name != prevPlace && distance(thisPoint, prevPlacePoint) > 0.2 {
					s.PointsOfInterest = append(s.PointsOfInterest, POI{nn.Name, s.Distance})
					prevPlace = nn.Name
					prevPlacePoint = thisPoint
				}
				prevPoint = thisPoint
				prevHeight = thisHeight
			}
		}
	}
	s.Finish = prevPlace
	return &s, nil
}
