package openname

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"strconv"

	"github.com/dhconnelly/rtreego"
)

type Record struct {
	ID                    string
	NamesUri              string
	Name                  string
	NameLang              string
	AltName               string
	AltNameLang           string
	Type                  string
	LocalType             string
	GeomX                 float64
	GeomY                 float64
	MostDetailViewRes     float64
	LeastDetailViewRes    float64
	MbrXMin               float64
	MbrYMin               float64
	MbrXMax               float64
	MbrYMax               float64
	PostcodeDistrict      string
	PostcodeDistrictUri   string
	PopulatedPlace        string
	PopulatedPlaceUri     string
	PopulatedPlaceType    string
	DistrictBorough       string
	DistrictBoroughUri    string
	DistrictBoroughType   string
	CountyUnitary         string
	ConutyUnitaryUri      string
	CountyUnitaryType     string
	Region                string
	RegionUri             string
	Country               string
	CountryUri            string
	RelativeSpatialObject string
	SameAsDbpedia         string
	SameAsGeonames        string
}

func (r *Record) Bounds() *rtreego.Rect {
	p := rtreego.Point{r.MbrXMin, r.MbrYMin}
	rect, err := rtreego.NewRect(p, []float64{r.MbrXMax - r.MbrXMin, r.MbrYMax - r.MbrYMin})
	if err != nil {
		panic(err)
	}
	return rect
}

func (r *Record) Area() float64 {
	return (r.MbrXMax - r.MbrXMin) * (r.MbrYMax - r.MbrYMin)
}

type Scanner struct {
	csvReader  *csv.Reader
	nextRecord *Record
	err        error
}

func NewScanner(r io.Reader) (*Scanner, error) {
	br := bufio.NewReader(r)
	err := skipBOM(br)
	if err != nil {
		return nil, err
	}
	return &Scanner{csvReader: csv.NewReader(br)}, nil
}

var BOM = [3]byte{0xef, 0xbb, 0xbf}

func skipBOM(br *bufio.Reader) error {
	xs, err := br.Peek(3)
	if err != nil {
		return err
	}
	if xs[0] == BOM[0] && xs[1] == BOM[1] && xs[2] == BOM[2] {
		br.Discard(3)
	}
	return nil
}

func (s *Scanner) Scan() bool {
	rawRecord, err := s.csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false
		}
		s.err = err
		return false
	}
	s.nextRecord, err = parseRecord(rawRecord)
	if err != nil {
		s.err = err
		return false
	}
	return true
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Record() *Record {
	return s.nextRecord
}

func parseRecord(xs []string) (*Record, error) {
	if len(xs) != 34 {
		return nil, csv.ErrFieldCount
	}
	record := Record{
		ID:                    xs[0],
		NamesUri:              xs[1],
		Name:                  xs[2],
		NameLang:              xs[3],
		AltName:               xs[4],
		AltNameLang:           xs[5],
		Type:                  xs[6],
		LocalType:             xs[7],
		GeomX:                 0,
		GeomY:                 0,
		MostDetailViewRes:     0,
		LeastDetailViewRes:    0,
		MbrXMin:               0,
		MbrYMin:               0,
		MbrXMax:               0,
		MbrYMax:               0,
		PostcodeDistrict:      xs[16],
		PostcodeDistrictUri:   xs[17],
		PopulatedPlace:        xs[18],
		PopulatedPlaceUri:     xs[19],
		PopulatedPlaceType:    xs[20],
		DistrictBorough:       xs[21],
		DistrictBoroughUri:    xs[22],
		DistrictBoroughType:   xs[23],
		CountyUnitary:         xs[24],
		ConutyUnitaryUri:      xs[25],
		CountyUnitaryType:     xs[26],
		Region:                xs[27],
		RegionUri:             xs[28],
		Country:               xs[29],
		CountryUri:            xs[30],
		RelativeSpatialObject: xs[31],
		SameAsDbpedia:         xs[32],
		SameAsGeonames:        xs[33],
	}

	for i, p := range []*float64{&record.GeomX, &record.GeomY, &record.MostDetailViewRes, &record.LeastDetailViewRes, &record.MbrXMin, &record.MbrYMin, &record.MbrXMax, &record.MbrYMax} {
		s := xs[i+8]
		if s != "" {
			var err error
			*p, err = strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
		}
	}
	return &record, nil
}
