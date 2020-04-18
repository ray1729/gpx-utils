package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/ray1729/gpx-utils/pkg/placenames"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("Usage: %s GPX_FILE_OR_DIRECTORY")
	}
	inFile := os.Args[1]
	info, err := os.Stat(inFile)
	if err != nil {
		log.Fatal(err)
	}
	gs, err := placenames.NewGPXSummarizer()
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		err = summarizeDirectory(gs, inFile)
	} else {
		err = summarizeSingleFile(gs, inFile)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func summarizeDirectory(gs *placenames.GPXSummarizer, dirName string) error {
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

func summarizeSingleFile(gs *placenames.GPXSummarizer, filename string) error {
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

func writeSummary(s *placenames.TrackSummary, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(s); err != nil {
		return err
	}
	return nil
}
