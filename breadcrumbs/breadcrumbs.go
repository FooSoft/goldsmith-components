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
	nameKey, parentKey, crumbsKey string

	allNodes   []*node
	namedNodes map[string]*node

	mtx sync.Mutex
}

func New() *breadcrumbs {
	return &breadcrumbs{
		nameKey:    "CrumbName",
		parentKey:  "CrumbParent",
		crumbsKey:  "Crumbs",
		namedNodes: make(map[string]*node),
	}
}

func (b *breadcrumbs) NameKey(key string) *breadcrumbs {
	b.nameKey = key
	return b
}

func (b *breadcrumbs) ParentKey(key string) *breadcrumbs {
	b.parentKey = key
	return b
}

func (b *breadcrumbs) CrumbsKey(key string) *breadcrumbs {
	b.crumbsKey = key
	return b
}

func (*breadcrumbs) Name() string {
	return "breadcrumbs"
}

func (*breadcrumbs) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
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
		if nodeDup, ok := b.namedNodes[nodeNameStr]; ok && nodeDup.File.ModTime().Unix() < node.File.ModTime().Unix() {
			node = nodeDup
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
