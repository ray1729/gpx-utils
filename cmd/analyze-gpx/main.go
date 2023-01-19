package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/dhconnelly/rtreego"

	"github.com/ray1729/gpx-utils/pkg/cafes"
	"github.com/ray1729/gpx-utils/pkg/placenames"
)

func main() {
	log.SetFlags(0)
	stopNames := flag.String("stops", "", "Source for refreshment stops")
	stopRect := flag.Float64("sr", placenames.DefaultGPXSummarizerConfig.CoffeeStopSearchRectangleSize, "Size (m) of the rectangle we search for coffee stops near the route")
	stopDupDist := flag.Float64("sdd", placenames.DefaultGPXSummarizerConfig.CoffeeStopDuplicateDistance, "Suppress recurrences of coffee stops within this distance (km)")
	dupDist := flag.Float64("dd", placenames.DefaultGPXSummarizerConfig.PointOfInterestDuplicateDistance, "Suppress recurrences of points of interest within this distance (km)")
	minDist := flag.Float64("md", placenames.DefaultGPXSummarizerConfig.PointOfInterestMinimumDistance, "Minimum distance (km) between points of interest")
	minSettlement := flag.String("ms", "Other Settlement", "Exclude populated places smaller than this (City, Town, Village, Hamlet, Other Settlement)")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Usage: %s [--stops=ctccambridge|cyclingmaps] [--min-dist X] [--min-settlement S] GPX_FILE_OR_DIRECTORY")
	}
	inFile := flag.Arg(0)
	info, err := os.Stat(inFile)
	if err != nil {
		log.Fatal(err)
	}
	var stops *rtreego.Rtree
	if *stopNames != "" {
		var err error
		stops, err = cafes.New().Get(*stopNames)
		if err != nil {
			log.Fatal(err)
		}
	}
	gs, err := placenames.NewGPXSummarizer(
		placenames.WithMinimumSettlement(*minSettlement),
		placenames.WithPointOfInterestMinimumDistance(*minDist),
		placenames.WithPointOfInterestDuplicateDistance(*dupDist),
		placenames.WithCoffeeStopSearchRectangleSize(*stopRect),
		placenames.WithCoffeeStopDuplicateDistance(*stopDupDist),
	)
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		err = summarizeDirectory(gs, stops, inFile)
	} else {
		err = summarizeSingleFile(gs, stops, inFile)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func summarizeDirectory(gs *placenames.GPXSummarizer, stops *rtreego.Rtree, dirName string) error {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || path.Ext(f.Name()) != ".gpx" {
			continue
		}
		filename := path.Join(dirName, f.Name())
		r, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("error opening %s for reading: %v", filename, err)
		}
		log.Printf("Analyzing %s", filename)
		summary, err := gs.SummarizeTrack(r, stops)
		if err != nil {
			return fmt.Errorf("error creating summary of GPX track %s: %v", filename, err)
		}
		outfile := filename[:len(filename)-4] + ".json"
		wc, err := os.OpenFile(outfile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("error creating output file %s: %v", outfile, err)
		}
		err = writeSummary(summary, wc)
		if err != nil {
			wc.Close()
			return fmt.Errorf("error marshalling JSON to %s: %v", outfile, err)
		}
		if err = wc.Close(); err != nil {
			return fmt.Errorf("error closing file %s: %v", outfile, err)
		}
	}
	return nil
}

func summarizeSingleFile(gs *placenames.GPXSummarizer, stops *rtreego.Rtree, filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening %s for reading: %v", filename, err)
	}
	summary, err := gs.SummarizeTrack(r, stops)
	if err != nil {
		return fmt.Errorf("error creating summary of GPX track %s: %v", filename, err)
	}
	if err = writeSummary(summary, os.Stdout); err != nil {
		return fmt.Errorf("error marshalling summary for %s: %v", filename, err)
	}
	return nil
}

func writeSummary(s *placenames.TrackSummary, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(s); err != nil {
		return err
	}
	return nil
}
