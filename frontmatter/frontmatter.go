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

package frontmatter

import (
	"path"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/gernest/front"
)

type frontMatter struct {
	matter *front.Matter
}

func New() goldsmith.Context {
	fm := &frontMatter{front.NewMatter()}
	fm.matter.Handle("---", front.YAMLHandler)
	return goldsmith.Context{fm, nil}
}

func (fm *frontMatter) TaskSingle(ctx goldsmith.Context, file goldsmith.File) goldsmith.File {
	ext := strings.ToLower(path.Ext(file.Path))
	if ext != ".md" && ext != ".markdown" {
		return file
	}

	front, body, err := fm.matter.Parse(&file.Buff)
	if err != nil {
		file.Err = err
		return file
	}

	file.Buff.Reset()
	file.Buff.Write([]byte(body))

	for key, value := range front {
		file.Meta[key] = value
	}

	return file
}
