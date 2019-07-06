// Package exif extracts EXIF metadata stored in JPG and PNG files. In addition
// to extracting the raw GPS parameters (such as latitude and longitude), it is
// also able of optionally performing city and country lookups using a
// geographical "lookuper" provider. The provider for GeoNames, an open
// geolocation database is provided in this package.
package exif

import (
	"io/ioutil"
	"time"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/dsoprea/go-exif"
)

// Exif chainable plugin context.
type Exif struct {
	dataKey  string
	lookuper Lookuper
}

type LookupData struct {
	City    string
	Country string
}

type Lookuper interface {
	Lookup(latitude, longitude float64) (*LookupData, error)
}

type geoData struct {
	Latitude  float64
	Longitude float64
	Altitude  int
	Timestamp time.Time

	Lookup *LookupData
}

type exifData struct {
	Geo *geoData
}

// New creates new instance of the Exif plugin.
func New() *Exif {
	return &Exif{dataKey: "Exif"}
}

func (plugin *Exif) Lookuper(lookuper Lookuper) *Exif {
	plugin.lookuper = lookuper
	return plugin
}

func (*Exif) Name() string {
	return "exif"
}

func (plugin *Exif) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.jpg", "**/*.jpeg", "**/*.png"), nil
}

func (plugin *Exif) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	var exifData exifData
	if geoData, err := plugin.extractGeo(inputFile); err == nil {
		exifData.Geo = geoData
	}

	inputFile.Meta[plugin.dataKey] = exifData
	context.DispatchFile(inputFile)
	return nil
}

// Based on https://godoc.org/github.com/dsoprea/go-exif#example-Ifd-GpsInfo
func (plugin *Exif) extractGeo(file *goldsmith.File) (*geoData, error) {
	rawFile, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	rawExif, err := exif.SearchAndExtractExif(rawFile)
	if err != nil {
		return nil, err
	}

	im := exif.NewIfdMapping()
	if err := exif.LoadStandardIfds(im); err != nil {
		return nil, err
	}

	ti := exif.NewTagIndex()
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return nil, err
	}

	ifd, err := index.RootIfd.ChildWithIfdPath(exif.IfdPathStandardGps)
	if err != nil {
		return nil, err
	}

	gi, err := ifd.GpsInfo()
	if err != nil {
		return nil, err
	}

	latitude := gi.Latitude.Decimal()
	longitude := gi.Longitude.Decimal()

	var lookupData *LookupData
	if plugin.lookuper != nil {
		lookupData, _ = plugin.lookuper.Lookup(latitude, longitude)
	}

	data := &geoData{
		latitude,
		longitude,
		gi.Altitude,
		gi.Timestamp,
		lookupData,
	}

	return data, nil
}
