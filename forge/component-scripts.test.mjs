import test from "node:test";
import assert from "node:assert/strict";
import { validateScriptOwnership } from "./components/build-assets.mjs";

test("component script ownership rejects omissions, duplicates and missing files", () => {
  const files = ["runtime/utils.js", "carousel.js"];
  const modules = { common: { scripts: [files[0]] }, carousel: { scripts: [files[1]] } };
  assert.doesNotThrow(() => validateScriptOwnership(modules, files));
  assert.throws(() => validateScriptOwnership({ common: modules.common }, files), /Unowned component script: carousel.js/);
  assert.throws(() => validateScriptOwnership({ ...modules, extra: modules.carousel }, files), /Duplicate script owner/);
  assert.throws(() => validateScriptOwnership({ extra: { scripts: ["missing.js"] } }, files), /Unknown component script: missing.js/);
});
