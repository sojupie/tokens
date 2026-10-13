// Optional browser-script checks: node --test internal/web/ui_test.cjs
const assert = require('node:assert/strict');
const {readFileSync} = require('node:fs');
const {join} = require('node:path');
const {test} = require('node:test');
const {runInNewContext} = require('node:vm');

function runScript(name, globals) {
  // The inline initialization runs before the deferred theme controls in HTML.
  if (name === 'theme.js') runInNewContext(readFileSync(join(__dirname, 'static', 'theme-init.js'), 'utf8'), globals);
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

test('WebMCP uses the collection path for both index and token downloads', async () => {
  const tools = [];
  const requests = [];
  runScript('webmcp.js', {
    document: {currentScript: {src: 'https://tokencatalog.dev/md3/static/webmcp.js?v=1'}, modelContext: {registerTool(tool) { tools.push(tool); }}},
    navigator: {}, window: {addEventListener() {}},
    location: {origin: 'https://tokencatalog.dev'}, AbortController, URL, console,
    fetch: async path => {
      requests.push(path);
      return {ok: true, json: async () => path.endsWith('/api/v1/subjects') ? {subjects: [{kind: 'component', slug: 'snackbar', jsonUrl: 'https://tokencatalog.dev/md3/downloads/component/snackbar.json?scope=all'}]} : {tokens: []}};
    },
  });
  await tools.find(tool => tool.name === 'get_tokens').execute({kind: 'component', slug: 'snackbar'});
  assert.deepEqual(requests, ['/md3/api/v1/subjects', '/md3/downloads/component/snackbar.json?scope=all']);
});

function prefetchHarness({speculation = true, native = true, connection = {}} = {}) {
  const listeners = {};
  const scheduled = new Map();
  const appended = [];
  let now = 0;
  let next = 0;
  const document = {
    currentScript: {src: 'https://tokens.test/md3/static/prefetch.js'},
    visibilityState: 'visible',
    addEventListener: (name, listener) => { listeners[name] = listener; },
    createElement: tag => ({tag, relList: {supports: () => native}}),
    head: {append: element => appended.push(element)},
  };
  runScript('prefetch.js', {
    document,
    navigator: {connection},
    location: new URL('https://tokens.test/md3/'),
    URL,
    HTMLScriptElement: {supports: () => speculation},
    setTimeout: (callback, delay) => { const id = ++next; scheduled.set(id, {callback, time: now + delay}); return id; },
    clearTimeout: id => scheduled.delete(id),
    window: {addEventListener: (name, listener) => { listeners[name] = listener; }},
  });
  function link(path, options = {}) {
    const anchor = {
      href: new URL(path, 'https://tokens.test').href,
      target: options.target || '',
      isConnected: true,
      hasAttribute: name => name === 'download' && options.download,
      closest: selector => selector === 'a[href]' ? anchor : options.htmx ? anchor : null,
      contains: node => node === anchor,
    };
    return anchor;
  }
  return {
    appended, document, connection, link,
    event: (name, target, extra = {}) => listeners[name]?.({target, pointerType: 'mouse', ...extra}),
    advance: milliseconds => {
      now += milliseconds;
      for (const [id, job] of scheduled) {
        if (job.time <= now) { scheduled.delete(id); job.callback(); }
      }
    },
  };
}

test('hover prefetch waits 65 ms, cancels brief hovers, and deduplicates destinations', () => {
  const h = prefetchHarness();
  const link = h.link('/md3/components/buttons');
  h.event('pointerover', link);
  h.advance(1);
  h.event('pointerout', link, {relatedTarget: null});
  h.advance(500);
  assert.equal(h.appended.length, 0);
  h.event('pointerover', link);
  h.advance(64);
  assert.equal(h.appended.length, 0);
  h.advance(1);
  assert.deepEqual(JSON.parse(h.appended[0].textContent).prefetch[0].urls, ['https://tokens.test/md3/components/buttons']);
  h.event('pointerover', link);
  h.advance(500);
  assert.equal(h.appended.length, 1);
});

test('prefetch skips external, download, API, fragment, HTMX and new-tab links', () => {
  const h = prefetchHarness();
  for (const link of [
    h.link('https://github.com/sojupie/tokens'),
    h.link('/md3/downloads/tokens.json'),
    h.link('/md3/api/v1/subjects'),
    h.link('/other/components/buttons'),
    h.link('/md3/components/buttons#table'),
    h.link('/md3/components/buttons', {download: true}),
    h.link('/md3/components/buttons', {htmx: true}),
    h.link('/md3/components/buttons', {target: '_blank'}),
  ]) {
    h.event('pointerover', link);
    h.advance(65);
  }
  assert.equal(h.appended.length, 0);
});

test('prefetch handles keyboard focus and native hints, with an eight-page limit', () => {
  const h = prefetchHarness({speculation: false});
  for (let index = 0; index < 10; index++) {
    h.event('focusin', h.link(`/md3/components/subject-${index}`));
    h.advance(65);
  }
  assert.equal(h.appended.length, 8);
  assert.equal(h.appended[0].rel, 'prefetch');
  assert.equal(h.appended[0].as, 'document');
});

test('prefetch respects data saving, slow connections, visibility and touch input', () => {
  for (const connection of [{saveData: true}, {effectiveType: '2g'}, {effectiveType: 'slow-2g'}]) {
    const h = prefetchHarness({connection});
    h.event('pointerover', h.link('/md3/styles'));
    h.advance(65);
    assert.equal(h.appended.length, 0);
  }
  const h = prefetchHarness();
  const link = h.link('/md3/styles');
  h.event('pointerover', link, {pointerType: 'touch'});
  h.advance(65);
  assert.equal(h.appended.length, 0);
  h.event('pointerover', link);
  h.connection.saveData = true;
  h.advance(65);
  assert.equal(h.appended.length, 0);
  h.connection.saveData = false;
  h.event('pointerover', link);
  h.document.visibilityState = 'hidden';
  h.event('visibilitychange');
  h.advance(65);
  assert.equal(h.appended.length, 0);
});

test('prefetch ignores event targets without an element closest method', () => {
  const h = prefetchHarness();
  assert.doesNotThrow(() => h.event('pointerover', h.document));
  assert.doesNotThrow(() => h.event('focusin', h.document));
  h.advance(65);
  assert.equal(h.appended.length, 0);
});

test('prefetch keeps the hover timer when moving between children of a link', () => {
  const h = prefetchHarness();
  const link = h.link('/md3/components/buttons');
  const child = {closest: () => link};
  link.contains = node => node === link || node === child;
  h.event('pointerover', link);
  h.advance(40);
  h.event('pointerout', link, {relatedTarget: child});
  h.event('pointerover', child);
  h.advance(25);
  assert.equal(h.appended.length, 1);
});

test('prefetch checks links again after the hover delay', () => {
  const h = prefetchHarness();
  const removed = h.link('/md3/components/buttons');
  h.event('pointerover', removed);
  removed.isConnected = false;
  h.advance(65);
  const changed = h.link('/md3/styles');
  h.event('pointerover', changed);
  changed.href = 'https://tokens.test/md3/downloads/tokens.json';
  h.advance(65);
  assert.equal(h.appended.length, 0);
});

test('agent handoff copies visible text and selects it when clipboard access fails', async () => {
  const field = {textContent: 'Use https://catalog.test/md3/mcp', focus() { this.focused = true; }};
  const status = {textContent: ''};
  const button = {hidden: true, dataset: {copyTarget: 'agent-mcp'}, closest: () => ({querySelector: () => status}), addEventListener(name, callback) { this.click = callback; }};
  let copied;
  let selected;
  let fail = false;
  const globals = {
    document: {querySelectorAll: () => [button], getElementById: () => field},
    navigator: {clipboard: {writeText: async text => { if (fail) throw new Error('denied'); copied = text; }}},
    window: {getSelection: () => ({selectAllChildren: element => { selected = element; }})},
  };
  runScript('agent-handoff.js', globals);
  assert.equal(button.hidden, false);
  await button.click();
  assert.equal(copied, field.textContent);
  assert.equal(status.textContent, 'Copied.');
  fail = true;
  await button.click();
  assert.equal(field.focused, true);
  assert.equal(selected, field);
  assert.match(status.textContent, /^Copy failed/);
  button.hidden = true;
  globals.navigator = {};
  runScript('agent-handoff.js', globals);
  assert.equal(button.hidden, true);
});
