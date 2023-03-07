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
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	write := flag.Bool("w", false, "write to file")
	flag.Parse()

	if err := run(*write, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "mdunwrap:", err)
		os.Exit(1)
	}
}

func run(write bool, args []string) error {
	switch {
	case len(args) == 0 && !write:
		// Simple stdin, stdout.
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		output := filter(input)
		if _, err := os.Stdout.Write(output); err != nil {
			return err
		}
	case len(args) == 0 && write:
		return errors.New("must include filenames with -w option")
	default:
		for _, fname := range args {
			flag := os.O_RDONLY
			if write {
				flag = os.O_RDWR
			}
			f, err := os.OpenFile(fname, flag, 0)
			if err != nil {
				return err
			}
			input, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}

			output := filter(input)
			if write {
				if _, err := f.Seek(0, io.SeekStart); err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
				if err := f.Truncate(0); err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
				if _, err := f.Write(output); err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
				if err := f.Close(); err != nil {
					return fmt.Errorf("%s: %w", fname, err)
				}
			} else {
				f.Close()
				if _, err := os.Stdout.Write(output); err != nil {
					return err
				}
			}
		}
	}
	return nil
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
