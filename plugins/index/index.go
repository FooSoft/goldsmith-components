// Package index creates pages for displaying directory listings.
package index

import (
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

// Index chainable plugin context.
type Index interface {
	goldsmith.Plugin
	goldsmith.Processor
	goldsmith.Finalizer

	// IndexFilename sets the name of the file to be created as the directory index (default: "index.html").
	IndexFilename(filename string) Index

	// FilesKey sets the metadata key used to access the files in the current directory (default: "Files").
	FilesKey(filename string) Index
}

// New creates a new instance of the Index plugin.
func New(meta map[string]interface{}) Index {
	return &index{
		indexName:   "index.html",
		indexMeta:   meta,
		filesKey:    "Files",
		dirsHandled: make(map[string]bool),
		dirLists:    make(map[string]*dirIndex),
	}
}

type dirIndex struct {
	entries   dirEntries
	indexFile *goldsmith.File
}

type dirEntry struct {
	Name  string
	Path  string
	IsDir bool
	File  *goldsmith.File
}

type index struct {
	indexName string
	filesKey  string
	indexMeta map[string]interface{}

	dirLists    map[string]*dirIndex
	dirsHandled map[string]bool

	mutex sync.Mutex
}

func (idx *index) IndexFilename(filename string) Index {
	idx.indexName = filename
	return idx
}

func (idx *index) FilesKey(key string) Index {
	idx.filesKey = key
	return idx
}

func (*index) Name() string {
	return "index"
}

func (idx *index) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	currentPath := inputFile.Path()
	currentIsDir := false

	for {
		if handled, _ := idx.dirsHandled[currentPath]; handled {
			break
		}

		idx.dirsHandled[currentPath] = true

		currentDir := path.Dir(currentPath)
		currentBase := path.Base(currentPath)

		list, ok := idx.dirLists[currentDir]
		if !ok {
			list = new(dirIndex)
			idx.dirLists[currentDir] = list
		}

		if !currentIsDir {
			if currentBase == idx.indexName {
				list.indexFile = inputFile
			} else {
				context.DispatchFile(inputFile)
			}
		}

		entry := dirEntry{Name: currentBase, Path: currentPath, IsDir: currentIsDir, File: inputFile}
		list.entries = append(list.entries, entry)

		if currentDir == "." {
			break
		}

		currentPath = currentDir
		currentIsDir = true
	}

	return nil
}

func (idx *index) Finalize(context *goldsmith.Context) error {
	for name, list := range idx.dirLists {
		sort.Sort(list.entries)

		indexFile := list.indexFile
		if indexFile == nil {
			indexFile = goldsmith.NewFileFromData(path.Join(name, idx.indexName), make([]byte, 0))
			for name, value := range idx.indexMeta {
				indexFile.SetValue(name, value)
			}
		}

		indexFile.SetValue(idx.filesKey, list.entries)
		context.DispatchFile(indexFile)
	}

	return nil
}

type dirEntries []dirEntry

func (d dirEntries) Len() int {
	return len(d)
}

func (d dirEntries) Less(i, j int) bool {
	d1, d2 := d[i], d[j]

	if d1.IsDir && !d2.IsDir {
		return true
	}
	if !d1.IsDir && d2.IsDir {
		return false
	}

	return strings.Compare(d1.Name, d2.Name) == -1
}

func (d dirEntries) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
