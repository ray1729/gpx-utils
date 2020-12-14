package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/ray1729/gpx-utils/pkg/placenames"
)

func main() {
	log.SetFlags(0)
	w := csv.NewWriter(os.Stdout)
	for _, filename := range os.Args[1:] {
		ts, err := readTrackSummary(filename)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(constructRow(path.Base(filename), ts))
	}
	w.Flush()
}

func constructRow(filename string, ts *placenames.TrackSummary) []string {
	row := make([]string, 5)
	row[0] = filename[:10]
	row[1] = ts.Start
	row[2] = ts.Finish
	row[3] = strconv.FormatFloat(ts.Distance, 'f', 1, 32)
	row[4] = strconv.FormatFloat(ts.Ascent, 'f', 0, 32)
	return row
}

func readTrackSummary(path string) (*placenames.TrackSummary, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ts placenames.TrackSummary
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, err
	}
	return &ts, nil
}
