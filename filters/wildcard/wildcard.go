package wildcard

import (
	"foosoft.net/projects/goldsmith"
	"github.com/bmatcuk/doublestar"
)

type Wildcard struct {
	wildcards []string
}

func New(wildcards ...string) *Wildcard {
	return &Wildcard{wildcards}
}

func (*Wildcard) Name() string {
	return "wildcard"
}

func (self *Wildcard) Accept(file *goldsmith.File) bool {
	filePath := file.Path()

	for _, wildcard := range self.wildcards {
		if matched, _ := doublestar.PathMatch(wildcard, filePath); matched {
			return true
		}
	}

	return false
}
