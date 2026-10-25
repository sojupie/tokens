const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');

function harness(type = 'radio', file = 'segmented-buttons.js') {
  const handlers = [];
  let clicks = 0;
  const input = {
    type, disabled: false, unavailable: false,
    matches(selector) { return selector === ':disabled' ? this.disabled : ['radio', 'checkbox'].includes(type); },
    closest() { return this.unavailable ? {} : null; },
    click() { clicks++; },
  };
  runInNewContext(readFileSync(`${__dirname}/behaviors/${file}`, 'utf8'), {
    document: { addEventListener(type, handler) { assert.equal(type, 'keydown'); handlers.push(handler); } },
  });
  return { input, get clicks() { return clicks; }, key(key = 'Enter', options = {}) {
    const event = { key, target: input, repeat: false, defaultPrevented: false, isComposing: false, preventDefault() { this.defaultPrevented = true; }, ...options };
    for (const handler of handlers) handler(event);
    return event.defaultPrevented;
  } };
}
for (const type of ['radio', 'checkbox']) {
  test(`Enter invokes native ${type} activation exactly once`, () => {
    const h = harness(type);
    assert.equal(h.key(), true);
    assert.equal(h.clicks, 1);
    assert.equal(h.key('Enter', { repeat: true }), true);
    assert.equal(h.clicks, 1);
  });
}
test('Space, arrows and Tab retain native behavior', () => {
  const h = harness();
  for (const key of [' ', 'ArrowRight', 'Tab']) assert.equal(h.key(key), false);
  assert.equal(h.clicks, 0);
});
test('disabled and unavailable segments do not activate', () => {
  const h = harness();
  h.input.disabled = true; assert.equal(h.key(), false);
  h.input.disabled = false; h.input.unavailable = true; assert.equal(h.key(), false);
  assert.equal(h.clicks, 0);
});
test('handled events and composition do not activate segments', () => {
  const h = harness();
  h.key('Enter', { defaultPrevented: true });
  assert.equal(h.key('Enter', { isComposing: true }), false);
  assert.equal(h.clicks, 0);
});
test('other controls and non-element event targets are untouched', () => {
  const h = harness('text');
  assert.equal(h.key(), false);
  assert.equal(h.key('Enter', { target: {} }), false);
  assert.equal(h.clicks, 0);
});

// Checkbox segments have no native directional navigation.
test('multi-select arrows move focus without changing selection', () => {
 const handlers=[],items=[];let focused;
 const group={querySelectorAll:()=>items};
 for(let index=0;index<3;index++)items.push({type:'checkbox',checked:index===1,
  matches:selector=>selector===':disabled'?false:true,
  closest:selector=>selector==='.md-segmented-buttons'?group:null,
  getClientRects:()=>[{}],focus(){focused=this;},click(){throw Error('arrow must not activate');}});
 const globals={document:{addEventListener:(type,handler)=>handlers.push(handler)},getComputedStyle:()=>({direction:'ltr',visibility:'visible'})};
 runInNewContext(readFileSync(`${__dirname}/behaviors/utils.js`,'utf8'),globals);
 runInNewContext(readFileSync(`${__dirname}/behaviors/segmented-buttons.js`,'utf8'),globals);
 handlers[0]({target:items[0],key:'ArrowRight',preventDefault(){}});
 assert.equal(focused,items[1]);assert.deepEqual(items.map(i=>i.checked),[false,true,false]);
 handlers[0]({target:items[0],key:'ArrowLeft',preventDefault(){}});assert.equal(focused,items[2]);
 globals.getComputedStyle=()=>({direction:'rtl',visibility:'visible'});
 handlers[0]({target:items[0],key:'ArrowLeft',preventDefault(){}});assert.equal(focused,items[1]);
});

test('switch Enter activation uses the native checkbox without repeats',()=>{
 const h=harness('checkbox','switches.js');assert.equal(h.key(),true);assert.equal(h.clicks,1);
 assert.equal(h.key('Enter',{repeat:true}),true);assert.equal(h.clicks,1);
 assert.equal(h.key(' '),false);h.input.disabled=true;assert.equal(h.key(),false);assert.equal(h.clicks,1);
});
