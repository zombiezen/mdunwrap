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

// Package markdown provides a [CommonMark] parser.
//
// [CommonMark]: https://commonmark.org/
package markdown

import (
	"bytes"
	"fmt"
	"io"
)

const (
	tabStopSize          = 4
	codeBlockIndentLimit = 4
)

type Parser struct {
	buf      []byte // current block being parsed
	offset   int64  // offset from beginning of stream to beginning of buf
	parsePos int    // parse position within buf
	lineno   int    // line number of parse position

	r   io.Reader
	err error // non-nil indicates there is no more data after end of buf
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		r: r,
	}
}

func Parse(source []byte) []*RootBlock {
	p := &Parser{
		buf: source,
		err: io.EOF,
	}
	var blocks []*RootBlock
	for {
		block, err := p.NextBlock()
		if err == io.EOF {
			return blocks
		}
		if err != nil {
			panic(err)
		}
		blocks = append(blocks, block)
	}
}

func (p *Parser) NextBlock() (*RootBlock, error) {
	// Keep going until we encounter a non-blank line.
	var line []byte
	for {
		line = p.readline()
		if len(line) == 0 {
			return nil, p.err
		}
		if !isBlankLine(line) {
			break
		}

		p.offset += int64(p.parsePos)
		p.buf = p.buf[p.parsePos:]
	}

	// Open root block.
	root := &RootBlock{
		StartLine:   p.lineno,
		StartOffset: p.offset,
	}

	root.Source = p.consume()
	return root, nil
}

// readline reads the next line of input, growing p.buf as necessary.
// It will return a zero-length slice if and only if it has reached the end of input.
// After calling readline, p.lineno will contain the current line's number.
func (p *Parser) readline() []byte {
	const (
		chunkSize    = 8 * 1024
		maxBlockSize = 1024 * 1024
	)

	eolEnd := -1
	for {
		// Check if we have a line ending available.
		if i := bytes.IndexAny(p.buf[p.parsePos:], "\r\n"); i >= 0 {
			eolStart := p.parsePos + i
			if p.buf[eolStart] == '\n' {
				eolEnd = eolStart + 1
				break
			}
			if eolStart+1 < len(p.buf) {
				// Carriage return with enough buffer for 1 byte lookahead.
				eolEnd = eolStart + 1
				if p.buf[eolEnd] == '\n' {
					eolEnd++
				}
				break
			}
			if p.err != nil {
				// Carriage return right before EOF.
				eolEnd = len(p.buf)
				break
			}
		}

		// If we don't have any more line ending available,
		// but we're at EOF, return everything we have.
		if p.err != nil {
			eolEnd = len(p.buf)
			break
		}

		// If we're already at the maximum block size,
		// then drop the line and pretend it's an EOF.
		if len(p.buf) >= maxBlockSize {
			p.lineno++
			p.buf = p.buf[:p.parsePos]
			p.err = fmt.Errorf("line %d: block too large", p.lineno)
			return nil
		}

		// Grab more data from the reader.
		newSize := len(p.buf) + chunkSize
		if newSize > maxBlockSize {
			newSize = maxBlockSize
		}
		if cap(p.buf) < newSize {
			newbuf := make([]byte, len(p.buf), newSize)
			copy(newbuf, p.buf)
			p.buf = newbuf
		}
		var n int
		n, p.err = p.r.Read(p.buf[len(p.buf):newSize])
		p.buf = p.buf[:len(p.buf)+n]
	}

	line := p.buf[p.parsePos:eolEnd]
	p.parsePos = eolEnd
	p.lineno++
	return line
}

func (p *Parser) consume() []byte {
	out := p.buf[:p.parsePos:p.parsePos]
	p.offset += int64(p.parsePos)
	p.buf = p.buf[p.parsePos:]
	p.parsePos = 0
	return out
}

type linePrefix struct {
	kind  BlockKind
	start int
	end   int
}

