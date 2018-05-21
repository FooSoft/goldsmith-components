// Copyright (c) 2016-2018 Alex Yatskov <alex@foosoft.net>
//
// BreadCrumbs generates metadata required to build navigation breadcrumbs.
package breadcrumbs

import (
	"fmt"
	"sync"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/filters/extension"
)

type Breadcrumbs interface {
	// NameKey sets the metadata key used to access the crumb name.
	NameKey(key string) Breadcrumbs

	// ParentKey sets the metadata key used to access the parent name.
	ParentKey(key string) Breadcrumbs

	// CrumbsKey sets the metadata key used to access information about crumbs.
	CrumbsKey(key string) Breadcrumbs

	Name() string
	Initialize(ctx goldsmith.Context) ([]goldsmith.Filter, error)
	Process(ctx goldsmith.Context, f goldsmith.File) error
	Finalize(ctx goldsmith.Context) error
}

type crumbs struct {
	Ancestors []*node
	Node      *node
}

type node struct {
	File     goldsmith.File
	Parent   *node
	Children []*node

	parentName string
}

type breadcrumbs struct {
	nameKey   string
	parentKey string
	crumbsKey string

	allNodes   []*node
	namedNodes map[string]*node

	mtx sync.Mutex
}

// Creates a new instance of the BreadCrumbs plugin.
func New() Breadcrumbs {
	return &breadcrumbs{
		nameKey:    "CrumbName",
		parentKey:  "CrumbParent",
		crumbsKey:  "Crumbs",
		namedNodes: make(map[string]*node),
	}
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

	node := &node{File: f, parentName: parentNameStr}
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
		var ancestors []*node
		for c := n.Parent; c != nil; c = c.Parent {
			ancestors = append([]*node{c}, ancestors...)
		}

		n.File.SetValue(b.crumbsKey, crumbs{ancestors, n})
		ctx.DispatchFile(n.File)
	}

	return nil
}
