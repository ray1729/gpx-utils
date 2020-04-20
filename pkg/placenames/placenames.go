package placenames

//go:generate mule -p placenames -o data.go placenames.bin

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"

	"github.com/dhconnelly/rtreego"
)

type NamedBoundary struct {
	Name string
	Type string
	Xmin float64
	Ymin float64
	Xmax float64
	Ymax float64
}

func (b *NamedBoundary) Bounds() *rtreego.Rect {
	r, err := rtreego.NewRect(rtreego.Point{b.Xmin, b.Ymin}, []float64{b.Xmax - b.Xmin, b.Ymax - b.Ymin})
	if err != nil {
		panic(err)
	}
	return r
}

func (b *NamedBoundary) Contains(p rtreego.Point) bool {
	if len(p) != 2 {
		panic("Expected a 2-dimensional point")
	}
	return p[0] >= b.Xmin && p[0] <= b.Xmax && p[1] >= b.Ymin && p[1] <= b.Ymax
}

// Restore reads bounded places in gob format and constructs an RTree index
func RestoreIndex() (*rtreego.Rtree, error) {
	data, err := dataResource()
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewReader(data))
	var objs []rtreego.Spatial
	for {
		var b NamedBoundary
		if err := dec.Decode(&b); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		objs = append(objs, &b)
	}
	rt := rtreego.NewTree(2, 25, 50, objs...)
	return rt, nil
}
