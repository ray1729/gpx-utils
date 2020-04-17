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

	"github.com/ray1729/gpx-utils/pkg/openname"
)

func main() {
	openNames := flag.String("opname", "", "Path to Ordnance Server Open Names zip archive")
	gpxFile := flag.String("gpx", "", "Path to GPX file")
	dirName := flag.String("dir", "", "Directory to scan for GPX files")
	flag.Parse()
	if *openNames == "" {
		log.Fatal("--opname is required")
	}
	if (*gpxFile == "" && *dirName == "") || (*gpxFile != "" && *dirName != "") {
		log.Fatal("exactly one of --dir or --gpx is required")
	}
	rt, err := openname.BuildIndex(*openNames)
	if err != nil {
		log.Fatal(err)
	}
	gs, err := openname.NewGPXSummarizer(rt)
	if err != nil {
		log.Fatal(err)
	}
	if *gpxFile != "" {
		err = summarizeSingleFile(gs, *gpxFile)
	} else {
		err = summarizeDirectory(gs, *dirName)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func summarizeDirectory(gs *openname.GPXSummarizer, dirName string) error {
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
		summary, err := gs.SummarizeTrack(r)
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

func summarizeSingleFile(gs *openname.GPXSummarizer, filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening %s for reading: %v", filename, err)
	}
	summary, err := gs.SummarizeTrack(r)
	if err != nil {
		return fmt.Errorf("error creating summary of GPX track %s: %v", filename, err)
	}
	if err = writeSummary(summary, os.Stdout); err != nil {
		return fmt.Errorf("error marshalling summary for %s: %v", filename, err)
	}
	return nil
}

func writeSummary(s *openname.TrackSummary, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(s); err != nil {
		return err
	}
	return nil
}
