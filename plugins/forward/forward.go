// Package forward allows to create simple redirections for pages that have moved.
package forward

import (
	"github.com/FooSoft/goldsmith"
)

// Forward plugin context.
type Forward struct {
	sourceMeta map[string]interface{}
	pathMap    map[string]string
	sourceKey  string
	targetKey  string
}

// New creates a new instance of the Forward plugin.
func New(sourceMeta map[string]interface{}) *Forward {
	return &Forward{
		sourceMeta: sourceMeta,
		pathMap:    make(map[string]string),
		sourceKey:  "PathOld",
		targetKey:  "PathNew",
	}
}

// AddPathMapping adds a single path mapping between an old path and a new path.
func (plugin *Forward) AddPathMapping(sourcePath, targetPath string) *Forward {
	plugin.pathMap[sourcePath] = targetPath
	return plugin
}

// PathMap sets multiple path mappings between old paths and new paths.
func (plugin *Forward) PathMap(pathMap map[string]string) *Forward {
	plugin.pathMap = pathMap
	return plugin
}

// SourceKey sets the metadata key used to access the old path (default: "PathOld").
func (plugin *Forward) SourceKey(key string) *Forward {
	plugin.sourceKey = key
	return plugin
}

// SourceKey sets the metadata key used to access the new path (default: "PathNew").
func (plugin *Forward) TargetKey(key string) *Forward {
	plugin.targetKey = key
	return plugin
}

func (*Forward) Name() string {
	return "forward"
}

func (plugin *Forward) Initialize(context *goldsmith.Context) error {
	for sourcePath, targetPath := range plugin.pathMap {
		sourceFile := context.CreateFileFromData(sourcePath, nil)
		for name, value := range plugin.sourceMeta {
			sourceFile.Meta[name] = value
		}

		sourceFile.Meta[plugin.targetKey] = targetPath
		context.DispatchFile(sourceFile)
	}

	return nil
}
