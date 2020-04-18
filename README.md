# GPX Utils

Utilities for analyzing and indexing GPX routes.

## Compiling placenames.bin

This step extracts bounding boxes for populated places from  the OS Open Names dataset, which is available for free download: https://www.ordnancesurvey.co.uk/business-government/products/open-map-names

    go run ./cmd/save-gob/... opname_csv_gb.zip ./pkg/placenames/placenames.bin

I have included a compiled extract in this repository so you can skip this step.

## Binary embedding

We use [mule](https://github.com/wlbr/mule) to embed the gob data in the compiled binaries:

    go get github.com/wlbr/mule

Make sure the `mule` command is on your PATH before running `go generate`.
    
## Compiling

    mkdir -p bin
    go generate ./...
    go build -o bin ./...

## Usage

### analyze-gpx

To analyze a single GPX track:

    ./bin/analyze-gpx FILENAME
    
This will write a JSON summary to STDOUT.

To analyze an entire directory:

    ./bin/analyze-gpx DIRNAME
    
This will scan the directory and, for each file with suffix `.gpx`, will output the analysis to a corresponding file with suffix `.json`.

### serve-rwgps

This will start a small server to analyze [RideWithGPS](https://ridewithgps.com/) tracks. 

    ./bin/serve-rwgps

It defaults to listening on port 8000, override by setting the `LISTEN_ADDR` environment variable:

    LISTEN_ADDR=127.0.0.1:3000 ./bin/serve-rwgps
    
Then to query a route:

    curl http://localhost:3000/rwgps?routeId=30165378  

## Attribution

Contains OS data © Crown copyright and database right 2018

## MIT License

Copyright © 2020 Raymond Miller

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.