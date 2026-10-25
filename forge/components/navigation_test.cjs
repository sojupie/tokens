const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');

// Exercise the actual destination key handler; browser checks cover DOM visibility
// and whether focus really moves on the rendered navigation components.
function navigation(options = {}) {
  const listeners = [];
  let active;
  const group = {
    classList: { contains: name => name === 'md-navigation-bar' && !!options.horizontal },
    querySelectorAll: () => items,
    direction: options.rtl ? 'rtl' : 'ltr',
  };
  const items = (options.items || [{}, {}, {}, {}, { hidden: true }, { hidden: true }]).map(flags => ({
    ...flags,
    closest(selector) {
      if (selector === '[data-navigation-item]') return this;
      if (selector === '[data-navigation-group]') return group;
      if (selector === '[inert]') return this.inert ? {} : null;
      return null;
    },
    getClientRects() { return this.hidden ? [] : [{}]; },
    matches() { return !!(this.disabled || this.ariaDisabled); },
    focus() { active = this; },
  }));
  const globals = {
    document: { readyState: 'loading', addEventListener(type, listener) {
      if (type === 'keydown') listeners.push(listener);
    } },
    ResizeObserver: class { observe() {} },
    getComputedStyle: element => ({ direction: element.direction || 'ltr', visibility: element.visibility || 'visible' }),
  };
  runInNewContext(readFileSync(`${__dirname}/behaviors/utils.js`, 'utf8'), globals);
  runInNewContext(readFileSync(`${__dirname}/behaviors/navigation.js`, 'utf8'), globals);
  return {
    key(index, key) {
      active = items[index];
      let prevented = false;
      listeners[0]({ target: active, key, preventDefault() { prevented = true; } });
      return { index: items.indexOf(active), prevented };
    },
  };
}

test('collapsed rail wraps visible destinations and Home/End exclude hidden items', () => {
  const nav = navigation();
  assert.equal(nav.key(3, 'ArrowDown').index, 0);
  assert.equal(nav.key(0, 'ArrowUp').index, 3);
  assert.equal(nav.key(0, 'End').index, 3);
  assert.equal(nav.key(3, 'Home').index, 0);
});

test('expanded rail includes newly visible secondary destinations', () => {
  const nav = navigation({ items: Array.from({ length: 6 }, () => ({})) });
  assert.equal(nav.key(3, 'ArrowDown').index, 4);
  assert.equal(nav.key(0, 'End').index, 5);
});

test('destination traversal skips disabled, inert and visibility-hidden controls', () => {
  const nav = navigation({ items: [{}, { disabled: true }, { ariaDisabled: true }, { inert: true }, { visibility: 'hidden' }, {}] });
  assert.equal(nav.key(0, 'ArrowDown').index, 5);
  assert.equal(nav.key(5, 'ArrowUp').index, 0);
});

test('horizontal RTL navigation retains direction and skips unavailable destinations', () => {
  const nav = navigation({ horizontal: true, rtl: true, items: [{}, { hidden: true }, {}, {}] });
  assert.equal(nav.key(0, 'ArrowLeft').index, 2);
  assert.equal(nav.key(0, 'ArrowRight').index, 3);
  assert.deepEqual(nav.key(0, 'ArrowDown'), { index: 0, prevented: false });
});

test('unavailable event targets do not consume keys or move focus', () => {
  const nav = navigation({ items: [{ disabled: true }] });
  assert.deepEqual(nav.key(0, 'Home'), { index: 0, prevented: false });
});
