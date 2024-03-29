// Copyright 2023 Ross Light
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// @deno-types="npm:@types/commonmark@0.27.5"
import * as commonmark from "npm:commonmark@0.30.0";

declare module "npm:commonmark@0.30.0" {
  interface Node {
    _isFenced?: boolean;
  }
}

export function filter(doc: string): string {
  const parser = new commonmark.Parser();
  const parsed = parser.parse(doc);
  const walker = parsed.walker();

  let event: commonmark.NodeWalkingStep | null;
  let firstBlock = true;
  const parts: string[] = [];
  const prefix: string[] = [];
  while ((event = walker.next())) {
    if (event.entering && isBlock(event.node)) {
      if (firstBlock) {
        firstBlock = false;
      } else if (
        event.node.type !== "item" || event.node.parent?.listTight === false
      ) {
        parts.push(...prefixTrimRight(prefix));
        parts.push("\n");
        parts.push(...prefix);
      }
    }
    switch (event.node.type) {
      case "text":
        parts.push(event.node.literal !== null ? unescapeText(event.node.literal) : "<null>");
        break;
      case "linebreak":
        parts.push("\\\n");
        parts.push(...prefix);
        break;
      case "softbreak":
        parts.push(" ");
        break;
      case "thematic_break":
        parts.push("---\n");
        break;
      case "emph":
        parts.push("*");
        break;
      case "strong":
        parts.push("**");
        break;
      case "code":
        parts.push("`");
        // TODO(soon): Escape characters as needed.
        parts.push(event.node.literal ?? "<null>");
        parts.push("`");
        break;
      case "link":
        if (event.entering) {
          parts.push("[");
        } else {
          parts.push("](");
          parts.push(event.node.destination ?? "");
          parts.push(")");
        }
        break;
      case "image":
        if (event.entering) {
          parts.push("![");
        } else {
          parts.push("](");
          parts.push(event.node.destination ?? "");
          parts.push(")");
        }
        break;
      case "heading":
        if (event.entering) {
          for (let i = 0; i < event.node.level; i++) {
            parts.push("#");
          }
          parts.push(" ");
        } else {
          parts.push("\n");
        }
        break;
      case "code_block":
        {
          if (event.node._isFenced) {
            parts.push("```");
            if (event.node.info) {
              parts.push(event.node.info);
            }
            parts.push("\n");
          }
          let contents = event.node.literal ?? "";
          if (contents.endsWith("\n")) {
            contents = contents.substring(0, contents.length - 1);
          }
          let first = true;
          for (const line of contents.split("\n")) {
            if (line === "") {
              if (!first || event.node._isFenced) {
                parts.push(...prefixTrimRight(prefix));
              } else {
                first = false;
              }
              parts.push("\n");
              continue;
            }

            if (!first || event.node._isFenced) {
              parts.push(...prefix);
            } else {
              first = false;
            }
            if (!event.node._isFenced) {
              parts.push("    ");
            }
            parts.push(line);
            parts.push("\n");
          }
          if (event.node._isFenced) {
            parts.push(...prefix);
            parts.push("```\n");
          }
        }
        break;
      case "list":
        if (event.entering) {
          firstBlock = true;
        } else {
          firstBlock = false;
        }
        break;
      case "item":
        if (event.entering) {
          if (event.node.listType === "bullet") {
            parts.push("- ");
            prefix.push("  ");
          } else {
            const numbering = event.node.listStart.toString();
            parts.push(numbering, event.node.listDelimiter, " ");
            prefix.push(" ".repeat(numbering.length + 2));
          }
          firstBlock = true;
        } else {
          firstBlock = false;
          prefix.pop();
        }
        break;
      case "block_quote":
        if (event.entering) {
          parts.push("> ");
          prefix.push("> ");
          firstBlock = true;
        } else {
          prefix.pop();
          firstBlock = false;
        }
        break;
      case "paragraph":
        if (!event.entering) {
          parts.push("\n");
        }
        break;
      case "document":
        // Do nothing.
        break;
      default:
        parts.push(
          `<${event.node.type} ${event.node.literal} (${event.node.info})>`,
        );
        break;
    }
  }
  return parts.join("");
}

function prefixTrimRight(prefix: string[]): string[] {
  for (let i = prefix.length - 1; i >= 0; i--) {
    const trimmed = prefix[i].trimEnd();
    if (trimmed !== "") {
      if (trimmed === prefix[i]) {
        if (i + 1 === prefix.length) {
          return prefix;
        }
        return prefix.slice(0, i + 1);
      } else {
        const newArray = prefix.slice(0, i);
        newArray.push(trimmed);
        return newArray;
      }
    }
  }
  return [];
}

const MUST_ESCAPE = new Set<string>([
  '![',
  '#',
  '&',
  '[',
  '*',
  '_',
  '|',
  '\\',
  '<',
  '`',
]);

function unescapeText(s: string): string {
  let start = 0;
  let result = "";
  for (let i = 0; i < s.length; i++) {
    for (const seq of MUST_ESCAPE) {
      if (s.startsWith(seq, i)) {
        if (start < i) {
          result += s.substring(start, i);
        }
        start = i + seq.length;
        for (let j = 0; j < s.length; j++) {
          result += "\\" + seq[j];
        }
        i += seq.length - 1;
        break;
      }
    }
  }
  if (start < s.length) {
    result += s.substring(start);
  }
  return result;
}

const BLOCK_TYPES = new Set<commonmark.NodeType>([
  "block_quote",
  "code_block",
  "custom_block",
  "heading",
  "html_block",
  "item",
  "list",
  "paragraph",
  "thematic_break",
]);

/** Reports whether a node is a block, excluding `document`. */
function isBlock(node: commonmark.Node): boolean {
  return BLOCK_TYPES.has(node.type);
}
