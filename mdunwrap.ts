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

import { parse as parseFlags } from "https://deno.land/std@0.179.0/flags/mod.ts";

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
  const parts: string[] = [];
  const prefix: string[] = [];
  while ((event = walker.next())) {
    switch (event.node.type) {
      case "text":
        parts.push(event.node.literal ?? "<null>");
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
      case "code_block":
        if (event.node._isFenced) {
          parts.push("```");
          if (event.node.info) {
            parts.push(event.node.info);
          }
          parts.push("\n");
          parts.push(event.node.literal ?? "");
          parts.push("```\n\n");
        } else {
          for (const line of (event.node.literal ?? "").split("\n")) {
            parts.push("    ");
            parts.push(line);
            parts.push("\n");
          }
        }
        break;
      case "paragraph":
        if (!event.entering) {
          parts.push("\n\n");
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

async function run(write: boolean, args: string[]): Promise<void> {
  if (args.length === 0 && !write) {
    // Simple stdin, stdout.
    const input = await readAllString(Deno.stdin.readable);
    const output = (new TextEncoder()).encode(filter(input));
    await Deno.stdout.write(output);
  } else if (args.length === 0 && write) {
    await Deno.stderr.write(
      new TextEncoder().encode("must include filenames with -w option"),
    );
  } else {
    for (const _fname of args) {
      // TODO(now)
    }
  }
}

async function readAllString(r: ReadableStream<Uint8Array>): Promise<string> {
  const parts: string[] = [];
  const stream = await r.pipeThrough(new TextDecoderStream());
  for await (const chunk of stream) {
    parts.push(chunk);
  }
  return parts.join("");
}

if (import.meta.main) {
  const flags = parseFlags(Deno.args, {
    boolean: ["w"],
    string: ["_"],
  });
  run(flags.w, flags._.map((a) => a.toString()));
}
