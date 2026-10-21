import test from "node:test";
import assert from "node:assert/strict";
import { readFile, readdir } from "node:fs/promises";
import { Script } from "node:vm";
import {
  buildComponentAssets,
  componentAssets,
} from "./components/build-assets.mjs";

test("embedded component assets match all maintained sources", async () => {
  await buildComponentAssets({ check: true });
  for (const kind of ["style", "behavior"]) {
    const directory = kind === "style" ? "styles" : "behaviors";
    const root = new URL("./components/", import.meta.url);
    const sources = JSON.parse(
      await readFile(new URL(`${kind}-sources.json`, root), "utf8"),
    );
    const files = (await readdir(new URL(`${directory}/`, root))).map(
      (file) => `${directory}/${file}`,
    );
    assert.equal(
      new Set(sources).size,
      sources.length,
      "sources appear only once",
    );
    assert.deepEqual(
      [...sources].sort(),
      files.sort(),
      "every source is included",
    );
  }
  const assets = await componentAssets();
  new Script(assets.get("static/components.js"));
});
