// Package collection groups related pages into named collections. This can be
// useful for presenting blog posts on your front page, and displaying summary
// information about other types of content on your website. It can be used in
// conjunction with the "pager" plugin to create large collections which are
// split across several pages.
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
// The metadata associated with this key can be either a single string or an array of strings.
func (plugin *Collection) CollectionKey(collectionKey string) *Collection {
	plugin.collectionKey = collectionKey
	return plugin
}

// GroupsKey sets the metadata key used to store information about collection groups (default: "Groups").
// This information is stored as a mapping of group names to contained files.
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

func (*Collection) Initialize(context *goldsmith.Context) error {
	context.Filter(wildcard.New("**/*.html", "**/*.htm"))
	return nil
}

func (plugin *Collection) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	plugin.mutex.Lock()
	defer func() {
		inputFile.Meta[plugin.groupsKey] = plugin.groups
		plugin.files = append(plugin.files, inputFile)
		plugin.mutex.Unlock()
	}()

	collectionRaw, ok := inputFile.Meta[plugin.collectionKey]
	if !ok {
		return nil
	}

	var collectionNames []string
	switch t := collectionRaw.(type) {
	case string:
		collectionNames = append(collectionNames, t)
	case []string:
		collectionNames = append(collectionNames, t...)
	}

	for _, collectionName := range collectionNames {
		files, _ := plugin.groups[collectionName]
		files = append(files, inputFile)
		plugin.groups[collectionName] = files
	}

	return nil
}

func (plugin *Collection) Finalize(context *goldsmith.Context) error {
	for _, files := range plugin.groups {
		fg := &fileSorter{files, plugin.comparer}
		sort.Sort(fg)
	}

	for _, file := range plugin.files {
		context.DispatchFile(file)
	}

	return nil
}

type fileSorter struct {
	files    []*goldsmith.File
	comparer Comparer
}

func (fs fileSorter) Len() int {
	return len(fs.files)
}

func (fs fileSorter) Swap(i, j int) {
	fs.files[i], fs.files[j] = fs.files[j], fs.files[i]
}

func (fs fileSorter) Less(i, j int) bool {
	if fs.comparer == nil {
		return strings.Compare(fs.files[i].Path(), fs.files[j].Path()) < 0
	}

	return fs.comparer(fs.files[i], fs.files[j])
}
