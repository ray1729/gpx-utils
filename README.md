# GPX Utils

Utilities for analyzing and indexing GPX routes.

## Compiling placenames.bin

This step extracts bounding boxes for populated places from  the OS Open Names dataset, which is available for free download: https://www.ordnancesurvey.co.uk/business-government/products/open-map-names

    go run ./cmd/save-gob/... opname_csv_gb.zip ./pkg/placenames/placenames.bin

I have included a the compiled extract in this repository so you can skip this step.

## Binary embedding

We use [mule](https://github.com/wlbr/mule) to embed the gob data in the compiled binaries:

    go get github.com/wlbr/mule

Make sure the `mule` command is on you PATH before running `go generate`.
    
## Compiling

    mudir -p bin
    go generate ./...
    go build -o bin ./...

## Attribution

Contains OS data © Crown copyright and database right 2018
