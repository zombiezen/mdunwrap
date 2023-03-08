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

package markdown_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/net/html"
	. "zombiezen.com/go/mdunwrap/internal/markdown"
)

var supportedSections = map[string]struct{}{
	"Thematic breaks": {},
}

func TestSpec(t *testing.T) {
	testsuiteData, err := os.ReadFile(filepath.Join("testdata", "spec-0.30.json"))
	if err != nil {
		t.Fatal(err)
	}
	var testsuite []struct {
		Markdown string
		HTML     string
		Example  int
		Section  string
	}
	if err := json.Unmarshal(testsuiteData, &testsuite); err != nil {
		t.Fatal(err)
	}

	for _, test := range testsuite {
		t.Run(fmt.Sprintf("Example%d", test.Example), func(t *testing.T) {
			if _, ok := supportedSections[test.Section]; !ok {
				t.Skipf("Section %q not implemented yet", test.Section)
			}
			blocks := Parse([]byte(test.Markdown))
			buf := new(bytes.Buffer)
			if err := RenderHTML(buf, blocks); err != nil {
				t.Error("RenderHTML:", err)
			}
			got := string(normalizeHTML(buf.Bytes()))
			want := string(normalizeHTML([]byte(test.HTML)))
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("-want +got:\n%s", diff)
			}
		})
	}
}

var whitespaceRE = regexp.MustCompile(`\s+`)

func normalizeHTML(b []byte) []byte {
	tok := html.NewTokenizer(bytes.NewReader(b))
	var output []byte
	last := html.StartTagToken
	var lastTag string
	inPre := false
	for {
		tt := tok.Next()
		switch tt {
		case html.ErrorToken:
			return output
		case html.TextToken:
			data := tok.Raw()
			afterTag := last == html.EndTagToken || last == html.StartTagToken
			afterBlockTag := afterTag && isBlockTag(lastTag)
			if afterTag && lastTag == "br" {
				data = bytes.TrimLeft(data, "\n")
			}
			if !inPre {
				data = whitespaceRE.ReplaceAll(data, []byte(" "))
			}
			if afterBlockTag && !inPre {
				if last == html.StartTagToken {
					data = bytes.TrimLeftFunc(data, unicode.IsSpace)
				} else if last == html.EndTagToken {
					data = bytes.TrimSpace(data)
				}
			}
			output = append(output, data...)
		case html.EndTagToken:
			tagBytes, _ := tok.TagName()
			tag := string(tagBytes)
			if tag == "pre" {
				inPre = false
			} else if isBlockTag(tag) {
				output = bytes.TrimRightFunc(output, unicode.IsSpace)
			}
			output = append(output, "</"...)
			output = append(output, tag...)
			output = append(output, ">"...)
			lastTag = tag
		case html.StartTagToken, html.SelfClosingTagToken:
			tagBytes, hasAttr := tok.TagName()
			tag := string(tagBytes)
			if tag == "pre" {
				inPre = true
			}
			if isBlockTag(tag) {
				output = bytes.TrimRightFunc(output, unicode.IsSpace)
			}
			output = append(output, "<"...)
			output = append(output, tag...)
			if hasAttr {
				// TODO(now)
				panic("bork")
			}
			output = append(output, ">"...)
			lastTag = tag
		case html.CommentToken:
			output = append(output, tok.Raw()...)
		}

		last = tt
		if tt == html.SelfClosingTagToken {
			last = html.EndTagToken
		}
	}
}

var blockTags = map[string]struct{}{
	"article":    {},
	"header":     {},
	"aside":      {},
	"hgroup":     {},
	"blockquote": {},
	"hr":         {},
	"iframe":     {},
	"body":       {},
	"li":         {},
	"map":        {},
	"button":     {},
	"object":     {},
	"canvas":     {},
	"ol":         {},
	"caption":    {},
	"output":     {},
	"col":        {},
	"p":          {},
	"colgroup":   {},
	"pre":        {},
	"dd":         {},
	"progress":   {},
	"div":        {},
	"section":    {},
	"dl":         {},
	"table":      {},
	"td":         {},
	"dt":         {},
	"tbody":      {},
	"embed":      {},
	"textarea":   {},
	"fieldset":   {},
	"tfoot":      {},
	"figcaption": {},
	"th":         {},
	"figure":     {},
	"thead":      {},
	"footer":     {},
	"tr":         {},
	"form":       {},
	"ul":         {},
	"h1":         {},
	"h2":         {},
	"h3":         {},
	"h4":         {},
	"h5":         {},
	"h6":         {},
	"video":      {},
	"script":     {},
	"style":      {},
}

func isBlockTag(tag string) bool {
	_, ok := blockTags[tag]
	return ok
}

func TestNormalizeHTML(t *testing.T) {
	tests := []struct {
		b    string
		want string
	}{
		{"<p>a  \t b</p>", "<p>a b</p>"},
		{"<p>a  \t\nb</p>", "<p>a b</p>"},
		{"<p>a  b</p>", "<p>a b</p>"},
		{" <p>a  b</p>", "<p>a b</p>"},
		{"<p>a  b</p> ", "<p>a b</p>"},
		{"\n\t<p>\n\t\ta  b\t\t</p>\n\t", "<p>a b</p>"},
		{"<i>a  b</i> ", "<i>a b</i> "},
		{"<br />", "<br>"},
		// {`<a title="bar" HREF="foo">x</a>`, `<a href="foo" title="bar">x</a>`},
		// {"&forall;&amp;&gt;&lt;&quot;", "\u2200&amp;&gt;&lt;&quot;"},
	}
	for _, test := range tests {
		if got := normalizeHTML([]byte(test.b)); string(got) != test.want {
			t.Errorf("normalizeHTML(%q) = %q; want %q", test.b, got, test.want)
		}
	}
}
