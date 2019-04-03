// Package collection groups related pages into named collections.
package collection

import (
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
)

// A Comparer callback function is used to sort files within a collection group.
type Comparer func(i, j *goldsmith.File) (less bool)

// Collection chainable plugin context.
type Collection struct {
	collectionKey string
	groupsKey     string
	comparer      Comparer

	groups map[string][]*goldsmith.File
	files  []*goldsmith.File
	mutex  sync.Mutex
}

// New creates a new instance of the Collection plugin.
func New() *Collection {
	return &Collection{
		collectionKey: "Collection",
		groupsKey:     "Groups",
		comparer:      nil,
		groups:        make(map[string][]*goldsmith.File),
	}
}

// CollectionKey sets the metadata key used to access the collection name (default: "Collection").
func (plugin *Collection) CollectionKey(collectionKey string) *Collection {
	plugin.collectionKey = collectionKey
	return plugin
}

// GroupsKey sets the metadata key used to store information about collection groups (default: "Groups").
// This information is stored as a mapping of group names to contained files (map[string][]goldsmith.File).
func (plugin *Collection) GroupsKey(groupsKey string) *Collection {
	plugin.groupsKey = groupsKey
	return plugin
}

// Comparer sets the function used to sort files in collection groups (default: sort by filenames).
func (plugin *Collection) Comparer(comparer Comparer) *Collection {
	plugin.comparer = comparer
	return plugin
}

func (*Collection) Name() string {
	return "collection"
}

func (*Collection) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.html", "**/*.htm"), nil
}

func (c *Collection) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	c.mutex.Lock()
	defer func() {
		inputFile.Meta[c.groupsKey] = c.groups
		c.files = append(c.files, inputFile)
		c.mutex.Unlock()
	}()

	collection, ok := inputFile.Meta[c.collectionKey]
	if !ok {
		return nil
	}

	var collections []string
	switch t := collection.(type) {
	case string:
		collections = append(collections, t)
	case []string:
		collections = append(collections, t...)
	}

	for _, collection := range collections {
		files, _ := c.groups[collection]
		files = append(files, inputFile)
		c.groups[collection] = files
	}

	return nil
}

func (c *Collection) Finalize(context *goldsmith.Context) error {
	for _, files := range c.groups {
		fg := &fileSorter{files, c.comparer}
		sort.Sort(fg)
	}

	for _, file := range c.files {
		context.DispatchFile(file)
	}

	return nil
}

type fileSorter struct {
	files    []*goldsmith.File
	comparer Comparer
}

func (f fileSorter) Len() int {
	return len(f.files)
}

func (f fileSorter) Swap(i, j int) {
	f.files[i], f.files[j] = f.files[j], f.files[i]
}

func (f fileSorter) Less(i, j int) bool {
	if f.comparer == nil {
		return strings.Compare(f.files[i].Path(), f.files[j].Path()) < 0
	}

	return f.comparer(f.files[i], f.files[j])
}
