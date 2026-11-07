import test from "node:test";
import assert from "node:assert/strict";
import { compileBehavior, validateScriptOwnership } from "./components/build-assets.ts";
import { Script } from "node:vm";
import { readFile } from "node:fs/promises";

test("component script ownership rejects omissions, duplicates and missing files", () => {
  const files = ["runtime/utils.js", "carousel.js"];
  const modules = { common: { scripts: [files[0]] }, carousel: { scripts: [files[1]] } };
  assert.doesNotThrow(() => validateScriptOwnership(modules, files));
  assert.throws(() => validateScriptOwnership({ common: modules.common }, files), /Unowned component script: carousel.js/);
  assert.throws(() => validateScriptOwnership({ ...modules, extra: modules.carousel }, files), /Duplicate script owner/);
  assert.throws(() => validateScriptOwnership({ extra: { scripts: ["missing.js"] } }, files), /Unknown component script: missing.js/);
});

test("classic behavior entries include their source imports without browser dependencies", async () => {
  const entries = JSON.parse(await readFile(new URL('./components/behavior-sources.json', import.meta.url), 'utf8'));
  const sources = entries.map(entry => typeof entry === 'string' ? entry : entry.source);
  const sourceFor = name => {
    const source = sources.find(source => source.startsWith(`behaviors/${name}.`));
    assert(source, `Missing source for ${name}`);
    return source;
  };
  for (const name of ["button-groups", "segmented-buttons", "navigation", "fab-menus", "split-buttons", "toolbars", "lists"]) {
    const source = sourceFor(name);
    const result = await compileBehavior(source);
    assert(result.inputs.some(file => file.endsWith("helpers/navigation.ts")));
    assert(result.inputs.some(file => file.endsWith(source)));
    assert(!result.code.includes("MaterialComponents"), "the control must not require the global helper provider");
    assert(!result.code.includes("snackbarTimer"), "unrelated helpers must not enter the chunk");
    new Script(result.code);
  }
  for (const name of ["sliders", "snackbars"]) {
    const result = await compileBehavior(sourceFor(name));
    assert(result.inputs.some(file => file.endsWith("helpers/controls.ts")));
    assert(!result.code.includes("MaterialComponents"));
    assert(!result.code.includes("navigationIndex"));
    new Script(result.code);
  }
});

test("behavior consumers do not depend on the compatibility helper global", async () => {
  const entries = JSON.parse(await readFile(new URL('./components/behavior-sources.json', import.meta.url), 'utf8'));
  for (const entry of entries) {
    const source = typeof entry === 'string' ? entry : entry.source;
    if (source === 'behaviors/utils.js') continue;
    const result = await compileBehavior(source);
    assert(!result.code.includes('MaterialComponents'), source);
  }
});
