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
import { readableStreamFromReader } from "https://deno.land/std@0.179.0/streams/mod.ts";

import { filter } from "./unwrap.ts";

async function run(write: boolean, args: string[]): Promise<void> {
  if (args.length === 0 && !write) {
    // Simple stdin, stdout.
    const input = await readAllString(Deno.stdin.readable);
    const output = (new TextEncoder()).encode(filter(input));
    await Deno.stdout.write(output);
  } else if (args.length === 0 && write) {
    await Deno.stderr.write(
      new TextEncoder().encode(
        "mdunwrap: must include filenames with -w option",
      ),
    );
  } else {
    const enc = new TextEncoder();
    for (const fname of args) {
      const f = await Deno.open(fname, { read: true, write });
      try {
        // TODO(someday): We should be able to use readAllString(f.readable),
        // but as of Deno 1.31.1, https://github.com/denoland/deno/issues/17828
        // causes its usage to automatically close the file.
        const input = await readAllString(
          readableStreamFromReader(nopCloser(f)),
        );

        const output = enc.encode(filter(input));
        if (write) {
          await f.seek(0, Deno.SeekMode.Start);
          await f.truncate();
          await f.write(output);
        } else {
          await Deno.stdout.write(output);
        }
      } finally {
        f.close();
      }
    }
  }
}

async function readAllString(r: ReadableStream<Uint8Array>): Promise<string> {
  const parts: string[] = [];
  const sink = new WritableStream<string>({
    write(chunk) {
      parts.push(chunk);
    },
  });
  await r.pipeThrough(new TextDecoderStream()).pipeTo(sink);
  return parts.join("");
}

function nopCloser(r: Deno.Reader): Deno.Reader & Deno.Closer {
  return {
    read(p) {
      return r.read(p);
    },
    close() {},
  };
}

if (import.meta.main) {
  const flags = parseFlags(Deno.args, {
    boolean: ["w"],
    string: ["_"],
  });
  run(flags.w, flags._.map((a) => a.toString()));
}
