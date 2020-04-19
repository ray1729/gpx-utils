package openname

import (
	"archive/zip"
	"fmt"
	"math"
	"strings"
)

type Handler func(*Record) error

type Filter func(*Record) bool

func (f Filter) Complement() Filter {
	return func(r *Record) bool {
		return !f(r)
	}
}

func FilterType(t string) Filter {
	return func(r *Record) bool {
		return r.Type == t
	}
}

func FilterLocalType(t string) Filter {
	return func(r *Record) bool {
		return r.LocalType == t
	}
}

func FilterWithinRadius(x, y, radius float64) Filter {
	return func(r *Record) bool {
		dx := x - r.GeomX
		dy := y - r.GeomY
		return math.Sqrt(dx*dx+dy*dy) <= radius
	}
}

func FilterAreaGt(a float64) Filter {
	return func(r *Record) bool {
		return r.Area() > a
	}
}

// ProcessFile reads the compressed OS Open Names data set and calls the handler for each record.
func ProcessFile(filename string, handler Handler, filters ...Filter) error {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return fmt.Errorf("error opening %s for reading: %v", filename, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !(strings.HasPrefix(f.Name, "DATA/") && strings.HasSuffix(f.Name, ".csv")) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("error opening %s: %v", filename, err)
		}
		s, err := NewScanner(rc)
		if err != nil {
			rc.Close()
			return fmt.Errorf("error reading %s: %v", f.Name, err)
		}
		for s.Scan() {
			r := s.Record()
			if wanted := applyFilters(r, filters); !wanted {
				continue
			}
			if err := handler(r); err != nil {
				return err
			}
		}
		if err = s.Err(); err != nil {
			rc.Close()
			return fmt.Errorf("error parsing %s: %v", f.Name, err)
		}
		rc.Close()
	}
	return nil
}

func applyFilters(r *Record, filters []Filter) bool {
	for _, f := range filters {
		if !f(r) {
			return false
		}
	}
	return true
}
