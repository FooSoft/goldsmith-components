// Package index creates pages for displaying directory listings.
package index

import (
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type dirIndex struct {
	entries   dirEntriesByName
	indexFile *goldsmith.File
}

type dirEntry struct {
	Name  string
	Path  string
	IsDir bool
	File  *goldsmith.File
}

type Index struct {
	indexName string
	filesKey  string
	indexMeta map[string]interface{}

	dirLists    map[string]*dirIndex
	dirsHandled map[string]bool
	mutex       sync.Mutex
}

// New creates a new instance of the Index plugin.
func New(meta map[string]interface{}) *Index {
	return &Index{
		indexName:   "index.html",
		indexMeta:   meta,
		filesKey:    "Files",
		dirsHandled: make(map[string]bool),
		dirLists:    make(map[string]*dirIndex),
	}
}

// IndexFilename sets the name of the file to be created as the directory index (default: "index.html").
func (plugin *Index) IndexFilename(filename string) *Index {
	plugin.indexName = filename
	return plugin
}

// FilesKey sets the metadata key used to access the files in the current directory (default: "Files").
func (plugin *Index) FilesKey(key string) *Index {
	plugin.filesKey = key
	return plugin
}

func (*Index) Name() string {
	return "index"
}

func (plugin *Index) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	currentPath := inputFile.Path()
	currentIsDir := false

	for {
		if handled, _ := plugin.dirsHandled[currentPath]; handled {
			break
		}

		plugin.dirsHandled[currentPath] = true

		currentDir := path.Dir(currentPath)
		currentBase := path.Base(currentPath)

		list, ok := plugin.dirLists[currentDir]
		if !ok {
			list = new(dirIndex)
			plugin.dirLists[currentDir] = list
		}

		if !currentIsDir {
			if currentBase == plugin.indexName {
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

func (plugin *Index) Finalize(context *goldsmith.Context) error {
	for name, list := range plugin.dirLists {
		sort.Sort(list.entries)

		indexFile := list.indexFile
		if indexFile == nil {
			indexFile = context.CreateFileFromData(path.Join(name, plugin.indexName), make([]byte, 0))
			for name, value := range plugin.indexMeta {
				indexFile.Meta[name] = value
			}
		}

		indexFile.Meta[plugin.filesKey] = list.entries
		context.DispatchFile(indexFile)
	}

	return nil
}

type dirEntriesByName []dirEntry

func (d dirEntriesByName) Len() int {
	return len(d)
}

func (d dirEntriesByName) Less(i, j int) bool {
	d1, d2 := d[i], d[j]

	if d1.IsDir && !d2.IsDir {
		return true
	}
	if !d1.IsDir && d2.IsDir {
		return false
	}

	return strings.Compare(d1.Name, d2.Name) == -1
}

func (d dirEntriesByName) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
