package main

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/ray1729/gpx-utils/pkg/openname"
	"github.com/ray1729/gpx-utils/pkg/placenames"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("Usage: %s INFILE OUTFILE", os.Args[0])
	}
	wc, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer wc.Close()
	enc := gob.NewEncoder(wc)
	err = openname.ProcessFile(
		os.Args[1],
		func(r *openname.Record) error {
			b := placenames.NamedBoundary{
				Name:   r.Name,
				Type:   r.LocalType,
				County: coalesce(r.CountyUnitary, r.DistrictBorough),
				Xmin:   r.MbrXMin,
				Ymin:   r.MbrYMin,
				Xmax:   r.MbrXMax,
				Ymax:   r.MbrYMax}
			return enc.Encode(b)
		},
		openname.FilterType("populatedPlace"),
		openname.FilterLocalType("Suburban Area").Complement(),
		openname.FilterAreaGt(0),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func coalesce(xs ...string) string {
	for _, x := range xs {
		if len(x) > 0 {
			return x
		}
	}
	return ""
}
