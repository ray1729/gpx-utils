package openname

import (
	"archive/zip"
	"log"
	"strings"

	"github.com/dhconnelly/rtreego"
)

func BuildIndex(filename string) (*rtreego.Rtree, error) {
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
		s, err := NewScanner(rc)
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
