// Package tags builds tag clouds from file metadata. This makes it easy to
// create lists of all files tagged with a specific tag, as well as to see all
// tags globally used on a site.
package tags

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/wildcard"
)

// TagInfo contains site-wide information about a particular tag.
type TagInfo struct {
	TaggedFiles filesByPath
	IndexFile   *goldsmith.File
	SafeName    string
	RawName     string
}

// TagState contains site-wide information about tags used on a site.
type TagState struct {
	CurrentTag  *TagInfo
	CurrentTags tagInfoByName
	TagsByName  *tagInfoByName
	TagsByCount *tagInfoByCount
}

// Tags chainable context.
type Tags struct {
	tagsKey  string
	stateKey string

	baseDir   string
	indexName string
	indexMeta map[string]interface{}

	info        map[string]*TagInfo
	infoByName  tagInfoByName
	infoByCount tagInfoByCount

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
		info:      make(map[string]*TagInfo),
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

func (*Tags) Initialize(context *goldsmith.Context) error {
	context.Filter(wildcard.New("**/*.html", "**/*.htm"))
	return nil
}

func (plugin *Tags) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	tagState := &TagState{
		TagsByName:  &plugin.infoByName,
		TagsByCount: &plugin.infoByCount,
	}

	plugin.mutex.Lock()
	defer func() {
		inputFile.Meta[plugin.stateKey] = tagState
		plugin.files = append(plugin.files, inputFile)
		plugin.mutex.Unlock()
	}()

	tagsArr, ok := inputFile.Meta[plugin.tagsKey].([]interface{})
	if !ok {
		return nil
	}

	for _, tag := range tagsArr {
		tagRaw, ok := tag.(string)
		if !ok {
			continue
		}

		tagSafe := safeTag(tagRaw)
		if len(tagSafe) == 0 {
			continue
		}

		var duplicate bool
		for _, tagState := range tagState.CurrentTags {
			if tagState.RawName == tagRaw {
				duplicate = true
				break
			}
		}

		if duplicate {
			continue
		}

		info, ok := plugin.info[tagRaw]
		if !ok {
			info = &TagInfo{
				SafeName: tagSafe,
				RawName:  tagRaw,
			}

			plugin.info[tagRaw] = info
		}
		info.TaggedFiles = append(info.TaggedFiles, inputFile)

		tagState.CurrentTags = append(tagState.CurrentTags, info)
	}

	sort.Sort(tagState.CurrentTags)

	return nil
}

func (plugin *Tags) Finalize(context *goldsmith.Context) error {
	for _, info := range plugin.info {
		sort.Sort(info.TaggedFiles)

		plugin.infoByName = append(plugin.infoByName, info)
		plugin.infoByCount = append(plugin.infoByCount, info)
	}

	sort.Sort(plugin.infoByName)
	sort.Sort(plugin.infoByCount)

	if plugin.indexMeta != nil {
		plugin.files = append(plugin.files, plugin.buildPages(context)...)
	}

	for _, file := range plugin.files {
		context.DispatchFile(file)
	}

	return nil
}

func (plugin *Tags) buildPages(context *goldsmith.Context) []*goldsmith.File {
	var files []*goldsmith.File
	for tag, info := range plugin.info {
		info.IndexFile = context.CreateFileFromData(plugin.tagPagePath(tag), nil)
		info.IndexFile.Meta[plugin.stateKey] = &TagState{
			CurrentTag:  info,
			TagsByName:  &plugin.infoByName,
			TagsByCount: &plugin.infoByCount,
		}
		for name, value := range plugin.indexMeta {
			info.IndexFile.Meta[name] = value
		}

		files = append(files, info.IndexFile)
	}

	return files
}

func (plugin *Tags) tagPagePath(tag string) string {
	return filepath.Join(plugin.baseDir, safeTag(tag), plugin.indexName)
}

func safeTag(tagRaw string) string {
	tagRaw = strings.TrimSpace(tagRaw)
	tagRaw = strings.ToLower(tagRaw)

	var valid bool
	var tagSafe string
	for _, c := range tagRaw {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			tagSafe += string(c)
			valid = true
		} else if valid {
			tagSafe += "-"
			valid = false
		}
	}

	return tagSafe
}

type tagInfoByCount []*TagInfo

func (info tagInfoByCount) Len() int {
	return len(info)
}

func (info tagInfoByCount) Swap(i, j int) {
	info[i], info[j] = info[j], info[i]
}

func (info tagInfoByCount) Less(i, j int) bool {
	if len(info[i].TaggedFiles) > len(info[j].TaggedFiles) {
		return true
	} else if len(info[i].TaggedFiles) == len(info[j].TaggedFiles) && strings.Compare(info[i].RawName, info[j].RawName) < 0 {
		return true
	}

	return false
}

type tagInfoByName []*TagInfo

func (info tagInfoByName) Len() int {
	return len(info)
}

func (info tagInfoByName) Swap(i, j int) {
	info[i], info[j] = info[j], info[i]
}

func (info tagInfoByName) Less(i, j int) bool {
	if strings.Compare(info[i].RawName, info[j].RawName) < 0 {
		return true
	} else if info[i].RawName == info[j].RawName && len(info[i].TaggedFiles) > len(info[j].TaggedFiles) {
		return true
	}

	return false
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
