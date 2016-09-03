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

package tree

import (
	"fmt"
	"sync"

	"github.com/FooSoft/goldsmith"
)

type treeInfo struct {
	Roots []*treeNode
	Node  *treeNode
}

type treeNode struct {
	File goldsmith.File

	Parent   *treeNode
	Children []*treeNode

	nodeName   string
	parentName string
}

type tree struct {
	nodeKey, parentKey, treeKey string

	allNodes   []*treeNode
	rootNodes  []*treeNode
	namedNodes map[string]*treeNode

	mtx sync.Mutex
}

func New() *tree {
	return &tree{
		nodeKey:    "node",
		parentKey:  "parent",
		treeKey:    "tree",
		namedNodes: make(map[string]*treeNode),
	}
}

func (t *tree) NodeKey(nodeKey string) *tree {
	t.nodeKey = nodeKey
	return t
}

func (t *tree) ParentKey(parentKey string) *tree {
	t.parentKey = parentKey
	return t
}

func (t *tree) TreeKey(treeKey string) *tree {
	t.treeKey = treeKey
	return t
}

func (*tree) Name() string {
	return "tree"
}

func (*tree) Initialize(ctx goldsmith.Context) ([]string, error) {
	return []string{"**/*.html", "**/*.htm"}, nil
}

func (t *tree) Process(ctx goldsmith.Context, f goldsmith.File) error {
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

	node := &treeNode{File: f, nodeName: nodeNameStr, parentName: parentNameStr}
	t.allNodes = append(t.allNodes, node)

	if len(nodeNameStr) > 0 {
		if _, ok := t.namedNodes[nodeNameStr]; ok {
			return fmt.Errorf("duplicate node name: %s", nodeNameStr)
		}

		t.namedNodes[nodeNameStr] = node
	}

	if len(parentNameStr) == 0 {
		t.rootNodes = append(t.rootNodes, node)
	}

	return nil
}

func (t *tree) Finalize(ctx goldsmith.Context) error {
	for _, n := range t.allNodes {
		if parent, ok := t.namedNodes[n.parentName]; ok {
			parent.Children = append(parent.Children, n)
			n.Parent = parent
		} else {
			return fmt.Errorf("undefined parent: %s", n.parentName)
		}
	}

	for _, n := range t.allNodes {
		n.File.SetValue(t.treeKey, treeInfo{Roots: t.rootNodes, Node: n})
		ctx.DispatchFile(n.File)
	}

	return nil
}
