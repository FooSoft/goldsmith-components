package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
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

	return &thumbnailPlugin{128, namer}
}

type thumbnailPlugin struct {
	dims  uint
	namer Namer
}

func (plugin *thumbnailPlugin) Dims(dims uint) Thumbnail {
	plugin.dims = dims
	return plugin
}

func (plugin *thumbnailPlugin) Namer(namer Namer) Thumbnail {
	plugin.namer = namer
	return plugin
}

func (*thumbnailPlugin) Name() string {
	return "thumbnail"
}

func (*thumbnailPlugin) Initialize(context goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".jpg", ".jpeg", ".gif", ".png")}, nil
}

func (plugin *thumbnailPlugin) Process(context goldsmith.Context, f goldsmith.File) error {
	defer context.DispatchFile(f)

	thumbPath, create := plugin.namer(f.Path(), plugin.dims)
	if !create {
		return nil
	}

	var (
		fn  goldsmith.File
		err error
	)

	if cached(context, f.Path(), thumbPath) {
		thumbPathDst := filepath.Join(context.DstDir(), thumbPath)
		fn, err = goldsmith.NewFileFromAsset(thumbPath, thumbPathDst)
		if err != nil {
			return err
		}
	} else {
		var err error
		fn, err = plugin.thumbnail(f, thumbPath)
		if err != nil {
			return err
		}
	}

	context.DispatchFile(fn)
	return nil
}

func (plugin *thumbnailPlugin) thumbnail(f goldsmith.File, thumbPath string) (goldsmith.File, error) {
	origImg, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	var thumbBuff bytes.Buffer
	thumbImg := resize.Thumbnail(plugin.dims, plugin.dims, origImg, resize.Bicubic)

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

func cached(context goldsmith.Context, srcPath, dstPath string) bool {
	srcPathFull := filepath.Join(context.SrcDir(), srcPath)
	srcStat, err := os.Stat(srcPathFull)
	if err != nil {
		return false
	}

	dstPathFull := filepath.Join(context.DstDir(), dstPath)
	dstStat, err := os.Stat(dstPathFull)
	if err != nil {
		return false
	}

	return dstStat.ModTime().Unix() >= srcStat.ModTime().Unix()
}
