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

package markdown

type RootBlock struct {
	Source      []byte
	StartLine   int
	StartOffset int64
	Block
}

type Block struct {
	Kind     BlockKind
	Start    int
	End      int
	Children []Child
}

type BlockKind uint16

const (
	ParagraphKind BlockKind = 1 + iota
	ThematicBreakKind
	ATXHeadingKind
	SetextHeadingKind
	IndentedCodeBlockKind
	FencedCodeBlockKind
	HTMLBlockKind
	LinkReferenceDefinitionKind
	BlockQuoteKind
	ListItemKind
	ListKind
)

type Inline struct {
	Kind  InlineKind
	Start int
	End   int
}

type InlineKind uint16

type Child struct {
	block Block
	typ   uint8
}

const (
	childTypeBlock = 1 + iota
	childTypeInline
)

func (c *Child) Block() Block {
	if c.typ != childTypeBlock {
		return Block{}
	}
	return c.block
}

func (c *Child) Inline() Inline {
	if c.typ != childTypeInline {
		return Inline{}
	}
	return Inline{
		Kind:  InlineKind(c.block.Kind),
		Start: c.block.Start,
		End:   c.block.End,
	}
}
