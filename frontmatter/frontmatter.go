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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/FooSoft/goldsmith"
	"github.com/naoina/toml"
)

type frontMatter struct {
}

func New() goldsmith.Plugin {
	return &frontMatter{}
}

func (*frontMatter) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	switch filepath.Ext(strings.ToLower(f.Path())) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

func (fm *frontMatter) Process(ctx goldsmith.Context, f goldsmith.File) error {
	meta, body, err := parse(f)
	if err != nil {
		return err
	}

	nf := goldsmith.NewFileFromData(f.Path(), body.Bytes())
	nf.Apply(f.Meta())
	nf.Apply(meta)
	ctx.DispatchFile(nf)

	return nil
}

func parse(input io.Reader) (map[string]interface{}, *bytes.Buffer, error) {
	const (
		yamlOpener = "---"
		yamlCloser = "---"
		tomlOpener = "+++"
		tomlCloser = "+++"
		jsonOpener = "{"
		jsonCloser = "}"
	)

	var (
		body, front bytes.Buffer
		closer      string
		meta        = make(map[string]interface{})
		scanner     = bufio.NewScanner(input)
		header      = true
	)

	for scanner.Scan() {
		line := scanner.Text()

		if header {
			if len(closer) == 0 {
				switch strings.TrimSpace(line) {
				case tomlOpener:
					closer = tomlCloser
				case yamlOpener:
					closer = yamlCloser
				case jsonOpener:
					closer = jsonCloser
					front.WriteString(jsonOpener)
				default:
					header = false
				}
			} else {
				switch strings.TrimSpace(line) {
				case closer:
					if closer == jsonCloser {
						front.WriteString(jsonCloser)
					}
					header = false
				default:
					front.Write([]byte(line + "\n"))
				}
			}
		} else {
			body.Write([]byte(line + "\n"))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	if header {
		return nil, nil, errors.New("unterminated front matter block")
	}

	switch closer {
	case tomlCloser:
		if err := toml.Unmarshal(front.Bytes(), meta); err != nil {
			return nil, nil, err
		}
	case yamlCloser:
		if err := yaml.Unmarshal(front.Bytes(), meta); err != nil {
			return nil, nil, err
		}
	case jsonCloser:
		if err := json.Unmarshal(front.Bytes(), meta); err != nil {
			return nil, nil, err
		}
	}

	return meta, &body, nil
}
