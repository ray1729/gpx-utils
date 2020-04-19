package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ray1729/gpx-utils/pkg/openname"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("Usage: %s OPNAME_CSV_ZIP", os.Args[0])
	}
	var records []*openname.Record
	openname.ProcessFile(
		os.Args[1],
		func(r *openname.Record) error {
			records = append(records, r)
			return nil
		},
		openname.FilterType("populatedPlace"),
		openname.FilterWithinRadius(544945, 258410, 20000),
	)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})
	for _, r := range records {
		fmt.Printf("%s,%s\n", r.Name, r.LocalType)
	}
}
