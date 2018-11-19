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
	"github.com/FooSoft/goldsmith-components/filters/extension"
	"github.com/nfnt/resize"
)

type Namer func(string, uint) (string, bool)

type Thumbnail interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor

	Size(size uint) Thumbnail
	Namer(namer Namer) Thumbnail
}

func New() Thumbnail {
	namer := func(path string, dims uint) (string, bool) {
		ext := filepath.Ext(path)
		body := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s-thumb.png", body), true
	}

	return &thumbnail{128, namer}
}

type thumbnail struct {
	size  uint
	namer Namer
}

func (t *thumbnail) Size(dims uint) Thumbnail {
	t.size = dims
	return t
}

func (t *thumbnail) Namer(namer Namer) Thumbnail {
	t.namer = namer
	return t
}

func (*thumbnail) Name() string {
	return "thumbnail"
}

func (*thumbnail) Initialize(context *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".jpg", ".jpeg", ".gif", ".png")}, nil
}

func (t *thumbnail) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	defer context.DispatchFile(inputFile, false)

	thumbPath, create := t.namer(inputFile.Path(), t.size)
	if !create {
		return nil
	}

	if outputFile := context.RetrieveCachedFile(thumbPath, inputFile.Path()); outputFile != nil {
		context.DispatchFile(outputFile, false)
		return nil
	}

	outputFile, err := t.thumbnail(inputFile, thumbPath)
	if err != nil {
		return err
	}

	context.DispatchFile(outputFile, true)
	return nil
}

func (t *thumbnail) thumbnail(inputFile *goldsmith.File, thumbPath string) (*goldsmith.File, error) {
	inputImage, _, err := image.Decode(inputFile)
	if err != nil {
		return nil, err
	}

	var thumbBuff bytes.Buffer
	thumbImage := resize.Thumbnail(t.size, t.size, inputImage, resize.Bicubic)

	switch filepath.Ext(thumbPath) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&thumbBuff, thumbImage, nil)
	case ".gif":
		err = gif.Encode(&thumbBuff, thumbImage, nil)
	case ".png":
		err = png.Encode(&thumbBuff, thumbImage)
	}

	return goldsmith.NewFileFromData(thumbPath, thumbBuff.Bytes()), nil
}
