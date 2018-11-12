// Package index creates pages for displaying directory listings.
package index

import (
	"path"
	"sort"
	"strings"
	"sync"
	"time"

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
		filename: "index.html",
		filesKey: "Files",
		meta:     meta,
		handled:  make(map[string]bool),
		dirs:     make(map[string]*dirSummary),
	}
}

type index struct {
	filename string
	filesKey string
	meta     map[string]interface{}

	dirs    map[string]*dirSummary
	handled map[string]bool
	dirsMtx sync.Mutex
}

func (idx *index) IndexFilename(filename string) Index {
	idx.filename = filename
	return idx
}

func (idx *index) FilesKey(key string) Index {
	idx.filesKey = key
	return idx
}

func (*index) Name() string {
	return "index"
}

func (idx *index) Process(ctx *goldsmith.Context, f *goldsmith.File) error {
	idx.dirsMtx.Lock()
	defer idx.dirsMtx.Unlock()

	curr := f.Path()
	leaf := true

	for {
		if handled, _ := idx.handled[curr]; handled {
			break
		}

		idx.handled[curr] = true

		dir := path.Dir(curr)
		base := path.Base(curr)

		summary, ok := idx.dirs[dir]
		if !ok {
			summary = new(dirSummary)
			idx.dirs[dir] = summary
		}

		if leaf {
			if base == idx.filename {
				summary.index = f
			} else {
				ctx.DispatchFile(f)
			}
		}

		entry := dirEntry{Name: base, Path: curr, IsDir: !leaf, File: f}
		summary.entries = append(summary.entries, entry)

		if dir == "." {
			break
		}

		curr = dir
		leaf = false
	}

	return nil
}

func (idx *index) Finalize(ctx *goldsmith.Context) error {
	for name, summary := range idx.dirs {
		sort.Sort(summary.entries)

		f := summary.index
		if f == nil {
			f = goldsmith.NewFileFromData(path.Join(name, idx.filename), make([]byte, 0), time.Now())
			for name, value := range idx.meta {
				f.SetValue(name, value)
			}
		}

		f.SetValue(idx.filesKey, summary.entries)
		ctx.DispatchFile(f)
	}

	return nil
}

type dirSummary struct {
	entries dirEntries
	index   *goldsmith.File
}

type dirEntry struct {
	Name  string
	Path  string
	IsDir bool
	File  *goldsmith.File
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
