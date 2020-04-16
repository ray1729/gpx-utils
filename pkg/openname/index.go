package openname

import (
	"archive/zip"
	"fmt"
	"strings"

	"github.com/dhconnelly/rtreego"
)

func BuildIndex(filename string) (*rtreego.Rtree, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening %s fo reading: %v", filename, err)
	}
	defer r.Close()
	rt := rtreego.NewTree(2, 25, 50)
	for _, f := range r.File {
		if !(strings.HasPrefix(f.Name, "DATA/") && strings.HasSuffix(f.Name, ".csv")) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("erorr opening %s: %v", filename, err)
		}
		s, err := NewScanner(rc)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("error reading %s: %v", f.Name, err)
		}
		for s.Scan() {
			r := s.Record()
			if r.Type == "populatedPlace" && r.MbrXMax != r.MbrXMin && r.MbrYMax != r.MbrYMin {
				rt.Insert(r)
			}
		}
		if err = s.Err(); err != nil {
			rc.Close()
			return nil, fmt.Errorf("error parsing %s: %v", f.Name, err)
		}
		rc.Close()
	}
	return rt, nil
}
