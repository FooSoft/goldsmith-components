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
	Geo  *geoData
	Time *time.Time
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
	rawFile, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return err
	}

	inputFile.Meta[plugin.dataKey] = plugin.extractExif(rawFile)
	context.DispatchFile(inputFile)

	return nil
}

func (plugin *Exif) extractExif(rawFile []byte) (exifData exifData) {
	rawExif, err := exif.SearchAndExtractExif(rawFile)
	if err != nil {
		return
	}

	im := exif.NewIfdMapping()
	if err := exif.LoadStandardIfds(im); err != nil {
		return
	}

	_, index, err := exif.Collect(im, exif.NewTagIndex(), rawExif)
	if err != nil {
		return
	}

	exifData.Time = extractTimestamp(index.RootIfd)
	if exifData.Geo = extractGeo(index.RootIfd); exifData.Geo != nil {
		if plugin.lookuper != nil {
			exifData.Geo.Lookup, _ = plugin.lookuper.Lookup(exifData.Geo.Latitude, exifData.Geo.Longitude)
		}
	}

	return
}

func extractGeo(rootIfd *exif.Ifd) *geoData {
	ifd, err := rootIfd.ChildWithIfdPath(exif.IfdPathStandardGps)
	if err != nil {
		return nil
	}

	gi, err := ifd.GpsInfo()
	if err != nil {
		return nil
	}

	return &geoData{gi.Latitude.Decimal(), gi.Longitude.Decimal(), gi.Altitude, gi.Timestamp, nil}
}

func extractTimestamp(rootIfd *exif.Ifd) *time.Time {
	if exifIfd, err := rootIfd.ChildWithIfdPath(exif.IfdPathStandardExif); err == nil {
		if timestamp := extractTimestampByName(exifIfd, "DateTimeOriginal"); timestamp != nil {
			return timestamp
		}
		if timestamp := extractTimestampByName(exifIfd, "DateTimeDigitized"); timestamp != nil {
			return timestamp
		}
	}

	return extractTimestampByName(rootIfd, "DateTime")
}

func extractTimestampByName(ifd *exif.Ifd, tagName string) *time.Time {
	results, err := ifd.FindTagWithName(tagName)
	if err != nil || len(results) == 0 {
		return nil
	}

	valueRaw, err := ifd.TagValue(results[0])
	if err != nil {
		return nil
	}

	valueStr, ok := valueRaw.(string)
	if !ok {
		return nil
	}

	timestamp, err := exif.ParseExifFullTimestamp(valueStr)
	if err != nil {
		return nil
	}

	return &timestamp
}
