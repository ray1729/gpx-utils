package placenames

import (
	"io"
	"math"
	"strings"
	"time"

	"github.com/dhconnelly/rtreego"
	"github.com/fofanov/go-osgb"
	"github.com/twpayne/go-gpx"

	"github.com/ray1729/gpx-utils/pkg/cafes"
)

type GPXSummarizer struct {
	poi   *rtreego.Rtree
	trans osgb.CoordinateTransformer
}

func NewGPXSummarizer() (*GPXSummarizer, error) {
	trans, err := osgb.NewOSTN15Transformer()
	if err != nil {
		return nil, err
	}
	rt, err := RestoreIndex()
	if err != nil {
		return nil, err
	}
	return &GPXSummarizer{poi: rt, trans: trans}, nil
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
	Type     string
	Distance float64
}

type RefreshmentStop struct {
	Name     string
	Url      string
	Distance float64
}

type TrackSummary struct {
	Name             string
	Direction        string
	Time             time.Time
	Link             string
	Start            string
	Finish           string
	Distance         float64
	Ascent           float64
	PointsOfInterest []POI
	RefreshmentStops []RefreshmentStop `json:",omitempty"`
}

func (gs *GPXSummarizer) SummarizeTrack(r io.Reader, stops *rtreego.Rtree) (*TrackSummary, error) {
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
	var prevStop *cafes.RefreshmentStop
	var start rtreego.Point
	var dN, dE float64

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
				nn, _ := gs.poi.NearestNeighbor(thisPoint).(*NamedBoundary)
				if init {
					start = thisPoint
					s.Start = nn.Name
					prevPlace = nn.Name
					prevPlacePoint = thisPoint
					prevPoint = thisPoint
					prevHeight = thisHeight
					s.PointsOfInterest = append(s.PointsOfInterest, POI{Name: nn.Name, Type: nn.Type, Distance: 0.0})
					init = false
					continue
				}
				s.Distance += distance(thisPoint, prevPoint)
				if ascent := thisHeight - prevHeight; ascent > 0 {
					s.Ascent += ascent
				}
				dE += thisPoint[0] - start[0]
				dN += thisPoint[1] - start[1]
				if nn.Contains(thisPoint) && nn.Name != prevPlace && distance(thisPoint, prevPlacePoint) > 0.2 {
					s.PointsOfInterest = append(s.PointsOfInterest, POI{Name: nn.Name, Type: nn.Type, Distance: s.Distance})
					prevPlace = nn.Name
					prevPlacePoint = thisPoint
				}
				if stops != nil {
					stop, ok := stops.NearestNeighbor(thisPoint).(*cafes.RefreshmentStop)
					if ok && stop.Contains(thisPoint) && (prevStop == nil || stop.Name != prevStop.Name) {
						s.RefreshmentStops = append(s.RefreshmentStops, RefreshmentStop{
							Name:     stop.Name,
							Url:      stop.Url,
							Distance: s.Distance,
						})
						prevStop = stop
					}
				}
				prevPoint = thisPoint
				prevHeight = thisHeight
			}
		}
	}
	s.Finish = prevPlace
	s.Direction = ComputeDirection(dE, dN)
	return &s, nil
}

func ComputeDirection(dE, dN float64) string {
	if dN == 0 {
		if dE >= 0 {
			return "east"
		}
		return "west"
	}
	t := math.Abs(dE) / math.Abs(dN)
	if dN > 0 {
		if t < math.Tan(math.Pi/8) {
			return "north"
		}
		if t < math.Tan(3*math.Pi/8) {
			if dE > 0 {
				return "north-east"
			}
			return "north-west"
		}
		if dE > 0 {
			return "east"
		}
		return "west"
	}
	if t < math.Tan(math.Pi/8) {
		return "south"
	}
	if t < math.Tan(3*math.Pi/8) {
		if dE > 0 {
			return "south-east"
		}
		return "south-west"
	}
	if dE > 0 {
		return "east"
	}
	return "west"
}
