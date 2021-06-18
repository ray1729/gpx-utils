package cafes

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dhconnelly/rtreego"
)

// Size (in metres) of the bounding box around a stop
const stopRectangleSize = 500

type RefreshmentStop struct {
	Name     string
	Url      string
	Easting  float64
	Northing float64
}

func (s *RefreshmentStop) Bounds() *rtreego.Rect {
	p := rtreego.Point{s.Easting, s.Northing}
	return p.ToRect(stopRectangleSize)
}

func (s *RefreshmentStop) Contains(p rtreego.Point) bool {
	if len(p) != 2 {
		panic("Expected a 2-dimensional point")
	}
	bounds := s.Bounds()
	for i := 0; i < 2; i++ {
		if p[i] < bounds.PointCoord(i) || p[i] > bounds.PointCoord(i)+bounds.LengthsCoord(i) {
			return false
		}
	}
	return true
}

// TTL cache based on "9.7 Example: Concurrent Non-Blocking Cache" from
// "The Go Programming Language", Alan A. A. Donovan and Brian W. Kernighan

type result struct {
	value *rtreego.Rtree
	err   error
}

type entry struct {
	res     result
	expires time.Time
	ready   chan struct{} // closed when res is ready
}

type Cache struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func New() *Cache {
	return &Cache{entries: make(map[string]*entry)}
}

func (c *Cache) Get(k string) (*rtreego.Rtree, error) {
	c.mu.Lock()
	e := c.entries[k]
	if e == nil || e.expires.Before(time.Now()) {
		e = &entry{ready: make(chan struct{}), expires: time.Now().Add(4 * time.Hour)}
		c.entries[k] = e
		c.mu.Unlock()
		e.res.value, e.res.err = FetchStops(k)
		close(e.ready)
	} else {
		c.mu.Unlock()
		<-e.ready
	}
	return e.res.value, e.res.err
}

var ErrInvalidStops = errors.New("invalid stops")

func FetchStops(k string) (*rtreego.Rtree, error) {
	switch k {
	case "ctccambridge":
		return FetchCtcCamIndex()
	case "cyclingmaps":
		return FetchCyclingMapsIndex()
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidStops, k)
	}
}
