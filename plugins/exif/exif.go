package exif

import (
	"io"
	"io/ioutil"
	"time"

	"github.com/FooSoft/geolookup"
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/dsoprea/go-exif"
)

// Exif chainable plugin context.
type Exif struct {
	exifKey   string
	geoData   io.Reader
	geoLookup *geolookup.Lookup
}

type geoLookup struct {
	City    *string
	Country *string
}

type geoData struct {
	Latitude  float64
	Longitude float64
	Altitude  int
	Timestamp time.Time
	Lookup    *geoLookup
}

// New creates new instance of the Exif plugin.
func New() *Exif {
	return &Exif{exifKey: "Exif"}
}

// Sets a reader to the geodata used to do geoip lookups.
func (plugin *Exif) Lookup(geoData io.Reader) *Exif {
	plugin.geoData = geoData
	return plugin
}

func (*Exif) Name() string {
	return "exif"
}

func (plugin *Exif) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	if plugin.geoData != nil {
		plugin.geoLookup = geolookup.New()
		if err := plugin.geoLookup.Load(plugin.geoData); err != nil {
			return nil, err
		}
	}

	return wildcard.New("**/*.jpg", "**/*.jpeg"), nil
}

func (plugin *Exif) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	defer context.DispatchFile(inputFile)
	inputFile.Meta[plugin.exifKey], _ = plugin.extractGeoData(inputFile)
	return nil
}

// Based on https://godoc.org/github.com/dsoprea/go-exif#example-Ifd-GpsInfo
func (plugin *Exif) extractGeoData(file *goldsmith.File) (*geoData, error) {
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

	var lookup *geoLookup
	if plugin.geoLookup != nil {
		if location := plugin.geoLookup.Find(gi.Latitude.Decimal(), gi.Longitude.Decimal()); location != nil {
			lookup = &geoLookup{&location.City, &location.Country}
		}
	}

	data := &geoData{
		gi.Latitude.Decimal(),
		gi.Longitude.Decimal(),
		gi.Altitude,
		gi.Timestamp,
		lookup,
	}

	return data, nil
}
