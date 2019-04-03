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

type Namer func(string, uint) (string, bool)

type Thumbnail struct {
	size  uint
	namer Namer
}

func New() *Thumbnail {
	namer := func(path string, dims uint) (string, bool) {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-thumb.png", body), true
	}

	return &Thumbnail{128, namer}
}

func (plugin *Thumbnail) Size(dims uint) *Thumbnail {
	plugin.size = dims
	return plugin
}

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

	thumbPath, create := plugin.namer(inputFile.Path(), plugin.size)
	if !create {
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
