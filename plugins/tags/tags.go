// Package tags builds tag clouds from file metadata. This makes it easy to
// create lists of all files tagged with a specific tag, as well as to see all
// tags globally used on a site.
package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
)

// TagInfo contains site-wide information about a particular tag.
type TagInfo struct {
	Files    filesByPath
	SafeName string
	RawName  string
	Path     string
}

// TagState contains site-wide information about tags used on a site.
type TagState struct {
	Index string
	Tags  []string
	Info  map[string]TagInfo
}

// Tags chainable context.
type Tags struct {
	tagsKey  string
	stateKey string

	baseDir   string
	indexName string
	indexMeta map[string]interface{}

	info  map[string]TagInfo
	files []*goldsmith.File
	mutex sync.Mutex
}

// New creates a new instance of the Tags plugin.
func New() *Tags {
	return &Tags{
		tagsKey:   "Tags",
		stateKey:  "TagState",
		baseDir:   "tags",
		indexName: "index.html",
		info:      make(map[string]TagInfo),
	}
}

// TagsKey sets the metadata key used to get the tags for this file, stored as a slice of strings (default: "Tags").
func (plugin *Tags) TagsKey(key string) *Tags {
	plugin.tagsKey = key
	return plugin
}

// StateKey sets the meatadata key used to store site-wide tag information (default: "TagState").
func (plugin *Tags) StateKey(key string) *Tags {
	plugin.stateKey = key
	return plugin
}

// IndexName sets the filename which will be used to create tag list files (default: "index.html").
func (plugin *Tags) IndexName(name string) *Tags {
	plugin.indexName = name
	return plugin
}

// IndexMeta sets the metadata which will be assigned to generated tag list files (default: {}).
func (plugin *Tags) IndexMeta(meta map[string]interface{}) *Tags {
	plugin.indexMeta = meta
	return plugin
}

// BaseDir sets the base directory used to generate tag list files (default: "tags").
func (plugin *Tags) BaseDir(dir string) *Tags {
	plugin.baseDir = dir
	return plugin
}

func (*Tags) Name() string {
	return "tags"
}

func (*Tags) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return wildcard.New("**/*.html", "**/*.htm"), nil
}

func (plugin *Tags) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	TagState := &TagState{Info: plugin.info}

	plugin.mutex.Lock()
	defer func() {
		inputFile.Meta[plugin.stateKey] = TagState
		plugin.files = append(plugin.files, inputFile)
		plugin.mutex.Unlock()
	}()

	tags, ok := inputFile.Meta[plugin.tagsKey]
	if !ok {
		return nil
	}

	tagsArr, ok := tags.([]interface{})
	if !ok {
		return nil
	}

	for _, tag := range tagsArr {
		tagStr, ok := tag.(string)
		if !ok {
			continue
		}

		TagState.Tags = append(TagState.Tags, tagStr)

		info, ok := plugin.info[tagStr]
		info.Files = append(info.Files, inputFile)
		if !ok {
			info.SafeName = safeTag(tagStr)
			info.RawName = tagStr
			info.Path = plugin.tagPagePath(tagStr)
		}

		plugin.info[tagStr] = info
	}

	sort.Strings(TagState.Tags)
	return nil
}

func (plugin *Tags) Finalize(context *goldsmith.Context) error {
	for _, meta := range plugin.info {
		sort.Sort(meta.Files)
	}

	if plugin.indexMeta != nil {
		for _, file := range plugin.buildPages(context, plugin.info) {
			context.DispatchFile(file)
		}
	}

	for _, file := range plugin.files {
		context.DispatchFile(file)
	}

	return nil
}

func (plugin *Tags) buildPages(context *goldsmith.Context, info map[string]TagInfo) (files []*goldsmith.File) {
	for tag := range info {
		tagFile := context.CreateFileFromData(plugin.tagPagePath(tag), nil)
		tagFile.Meta[plugin.tagsKey] = TagState{Index: tag, Info: plugin.info}
		for name, value := range plugin.indexMeta {
			tagFile.Meta[name] = value
		}

		files = append(files, tagFile)
	}

	return
}

func (plugin *Tags) tagPagePath(tag string) string {
	return filepath.Join(plugin.baseDir, safeTag(tag), plugin.indexName)
}

func safeTag(tag string) string {
	return strings.ToLower(strings.Replace(tag, " ", "-", -1))
}

type filesByPath []*goldsmith.File

func (f filesByPath) Len() int {
	return len(f)
}

func (f filesByPath) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f filesByPath) Less(i, j int) bool {
	return strings.Compare(f[i].Path(), f[j].Path()) < 0
}
