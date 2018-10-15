package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

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
	tagsKey, stateKey string

	baseDir   string
	indexName string
	indexMeta map[string]interface{}

	info  map[string]tagInfo
	files []goldsmith.File
	mtx   sync.Mutex
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

func (*tags) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (t *tags) Process(ctx goldsmith.Context, f goldsmith.File) error {
	tagState := &tagState{Info: t.info}

	t.mtx.Lock()
	defer func() {
		f.SetValue(t.stateKey, tagState)
		t.files = append(t.files, f)
		t.mtx.Unlock()
	}()

	tags, ok := f.Value(t.tagsKey)
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

		item, ok := t.info[tagStr]
		item.Files = append(item.Files, f)
		if !ok {
			item.SafeName = safeTag(tagStr)
			item.RawName = tagStr
			item.Path = t.tagPagePath(tagStr)
		}

		t.info[tagStr] = item
	}

	sort.Strings(tagState.Tags)
	return nil
}

func (t *tags) Finalize(ctx goldsmith.Context) error {
	for _, meta := range t.info {
		sort.Sort(meta.Files)
	}

	if t.indexMeta != nil {
		for _, f := range t.buildPages(ctx, t.info) {
			ctx.DispatchFile(f)
		}
	}

	for _, f := range t.files {
		ctx.DispatchFile(f)
	}

	return nil
}

func (t *tags) buildPages(ctx goldsmith.Context, info map[string]tagInfo) (files []goldsmith.File) {
	for tag := range info {
		f := goldsmith.NewFileFromData(t.tagPagePath(tag), nil, time.Now())
		f.SetValue(t.tagsKey, tagState{Index: tag, Info: t.info})
		for name, value := range t.indexMeta {
			f.SetValue(name, value)
		}

		files = append(files, f)
	}

	return
}

func (t *tags) tagPagePath(tag string) string {
	return filepath.Join(t.baseDir, safeTag(tag), t.indexName)
}

func safeTag(tag string) string {
	return strings.ToLower(strings.Replace(tag, " ", "-", -1))
}

type files []goldsmith.File

func (f files) Len() int {
	return len(f)
}

func (f files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f files) Less(i, j int) bool {
	return strings.Compare(f[i].Path(), f[j].Path()) < 0
}
