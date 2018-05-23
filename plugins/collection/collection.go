// Package collection groups related pages into named collections.
package collection

import (
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// Collection chainable plugin context.
type Collection interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
	goldsmith.Finalizer

	// CollectionKey sets the metadata key used to access the collection name (default: "Collection").
	CollectionKey(collKey string) Collection

	// GroupsKey sets the metadata key used to store information about collection groups (default: "Groups").
	// This information is stored as a mapping of group names to contained files (map[string][]goldsmith.File).
	GroupsKey(groupsKey string) Collection

	// Comparer sets the function used to sort files in collection groups (default: sort by filenames).
	Comparer(comp Comparer) Collection
}

// A Comparer callback function is used to sort files within a collection group.
type Comparer func(i, j goldsmith.File) (less bool)

// New creates a new instance of the Collection plugin.
func New() Collection {
	return &collection{
		collKey:   "Collection",
		groupsKey: "Groups",
		comp:      nil,
		groups:    make(map[string][]goldsmith.File),
	}
}

type collection struct {
	collKey   string
	groupsKey string

	comp   Comparer
	groups map[string][]goldsmith.File
	files  []goldsmith.File

	mtx sync.Mutex
}

func (c *collection) CollectionKey(collKey string) Collection {
	c.collKey = collKey
	return c
}

func (c *collection) GroupsKey(groupsKey string) Collection {
	c.groupsKey = groupsKey
	return c
}

func (c *collection) Comparer(comp Comparer) Collection {
	c.comp = comp
	return c
}

func (*collection) Name() string {
	return "collection"
}

func (*collection) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (c *collection) Process(ctx goldsmith.Context, f goldsmith.File) error {
	c.mtx.Lock()
	defer func() {
		f.SetValue(c.groupsKey, c.groups)
		c.files = append(c.files, f)
		c.mtx.Unlock()
	}()

	coll, ok := f.Value(c.collKey)
	if !ok {
		return nil
	}

	var collStrs []string
	switch t := coll.(type) {
	case string:
		collStrs = append(collStrs, t)
	case []string:
		collStrs = append(collStrs, t...)
	}

	for _, collStr := range collStrs {
		files, _ := c.groups[collStr]
		files = append(files, f)
		c.groups[collStr] = files
	}

	return nil
}

func (c *collection) Finalize(ctx goldsmith.Context) error {
	for _, files := range c.groups {
		fg := &fileGroup{files, c.comp}
		sort.Sort(fg)
	}

	for _, f := range c.files {
		ctx.DispatchFile(f)
	}

	return nil
}

type fileGroup struct {
	Files []goldsmith.File
	comp  Comparer
}

func (f fileGroup) Len() int {
	return len(f.Files)
}

func (f fileGroup) Swap(i, j int) {
	f.Files[i], f.Files[j] = f.Files[j], f.Files[i]
}

func (f fileGroup) Less(i, j int) bool {
	if f.comp == nil {
		return strings.Compare(f.Files[i].Path(), f.Files[j].Path()) < 0
	}

	return f.comp(f.Files[i], f.Files[j])
}
