/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package breadcrumbs

import (
	"fmt"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type bcInfo struct {
	Ancestors []*bcNode
	Node      *bcNode
}

type bcNode struct {
	File     goldsmith.File
	Parent   *bcNode
	Children []*bcNode

	parentName string
}

type breadcrumbs struct {
	nodeKey, parentKey, breadcrumbsKey string

	allNodes   []*bcNode
	rootNodes  []*bcNode
	namedNodes map[string]*bcNode

	mtx sync.Mutex
}

func New() *breadcrumbs {
	return &breadcrumbs{
		nodeKey:        "Node",
		parentKey:      "Parent",
		breadcrumbsKey: "Breadcrumbs",
		namedNodes:     make(map[string]*bcNode),
	}
}

func (t *breadcrumbs) NodeKey(key string) *breadcrumbs {
	t.nodeKey = key
	return t
}

func (t *breadcrumbs) ParentKey(key string) *breadcrumbs {
	t.parentKey = key
	return t
}

func (t *breadcrumbs) BreadcrumbsKey(breadcrumbsKey string) *breadcrumbs {
	t.breadcrumbsKey = breadcrumbsKey
	return t
}

func (*breadcrumbs) Name() string {
	return "breadcrumbs"
}

func (*breadcrumbs) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (t *breadcrumbs) Process(ctx goldsmith.Context, f goldsmith.File) error {
	var parentNameStr string
	if parentName, ok := f.Value(t.parentKey); ok {
		parentNameStr, _ = parentName.(string)
	}

	var nodeNameStr string
	if nodeName, ok := f.Value(t.nodeKey); ok {
		nodeNameStr, _ = nodeName.(string)
	}

	t.mtx.Lock()
	defer t.mtx.Unlock()

	node := &bcNode{File: f, parentName: parentNameStr}
	t.allNodes = append(t.allNodes, node)

	if len(nodeNameStr) > 0 {
		if _, ok := t.namedNodes[nodeNameStr]; ok {
			return fmt.Errorf("duplicate node: %s", nodeNameStr)
		}

		t.namedNodes[nodeNameStr] = node
	}

	return nil
}

func (t *breadcrumbs) Finalize(ctx goldsmith.Context) error {
	for _, n := range t.allNodes {
		if len(n.parentName) == 0 {
			continue
		}

		if parent, ok := t.namedNodes[n.parentName]; ok {
			parent.Children = append(parent.Children, n)
			n.Parent = parent
		} else {
			return fmt.Errorf("undefined parent: %s", n.parentName)
		}
	}

	for _, n := range t.allNodes {
		var ancestors []*bcNode
		for c := n.Parent; c != nil; c = c.Parent {
			ancestors = append([]*bcNode{c}, ancestors...)
		}

		n.File.SetValue(t.breadcrumbsKey, bcInfo{ancestors, n})
		ctx.DispatchFile(n.File)
	}

	return nil
}
