import test from "node:test";
import assert from "node:assert/strict";
import { compileBehavior, validateScriptOwnership } from "./components/build-assets.mjs";
import { Script } from "node:vm";

test("component script ownership rejects omissions, duplicates and missing files", () => {
  const files = ["runtime/utils.js", "carousel.js"];
  const modules = { common: { scripts: [files[0]] }, carousel: { scripts: [files[1]] } };
  assert.doesNotThrow(() => validateScriptOwnership(modules, files));
  assert.throws(() => validateScriptOwnership({ common: modules.common }, files), /Unowned component script: carousel.js/);
  assert.throws(() => validateScriptOwnership({ ...modules, extra: modules.carousel }, files), /Duplicate script owner/);
  assert.throws(() => validateScriptOwnership({ extra: { scripts: ["missing.js"] } }, files), /Unknown component script: missing.js/);
});

test("classic behavior entries include their source imports without browser dependencies", async () => {
  for (const name of ["button-groups", "segmented-buttons"]) {
    const result = await compileBehavior(`behaviors/${name}.js`);
    assert(result.inputs.some(file => file.endsWith("helpers/navigation.js")));
    assert(result.inputs.some(file => file.endsWith(`behaviors/${name}.js`)));
    assert(!result.code.includes("MaterialComponents"), "the control must not require the global helper provider");
    assert(!result.code.includes("snackbarTimer"), "unrelated helpers must not enter the chunk");
    new Script(result.code);
  }
});
