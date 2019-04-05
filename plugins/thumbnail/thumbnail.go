// Package thumbnail automatically generates thumbnails for a variety of common
// image formats and saves them to a user-configurable path.
package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/nfnt/resize"
)

// Namer callback function builds thumbnail file paths based on the original file path.
// An empty path can be returned if a thumbnail should not be generated for the current file.
type Namer func(string, uint) string

// Thumbnail chainable context.
type Thumbnail struct {
	size  uint
	namer Namer
}

// New creates a new instance of the Thumbnail plugin.
func New() *Thumbnail {
	namer := func(path string, dims uint) string {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-thumb.png", body)
	}

	return &Thumbnail{128, namer}
}

// Size sets the desired maximum pixel size of generated thumbnails (default: 128).
func (plugin *Thumbnail) Size(dims uint) *Thumbnail {
	plugin.size = dims
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

func (*Thumbnail) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.jpg", "**/*.jpeg", "**/*.gif", "**/*.png"), nil
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
	inputImage, _, err := image.Decode(inputFile)
	if err != nil {
		return nil, err
	}

	var thumbBuff bytes.Buffer
	thumbImage := resize.Thumbnail(plugin.size, plugin.size, inputImage, resize.Bicubic)

	switch filepath.Ext(thumbPath) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&thumbBuff, thumbImage, nil)
	case ".gif":
		err = gif.Encode(&thumbBuff, thumbImage, nil)
	case ".png":
		err = png.Encode(&thumbBuff, thumbImage)
	}

	return context.CreateFileFromData(thumbPath, thumbBuff.Bytes()), nil
}
