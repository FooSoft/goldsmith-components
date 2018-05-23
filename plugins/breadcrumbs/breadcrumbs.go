// Package breadcrumbs generates metadata required to build navigation breadcrumbs.
package breadcrumbs

import (
	"fmt"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

// Breadcrumbs chainable plugin context.
type Breadcrumbs interface {
	goldsmith.Plugin
	goldsmith.Initializer
	goldsmith.Processor
	goldsmith.Finalizer

	// NameKey sets the metadata key used to access the crumb name (default: "CrumbName").
	NameKey(key string) Breadcrumbs

	// ParentKey sets the metadata key used to access the parent name (default: "CrumbParent").
	ParentKey(key string) Breadcrumbs

	// CrumbsKey sets the metadata key used to store information about crumbs (default: "Crumbs").
	CrumbsKey(key string) Breadcrumbs
}

// A Crumb provides organizational information about this node and ones before it.
type Crumb struct {
	Ancestors []*Node
	Node      *Node
}

// A Node represents information about a specific file in the site's structure.
type Node struct {
	File     goldsmith.File
	Parent   *Node
	Children []*Node

	parentName string
}

// New creates a new instance of the breadcrumbs plugin.
func New() Breadcrumbs {
	return &breadcrumbs{
		nameKey:    "CrumbName",
		parentKey:  "CrumbParent",
		crumbsKey:  "Crumbs",
		namedNodes: make(map[string]*Node),
	}
}

type breadcrumbs struct {
	nameKey   string
	parentKey string
	crumbsKey string

	allNodes   []*Node
	namedNodes map[string]*Node

	mtx sync.Mutex
}

func (b *breadcrumbs) NameKey(key string) Breadcrumbs {
	b.nameKey = key
	return b
}

func (b *breadcrumbs) ParentKey(key string) Breadcrumbs {
	b.parentKey = key
	return b
}

func (b *breadcrumbs) CrumbsKey(key string) Breadcrumbs {
	b.crumbsKey = key
	return b
}

func (*breadcrumbs) Name() string {
	return "breadcrumbs"
}

func (*breadcrumbs) Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error) {
	return []goldsmith.Filter{extension.New(".html", ".htm")}, nil
}

func (b *breadcrumbs) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var parentNameStr string
	if parentName, ok := f.Value(b.parentKey); ok {
		parentNameStr, _ = parentName.(string)
	}

	var nodeNameStr string
	if nodeName, ok := f.Value(b.nameKey); ok {
		nodeNameStr, _ = nodeName.(string)
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()

	node := &Node{File: f, parentName: parentNameStr}
	b.allNodes = append(b.allNodes, node)

	if len(nodeNameStr) > 0 {
		if _, ok := b.namedNodes[nodeNameStr]; ok {
			return fmt.Errorf("duplicate node: %s", nodeNameStr)
		}

		b.namedNodes[nodeNameStr] = node
	}

	return nil
}

func (b *breadcrumbs) Finalize(ctx goldsmith.Context) error {
	for _, n := range b.allNodes {
		if len(n.parentName) == 0 {
			continue
		}

		if parent, ok := b.namedNodes[n.parentName]; ok {
			parent.Children = append(parent.Children, n)
			n.Parent = parent
		} else {
			return fmt.Errorf("undefined parent: %s", n.parentName)
		}
	}

	for _, n := range b.allNodes {
		var ancestors []*Node
		for c := n.Parent; c != nil; c = c.Parent {
			ancestors = append([]*Node{c}, ancestors...)
		}

		n.File.SetValue(b.crumbsKey, Crumb{ancestors, n})
		ctx.DispatchFile(n.File)
	}

	return nil
}
