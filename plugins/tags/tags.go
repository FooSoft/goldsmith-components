package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

type Tags interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
	goldsmith.Finalizer

	TagsKey(key string) Tags
	StateKey(key string) Tags
	IndexName(name string) Tags
	IndexMeta(meta map[string]interface{}) Tags
	BaseDir(dir string) Tags
}

func New() Tags {
	return &tags{
		tagsKey:   "Tags",
		stateKey:  "TagState",
		baseDir:   "tags",
		indexName: "index.html",
		info:      make(map[string]tagInfo),
	}
}

type tags struct {
	tagsKey  string
	stateKey string

	baseDir   string
	indexName string
	indexMeta map[string]interface{}

	info  map[string]tagInfo
	files []*goldsmith.File
	mutex sync.Mutex
}

type tagInfo struct {
	Files    files
	SafeName string
	RawName  string
	Path     string
}

type tagState struct {
	Index string
	Tags  []string
	Info  map[string]tagInfo
}

func (t *tags) TagsKey(key string) Tags {
	t.tagsKey = key
	return t
}

func (t *tags) StateKey(key string) Tags {
	t.stateKey = key
	return t
}

func (t *tags) IndexName(name string) Tags {
	t.indexName = name
	return t
}

func (t *tags) IndexMeta(meta map[string]interface{}) Tags {
	t.indexMeta = meta
	return t
}

func (t *tags) BaseDir(dir string) Tags {
	t.baseDir = dir
	return t
}

func (*tags) Name() string {
	return "tags"
}

func (*tags) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (t *tags) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	tagState := &tagState{Info: t.info}

	t.mutex.Lock()
	defer func() {
		inputFile.Meta[t.stateKey] = tagState
		t.files = append(t.files, inputFile)
		t.mutex.Unlock()
	}()

	tags, ok := inputFile.Meta[t.tagsKey]
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

		info, ok := t.info[tagStr]
		info.Files = append(info.Files, inputFile)
		if !ok {
			info.SafeName = safeTag(tagStr)
			info.RawName = tagStr
			info.Path = t.tagPagePath(tagStr)
		}

		t.info[tagStr] = info
	}

	sort.Strings(tagState.Tags)
	return nil
}

func (t *tags) Finalize(context *goldsmith.Context) error {
	for _, meta := range t.info {
		sort.Sort(meta.Files)
	}

	if t.indexMeta != nil {
		for _, file := range t.buildPages(context, t.info) {
			context.DispatchFile(file)
		}
	}

	for _, file := range t.files {
		context.DispatchFile(file)
	}

	return nil
}

func (t *tags) buildPages(context *goldsmith.Context, info map[string]tagInfo) (files []*goldsmith.File) {
	for tag := range info {
		tagFile := context.CreateFileFromData(t.tagPagePath(tag), nil)
		tagFile.Meta[t.tagsKey] = tagState{Index: tag, Info: t.info}
		for name, value := range t.indexMeta {
			tagFile.Meta[name] = value
		}

		files = append(files, tagFile)
	}

	return
}

func (t *tags) tagPagePath(tag string) string {
	return filepath.Join(t.baseDir, safeTag(tag), t.indexName)
}

func safeTag(tag string) string {
	return strings.ToLower(strings.Replace(tag, " ", "-", -1))
}

type files []*goldsmith.File

func (f files) Len() int {
	return len(f)
}

func (f files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f files) Less(i, j int) bool {
	return strings.Compare(f[i].Path(), f[j].Path()) < 0
}
