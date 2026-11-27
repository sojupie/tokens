import test from "node:test";
import assert from "node:assert/strict";
import { componentRadius, checkButtonRendering } from "./conformance/browser/check-button-rendering.mjs";
import { checkComponentRendering } from "./conformance/browser/check-component-rendering.mjs";
import { runInNewContext } from "node:vm";
test("component corners resolve bounded pixel, full and percentage radii", () => {
  assert.equal(componentRadius("12px", 180, 40), 12);
  assert.equal(componentRadius("calc(infinity * 1px)", 180, 40), 20);
  assert.equal(componentRadius("min(50%, 20px)", 180, 40), 20);
  assert.equal(componentRadius("50%", 44, 40), 20);
  assert.equal(componentRadius("0px", 44, 40), 0);
  assert.ok(Number.isNaN(componentRadius("unresolved", 44, 40)));
});

test('rendering checks reject empty, hidden and unknown-family fixtures', () => {
  for (const check of [checkButtonRendering, checkComponentRendering]) {
    for (const nodes of [[], [{ getClientRects: () => [] }]]) {
      const result = runInNewContext('(' + check.toString() + ')()', {
        document: { querySelectorAll: () => nodes, documentElement: { dataset: {} } },
      });
      assert.ok(result.failureCount > 0);
      assert.equal(result.failures.at(-1).property, 'fixtures');
    }
    assert.throws(() => runInNewContext('(' + check.toString() + ')(' + JSON.stringify({ minimumCases: 0, minimumConfigurations: 0 }) + ')'), /at least one/);
  }
  const result = runInNewContext('(' + checkComponentRendering.toString() + ')({requiredFamilies:["list"]})', {
    document: { querySelectorAll: selector => selector === '[data-component-case]' ? [{ getClientRects: () => [{}], dataset: { family: 'unknown' }, getAttribute: () => null, querySelector: () => null }] : [] },
    getComputedStyle: () => ({ getPropertyValue: () => '' }),
  });
  assert.ok(result.failureCount > 0);
  assert.deepEqual(Array.from(result.failures.at(-1).missingFamilies), ['list']);
});
