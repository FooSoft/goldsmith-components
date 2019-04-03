// Package frontmatter extracts the metadata stored in your files. This
// metadata can include any information you want, but typically contains the
// page title, creation date, tags, layout template, and more. There are no
// requirements about what fields must be present; this is entirely up to you.
//
// +++
// Title = "My homepage"
// Tags = ["best", "page", "ever"]
// +++
//
// Metadata in YAML format is enclosed by three minus (-) characters:
//
// ---
// Title: "My homepage"
// Tags:
//   - "best"
//   - "page"
//   - "ever"
// ---
//
// Metadata in JSON format is enclosed by brace characters ({ and }):
//
// {
//     "Title": "My homepage",
//     "Tags": ["best", "page", "ever"]
// }
//
// Normal page content immediately follows the metadata section. The metadata
// section is stripped after processed by this plugin.
package frontmatter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
	"github.com/naoina/toml"
	"gopkg.in/yaml.v2"
)

// Frontmatter chainable plugin context.
type FrontMatter struct{}

// New creates a new instance of the Frontmatter plugin.
func New() *FrontMatter {
	return new(FrontMatter)
}

func (*FrontMatter) Name() string {
	return "frontmatter"
}

func (*FrontMatter) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.md", "**/*.markdown", "**/*.rst", "**/*.txt", "**/*.html", "**/*.htm"), nil
}

func (*FrontMatter) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	meta, body, err := parse(inputFile)
	if err != nil {
		return err
	}

	outputFile := context.CreateFileFromData(inputFile.Path(), body.Bytes())
	outputFile.Meta = inputFile.Meta
	for name, value := range meta {
		outputFile.Meta[name] = value
	}

	context.DispatchFile(outputFile)
	return nil
}

func parse(reader io.Reader) (map[string]interface{}, *bytes.Buffer, error) {
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
	)

	meta := make(map[string]interface{})
	scanner := bufio.NewScanner(reader)
	header := false
	first := true

	for scanner.Scan() {
		line := scanner.Text()

		if first {
			first = false

			if len(closer) == 0 {
				switch strings.TrimSpace(line) {
				case tomlOpener:
					header = true
					closer = tomlCloser
				case yamlOpener:
					header = true
					closer = yamlCloser
				case jsonOpener:
					header = true
					closer = jsonCloser
					front.WriteString(jsonOpener)
				}
			}

			if header {
				continue
			}
		}

		if header {
			switch strings.TrimSpace(line) {
			case closer:
				header = false
				if closer == jsonCloser {
					front.WriteString(jsonCloser)
				}
			default:
				front.Write([]byte(line + "\n"))
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
		if err := json.Unmarshal(front.Bytes(), &meta); err != nil {
			return nil, nil, err
		}
	}

	return meta, &body, nil
}
