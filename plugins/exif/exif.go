package exif

import (
	"io"
	"time"

	"github.com/FooSoft/geolookup"
	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	goexif "github.com/rwcarlsen/goexif/exif"
)

// Exif chainable plugin context.
type Exif struct {
	exifKey   string
	geoData   io.Reader
	geoLookup *geolookup.Lookup
}

type data struct {
	Time      *time.Time
	Latitude  *float64
	Longitude *float64
	City      *string
	Country   *string
}

// New creates new instance of the Exif plugin.
func New() *Exif {
	return &Exif{exifKey: "Exif"}
}

// Sets a reader to the geodata used to do geoip lookups.
func (plugin *Exif) GeoData(geoData io.Reader) *Exif {
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
	data := new(data)
	inputFile.Meta[plugin.exifKey] = data
	defer context.DispatchFile(inputFile)

	x, err := goexif.Decode(inputFile)
	if err != nil {
		return nil
	}

	if time, err := x.DateTime(); err == nil {
		data.Time = &time
	}

	if latitude, longitude, err := x.LatLong(); err == nil {
		data.Latitude = &latitude
		data.Longitude = &longitude

		if plugin.geoLookup != nil {
			if location := plugin.geoLookup.Find(latitude, longitude); location != nil {
				data.City = &location.City
				data.Country = &location.Country
			}
		}
	}

	return nil
}
