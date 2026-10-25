const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');
const out = { exports: {} };
runInNewContext(readFileSync(`${__dirname}/behaviors/navigation.js`, 'utf8'), { module: out });
test('drawer dismissal follows the logical anchor and rejects taps, vertical scroll and small drags', () => {
  const {drawerDismissed: swipe} = out.exports;
  assert.equal(swipe(-90, 8, 320, false), true);
  assert.equal(swipe(90, 8, 320, true), true);
  assert.equal(swipe(90, 8, 320, false), false);
  assert.equal(swipe(-90, 8, 320, true), false);
  assert.equal(swipe(-20, 0, 320, false), false);
  assert.equal(swipe(-90, 120, 320, false), false);
  assert.equal(swipe(0, 0, 320, false), false);
});
