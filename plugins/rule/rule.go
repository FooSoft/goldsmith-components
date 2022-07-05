package rule

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"sync"

	"foosoft.net/projects/goldsmith"
	"github.com/BurntSushi/toml"
	"github.com/bmatcuk/doublestar"
)

type rule struct {
	Patterns []string
	Props    map[string]goldsmith.Prop
	Drop     bool
}

func (self *rule) rebase(inputFile *goldsmith.File) error {
	for i, path := range self.Patterns {
		if filepath.IsAbs(path) {
			return errors.New("rule paths must be relative")
		}

		self.Patterns[i] = filepath.Join(inputFile.Dir(), path)
	}

	return nil
}

func (self *rule) apply(inputFile *goldsmith.File) bool {
	var matched bool
	for _, pattern := range self.Patterns {
		if match, err := doublestar.PathMatch(pattern, inputFile.Path()); match && err == nil {
			matched = true
			break
		}
	}

	if matched {
		if self.Drop {
			return false
		}

		for name, value := range self.Props {
			inputFile.SetProp(name, value)
		}
	}

	return true
}

type ruleSet struct {
	Rules []*rule
}

func (self *ruleSet) rebase(inputFile *goldsmith.File) error {
	for _, rule := range self.Rules {
		if err := rule.rebase(inputFile); err != nil {
			return err
		}
	}

	return nil
}

func newRuleSet(inputFile *goldsmith.File) (*ruleSet, error) {
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}

	var ruleSet ruleSet
	if err := toml.Unmarshal(data, &ruleSet); err != nil {
		return nil, err
	}

	if err := ruleSet.rebase(inputFile); err != nil {
		return nil, err
	}

	return &ruleSet, nil
}

func (self *ruleSet) apply(inputFile *goldsmith.File) bool {
	for _, rule := range self.Rules {
		if !rule.apply(inputFile) {
			return false
		}
	}

	return true
}

// Rule chainable context.
type Rule struct {
	filename string

	ruleSets   []*ruleSet
	inputFiles []*goldsmith.File
	mutex      sync.Mutex
}

// New creates a new instance of the Rule plugin.
func New() *Rule {
	return &Rule{filename: "rules.toml"}
}

func (self *Rule) Filename(filename string) *Rule {
	self.filename = filename
	return self
}

func (*Rule) Name() string {
	return "rule"
}

func (self *Rule) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	if inputFile.Name() == self.filename {
		ruleSet, err := newRuleSet(inputFile)
		if err != nil {
			return err
		}

		self.mutex.Lock()
		self.ruleSets = append(self.ruleSets, ruleSet)
		self.mutex.Unlock()
	} else {
		self.mutex.Lock()
		self.inputFiles = append(self.inputFiles, inputFile)
		self.mutex.Unlock()
	}

	return nil
}

func (self *Rule) Finalize(context *goldsmith.Context) error {
	for _, inputFile := range self.inputFiles {
		var block bool
		for _, ruleSet := range self.ruleSets {
			if !ruleSet.apply(inputFile) {
				block = true
				break
			}
		}

		if !block {
			context.DispatchFile(inputFile)
		}
	}

	return nil
}
