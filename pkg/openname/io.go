package openname

import (
	"archive/zip"
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	"github.com/ray1729/gpx-utils/pkg/placenames"
)

// ProcessFile reads the compressed OS Open Names data set and calls the handler for each record.
func ProcessFile(filename string, handler func(*Record) error) error {
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
			if r.Type == "populatedPlace" && r.MbrXMax != r.MbrXMin && r.MbrYMax != r.MbrYMin {
				if err := handler(r); err != nil {
					return err
				}
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

// Save processes the OS OpenNames zip file and outputs bounded places in gob format.
func Save(inFile string, outFile string) error {
	wc, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer wc.Close()
	enc := gob.NewEncoder(wc)
	err = ProcessFile(inFile, func(r *Record) error {
		b := placenames.NamedBoundary{
			Name: r.Name,
			Xmin: r.MbrXMin,
			Ymin: r.MbrYMin,
			Xmax: r.MbrXMax,
			Ymax: r.MbrYMax}
		return enc.Encode(b)
	})
	return err
}
