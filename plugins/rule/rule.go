package rule

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"foosoft.net/projects/goldsmith"
	"github.com/BurntSushi/toml"
	"github.com/bmatcuk/doublestar"
)

type rule struct {
	Match   []string
	Unmatch []string
	Props   map[string]goldsmith.Prop

	baseDir string
}

func (self *rule) matches(inputFile *goldsmith.File) bool {
	patternDir := filepath.Join(self.baseDir, "**")
	if match, err := doublestar.PathMatch(patternDir, inputFile.Path()); !match || err != nil {
		return false
	}

	for _, pattern := range self.Match {
		patternAbs := filepath.Join(self.baseDir, pattern)
		if match, err := doublestar.PathMatch(patternAbs, inputFile.Path()); match && err == nil {
			return true
		}
	}

	for _, pattern := range self.Unmatch {
		patternAbs := filepath.Join(self.baseDir, pattern)
		if match, err := doublestar.PathMatch(patternAbs, inputFile.Path()); match && err == nil {
			return false
		}
	}

	return len(self.Unmatch) > 0
}

func (self *rule) apply(inputFile *goldsmith.File) bool {
	if self.matches(inputFile) {
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
