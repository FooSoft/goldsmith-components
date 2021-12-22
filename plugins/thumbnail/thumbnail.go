// Package thumbnail automatically generates thumbnails for a variety of common
// image formats and saves them to a user-configurable path.
package thumbnail

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/disintegration/imaging"
)

// Namer callback function builds thumbnail file paths based on the original file path.
// An empty path can be returned if a thumbnail should not be generated for the current file.
type Namer func(string, int) string

// Desired thumbnailing styles.
type Style int

const (
	Fit Style = iota
	Crop
	Pad
)

// Thumbnail chainable context.
type Thumbnail struct {
	size  int
	style Style
	color color.Color
	namer Namer
}

// New creates a new instance of the Thumbnail plugin.
func New() *Thumbnail {
	namer := func(path string, dims int) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-thumb.png", body)
	}

	return &Thumbnail{
		128,
		Fit,
		color.Transparent,
		namer,
	}
}

// Size sets the desired maximum pixel size of generated thumbnails (default: 128).
func (plugin *Thumbnail) Size(dims int) *Thumbnail {
	plugin.size = dims
	return plugin
}

// Style sets the desired thumbnailing style.
func (plugin *Thumbnail) Style(style Style) *Thumbnail {
	plugin.style = style
	return plugin
}

// Namer sets the callback used to build paths for thumbnail files.
// Default naming appends "-thumb" to the path and changes extension to PNG,
// for example "file.jpg" becomes "file-thumb.png".
func (plugin *Thumbnail) Namer(namer Namer) *Thumbnail {
	plugin.namer = namer
	return plugin
}

func (*Thumbnail) Name() string {
	return "thumbnail"
}

func (*Thumbnail) Initialize(context *goldsmith.Context) error {
	context.Filter(wildcard.New("**/*.jpg", "**/*.jpeg", "**/*.gif", "**/*.png"))
	return nil
}

func (plugin *Thumbnail) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	defer context.DispatchFile(inputFile)

	thumbPath := plugin.namer(inputFile.Path(), plugin.size)
	if len(thumbPath) == 0 {
		return nil
	}

	if outputFile := context.RetrieveCachedFile(thumbPath, inputFile); outputFile != nil {
		outputFile.Meta = inputFile.Meta
		context.DispatchFile(outputFile)
		return nil
	}

	outputFile, err := plugin.thumbnail(context, inputFile, thumbPath)
	if err != nil {
		return err
	}

	context.DispatchAndCacheFile(outputFile, inputFile)
	return nil
}

func (plugin *Thumbnail) thumbnail(context *goldsmith.Context, inputFile *goldsmith.File, thumbPath string) (*goldsmith.File, error) {
	var thumbFormat imaging.Format
	switch strings.ToLower(filepath.Ext(thumbPath)) {
	case ".jpg", ".jpeg":
		thumbFormat = imaging.JPEG
	case ".gif":
		thumbFormat = imaging.GIF
	case ".png":
		thumbFormat = imaging.PNG
	default:
		return nil, errors.New("unsupported image format")
	}

	thumbImage, err := imaging.Decode(inputFile)
	if err != nil {
		return nil, err
	}

	switch plugin.style {
	case Fit:
		thumbImage = imaging.Fit(thumbImage, plugin.size, plugin.size, imaging.Lanczos)
	case Crop:
		thumbImage = imaging.Fill(thumbImage, plugin.size, plugin.size, imaging.Center, imaging.Lanczos)
	case Pad:
		thumbImage = imaging.Fit(thumbImage, plugin.size, plugin.size, imaging.Lanczos)
		thumbImage = imaging.PasteCenter(imaging.New(plugin.size, plugin.size, plugin.color), thumbImage)
	default:
		return nil, errors.New("unsupported thumbnailing style")
	}

	var thumbBuff bytes.Buffer
	if err := imaging.Encode(&thumbBuff, thumbImage, thumbFormat); err != nil {
		return nil, err
	}

	return context.CreateFileFromData(thumbPath, thumbBuff.Bytes()), nil
}
