import test from "node:test";
import assert from "node:assert/strict";
import { componentRadius } from "./conformance/check-button-rendering.mjs";
test("component corners resolve bounded pixel, full and percentage radii", () => {
  assert.equal(componentRadius("12px", 180, 40), 12);
  assert.equal(componentRadius("calc(infinity * 1px)", 180, 40), 20);
  assert.equal(componentRadius("min(50%, 20px)", 180, 40), 20);
  assert.equal(componentRadius("50%", 44, 40), 20);
  assert.equal(componentRadius("0px", 44, 40), 0);
  assert.ok(Number.isNaN(componentRadius("unresolved", 44, 40)));
});
