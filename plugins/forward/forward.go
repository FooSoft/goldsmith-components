package forward

import (
	"github.com/FooSoft/goldsmith"
)

type Forward struct {
	sourceMeta map[string]interface{}
	pathMap    map[string]string
	sourceKey  string
	targetKey  string
}

func New(sourceMeta map[string]interface{}) *Forward {
	return &Forward{
		sourceMeta: sourceMeta,
		pathMap:    make(map[string]string),
		sourceKey:  "PathOld",
		targetKey:  "PathNew",
	}
}

func (plugin *Forward) AddPathMapping(sourcePath, targetPath string) *Forward {
	plugin.pathMap[sourcePath] = targetPath
	return plugin
}

func (plugin *Forward) PathMap(pathMap map[string]string) *Forward {
	plugin.pathMap = pathMap
	return plugin
}

func (plugin *Forward) SourceKey(key string) *Forward {
	plugin.sourceKey = key
	return plugin
}

func (plugin *Forward) TargetKey(key string) *Forward {
	plugin.targetKey = key
	return plugin
}

func (*Forward) Name() string {
	return "forward"
}

func (plugin *Forward) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	for sourcePath, targetPath := range plugin.pathMap {
		sourceFile := context.CreateFileFromData(sourcePath, nil)
		for name, value := range plugin.sourceMeta {
			sourceFile.Meta[name] = value
		}

		sourceFile.Meta[plugin.targetKey] = targetPath
		context.DispatchFile(sourceFile)
	}

	return nil, nil
}
