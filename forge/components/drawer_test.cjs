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

test('drawer openers resolve their explicit target outside component demos', () => {
 const drawer = {id:'site-navigation'};
 const doc = {getElementById(id){assert.equal(id,'site-navigation');return drawer;}};
 assert.equal(out.exports.drawerForTrigger({getAttribute(){return 'site-navigation';},closest(){throw Error('must not require demo wrapper');}},doc),drawer);
 const legacy={querySelector(selector){assert.equal(selector,'dialog');return drawer;}};
 assert.equal(out.exports.drawerForTrigger({getAttribute(){return null;},closest(){return legacy;}},doc),drawer);
 assert.equal(out.exports.drawerForTrigger({getAttribute(){return null;},closest(){return null;}},doc),undefined);
});
