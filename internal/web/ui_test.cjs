// Optional browser-script checks: node --test internal/web/ui_test.cjs
const assert = require('node:assert/strict');
const {readFileSync} = require('node:fs');
const {join} = require('node:path');
const {test} = require('node:test');
const {runInNewContext} = require('node:vm');

function runScript(name, globals) {
  runInNewContext(readFileSync(join(__dirname, 'static', name), 'utf8'), globals);
}

test('collection download buttons select the matching scope', () => {
  const buttons = ['all', 'components', 'styles'].map(scope => {
    const button = new EventTarget();
    button.dataset = {siteDownloadOpen: scope};
    return button;
  });
  let selected;
  let opened = 0;
  const dialog = {
    querySelector: selector => ({set checked(value) { if (value) selected = selector; }}),
    showModal: () => { opened++; },
  };
  runScript('token-table.js', {
    document: {
      querySelectorAll: selector => selector === '[data-site-download-open]' ? buttons : [],
      querySelector: selector => { assert.equal(selector, '#tokens-site-download'); return dialog; },
    },
  });
  for (const button of buttons) {
    button.dispatchEvent(new Event('click'));
    assert.equal(selected, `input[value="${button.dataset.siteDownloadOpen}"]`);
  }
  assert.equal(opened, 3);
});

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

test('WebMCP registers usable read-only tools and cleans up', async () => {
  const tools = [];
  const events = new Map();
  let signal;
  const index = {subjects: [{kind: 'component', slug: 'snackbar', name: 'Snackbar', jsonUrl: 'https://tokens.example/downloads/component/snackbar.json?scope=all'}]};
  let changes = [{subject: 'snackbar'}, {subject: 'buttons'}];
  const globals = {
    document: {modelContext: {registerTool(tool, options) { tools.push(tool); signal = options.signal; return Promise.resolve(); }}},
    navigator: {},
    window: {addEventListener(name, callback) { events.set(name, callback); }},
    location: {origin: 'https://tokens.example'},
    AbortController, URL, console,
    fetch: async path => ({ok: true, json: async () => {
      if (path === '/api/v1/subjects') return structuredClone(index);
      if (path === '/api/v1/changes') return {changes};
      assert.equal(path, '/downloads/component/snackbar.json?scope=all');
      return {tokens: [{name: 'md.comp.snackbar.container.color'}]};
    }}),
  };
  runScript('webmcp.js', globals);
  assert.deepEqual(tools.map(tool => tool.name), ['list_subjects', 'get_tokens', 'list_changes']);
  assert.equal(JSON.parse((await tools[0].execute({query: 'snack'})).content[0].text).subjects.length, 1);
  assert.equal(JSON.parse((await tools[0].execute({query: 'missing'})).content[0].text).subjects.length, 0);
  assert.equal(JSON.parse((await tools[1].execute({kind: 'component', slug: 'snackbar'})).content[0].text).tokens.length, 1);
  await assert.rejects(tools[1].execute({kind: 'component', slug: '../missing'}), /Subject not found/);
  assert.equal(JSON.parse((await tools[2].execute({subject: 'snackbar'})).content[0].text).changes.length, 1);
  changes = null;
  assert.deepEqual(JSON.parse((await tools[2].execute({subject: 'snackbar'})).content[0].text).changes, []);
  assert.deepEqual(JSON.parse((await tools[2].execute({})).content[0].text).changes, []);
  events.get('pagehide')({persisted: true});
  assert.equal(signal.aborted, false);
  events.get('pagehide')({persisted: false});
  assert.equal(signal.aborted, true);
});

test('WebMCP is optional and supports the navigator fallback', () => {
  runScript('webmcp.js', {document: {}, navigator: {}});
  let count = 0;
  runScript('webmcp.js', {
    document: {}, navigator: {modelContext: {registerTool() { count++; }}},
    AbortController, console, window: {addEventListener() {}},
  });
  assert.equal(count, 3);
});