func identifyLine(dst []linePrefix, line []byte) []linePrefix {
	for pos := 0; ; {
		start := pos
		if isBlankLine(line[pos:]) {
			return dst
		}

		// Consume leading indentation.
		indent := 0
		for indent < codeBlockIndentLimit && indent < len(line) && line[indent] == ' ' {
			indent++
		}
		pos += indent
		if indent < codeBlockIndentLimit && indent < len(line) && line[indent] == '\t' {
			// We only need to consume a single tab
			// since it will automatically trigger an indented code block.
			indent += tabStopSize
			pos++
		}
		if indent >= codeBlockIndentLimit {
			return append(dst, linePrefix{
				kind:  IndentedCodeBlockKind,
				start: start,
				end:   pos,
			})
		}

		// Now start checking for indicators.
		if pos < len(line) && line[pos] == '>' {
			dst = append(dst, linePrefix{
				kind:  BlockQuoteKind,
				start: pos,
				end:   pos + 1,
			})
			pos++
			if pos < len(line) && line[pos] == ' ' {
				pos++
			}
			continue
		}
		if end := parseThematicBreak(line[pos:]); end >= 0 {
			return append(dst, linePrefix{
				kind:  ThematicBreakKind,
				start: pos,
				end:   pos + end,
			})
		}
		if h := parseATXHeading(line[pos:]); h.level != 0 {
			return append(dst, linePrefix{
				kind:  ATXHeadingKind,
				start: pos,
				end:   pos + h.level,
			})
		}

		// Did not match anything else. Assume it's a paragraph.
		return append(dst, linePrefix{
			kind:  ParagraphKind,
			start: pos,
			end:   pos,
		})
	}
}

func isBlankLine(line []byte) bool {
	for _, b := range line {
		if !(b == '\r' || b == '\n' || b == ' ' || b == '\t') {
			return false
		}
	}
	return true
}

// parseThematicBreak attempts to parse the line as a [thematic break].
// It returns the end of the thematic break characters
// or -1 if the line is not a thematic break.
// parseThematicBreak assumes that the caller has stripped any leading indentation.
//
// [thematic break]: https://spec.commonmark.org/0.30/#thematic-breaks
func parseThematicBreak(line []byte) (end int) {
	const chars = "-_*"
	n := 0
	var want byte
	for i, b := range line {
		switch b {
		case '-', '_', '*':
			if n == 0 {
				want = b
			} else if b != want {
				return -1
			}
			n++
			end = i + 1
		case ' ', '\t', '\r', '\n':
			// Ignore
		default:
			return -1
		}
	}
	if n < 3 {
		return -1
	}
	return end
}

type atxHeading struct {
	level        int // 1-6
	contentStart int
	contentEnd   int
}

// parseATXHeading attempts to parse the line as an [ATX heading].
// The level is zero if the line is not an ATX heading.
// parseATXHeading assumes that the caller has stripped any leading indentation.
//
// [ATX heading]: https://spec.commonmark.org/0.30/#atx-headings
func parseATXHeading(line []byte) atxHeading {
	var h atxHeading
	for h.level < len(line) && line[h.level] == '#' {
		h.level++
	}
	if h.level == 0 || h.level > 6 {
		return atxHeading{}
	}

	// Consume required whitespace before heading.
	i := h.level
	if i >= len(line) || line[i] == '\n' || line[i] == '\r' {
		h.contentStart = i
		h.contentEnd = i
		return h
	}
	if !(line[i] == ' ' || line[i] == '\t') {
		return atxHeading{}
	}
	i++

	// Advance past leading whitespace.
	for i < len(line) && line[i] == ' ' || line[i] == '\t' {
		i++
	}
	h.contentStart = i

	// Find end of heading line. Skip past trailing spaces.
	h.contentEnd = len(line)
	hitHash := false
scanBack:
	for ; h.contentEnd > h.contentStart; h.contentEnd-- {
		switch line[h.contentEnd-1] {
		case '\r', '\n':
			// Skip past EOL.
		case ' ', '\t':
			if isEndEscaped(line[:h.contentEnd-1]) {
				break scanBack
			}
		case '#':
			hitHash = true
			break scanBack
		default:
			break scanBack
		}
	}
	if !hitHash {
		return h
	}

	// We've encountered one hashmark '#'.
	// Consume all of them, unless they are preceded by a space or tab.
scanTrailingHashes:
	for i := h.contentEnd - 1; ; i-- {
		if i <= h.contentStart {
			h.contentEnd = h.contentStart
			break
		}
		switch line[i] {
		case '#':
			// Keep going.
		case ' ', '\t':
			h.contentEnd = i + 1
			break scanTrailingHashes
		default:
			return h
		}
	}
	// We've hit the end of hashmarks. Trim trailing whitespace.
	for ; h.contentEnd > h.contentStart; h.contentEnd-- {
		if b := line[h.contentEnd-1]; !(b == ' ' || b == '\t') || isEndEscaped(line[:h.contentEnd-1]) {
			break
		}
	}
	return h
}

// isEndEscaped reports whether s ends with an odd number of backslashes.
func isEndEscaped(s []byte) bool {
	n := 0
	for ; n < len(s); n++ {
		if s[len(s)-n-1] != '\\' {
			break
		}
	}
	return n%2 == 1
}
