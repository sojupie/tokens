import test from "node:test";
import assert from "node:assert/strict";
import { readFile, readdir } from "node:fs/promises";
import { buildCatalogAssets, catalogAssets } from "../catalog/web/build-assets.ts";
import { Script } from "node:vm";
import {
  buildComponentAssets,
  componentAssets,
} from "./components/build-assets.ts";

test("embedded component assets match all maintained sources", async () => {
  await buildComponentAssets({ check: true });
  await buildCatalogAssets({ check: true });
  for (const kind of ["style", "behavior"]) {
    const directory = kind === "style" ? "styles" : "behaviors";
    const root = new URL("./components/", import.meta.url);
    const sources = JSON.parse(
      await readFile(new URL(`${kind}-sources.json`, root), "utf8"),
    ).map(entry => typeof entry === 'string' ? entry : entry.source);
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
  for (const [file, content] of assets) {
    if (file.endsWith(".js")) new Script(content, { filename: file });
  }
  for (const name of ["chips", "switches", "button-groups", "segmented-buttons"]) {
    assert(assets.has(`static/runtime/${name}.js`), "TypeScript must keep public JavaScript paths");
    assert(!assets.has(`static/runtime/${name}.ts`));
  }
  const catalog = await catalogAssets();
  assert(catalog.has("asset-loader.js"));
  for (const [file, content] of catalog) new Script(content, { filename: file });
});
