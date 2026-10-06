// Optional browser-script checks: node --test internal/web/ui_test.cjs
const assert = require('node:assert/strict');
const {readFileSync} = require('node:fs');
const {join} = require('node:path');
const {test} = require('node:test');
const {runInNewContext} = require('node:vm');

function runScript(name, globals) {
  runInNewContext(readFileSync(join(__dirname, 'static', name), 'utf8'), globals);
}

test('theme buttons restore, switch and persist settings independently', () => {
  class Button extends EventTarget {
    constructor(dataset = {}) { super(); this.dataset = dataset; this.disabled = true; this.attributes = {}; }
    setAttribute(name, value) { this.attributes[name] = value; }
  }
  const theme = new Button();
  const contrasts = ['default', 'medium', 'high'].map(appContrast => new Button({appContrast}));
  const icons = ['light', 'dark'].map(themeIcon => ({dataset: {themeIcon}}));
  const root = {dataset: {}};
  let saved = JSON.stringify({theme: 'dark', contrast: 'medium'});
  let ready;
  runScript('theme.js', {
    document: {
      documentElement: root,
      addEventListener: (_, listener) => { ready = listener; },
      querySelector: () => theme,
      querySelectorAll: selector => selector === '[data-app-contrast]' ? contrasts : icons,
    },
    localStorage: {getItem: () => saved, setItem: (_, value) => { saved = value; }},
  });
  assert.equal(root.dataset.theme, 'dark');
  ready();
  assert.equal(theme.attributes['aria-pressed'], 'true');
  assert.equal(contrasts[1].attributes['aria-pressed'], 'true');
  assert.equal(icons[0].hidden, true);
  theme.dispatchEvent(new Event('click'));
  contrasts[2].dispatchEvent(new Event('click'));
  assert.deepEqual(JSON.parse(saved), {theme: 'light', contrast: 'high'});
  assert.equal(theme.attributes['aria-pressed'], 'false');
  assert.equal(icons[1].hidden, true);
  assert.equal(contrasts[1].attributes['aria-pressed'], 'false');
  assert.equal(contrasts[2].attributes['aria-pressed'], 'true');
  assert.equal(theme.disabled, false);
});

test('only backdrop clicks cancel and close a modal', () => {
  const dialog = new EventTarget();
  dialog.getBoundingClientRect = () => ({left: 100, right: 300, top: 100, bottom: 300});
  let closed = 0;
  let cancelled = 0;
  dialog.close = () => { closed++; };
  dialog.addEventListener('cancel', () => { cancelled++; });
  runScript('token-table.js', {
    Event,
    document: {querySelectorAll: selector => selector === 'dialog' ? [dialog] : []},
  });
  const click = (x, y) => dialog.dispatchEvent(Object.assign(new Event('click'), {clientX: x, clientY: y}));
  click(150, 150);
  assert.equal(closed, 0);
  click(50, 150);
  assert.equal(closed, 1);
  assert.equal(cancelled, 1);
  dialog.addEventListener('cancel', event => event.preventDefault());
  click(350, 150);
  assert.equal(closed, 1);
});
