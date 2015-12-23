/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

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
	"github.com/nfnt/resize"
)

type Namer func(path string, dims uint) (string, bool)

type thumbnail struct {
	dims  uint
	namer Namer
}

func New(dims uint, namer Namer) goldsmith.Plugin {
	return &thumbnail{dims, namer}
}

func (*thumbnail) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return true
	default:
		return false
	}
}

func (t *thumbnail) Process(ctx goldsmith.Context, f goldsmith.File) error {
	if t.cached(ctx, f.Path(), f.Path()) {
		ctx.ReferenceFile(f.Path())
	} else {
		defer ctx.DispatchFile(f)
	}

	thumbPath, create := t.thumbName(f.Path())
	if !create {
		return nil
	}

	if t.cached(ctx, f.Path(), thumbPath) {
		ctx.ReferenceFile(thumbPath)
		return nil
	}

	fn, err := t.thumbnail(f, thumbPath)
	if err != nil {
		return err
	}
	ctx.DispatchFile(fn)

	return nil
}

func (t *thumbnail) thumbName(path string) (string, bool) {
	if t.namer != nil {
		return t.namer(path, t.dims)
	}

	ext := filepath.Ext(path)
	body := strings.TrimSuffix(path, ext)

	return fmt.Sprintf("%s-thumb.png", body), true
}

func (t *thumbnail) cached(ctx goldsmith.Context, srcPath, dstPath string) bool {
	srcPathFull := filepath.Join(ctx.SrcDir(), srcPath)
	srcStat, err := os.Stat(srcPathFull)
	if err != nil {
		return false
	}

	dstPathFull := filepath.Join(ctx.DstDir(), dstPath)
	dstStat, err := os.Stat(dstPathFull)
	if err != nil {
		return false
	}

	return dstStat.ModTime().Unix() >= srcStat.ModTime().Unix() && dstStat.Size() == srcStat.Size()
}

func (t *thumbnail) thumbnail(f goldsmith.File, thumbPath string) (goldsmith.File, error) {
	origImg, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	var thumbBuff bytes.Buffer
	thumbImg := resize.Thumbnail(t.dims, t.dims, origImg, resize.Bicubic)

	switch filepath.Ext(strings.ToLower(thumbPath)) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&thumbBuff, thumbImg, nil)
	case ".gif":
		err = gif.Encode(&thumbBuff, thumbImg, nil)
	case ".png":
		err = png.Encode(&thumbBuff, thumbImg)
	}

	return goldsmith.NewFileFromData(thumbPath, thumbBuff.Bytes()), nil
}
