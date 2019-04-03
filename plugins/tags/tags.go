package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
)

type tagInfo struct {
	Files    filesByPath
	SafeName string
	RawName  string
	Path     string
}

type tagState struct {
	Index string
	Tags  []string
	Info  map[string]tagInfo
}

type Tags struct {
	tagsKey  string
	stateKey string

	baseDir   string
	indexName string
	indexMeta map[string]interface{}

	info  map[string]tagInfo
	files []*goldsmith.File
	mutex sync.Mutex
}

func New() *Tags {
	return &Tags{
		tagsKey:   "Tags",
		stateKey:  "TagState",
		baseDir:   "tags",
		indexName: "index.html",
		info:      make(map[string]tagInfo),
	}
}

func (plugin *Tags) TagsKey(key string) *Tags {
	plugin.tagsKey = key
	return plugin
}

func (plugin *Tags) StateKey(key string) *Tags {
	plugin.stateKey = key
	return plugin
}

func (plugin *Tags) IndexName(name string) *Tags {
	plugin.indexName = name
	return plugin
}

func (plugin *Tags) IndexMeta(meta map[string]interface{}) *Tags {
	plugin.indexMeta = meta
	return plugin
}

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
	tagState := &tagState{Info: plugin.info}

	plugin.mutex.Lock()
	defer func() {
		inputFile.Meta[plugin.stateKey] = tagState
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

		tagState.Tags = append(tagState.Tags, tagStr)

		info, ok := plugin.info[tagStr]
		info.Files = append(info.Files, inputFile)
		if !ok {
			info.SafeName = safeTag(tagStr)
			info.RawName = tagStr
			info.Path = plugin.tagPagePath(tagStr)
		}

		plugin.info[tagStr] = info
	}

	sort.Strings(tagState.Tags)
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

func (plugin *Tags) buildPages(context *goldsmith.Context, info map[string]tagInfo) (files []*goldsmith.File) {
	for tag := range info {
		tagFile := context.CreateFileFromData(plugin.tagPagePath(tag), nil)
		tagFile.Meta[plugin.tagsKey] = tagState{Index: tag, Info: plugin.info}
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
