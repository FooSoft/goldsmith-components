// Package breadcrumbs generates metadata required to build navigation breadcrumbs.
package breadcrumbs

import (
	"fmt"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// A Crumb provides organizational information about this node and ones before it.
type Crumb struct {
	Ancestors []*Node
	Node      *Node
}

// A Node represents information about a specific file in the site's structure.
type Node struct {
	File     *goldsmith.File
	Parent   *Node
	Children []*Node

	parentName string
}

// Breadcrumbs chainable plugin context.
type Breadcrumbs struct {
	nameKey   string
	parentKey string
	crumbsKey string

	allNodes   []*Node
	namedNodes map[string]*Node
	mutex      sync.Mutex
}

// New creates a new instance of the Breadcrumbs plugin.
func New() *Breadcrumbs {
	return &Breadcrumbs{
		nameKey:    "CrumbName",
		parentKey:  "CrumbParent",
		crumbsKey:  "Crumbs",
		namedNodes: make(map[string]*Node),
	}
}

// NameKey sets the metadata key used to access the crumb name (default: "CrumbName").
func (plugin *Breadcrumbs) NameKey(key string) *Breadcrumbs {
	plugin.nameKey = key
	return plugin
}

// ParentKey sets the metadata key used to access the parent name (default: "CrumbParent").
func (plugin *Breadcrumbs) ParentKey(key string) *Breadcrumbs {
	plugin.parentKey = key
	return plugin
}

// CrumbsKey sets the metadata key used to store information about crumbs (default: "Crumbs").
func (plugin *Breadcrumbs) CrumbsKey(key string) *Breadcrumbs {
	plugin.crumbsKey = key
	return plugin
}

func (*Breadcrumbs) Name() string {
	return "breadcrumbs"
}

func (*Breadcrumbs) Initialize(context *goldsmith.Context) (goldsmith.Filter, error) {
	return extension.New(".html", ".htm"), nil
}

func (plugin *Breadcrumbs) Process(context *goldsmith.Context, inputFile *goldsmith.File) error {
	var parentNameStr string
	if parentName, ok := inputFile.Meta[plugin.parentKey]; ok {
		parentNameStr, _ = parentName.(string)
	}

	var nodeNameStr string
	if nodeName, ok := inputFile.Meta[plugin.nameKey]; ok {
		nodeNameStr, _ = nodeName.(string)
	}

	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	node := &Node{File: inputFile, parentName: parentNameStr}
	plugin.allNodes = append(plugin.allNodes, node)

	if len(nodeNameStr) > 0 {
		if _, ok := plugin.namedNodes[nodeNameStr]; ok {
			return fmt.Errorf("duplicate node: %s", nodeNameStr)
		}

		plugin.namedNodes[nodeNameStr] = node
	}

	return nil
}

func (plugin *Breadcrumbs) Finalize(context *goldsmith.Context) error {
	for _, node := range plugin.allNodes {
		if len(node.parentName) == 0 {
			continue
		}

		if parentNode, ok := plugin.namedNodes[node.parentName]; ok {
			parentNode.Children = append(parentNode.Children, node)
			node.Parent = parentNode
		} else {
			return fmt.Errorf("undefined parent: %s", node.parentName)
		}
	}

	for _, node := range plugin.allNodes {
		var ancestors []*Node
		for currentNode := node.Parent; currentNode != nil; currentNode = currentNode.Parent {
			ancestors = append([]*Node{currentNode}, ancestors...)
		}

		node.File.Meta[plugin.crumbsKey] = Crumb{ancestors, node}
		context.DispatchFile(node.File)
	}

	return nil
}
