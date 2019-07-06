package geotag

import (
	"io"

	"github.com/FooSoft/geolookup"
)

type LookuperGeonames struct {
	lookup *geolookup.Lookup
}

func NewLookuperGeonames(reader io.Reader) (*LookuperGeonames, error) {
	lookup := geolookup.New()
	if err := lookup.Load(reader); err != nil {
		return nil, err
	}

	return &LookuperGeonames{lookup}, nil
}

func NewLookuperGeonamesFile(path string) (*LookuperGeonames, error) {
	lookup := geolookup.New()
	if err := lookup.LoadFile(path); err != nil {
		return nil, err
	}

	return &LookuperGeonames{lookup}, nil
}

func (lookuper *LookuperGeonames) Lookup(latitude, longitude float64) (*LookupData, error) {
	location, err := lookuper.lookup.Find(latitude, longitude)
	if err != nil {
		return nil, err
	}

	return &LookupData{location.City, location.Country}, nil
}
