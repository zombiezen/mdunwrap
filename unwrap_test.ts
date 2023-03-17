import { assertEquals } from "https://deno.land/std@0.179.0/testing/asserts.ts";
import * as path from "https://deno.land/std@0.179.0/path/mod.ts";
import { filter } from "./unwrap.ts";

Deno.test(async function filterTest(t) {
  const dataDirectory = "testdata";
  const suffix = ".in.md";
  const contents = Deno.readDir(dataDirectory);
  for await (const ent of contents) {
    if (!ent.isFile || ent.name.startsWith(".") || !ent.name.endsWith(suffix)) {
      continue;
    }
    const base = ent.name.substring(0, ent.name.length - suffix.length);
    await t.step(base, async () => {
      const input = await Deno.readTextFile(path.join(dataDirectory, ent.name));
      const want = await Deno.readTextFile(
        path.join(dataDirectory, base + ".out.md"),
      );
      const got = filter(input);
      assertEquals(got, want);
    });
  }
});
