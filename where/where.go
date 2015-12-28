/*
 * Copyright (c) 2015 Alex Yatskov <alex@foosoft.net>
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

package where

import "github.com/FooSoft/goldsmith"

type Filter func(f goldsmith.File) bool

type where struct {
	f Filter
	p goldsmith.Plugin
}

func New(f Filter, p goldsmith.Plugin) goldsmith.Plugin {
	return &where{f, p}
}

func (w *where) Initialize(ctx goldsmith.Context) error {
	if init, ok := w.p.(goldsmith.Initializer); ok {
		return init.Initialize(ctx)
	}

	return nil
}

func (w *where) Accept(ctx goldsmith.Context, f goldsmith.File) bool {
	if !w.f(f) {
		return false
	}

	if accept, ok := w.p.(goldsmith.Accepter); ok {
		return accept.Accept(ctx, f)
	}

	return true
}

func (w *where) Finalize(ctx goldsmith.Context) error {
	if fin, ok := w.p.(goldsmith.Finalizer); ok {
		return fin.Finalize(ctx)
	}

	return nil
}

func (w *where) Process(ctx goldsmith.Context, f goldsmith.File) error {
	if proc, ok := w.p.(goldsmith.Processor); ok {
		return proc.Process(ctx, f)
	}

	ctx.DispatchFile(f)
	return nil
}
