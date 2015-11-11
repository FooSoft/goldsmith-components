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
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/nfnt/resize"
)

type Namer func(path string, dims uint) (string, bool)

type thumbnail struct {
	dims  uint
	namer Namer
}

func New(dims uint, namer Namer) (goldsmith.Chainer, error) {
	return &thumbnail{dims, namer}, nil
}

func (*thumbnail) Accept(file *goldsmith.File) bool {
	switch filepath.Ext(file.Path) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return true
	default:
		return false
	}
}

func (t *thumbnail) thumbName(path string) (string, bool) {
	if t.namer != nil {
		return t.namer(path, t.dims)
	}

	ext := filepath.Ext(path)
	body := strings.TrimSuffix(path, ext)

	return fmt.Sprintf("%s-thumb.png", body), true
}

func (t *thumbnail) cached(ctx goldsmith.Context, origPath, thumbPath string) bool {
	thumbPathFull := filepath.Join(ctx.DstDir(), thumbPath)
	thumbStat, err := os.Stat(thumbPathFull)
	if err != nil {
		return false
	}

	origPathFull := filepath.Join(ctx.SrcDir(), origPath)
	origStat, err := os.Stat(origPathFull)
	if err != nil {
		return false
	}

	return origStat.ModTime().Unix() <= thumbStat.ModTime().Unix()
}

func (t *thumbnail) thumbnail(ctx goldsmith.Context, origFile *goldsmith.File, thumbPath string) (*goldsmith.File, error) {
	origImg, _, err := image.Decode(bytes.NewReader(origFile.Buff.Bytes()))
	if err != nil {
		return nil, err
	}

	thumbImg := resize.Thumbnail(t.dims, t.dims, origImg, resize.Bicubic)
	thumbFile := goldsmith.NewFile(thumbPath)

	switch filepath.Ext(thumbPath) {
	case ".jpeg":
		fallthrough
	case ".jpg":
		thumbFile.Err = jpeg.Encode(&thumbFile.Buff, thumbImg, nil)
	case ".gif":
		thumbFile.Err = gif.Encode(&thumbFile.Buff, thumbImg, nil)
	case ".png":
		thumbFile.Err = png.Encode(&thumbFile.Buff, thumbImg)
	}

	return thumbFile, nil
}

func (t *thumbnail) Chain(ctx goldsmith.Context, input, output chan *goldsmith.File) {
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()
		close(output)
	}()

	for file := range input {
		wg.Add(1)
		go func(f *goldsmith.File) {
			defer func() {
				output <- f
				wg.Done()
			}()

			thumbPath, create := t.thumbName(f.Path)
			if !create {
				return
			}

			if t.cached(ctx, f.Path, thumbPath) {
				output <- goldsmith.NewFileRef(thumbPath)
				return
			}

			var tn *goldsmith.File
			if tn, f.Err = t.thumbnail(ctx, f, thumbPath); f.Err == nil {
				output <- tn
			}
		}(file)
	}
}
