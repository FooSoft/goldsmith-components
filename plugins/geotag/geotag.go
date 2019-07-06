package geotag

import (
	"io/ioutil"
	"time"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/dsoprea/go-exif"
)

// GeoTag chainable plugin context.
type GeoTag struct {
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

// New creates new instance of the GeoTag plugin.
func New() *GeoTag {
	return &GeoTag{dataKey: "GeoTag"}
}

func (plugin *GeoTag) Lookuper(lookuper Lookuper) *GeoTag {
	plugin.lookuper = lookuper
	return plugin
}

func (*GeoTag) Name() string {
	return "geotag"
}

func (plugin *GeoTag) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.jpg", "**/*.jpeg"), nil
}

func (plugin *GeoTag) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	defer context.DispatchFile(inputFile)
	inputFile.Meta[plugin.dataKey], _ = plugin.extractExif(inputFile)
	return nil
}

// Based on https://godoc.org/github.com/dsoprea/go-exif#example-Ifd-GpsInfo
func (plugin *GeoTag) extractExif(file *goldsmith.File) (*geoData, error) {
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
