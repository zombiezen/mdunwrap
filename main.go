// Copyright 2023 Ross Light
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		 https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	fmt.Println("Hello, World!")
}

func filter(doc []byte) []byte {
	docNode := goldmark.DefaultParser().Parse(text.NewReader(doc))
	var buf []byte
	ast.Walk(docNode, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.Kind() {
		case ast.KindDocument:
		case ast.KindParagraph:
			if !entering {
				buf = append(buf, '\n')
			}
		case ast.KindText:
			if entering {
				n := n.(*ast.Text)
				buf = append(buf, n.Text(doc)...)
				if n.SoftLineBreak() {
					buf = append(buf, ' ')
				} else if n.HardLineBreak() {
					buf = append(buf, '\n')
				}
			}
		default:
			if entering {
				buf = append(buf, n.Text(doc)...)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return buf
}
