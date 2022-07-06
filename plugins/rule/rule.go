package rule

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"foosoft.net/projects/goldsmith"
	"foosoft.net/projects/goldsmith-components/filters/operator"
	"foosoft.net/projects/goldsmith-components/filters/wildcard"
	"github.com/BurntSushi/toml"
)

type rule struct {
	Accept []string
	Reject []string
	Props  map[string]goldsmith.Prop

	baseDir string
}

func (self *rule) accept(inputFile *goldsmith.File) bool {
	if !wildcard.New(filepath.Join(self.baseDir, "**")).Accept(inputFile) {
		return false
	}

	var acceptPaths []string
	for _, path := range self.Accept {
		acceptPaths = append(acceptPaths, filepath.Join(self.baseDir, path))
	}

	if wildcard.New(acceptPaths...).Accept(inputFile) {
		return true
	}

	var rejectPaths []string
	for _, path := range self.Reject {
		rejectPaths = append(rejectPaths, filepath.Join(self.baseDir, path))
	}

	if len(rejectPaths) == 0 {
		return false
	}

	return operator.Not(wildcard.New(rejectPaths...)).Accept(inputFile)
}

func (self *rule) apply(inputFile *goldsmith.File) bool {
	if self.accept(inputFile) {
		if len(self.Props) == 0 {
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

func newRuleSet(inputFile *goldsmith.File) (*ruleSet, error) {
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}

	var ruleSet ruleSet
	if err := toml.Unmarshal(data, &ruleSet); err != nil {
		return nil, err
	}

	for _, rule := range ruleSet.Rules {
		rule.baseDir = inputFile.Dir()
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
