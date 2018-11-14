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

	Dims(dims uint) Thumbnail
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
	dims  uint
	namer Namer
}

func (t *thumbnail) Dims(dims uint) Thumbnail {
	t.dims = dims
	return t
}

func (t *thumbnail) Namer(namer Namer) Thumbnail {
	t.namer = namer
	return t
}

func (*thumbnail) Name() string {
	return "thumbnail"
}

func (*thumbnail) Initialize(ctx *goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".jpg", ".jpeg", ".gif", ".png")}, nil
}

func (t *thumbnail) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	defer ctx.DispatchFile(f)

	thumbPath, create := t.namer(f.Path(), t.dims)
	if !create {
		return nil
	}

	fn, err := t.thumbnail(f, thumbPath)
	if err != nil {
		return err
	}

	ctx.CacheFile(f, fn)
	ctx.DispatchFile(fn)
	return nil
}

func (t *thumbnail) thumbnail(f *goldsmith.File, thumbPath string) (*goldsmith.File, error) {
	origImg, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	var thumbBuff bytes.Buffer
	thumbImg := resize.Thumbnail(t.dims, t.dims, origImg, resize.Bicubic)

	switch filepath.Ext(thumbPath) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&thumbBuff, thumbImg, nil)
	case ".gif":
		err = gif.Encode(&thumbBuff, thumbImg, nil)
	case ".png":
		err = png.Encode(&thumbBuff, thumbImg)
	}

	return goldsmith.NewFileFromData(thumbPath, thumbBuff.Bytes(), f.ModTime()), nil
}
