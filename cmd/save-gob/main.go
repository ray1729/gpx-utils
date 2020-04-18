package main

import (
	"log"
	"os"

	"github.com/ray1729/gpx-utils/pkg/openname"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("Usage: %s INFILE OUTFILE", os.Args[0])
	}
	if err := openname.Save(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
