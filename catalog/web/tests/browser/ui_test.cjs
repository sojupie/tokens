const { existsSync } = require('node:fs');
// Optional browser-script checks: npm run test:ui
const assert = require('node:assert/strict');
const {readFileSync} = require('node:fs');
const {join} = require('node:path');
const {test} = require('node:test');
const {runInNewContext} = require('node:vm');

function runScript(name, globals) {
  globals.document.addEventListener ||= () => {};
  globals.AbortController ||= AbortController;
  globals.ResizeObserver ||= class {observe(){} unobserve(){}};
  // The inline initialization runs before the deferred theme controls in HTML.
  if (name === 'theme.js') runInNewContext(readFileSync(scriptPath('theme-init.js'), 'utf8'), globals);
  runInNewContext(readFileSync(scriptPath(name), 'utf8'), globals);
}

test('behavior audit compares declarations without changing inert discovery or eager script handling', () => {
  const globals={module:{exports:{}}};
  runInNewContext(readFileSync(scriptPath('asset-loader.js'),'utf8'),globals);
  const {auditBehaviors}=globals.module.exports;
  const previewNode={closest:()=>({})},thumbnail={closest:()=>({})},visible={closest:()=>null};
  const matches={'.md-button-group':[previewNode,thumbnail],'.md-carousel':[thumbnail],'.md-loading':[visible],'.md-progress':[thumbnail]};
  let declarations=['runtime/navigation.js','runtime/button-groups.js','loading.js','progress.js','carousel.js'];
  const root={querySelector:()=>({contains:node=>node===previewNode}),querySelectorAll:selector=>{
    if(selector==='script[data-component-behaviors]')return [{textContent:JSON.stringify(declarations)}];
    if(selector==='script[src]:not([data-component-script])')return [{getAttribute:()=>'/static/runtime/navigation.js?v=hash'}];
    return matches[selector]||[];
  }};
  let audit=JSON.parse(JSON.stringify(auditBehaviors(root)));
  assert.deepEqual(audit.undeclared,[]);
  assert.deepEqual(audit.declarationOnly,['carousel.js'],'inert thumbnails declare ownership without activating behavior');
  assert(!audit.declared.includes('runtime/navigation.js'),'core scripts are already delivered');
  assert(audit.discovered.includes('progress.js'),'preserve the current progress exception');
  declarations=[];
  audit=JSON.parse(JSON.stringify(auditBehaviors(root)));
  assert.deepEqual(audit.undeclared,['loading.js','progress.js','runtime/button-groups.js']);
});

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
  const sources = ['light-high','dark-medium','dark-high'].map(imageAppearance => ({dataset:{imageAppearance},media:'not all'}));
  const root = {dataset: {}};
  let saved = JSON.stringify({theme: 'dark', contrast: 'medium'});
  let ready;
  runScript('theme.js', {
    document: {
      documentElement: root,
      addEventListener: (_, listener) => { ready = listener; },
      querySelector: () => theme,
      querySelectorAll: selector => selector === '[data-app-contrast]' ? contrasts : selector === 'source[data-image-appearance]' ? sources : icons,
    },
    localStorage: {getItem: () => saved, setItem: (_, value) => { saved = value; }},
  });
  assert.equal(root.dataset.theme, 'dark');
  ready();
  assert.equal(theme.attributes['aria-pressed'], 'true');
  assert.equal(contrasts[1].attributes['aria-pressed'], 'true');
  assert.equal(sources[1].media, 'all');
  assert.equal(icons[0].hidden, true);
  theme.dispatchEvent(new Event('click'));
  contrasts[2].dispatchEvent(new Event('click'));
  assert.deepEqual(JSON.parse(saved), {theme: 'light', contrast: 'high'});
  assert.equal(sources[0].media, 'all');assert.equal(sources[1].media, 'not all');
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
    document: {querySelectorAll: selector => selector === 'dialog:not(.md-sheet)' ? [dialog] : []},
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

function prefetchHarness({connection = {}, fail = false} = {}) {
  const listeners = {};
  const scheduled = new Map();
  const appended = [];
  let extension;
  let now = 0;
  let next = 0;
  const document = {
    currentScript: {src: 'https://tokens.test/md3/static/prefetch.js'},
    visibilityState: 'visible',
    body: {},
    addEventListener: (name, listener) => { listeners[name] = listener; },
  };
  const internal={
    attributeValue:(elt,name)=>elt.attrs[name], parseTriggerSpecs:name=>[{name}],
    determineMethodAndAction:()=>({method:'GET'}),
    createRequestContext:elt=>({target:document.body,request:{action:elt.href,method:'GET',headers:{}}}),
    collectFormData:()=>new URLSearchParams(),getAttributeObject:()=>undefined,
  };
  const globals={document,navigator:{connection},location:new URL('https://tokens.test/md3/'),URL,URLSearchParams,CustomEvent,
    htmx:{config:{preload:{autoBoost:false}},registerExtension(name,value){extension=value;value.init(internal)}},
    fetch:(url,request)=>{const response={status:200};appended.push({url,request,response});return fail?Promise.reject(new Error("offline")):Promise.resolve(response)},
    Date:{now:()=>now},
    setTimeout:(callback,delay)=>{const id=++next;scheduled.set(id,{callback,time:now+delay});return id},
    clearTimeout:id=>scheduled.delete(id),
    window:{addEventListener:(name,listener)=>{listeners[name]=listener}},
  };
  runScript('hx-preload.min.js',globals);runScript('prefetch.js',globals);
  function link(path, options = {}) {
    const anchor = {
      href: new URL(path, 'https://tokens.test').href,
      target: options.target || '',
      isConnected: true,attrs:{},_htmx:{boosted:true},listeners:{},
      matches:()=>true,setAttribute(name,value){this.attrs[name]=value},
      addEventListener(name,fn){this.listeners[name]=fn},removeEventListener(name){delete this.listeners[name]},
      dispatchEvent(event){this.listeners[event.type]?.(event);return true},
      hasAttribute: name => name === 'download' && options.download,
      closest: selector => selector === 'a[href]' ? anchor : (options.htmx || options.unboosted) ? anchor : null,
      contains: node => node === anchor,
    };
    listeners['htmx:beforeProcessNode']?.({target:anchor,detail:{elt:anchor}});
    extension.htmx_after_init(anchor);
    return anchor;
  }
  return {
    appended, document, connection, link, extension, location:globals.location,
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
  assert.equal(h.appended[0].url, 'https://tokens.test/md3/components/buttons');
  assert.equal(h.appended[0].request.headers['HX-Request-Type'],'full');
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

test('prefetch handles keyboard focus through the upstream extension, with an eight-page limit', () => {
  const h = prefetchHarness();
  for (let index = 0; index < 10; index++) {
    h.event('focusin', h.link(`/md3/components/subject-${index}`));
    h.advance(65);
  }
  assert.equal(h.appended.length, 8);
  assert.equal(h.appended[0].url,'https://tokens.test/md3/components/subject-0');
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
  const button = {hidden: true, dataset: {copyTarget: 'agent-mcp'}, closest: selector => selector === '.tokens-card-content' ? {querySelector: () => status} : null, addEventListener(name, callback) { this.click = callback; }};
  let copied;
  let selected;
  let fail = false;
  const handlers={},timers=new Map();let timerID=0;
  const globals = {
    setTimeout(fn,ms){assert.equal(ms,1500);timers.set(++timerID,fn);return timerID;},
    clearTimeout(id){timers.delete(id);},
    document: {querySelectorAll: () => [button], getElementById: () => field, addEventListener:(name,fn)=>handlers[name]=fn},
    navigator: {clipboard: {writeText: async text => { if (fail) throw new Error('denied'); copied = text; }}},
    window: {getSelection: () => ({selectAllChildren: element => { selected = element; }})},
  };
  runScript('agent-handoff.js', globals);
  assert.equal(button.hidden, false);
  await button.click();
  assert.equal(copied, field.textContent);
  assert.equal(status.textContent, 'Copied.');
  const first=timerID;
  await button.click();assert.equal(timers.has(first),false,'copying again replaces the timeout');
  timers.get(timerID)();assert.equal(status.textContent,'','success clears after 1.5 seconds');
  await button.click();handlers.pointerdown();assert.equal(status.textContent,'','another pointer interaction clears success');
  await button.click();handlers.keydown();assert.equal(status.textContent,'','keyboard interaction also clears success');
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

test('table modes preserve preview preference and folder state, with matching action icons', async () => {
  const classes = new Set();
  const listeners = {};
  const groups = [{open: true}, {open: false}];
  const controls = Object.fromEntries(['previews', 'view', 'expand'].map(action => {
    const attributes = {};
    const icons = {on: {hidden: false}, off: {hidden: true}};
    const tooltipText = {}; const tooltip = {querySelector: () => tooltipText};
    return [action, {
      dataset: {action}, attributes, icons, tooltip, tooltipText, closest: () => tooltip,
      setAttribute: (key, value) => { attributes[key] = value; },
      getAttribute: key => attributes[key],
      querySelector: selector => icons[selector.includes('"off"') ? 'off' : 'on'],
    }];
  }));
  const menu = {addEventListener() {}, matches: () => false};
  const clicks = [];
  const browser = {
    classList: {
      contains: value => classes.has(value),
      toggle: value => { if (classes.has(value)) { classes.delete(value); return false; } classes.add(value); return true; },
    },
    querySelector: selector => {
      if (selector === 'form') return {addEventListener(){}};
      if (selector === '.tokens-context-dialog') return null;
      if (selector === '.tokens-token-set') return {};
      if (selector === '.tokens-token-set-menu') return menu;
      return controls[selector.match(/data-action="([^"]+)"/)[1]];
    },
    querySelectorAll: selector => selector === '.tokens-token-folder' ? groups : [],
    addEventListener: (name, handler) => {
      if (name === 'click') clicks.push(handler);
      else listeners[name] = handler;
    },
  };
  let copied, failClipboard = false, copyField, copyFallback = true, clearCopyStatus;
  const clipboard = {writeText: async value => { if (failClipboard) throw Error('denied'); copied = value; }};
  const navigator = {clipboard};
  runScript('token-table.js', {
    document: {
      querySelectorAll: selector => selector === '.tokens-token-browser' ? [browser] : [],
      createElement: () => ({style:{},focus(){},select(){},remove(){copyField=undefined;}}),
      execCommand: command => { assert.equal(command,'copy'); copied=copyField.value; return copyFallback; },
    },
    addEventListener() {}, navigator,
    setTimeout: (fn, delay) => { assert.equal(delay,1500); clearCopyStatus=fn; return 1; },
    clearTimeout() {},
  });
  const dispatchClick = closest => Promise.all(clicks.map(handler => handler({target: {closest}})));
  const click = action => dispatchClick(selector => selector === 'button[data-action]' ? controls[action] : null);
  await click('previews');
  assert.equal(classes.has('tokens-no-previews'), true);
  assert.equal(controls.previews.icons.off.hidden, false);
  await click('view');
  assert.equal(classes.has('tokens-grid-view'), true);
  assert.equal(controls.previews.hidden, true);
  assert.equal(controls.view.icons.off.hidden, false);
  assert.equal(controls.view.tooltipText.textContent, 'Show table view');
  assert.equal(controls.previews.tooltip.hidden, true);
  assert.deepEqual(groups.map(group => group.open), [true, false]);
  await click('view');
  assert.equal(controls.previews.hidden, false);
  assert.equal(classes.has('tokens-no-previews'), true);
  await click('expand');
  assert.deepEqual(groups.map(group => group.open), [true, true]);
  assert.equal(controls.expand.icons.off.hidden, false);
  await click('view');
  await click('expand');
  assert.equal(classes.has('tokens-grid-view'), true);
  assert.deepEqual(groups.map(group => group.open), [false, false]);
  const status = {};
  const copy = {dataset: {copyReference: 'md.sys.color.primary'}, focus(){}, closest: () => ({querySelector: () => status,append: field => {copyField=field;}})};
  const copyReference = () => dispatchClick(selector => selector === '[data-copy-reference]' ? copy : null);
  await copyReference();
  assert.equal(copied, 'md.sys.color.primary');
  assert.equal(status.textContent, 'Copied.');
  clearCopyStatus();
  assert.equal(status.textContent, '');
  failClipboard = true;
  copied = undefined;
  await copyReference();
  assert.equal(copied,'md.sys.color.primary');
  assert.equal(status.textContent,'Copied.');
  assert.equal(copyField,undefined);
  navigator.clipboard = undefined;
  copied = undefined;
  copy.dataset.copyReference = 'md.ref.palette.primary40';
  await copyReference();
  assert.equal(copied,'md.ref.palette.primary40');
  assert.equal(status.textContent,'Copied.');
  copyFallback = false;
  await copyReference();
  assert.equal(status.textContent,'Could not copy.');
  assert.equal(copyField,undefined);
});

test('history restoration binds replacement download controls once', () => {
  const events = new Map();
  let buttons = [];
  let fallbackStyles = [];
  let opened = 0;
  const dialog = {querySelector: () => ({checked: false}), showModal: () => { opened++; }};
  runScript('token-table.js', {document: {
    addEventListener: (name, callback) => events.set(name, callback),
    querySelectorAll: selector => selector === '[data-site-download-open]' ? buttons : selector === 'noscript style' ? fallbackStyles : [],
    querySelector: () => dialog,
  }});
  const restored = new EventTarget(); restored.dataset = {siteDownloadOpen: 'all'};
  buttons = [restored];
  fallbackStyles = [{remove: () => { fallbackStyles = []; }}];
  events.get('htmx:afterSwap')();
  events.get('htmx:historyRestore')();
  assert.equal(fallbackStyles.length, 0);
  restored.dispatchEvent(new Event('click'));
  assert.equal(opened, 1);
  const replacement = new EventTarget(); replacement.dataset = {siteDownloadOpen: 'styles'};
  buttons = [replacement];
  events.get('htmx:historyRestore')();
  replacement.dispatchEvent(new Event('click'));
  assert.equal(opened, 2);
});

test('theme controls survive body restoration without duplicate click handlers', () => {
  class Button extends EventTarget {
    constructor(dataset = {}) { super(); this.dataset = dataset; this.disabled = true; this.attributes = {}; }
    setAttribute(name, value) { this.attributes[name] = value; }
  }
  const events = new Map(), root = {dataset: {}};
  let theme = new Button(), contrasts = ['default','medium','high'].map(appContrast => new Button({appContrast}));
  runScript('theme.js', {localStorage: {getItem: () => null, setItem() {}}, document: {
    documentElement: root,
    addEventListener: (name, callback) => events.set(name, callback),
    querySelector: () => theme,
    querySelectorAll: selector => selector === '[data-app-contrast]' ? contrasts : [],
  }});
  events.get('DOMContentLoaded')();
  theme = new Button(); contrasts = ['default','medium','high'].map(appContrast => new Button({appContrast}));
  events.get('htmx:afterSwap')();
  events.get('htmx:historyRestore')();
  theme.dispatchEvent(new Event('click'));
  assert.equal(root.dataset.theme, 'dark');
  contrasts[2].dispatchEvent(new Event('click'));
  assert.equal(root.dataset.contrast, 'high');
  assert.equal(contrasts[2].attributes['aria-pressed'], 'true');
  assert.equal(theme.disabled, false);
});

test('export disclosures receive local open state before the fragment is inserted', () => {
 const events={},live=[{open:false,dataset:{}},{open:true,dataset:{}}],incoming=[{open:true},{open:false}];
 const panel={querySelectorAll:()=>live,contains:()=>false};
 const sourceElement={id:'tokens-export-form',closest:()=>panel};
 runScript('export.js',{document:{readyState:'loading',addEventListener:(name,fn)=>events[name]=fn},location:{search:''}});
 events['htmx:before:swap']({detail:{ctx:{sourceElement},tasks:[{target:{id:'tokens-export-fields'},fragment:{querySelectorAll:()=>incoming}}]}});
 assert.deepEqual(incoming.map(detail=>detail.open),[false,true]);
 // Changes made while the request was pending are read immediately before insertion.
 live[1].dataset.expanding='false';
 events['htmx:before:swap']({detail:{ctx:{sourceElement},tasks:[{target:{id:'tokens-export-fields'},fragment:{querySelectorAll:()=>incoming}}]}});
 assert.deepEqual(incoming.map(detail=>detail.open),[false,false]);
});

test('export changes invalidate download and retain native submission and cancellation', () => {
 const events={},requests=[],timers=[];
 function panel(){
  const fields=Object.fromEntries(Object.entries({kind:'component',slug:'buttons',set:'filled',q:'icon',content:'tokens',scope:'view',format:'css','context.Theme':'Dark'}).map(([name,value])=>[name,{name,value,type:'hidden'}]));
  fields._download={name:'_download',value:''};
  const form={id:'tokens-export-form',fields,listeners:{},addEventListener(name,fn){this.listeners[name]=fn;},querySelector:()=>fields._download,setAttribute(){},querySelectorAll:()=>Object.values(fields),closest:()=>result};
  const nodes={'form':form,'[data-export-confirmation]':{},'#tokens-export-submit-pending':{busy:false,classList:{contains(){return current.nodes['#tokens-export-submit-pending'].busy;}}},'[data-export-counts]':{},'[data-export-error]':{},'#tokens-export-fields':{dataset:{exportReady:'true',exportLabel:'Download ZIP'}},'[data-export-submit]':{attributes:{},label:{},setAttribute(name,value){this.attributes[name]=value;},querySelector(){return this.label;}}};
  const listeners={};const result={dataset:{defaultScope:'subject'},nodes,querySelector:s=>nodes[s],querySelectorAll:()=>[],addEventListener:(name,fn)=>listeners[name]=fn,contains:()=>false,showModal(){this.open=true;},close(){this.open=false;listeners.close();}};return result;
 }
 let current;current=panel();const original=current;
 class FormData extends Map {constructor(form){super(Object.entries(form.fields).map(([name,input])=>[name,input.value]));}}
 const document={cookie:'',readyState:'complete',addEventListener:(name,fn)=>events[name]=fn,getElementById:()=>current,querySelector:()=>null};
 runScript('export.js',{document,setInterval:fn=>{timers.push(fn);return timers.length;},clearInterval(){},location:{search:''},URLSearchParams,FormData,htmx:{trigger:(form,event)=>requests.push({form,event})}});
 current=panel();events['htmx:historyRestore']();let focused=false;
 const trigger={dataset:{exportScope:'view'},focus(){focused=true;}};
 events.click({target:{closest:selector=>selector==='[data-export-open]'?trigger:null}});
 assert.equal(current.open,true);assert.equal(original.open,undefined);
 assert.equal(requests[0].event,'export-refresh');assert.equal(requests[0].form,current.nodes.form);
 assert.equal(current.nodes.form.fields.scope.value,'view');
 const ctx={sourceElement:current.nodes.form};
 events['htmx:before:request']({detail:{ctx}});
 assert.equal(current.nodes['[data-export-confirmation]'].disabled,true,'a changed selection is unconfirmed until the server returns an action');
 assert.equal(current.nodes['[data-export-submit]'].attributes['aria-busy'],'false','validation must not activate the purple loading appearance');
 events['htmx:error']({detail:{ctx}});
 assert.equal(current.nodes['[data-export-error]'].hidden,false);
 assert.match(current.nodes['[data-export-error]'].textContent,/Could not check/);
 current.nodes['[data-export-confirmation]'].disabled=false;
 const form=current.nodes.form,button=current.nodes['[data-export-submit]'];
 form.listeners.submit({preventDefault(){throw Error('native download was intercepted');}});
 assert.equal(form.target,'tokens-export-transfer');
 assert.equal(button.attributes['aria-busy'],'true');assert.equal(button.disabled,true);
 const token=form.fields._download.value;assert(token);
 let duplicate=false;form.listeners.submit({preventDefault(){duplicate=true;}});assert(duplicate);
 document.cookie=`tokens-download-${token}=ready`;timers.at(-1)();
 assert.equal(button.attributes['aria-busy'],'false');assert.equal(button.disabled,false);
 form.listeners.submit({preventDefault(){}});
 document.cookie=`tokens-download-${form.fields._download.value}=error`;timers.at(-1)();
 assert.match(current.nodes['[data-export-error]'].textContent,/Could not prepare/);
 assert.equal(button.attributes['aria-busy'],'false');
 current.close();assert.equal(focused,true);assert.equal(requests.at(-1).event,'htmx:abort');
 const source=readFileSync(scriptPath('export.js'),'utf8');
 assert(!source.includes('fetch('));assert(!source.includes('createObjectURL'),'native attachment must not be buffered in JavaScript');
});

test('rapid export changes retain draft values from temporarily disabled child groups',()=>{
 const events={},values=new Map();
 const form={querySelectorAll:()=>[
  {name:'format',value:'css',type:'checkbox',checked:true},
  {name:'format',value:'dtcg',type:'checkbox',checked:false},
  {name:'context.Platform',value:'Web',type:'radio',checked:true},
 ]};
 runScript('export.js',{document:{readyState:'loading',addEventListener:(name,fn)=>events[name]=fn},location:{search:''}});
 events['htmx:config:request']({detail:{ctx:{sourceElement:{closest:()=>form},request:{body:{append:(n,v)=>values.set(n,v)}}}}});
 assert.deepEqual([...values],[['format','css'],['context.Platform','Web']]);
});

test('mosaic artwork pauses for the control, reduced motion, and offscreen visibility', () => {
  const events={}, reduced={matches:false,addEventListener(_,handler){this.change=handler;},removeEventListener(){}},control={checked:false,addEventListener(_,handler){this.change=handler;},removeEventListener(){}};
  let paused=false,observe;
  const animationStates=[];const animationOptions=[],animationFrames=[];
  function animate(frames,options){animationOptions.push(options);animationFrames.push(frames);const animation={state:'running',pause(){this.state='paused';},play(){this.state='running';},cancel(){this.state='idle';}};animationStates.push(animation);return animation;}
  const artwork={animate,querySelectorAll:selector=>Array.from({length:selector==='.tokens-art-column'?3:9},()=>({animate})),parentElement:{querySelector:()=>control},toggleAttribute(name,value){assert.equal(name,'data-paused');paused=value;}};
  const document={readyState:'complete',timeline:{currentTime:1234},hidden:false,querySelector:()=>artwork,addEventListener(type,handler){events[type]=handler;},removeEventListener(){}};
  runScript('artwork.js',{document,getComputedStyle:()=>({getPropertyValue:key=>({'--_art-spring-damping':'.9','--_art-spring-stiffness':'300','--_art-turn-duration':'300ms','--_art-turn-easing':'cubic-bezier(.2,0,0,1)','--_art-shape-damping':'.9','--_art-shape-stiffness':'1400','--_art-corner-small':'8px','--_art-corner-large':'16px','--_art-corner-base':'28px','--_art-corner-soft':'48px'}[key])}),matchMedia:()=>reduced,IntersectionObserver:class{constructor(callback){observe=callback;}observe(){}disconnect(){}}});
  assert.equal(paused,false);
  assert.equal(animationStates.length,22);
  assert.ok(animationStates.every(a=>a.startTime===1234),'all tracks share the same scene clock');
  assert.ok(animationOptions.every(options=>!options.delay),'grid, corners and stripe turns have no stagger');
  animationFrames.slice(4,13).flat().forEach(frame=>{assert.ok([0,90].includes(parseFloat(frame['--_art-angle'])));assert.equal(frame.easing,'cubic-bezier(.2,0,0,1)','stripe turns use the non-overshooting token curve');});
  animationOptions.slice(4,13).forEach((rotation,index)=>{const layout=animationOptions[1+Math.floor(index/3)];assert.equal(rotation.duration,layout.duration);assert.equal(rotation.delay,layout.delay,'stripes start with their owning column resize');});
  [0,4,8,2,6,1,7,3,5].forEach((tile,slot)=>{const corner=animationOptions[13+slot],layout=animationOptions[1+Math.floor(tile/3)];assert.equal(corner.duration,layout.duration);assert.equal(corner.delay,layout.delay,'corners start with their owning column resize');});
  control.checked=true;control.change();assert.equal(paused,true);assert.ok(animationStates.every(a=>a.state==='paused'));
  control.checked=false;control.change();assert.equal(paused,false);
  reduced.matches=true;reduced.change();assert.equal(paused,true);
  reduced.matches=false;reduced.change();assert.equal(paused,false);
  observe([{isIntersecting:false}]);assert.equal(paused,true);
  observe([{isIntersecting:true}]);assert.equal(paused,false);
  document.hidden=true;events.visibilitychange();assert.equal(paused,true);
  events['htmx:afterSwap']();assert.equal(paused,true);
});

 test('hero transitions use the supplied spatial spring and hold after settling', () => {
  const module={exports:{}};
  runInNewContext(readFileSync(scriptPath('artwork.js'),'utf8'),{module});
  const {artworkSpring,artworkFrames}=module.exports;
  const slow=artworkSpring(.9,300),fast=artworkSpring(.9,1400);
  assert.ok(slow.duration>fast.duration);
  assert.ok(slow.duration<1000);
  assert.ok(Math.max(...slow.samples)>1,'spatial spring retains its small overshoot');
  assert.equal(slow.samples.at(-1),1,'finish exactly at the destination');
  assert.equal(artworkSpring(0,300),null,'missing tokens must not invent motion');
  const frames=artworkFrames([[1,2],[2,1]],'gridTemplateColumns',slow);
  assert.equal(frames[0].gridTemplateColumns,'1.000000fr 2.000000fr');
  const arrival=frames[slow.samples.length-1],hold=frames[slow.samples.length];
  assert.equal(arrival.gridTemplateColumns,hold.gridTemplateColumns);
  assert.ok(hold.offset-arrival.offset>.4,'settling is separate from the loop hold');
  assert.equal(frames.at(-1).offset,1);
 });

test('internal stripe turns hold on resize beats without a loop jump', () => {
  const module={exports:{}};
  runInNewContext(readFileSync(scriptPath('artwork.js'),'utf8'),{module});
  const {artworkSpring,artworkFrames}=module.exports;
  const spring={samples:[0,1],duration:300},frames=artworkFrames([[0],[0],[90],[90]],'--_art-angle',spring,8000,'deg');
  assert.equal(frames[0]['--_art-angle'],'0.000000deg');
  assert.equal(frames.at(-1)['--_art-angle'],'0deg');
  for(let scene=0;scene<4;scene++){
    const start=frames[scene*spring.samples.length],end=frames[(scene+1)*spring.samples.length-1];
    assert.equal(start.offset,scene/4);
    assert.ok(end.offset<(scene+1)/4,'turn settles before the next resize beat');
    assert.equal(parseFloat(end['--_art-angle']),parseFloat(frames[(scene+1)*spring.samples.length]['--_art-angle']));
  }
  assert.ok(frames.every(frame=>parseFloat(frame['--_art-angle'])>=0&&parseFloat(frame['--_art-angle'])<=90),'stripe rotation never overshoots');
});

test('corner morphs use pixel radii and finish at their exact token values', () => {
  const module={exports:{}};
  runInNewContext(readFileSync(scriptPath('artwork.js'),'utf8'),{module});
  const {artworkSpring,artworkFrames}=module.exports;
  const frames=artworkFrames([[28,28,28,28],[48,8,48,8]],'borderRadius',artworkSpring(.9,1400),24000,'px');
  assert.equal(frames[0].borderRadius,'28.000000px 28.000000px 28.000000px 28.000000px');
  assert.equal(frames.find(frame=>frame.offset===.5).borderRadius,'48.000000px 8.000000px 48.000000px 8.000000px');
  assert.equal(frames.at(-1).borderRadius,'28px 28px 28px 28px');
});

test('hero corners follow leading edges and flatten compressed sides', () => {
  const module={exports:{}};
  runInNewContext(readFileSync(scriptPath('artwork.js'),'utf8'),{module});
  const corners=module.exports.artworkDirectionalCorners,from={left:0,right:1,top:0,bottom:1},radii=[8,16,28,48];
  const values=to=>Array.from(corners(from,to,radii));
  assert.deepEqual(values({left:0,right:1,top:-1,bottom:0}),[48,48,8,8],'upwards rounds the top');
  assert.deepEqual(values({left:1,right:2,top:0,bottom:1}),[8,48,48,8],'rightwards rounds the right');
  assert.deepEqual(values({left:-1,right:0,top:0,bottom:1}),[48,8,8,48],'leftwards rounds the left');
  assert.deepEqual(values({left:0,right:1,top:1,bottom:2}),[8,8,48,48],'downwards rounds the bottom');
  assert.deepEqual(values({left:.25,right:.75,top:0,bottom:1}),[8,8,8,8],'compression flattens corners');
  assert.deepEqual(values(from),[28,28,28,28]);
});

test('Canvas artwork uses the DOM spring samples and keeps all tiles in bounds',()=>{
  const core={exports:{}},canvas={exports:{}};
  runInNewContext(readFileSync(scriptPath('artwork.js'),'utf8'),{module:core});
  runInNewContext(readFileSync(scriptPath('artwork-canvas.js'),'utf8'),{module:canvas});
  const spring=core.exports.artworkSpring(.9,300),track=[[1.2,.8,1],[.65,1.45,.9],[1,.65,1.35],[1.5,.85,.65]];
  const frames=core.exports.artworkFrames(track,'gridTemplateColumns',spring);
  frames.slice(0,spring.samples.length).forEach((frame,i)=>{
    const values=canvas.exports.trackValues(track,0,core.exports.artworkProgress(spring,i*spring.duration/(spring.samples.length-1)));
    frame.gridTemplateColumns.split(' ').forEach((value,j)=>assert.ok(Math.abs(parseFloat(value)-values[j])<.000001));
  });
  for(const width of [240,480,960])for(let time=0;time<8000;time+=50){
    const scene=Math.floor(time/2000),p=core.exports.artworkProgress(spring,time%2000),values=canvas.exports.trackValues(track,scene,p);
    const boxes=canvas.exports.tileBounds(width,448,8,values,[values,values,values]);
    assert.equal(boxes.length,9);
    assert.ok(boxes.every(b=>b.width>0&&b.height>0&&b.x>=0&&b.y>=0&&b.x+b.width<=width+.00001&&b.y+b.height<=448+.00001));
  }
  const ease=canvas.exports.turnProgress;
  assert.equal(ease(0,300,'cubic-bezier(.2,0,0,1)'),0);
  assert.equal(ease(300,300,'cubic-bezier(.2,0,0,1)'),1);
  let prior=0;for(let t=0;t<=300;t++){const value=ease(t,300,'cubic-bezier(.2,0,0,1)');assert.ok(value>=prior&&value<=1);prior=value;}
});

test('Canvas artwork starts each scene at its first painted frame after a late wake-up', () => {
  let clock = 0, frame, timer;
  const times = [];
  const ctx = new Proxy({}, {
    get: (_, name) => name === 'createLinearGradient'
      ? () => ({addColorStop() {}}) : () => {},
  });
  const canvas = {getContext: () => ctx};
  const root = {
    querySelector: () => canvas,
    querySelectorAll: () => Array.from({length: 9}, () => ({append() {}})),
    getBoundingClientRect: () => ({width: 480, height: 448}),
    setAttribute() {}, removeAttribute() {},
  };
  const scope = {
    document: {
      createElement: () => ({style: {}, remove() {}}),
      documentElement: {},
    },
    window: {devicePixelRatio: 1, addEventListener() {}, removeEventListener() {}},
    performance: {now: () => clock},
    getComputedStyle: () => ({color: '#fff', gap: '8px'}),
    ResizeObserver: class {observe() {} disconnect() {}},
    MutationObserver: class {observe() {} disconnect() {}},
    requestAnimationFrame: fn => {frame = fn; return 1;},
    cancelAnimationFrame: () => {frame = null;},
    setTimeout: fn => {timer = fn; return 2;},
    clearTimeout: () => {timer = null;},
  };
  runScript('artwork-canvas.js', scope);
  const track = Array.from({length: 4}, () => [1, 1, 1]);
  const renderer = scope.TokenArtworkCanvas(root, {
    tracks: [track, track, track, track],
    spring: {duration: 900}, shapeSpring: {duration: 500},
    corners: [8, 16, 28, 48],
    turnDuration: 300, turnEasing: 'linear', loop: 8000,
    progress: (_, time) => {times.push(time); return Math.min(1, time / 900);},
    cornerStates: () => Array.from({length: 4}, () => [28, 28, 28, 28]),
  });
  renderer.play();
  clock = 1000;
  frame(); // Settled: sleep until the 2000 ms scene boundary.
  times.length = 0;
  clock = 2140;
  timer(); // Simulate a timer delivered 140 ms late.
  assert.equal(times.length, 0, 'timer must wait for a paint before restarting motion');
  clock = 2190;
  frame(); // The paint itself is delayed another 50 ms.
  assert.equal(times.at(-1), 0, 'first painted frame must not skip into the new scene');
  clock += 1000 / 60;
  frame();
  assert.ok(Math.abs(times.at(-1) - 1000 / 60) < .001, 'next frame advances normally');
  renderer.pause();
  assert.equal(frame, null);
  assert.equal(timer, null);
  renderer.dispose();
});

test('pointer focus returns to visible keyboard focus on keyboard input',()=>{
 const events={},attributes=new Set();
 runScript('site-ui.js',{
  document:{readyState:'loading',documentElement:{setAttribute:name=>attributes.add(name),removeAttribute:name=>attributes.delete(name)},addEventListener:(name,fn)=>events[name]=fn},
  window:{addEventListener(){}},matchMedia:()=>({matches:false,addEventListener(){}})
 });
 events.pointerdown();assert(attributes.has('data-pointer-focus'));
 events.keydown({key:'Tab'});assert(!attributes.has('data-pointer-focus'));
 events.pointerdown();events.keydown({key:'c',metaKey:true});assert(attributes.has('data-pointer-focus'));
 events.keydown({key:'ArrowDown'});assert(!attributes.has('data-pointer-focus'));
});

test('returning to a gallery restores its saved offset instantly without restoring on filters',async()=>{
 const events={},windows={},storage=new Map(),scrolls=[];
 const location={pathname:'/md3/components',search:''},window={scrollY:780,addEventListener:(name,fn)=>windows[name]=fn,scrollTo:options=>scrolls.push(options)};
 runScript('site-ui.js',{
  document:{readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,getElementById:()=>({})},window,location,
  sessionStorage:{setItem:(key,value)=>storage.set(key,value),getItem:key=>storage.get(key)??null},
  matchMedia:()=>({matches:false,addEventListener(){}}),requestAnimationFrame:fn=>fn(),Promise
 });
 events['htmx:before:request']({detail:{ctx:{request:{headers:{'HX-Boosted':'true'}}}}});
 assert.equal(storage.get('tokens-scroll:/md3/components'),'780');
 assert.equal(windows.pageshow,undefined,'full refresh uses browser scroll restoration');
 location.pathname='/md3/components/cards';windows.pagehide();
 location.pathname='/md3/components';events['htmx:after:swap']({detail:{ctx:{request:{headers:{'HX-Boosted':'true'}}}}});await Promise.resolve();
 assert.equal(scrolls.length,1);assert.equal(scrolls[0].top,780);assert.equal(scrolls[0].behavior,'instant');
 events['htmx:after:swap']({detail:{ctx:{request:{headers:{}}}}});await Promise.resolve();assert.equal(scrolls.length,1);
});

test('image previews reveal after decode and keep their placeholder on failure',async()=>{
 const events={},attributes=new Set(); let decoded;
 const picture={setAttribute:name=>attributes.add(name),removeAttribute:name=>attributes.delete(name)};
 class TestImage {complete=false;naturalWidth=0;closest(){return picture;}decode(){return new Promise(resolve=>decoded=resolve);}}
 const image=new TestImage();
 runScript('thumbnail-images.js',{HTMLImageElement:TestImage,document:{readyState:'complete',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[image]}});
 assert(attributes.has('data-image-loading'));
 events.load({target:image});assert(attributes.has('data-image-loading'));
 decoded();await new Promise(setImmediate);
 assert(!attributes.has('data-image-loading'));
 attributes.add('data-image-loading');image.decode=()=>Promise.reject(new Error('failed'));
 events.load({target:image});await new Promise(setImmediate);
 assert(attributes.has('data-image-loading'));
});

test('shared header establishes desktop docking before deferred scripts run', () => {
  const script = readFileSync(join(__dirname, '../../../../forge/site/site-init.js'), 'utf8');
  for (const [breakpoint, width, docked] of [[1100,599,false],[1100,600,true],[1100,1099,true],[1100,1100,true],[1440,1200,true],[1440,1440,true]]) {
    const drawer = {dataset:{variant:'modal'},setAttribute(name,value){this[name]=value;}};
    const appearance={},placed=[];
    const header = {dataset:{drawerBreakpoint:String(breakpoint)},querySelector:selector=>selector==='.tokens-drawer-appearance'?appearance:selector==='#tokens-site-menu'?drawer:null,append:node=>placed.push(node)};
    const root = {dataset:{}};
    runInNewContext(script, {
      document:{currentScript:{parentElement:header},documentElement:root},
      matchMedia:query=>({matches:width>=Number(query.match(/\d+/)[0])}),
    });
    assert.equal(header.dataset.docked,String(docked));
    assert.equal(root.dataset.siteDrawer,docked?'expanded':'modal');
    assert.equal(drawer.dataset.variant,docked?'standard':'modal');
    assert.equal(Object.hasOwn(drawer,'open'),docked);
    assert.equal(drawer.role,docked?'navigation':undefined);
    assert.equal(placed.length,docked?1:0,'appearance controls start in their final responsive container before paint');
    if(docked)assert.equal(placed[0],appearance);
  }
});

test('disclosure expansion retargets a reversed animation and honors reduced motion', async()=>{
 const events={},attributes=new Set(),animations=[];let reduced=false,rendered;
 const body={inert:false};
 const detail={open:false,style:{},children:[],getBoundingClientRect(){return {height:rendered??(this.open?220:60)};},setAttribute:k=>attributes.add(k),removeAttribute:k=>attributes.delete(k),animate(frames){let resolve,reject;const finished=new Promise((a,b)=>{resolve=a;reject=b;});const animation={frames,finished,finish:resolve,cancel(){this.cancelled=true;reject(Error('cancelled'));}};animations.push(animation);return animation;}};
 const toggle={setAttribute(){}};
 const summary={parentElement:detail,querySelector:()=>toggle,getBoundingClientRect:()=>({height:58})};detail.children=[summary,body];
 runScript('site-ui.js',{
  document:{readyState:'loading',addEventListener:(name,fn)=>(events[name]??=[]).push(fn)},window:{addEventListener(){}},
  matchMedia:()=>({matches:reduced}),getComputedStyle:()=>({borderTopWidth:'1',borderBottomWidth:'1'}),MaterialMotion:{timing:()=>({duration:450,easing:'linear'})}
 });
 const click=(icon=true)=>events.click[0]({target:{closest:selector=>selector==='[data-disclosure-toggle]'?(icon?toggle:null):summary},preventDefault(){}});
 const opening=click(false);assert.equal(detail.open,true);assert.equal(animations[0].frames[1].height,'220px');
 rendered=130;const closing=click();assert.equal(animations[0].cancelled,true);assert.equal(animations[1].frames[0].height,'130px');assert.equal(animations[1].frames[1].height,'60px');assert.equal(body.inert,true);
 animations[1].finish();await Promise.all([opening,closing]);assert.equal(detail.open,false);assert.equal(body.inert,false);assert.equal(attributes.size,0);
 reduced=true;rendered=undefined;await click();assert.equal(detail.open,true);assert.equal(animations.length,2,'reduced motion changes state without a height animation');
});

test('docked shell follows visual viewport frames without reopening or retaining a small-height gap',()=>{
 const styles=new Map(),events={},wideEvents={},frames=[];
 let shows=0,closes=0;
 const body={blur(){}},opener={setAttribute(){}};
 const drawer={open:true,dataset:{variant:'standard'},setAttribute(){},removeAttribute(){},addEventListener(){},show(){this.open=true;shows++;},close(){this.open=false;closes++;}};
 const header={addEventListener(){},querySelectorAll:()=>[],dataset:{drawerBreakpoint:'1100'},style:{setProperty:(key,value)=>styles.set(key,value),removeProperty:key=>styles.delete(key)},querySelector:selector=>selector==='#tokens-site-menu'?drawer:selector==='[data-navigation-open]'?opener:null,getBoundingClientRect:()=>({height:64})};
 const wide={matches:true,addEventListener:(name,fn)=>wideEvents[name]=fn};
 const viewport={height:700,offsetTop:0,scale:1,addEventListener:(name,fn)=>events[name]=fn};
 runInNewContext(readFileSync(join(__dirname,'../../../../forge/site/site.js'),'utf8'),{
  document:{addEventListener(){},getElementById:()=>header,body,activeElement:body,documentElement:{dataset:{},style:{setProperty(){}}}},
  window:{visualViewport:viewport},matchMedia:()=>wide,
  requestAnimationFrame:fn=>{frames.push(fn);return 1;},ResizeObserver:class {observe(){}},
 });
 assert.equal(styles.get('--site-drawer-height'),'700px');
 for(const height of [712,724,736,748,760,748,736,724,712,700]){
  viewport.height=height;events.resize();events.scroll();
  assert.equal(frames.length,1,'viewport events in the same frame are coalesced');
  frames.shift()();
  assert.equal(styles.get('--site-drawer-height'),`${height}px`);
 }
 viewport.offsetTop=12;events.scroll();frames.shift()();
 assert.equal(styles.has('--site-drawer-top'),false,'elastic-scroll offsets must not translate a fixed drawer');
 assert.equal(styles.get('--site-drawer-height'),'700px');
 assert.equal(shows,0);assert.equal(closes,0,'startup and toolbar scrolling preserve the already docked dialog');
 viewport.scale=2;events.resize();frames.shift()();
 assert.equal(styles.has('--site-drawer-height'),false,'pinch zoom preserves normal layout sizing');
 viewport.scale=1;wide.matches=false;wideEvents.change();
 assert.equal(styles.has('--site-drawer-height'),false);
 assert.equal(styles.has('--site-drawer-top'),false);
 assert.equal(drawer.open,false);assert.equal(header.dataset.docked,'false');
 wide.matches=true;viewport.height=800;wideEvents.change();
 assert.equal(styles.get('--site-drawer-height'),'800px');assert.equal(drawer.open,true);
});

test('playground comparison group updates its form value only when the mode changes', () => {
  const listeners = {};
  const input = new EventTarget();
  input.value = 'interactive';
  let changes = 0;
  input.addEventListener('change', () => changes++);
  const group = {
    closest: () => null,
    hasAttribute: name => name === 'data-preview-compare',
    parentElement: {querySelector: () => input},
  };
  const button = {dataset: {previewComparison: 'variants'}, closest: () => group};
  runScript('component-demos.js', {
    document: {addEventListener: (name, listener) => {listeners[name] = listener;}}, Event,
  });
  listeners['md-button-group-action']({target: button});
  assert.equal(input.value, 'variants');
  assert.equal(changes, 1);
  listeners['md-button-group-action']({target: button});
  assert.equal(changes, 1);
});

test('site navigation drills into a single mobile list and preserves its position across swaps',()=>{
 let width=390;
 const handlers={},docHandlers={},root={dataset:{},style:{setProperty(){}}};
 const focus=[],body={blur(){}},opener={setAttribute(){}};
 const panels=['/components','/styles'].map(path=>({dataset:{sitePanel:path},hidden:false,attributes:{},toggleAttribute(name,value){if(value)this.attributes[name]='';else delete this.attributes[name];},removeAttribute(name){delete this.attributes[name];}}));
 const buttons=panels.map(p=>({dataset:{siteSection:p.dataset.sitePanel},setAttribute(){},focus(){focus.push(this.dataset.siteSection);}}));
 const back={focus(){focus.push('back');}},main={focus(){focus.push('main');}};
 let node={dataset:{activeSection:'/components'}},lists=panels.map(p=>({dataset:{siteScroll:p.dataset.sitePanel},scrollTop:0,getClientRects:()=>p.hidden?[]:[{}],querySelector:()=>null,matches:()=>true}));
 const drawer={open:false,dataset:{},close(){this.open=false;},removeAttribute(){},setAttribute(){},addEventListener(){}};
 const header={contains:target=>target===drawer,dataset:{drawerBreakpoint:'1100'},style:{removeProperty(){}},getBoundingClientRect:()=>({height:64}),addEventListener:(name,fn)=>handlers[name]=fn,
 querySelectorAll:selector=>selector==='[data-site-panel]'?panels:selector==='[data-site-section]'?buttons:lists,
 querySelector:selector=>selector==='#tokens-site-menu'?drawer:selector==='[data-navigation-open]'?opener:selector==='#tokens-site-appearance'?null:selector==='[data-site-navigation]'?node:selector.includes('[data-site-back]')?back:selector==='.tokens-site-main a'?main:selector.startsWith('[data-site-section=')?buttons.find(b=>selector.includes(b.dataset.siteSection)):selector.startsWith('[data-site-panel=')?panels.find(p=>selector.includes(p.dataset.sitePanel)):null};
 runInNewContext(readFileSync(join(__dirname,'../../../../forge/site/site.js'),'utf8'),{
  document:{getElementById:()=>header,addEventListener:(name,fn)=>docHandlers[name]=fn,documentElement:root,body,activeElement:body},
  window:{},URL,location:{href:'http://localhost/components',pathname:'/components'},matchMedia:query=>({get matches(){return width>=Number(query.match(/\d+/)[0]);},addEventListener(){}}),ResizeObserver:class{observe(){}},
 });
 let prevented=false;handlers.click({target:{closest:selector=>selector==='[data-site-section]'?{dataset:{siteSection:'/components'}}:null},preventDefault(){prevented=true;},stopPropagation(){}});
 assert.equal(prevented,false,'the document section attribute must not intercept ordinary links or appearance controls');
 assert.equal(header.dataset.mobileLevel,'section');assert.equal(panels[0].hidden,false);assert.equal(panels[1].hidden,true);
 const click=(section,isBack=false)=>handlers.click({target:{closest:selector=>selector==='.tokens-appearance-toggle'?null:selector==='button[data-site-section]'?section:isBack?back:null},preventDefault(){},stopPropagation(){}});
 click(buttons[1]);assert.equal(panels[0].hidden,true);assert.equal(panels[1].hidden,false);assert.equal(focus.at(-1),'back');
 click(null,true);assert.equal(header.dataset.mobileLevel,'main');assert.ok(panels.every(p=>p.hidden));assert.equal(focus.at(-1),'/styles');
 click(buttons[0]);lists[0].scrollTop=270;handlers.scroll({target:lists[0]});
 assert.equal(panels[0].attributes['data-site-enter'],'');
 docHandlers['htmx:beforeSwap']();assert.equal(panels[0].attributes['data-site-enter'],undefined,'navigation must not replay the section entrance');lists=lists.map(l=>({...l,scrollTop:0}));node={dataset:{activeSection:'/components'}};docHandlers['htmx:afterSwap']();
 assert.equal(lists[0].scrollTop,270,'navigation updates must retain the existing list position');
 assert.equal(header.dataset.mobileLevel,'section');
 node.dataset.activeSection='/styles';node.dataset.activePage='/styles/shape';docHandlers['htmx:afterSwap']();
 assert.equal(panels[1].hidden,false,'morphing retains the navigation node but must refresh changed page metadata');
 node.dataset.activeSection='/components';node.dataset.activePage='/components';docHandlers['htmx:afterSwap']();
 width=900;
 const railLink={href:'/styles',dataset:{siteRailSection:'/styles'}};
 prevented=false;
 handlers.click({target:{closest:selector=>selector==='[data-site-rail-section]'?railLink:null},preventDefault(){prevented=true;},stopPropagation(){}});
 assert.equal(prevented,false,'rail links navigate directly at medium widths');
 handlers.pointerover({target:{closest:()=>railLink},pointerType:'mouse'});
 assert.equal(header.dataset.panelOpen,'true');assert.equal(panels[1].hidden,false);
 handlers.pointerleave();assert.equal(header.dataset.panelOpen,'false');assert.equal(panels[1].hidden,false,'overlay exit keeps the hovered section visible through its fade');
 handlers.focusin({target:{closest:()=>railLink}});assert.equal(header.dataset.panelOpen,'true');
 handlers.keydown({key:'Escape',preventDefault(){}});assert.equal(header.dataset.panelOpen,'false');
 const currentLink={href:'/components',dataset:{siteRailSection:'/components'}};
 const activate=link=>{
  let prevented=false;
  handlers.click({target:{closest:selector=>selector==='[data-site-rail-section]'?link:null},preventDefault(){prevented=true;},stopPropagation(){}});
  return prevented;
 };
 docHandlers.pointerdown({pointerType:'touch',target:{}});
 handlers.focusin({type:'focusin',target:{closest:()=>currentLink}});
 assert.equal(header.dataset.panelOpen,'false','touch focus must not impersonate hover');
 assert.equal(activate(currentLink),true,'the current page must not navigate again');
 assert.equal(header.dataset.panelOpen,'true');
 handlers.pointerleave();assert.equal(header.dataset.panelOpen,'true','a tapped drawer stays open after pointer exit');
 assert.equal(activate(currentLink),true);assert.equal(header.dataset.panelOpen,'false','a second tap closes the drawer');
 assert.equal(activate(railLink),false,'another destination still navigates');
 node={dataset:{activeSection:'/styles'}};docHandlers['htmx:afterSwap']();
 assert.equal(header.dataset.panelOpen,'true','touch navigation retains its drawer on arrival');
 docHandlers.pointerdown({pointerType:'touch',target:{}});
 assert.equal(header.dataset.panelOpen,'false','outside taps dismiss the latched drawer');
 handlers.keydown({key:'Escape',preventDefault(){}});
 width=1440;node={dataset:{activeSection:'/components'}};docHandlers['htmx:afterSwap']();
 assert.equal(activate(currentLink),true);assert.equal(drawer.inert,false,'wide current-page activation retains the persistent drawer');
 handlers.pointerover({target:{closest:()=>railLink},pointerType:'mouse'});
 assert.equal(panels[1].hidden,false,'wide windows also reveal the hovered section');
 assert.equal(root.dataset.siteSection,'/components','hover must not change reserved content width');
 const noSubpages={dataset:{siteRailSection:'/changes'}};
 handlers.pointerover({target:{closest:()=>noSubpages},pointerType:'mouse'});
 assert.equal(header.dataset.panelOpen,'true');assert.equal(panels[1].hidden,false,'destinations without subpages retain the last shown drawer');
 handlers.pointerleave();assert.equal(panels[1].hidden,false,'leaving retains the last shown section rather than restoring the selected page');assert.equal(drawer.inert,false);
 node={dataset:{activeSection:''}};docHandlers['htmx:afterSwap']();
 assert.equal(panels[1].hidden,false,'navigation to a page without subpages remembers the last drawer content');
 handlers.pointerover({target:{closest:()=>railLink},pointerType:'mouse'});
 assert.equal(header.dataset.panelOpen,'true');assert.equal(root.dataset.siteSection,'','a page without a drawer retains its width');assert.equal(drawer.inert,false);
 handlers.pointerleave();assert.equal(drawer.inert,true);assert.equal(panels[1].hidden,false,'retain outgoing content for the exit fade');
});

test('component styles load once in persistent manifest order before swaps proceed',async()=>{
 const links=[],events={};
 class Link {
  constructor(path,order){this.path=path;this.dataset={componentOrder:order};this.listeners={};this.sheet=null;}
  get href(){return new URL(this.path,'https://example.test/md3/').href;}
  getAttribute(name){return name==='href'?this.path:null;}
  removeAttribute(){}
  addEventListener(name,fn){this.listeners[name]=fn;}
  cloneNode(){return new Link(this.path,this.dataset.componentOrder);}
  remove(){const at=links.indexOf(this);if(at>=0)links.splice(at,1);}
 }
 const head={querySelectorAll:()=>links,insertBefore(link,next){const at=links.indexOf(next);links.splice(at<0?links.length:at,0,link);}};
 const document={baseURI:'https://example.test/md3/',head,currentScript:{src:'https://example.test/md3/static/asset-loader.js',dataset:{sources:'{}'}},querySelectorAll:()=>[],getElementById:()=>null,addEventListener:(name,fn)=>events[name]=fn,dispatchEvent(){}};
 const globals={document,URL,CustomEvent:class{},console};runScript('asset-loader.js',globals);
 const incoming={querySelectorAll:selector=>selector.startsWith('link[')?[new Link('/md3/b.css?v=1','0003'),new Link('/md3/a.css?v=1','0001')]:[]};
 let done=false;const ready=globals.TokenAssets.prepareStyles(incoming).then(()=>done=true);
 const repeated=globals.TokenAssets.prepareStyles(incoming);
 assert.deepEqual(links.map(link=>link.path),['/md3/a.css?v=1','/md3/b.css?v=1']);
 await Promise.resolve();assert.equal(done,false,'new content must wait for all CSS');
 links[0].listeners.load();await Promise.resolve();assert.equal(done,false);
 links[1].listeners.load();await ready;await repeated;assert.equal(done,true);
 await globals.TokenAssets.prepareStyles(incoming);assert.equal(links.length,2,'loaded styles must be reused');
 assert.equal(globals.TokenAssets.stylesChanged({querySelectorAll:()=>[new Link('/md3/a.css?v=2','0001')]}),true);
 const failed=globals.TokenAssets.prepareStyles({querySelectorAll:selector=>selector.startsWith('link[')?[new Link('/md3/c.css?v=1','0005')]:[]});
 links.at(-1).listeners.error();await assert.rejects(failed,/Could not load/);
 assert.equal(links.length,2,'failed stylesheet must be removed so it can be retried');
});

test('HTMX response preparation waits for component CSS and strips only dependency links',async()=>{
 const events={},body={},head={querySelectorAll:()=>[]};let release,prepared=false;
 const gate=new Promise(resolve=>release=resolve);
 const document={readyState:'loading',body,head,addEventListener:(name,fn)=>events[name]=fn};
 const TokenAssets={stylesChanged:()=>false,prepareStyles:async()=>{await gate;prepared=true;}};
 runScript('site-ui.js',{document,window:{addEventListener(){}},TokenAssets,DOMParser:class{parseFromString(){return {head,querySelectorAll:()=>[]};}}});
 const html='<main><link rel="stylesheet" data-component-stylesheet href="/split.css"><a href="/next">Next</a></main>';
 const ctx={target:body,response:{status:200,raw:{text:async()=>html}},request:{}};
 events['htmx:head:before:merge']({detail:{ctx},preventDefault(){throw Error('full head merge was cancelled');}});
 let result;const pending=ctx.response.raw.text().then(text=>result=text);
 await Promise.resolve();assert.equal(result,undefined,'htmx must still have the old page while CSS loads');
 release();await pending;assert.equal(prepared,true);assert.equal(result,'<main><a href="/next">Next</a></main>');
 let protectedStyles=false;events['htmx:head:before:remove']({detail:{headElement:{matches:selector=>selector.includes('[data-component-stylesheet]')}},preventDefault(){protectedStyles=true;}});
 assert.equal(protectedStyles,true);
 const partial={target:{},response:{status:200,raw:{text:async()=>html}},request:{}};let skippedHeadMerge=false;
 events['htmx:head:before:merge']({detail:{ctx:partial},preventDefault(){skippedHeadMerge=true;}});
 assert.equal(skippedHeadMerge,true);assert.equal(await partial.response.raw.text(),'<main><a href="/next">Next</a></main>');
 TokenAssets.stylesChanged=()=>true;
 const changed={target:body,response:{status:200,raw:{text:async()=>html}},request:{}};
 events['htmx:head:before:merge']({detail:{ctx:changed},preventDefault(){}});
 assert.equal(await changed.response.raw.text(),'');assert.equal(changed.catalogAssetChange,true);
});

test('per-node HTMX link policy does not rescan processed containers',()=>{
 const events={};
 runScript('site-ui.js',{document:{readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelector:()=>({src:'https://example.test/static/site-ui.js'})},window:{addEventListener(){}},URL,location:{origin:'https://example.test'}});
 events['htmx:beforeProcessNode']({detail:{elt:{matches:()=>false,querySelectorAll(){throw Error('must not rescan descendants');}}}});
 for(const href of ['https://example.test/components/cards','https://example.test/download','https://outside.test/']){
  const attributes={};
  const link={href,matches:()=>true,querySelectorAll:()=>[],setAttribute:(name,value)=>attributes[name]=value};
  events['htmx:beforeProcessNode']({detail:{elt:link}});
  assert.equal(attributes['hx-boost'],href.includes('/components/')?undefined:'false');
 }
});

function behaviorLoader({css=false,preview=true,eager=[],declared=['runtime/button-groups.js','runtime/segmented-buttons.js']}={}) {
 const nodes=[],events={},dispatched=[],errors=[],matches=new Map();
 const declarations=new Set(declared);
 const grid={inert:true},region={inert:false,isConnected:true,busy:false,error:null,
  setAttribute(){this.busy=true},removeAttribute(){this.busy=false},
  contains:()=>true,querySelector(selector){return selector==='.tokens-preview-grid'?grid:this.error},
  append(node){this.error=node}};
 let current=preview?region:null;
 const style={sheet:css?null:{},dataset:{},listeners:{},getAttribute:()=>'/style.css?v=1',addEventListener(name,fn){this.listeners[name]=fn},remove(){}};
 const document={baseURI:'https://example.test/',currentScript:{src:'https://example.test/static/asset-loader.js',dataset:{sources:JSON.stringify(Object.fromEntries(['runtime/button-groups.js','runtime/segmented-buttons.js','dialogs.js','dialog-demo.js','component-demos.js'].map(name=>[name,'/static/'+name+'?v=1']))) }},
  head:{querySelectorAll:()=>[style],append(node){nodes.push(node)}},
  querySelector:()=>current,getElementById:()=>current,
  querySelectorAll:selector=>selector==='script[data-component-behaviors]'?[{textContent:JSON.stringify([...declarations])}]:selector==='script[src]:not([data-component-script])'?eager.map(name=>({getAttribute:()=>'/static/'+name+'?v=1'})):matches.get(selector)||[],
  createElement(tag){return {tag,dataset:{},async:true,remove(){if(region.error===this)region.error=null;const index=nodes.indexOf(this);if(index>=0)nodes.splice(index,1)},setAttribute(){}}},
  addEventListener(name,fn){events[name]=fn},dispatchEvent(event){dispatched.push(event.type)}};
 const context={document,URL,CustomEvent:class{constructor(type){this.type=type}},console:{error:error=>errors.push(error)}};
 const add=selector=>matches.set(selector,[{closest:()=>null}]);
 add('.md-button-group');add('.md-segmented-button');
 runScript('asset-loader.js',context);
 return {api:context.TokenAssets,declarations,nodes,region,grid,events,dispatched,errors,style,add,
  remove:selector=>matches.delete(selector),scripts:()=>nodes.filter(node=>node.tag==='script'),
  hints:()=>nodes.filter(node=>node.tag==='link'),
  replace(){region.isConnected=false;current={...region,inert:false,isConnected:true,error:null};events['htmx:afterSwap']();return current}};
}

const settleBehaviorLoader=()=>new Promise(resolve=>setImmediate(resolve));

test('behavior downloads are hinted together while CSS gates ordered script execution',async()=>{
 const h=behaviorLoader({css:true});
 assert.deepEqual(h.hints().map(node=>[node.rel,node.as]),[['preload','script'],['preload','script']]);
 assert(h.hints().every(node=>node.href.endsWith('?v=1')));
 assert.equal(h.scripts().length,0,'preloading must not execute controllers before CSS');
 assert.equal(h.region.inert,true);
 h.style.sheet={};h.style.listeners.load();await settleBehaviorLoader();
 assert.equal(h.scripts().length,1);assert(h.scripts()[0].src.includes('button-groups.js'));
 h.scripts()[0].onload();await settleBehaviorLoader();
 assert.equal(h.scripts().length,2);assert(h.scripts()[1].src.includes('segmented-buttons.js'));
 assert(h.scripts().every(node=>node.async===false));
 h.scripts()[1].onload();await h.api.ready;
 assert.equal(h.region.inert,false);assert.equal(h.grid.inert,false);
 assert.deepEqual(h.dispatched,['tokens:styles-ready','tokens:assets-ready']);
 h.events['htmx:afterSwap']();await h.api.ready;
 assert.equal(h.hints().length,2);assert.equal(h.scripts().length,2,'activation reuses already evaluated scripts');
});

test('declared children load even when their markup is created by an earlier controller',async()=>{
 const h=behaviorLoader({preview:false,declared:['runtime/button-groups.js','dialogs.js','dialog-demo.js']});h.remove('.md-segmented-button');
 await settleBehaviorLoader();
 h.add('dialog[data-dialog-surface],dialog[data-dialog-preview]');h.add('.md-dialog-demo');
 h.scripts()[0].onload();await settleBehaviorLoader();
 assert(h.scripts()[1].src.includes('/dialogs.js?'));
 h.scripts()[1].onload();await settleBehaviorLoader();
 assert(h.scripts()[2].src.includes('/dialog-demo.js?'));
 h.scripts()[2].onload();await h.api.ready;
 assert.equal(h.hints().length,3,'parent declarations preload child behavior before its markup exists');
});

test('failed behavior releases its preload and can be retried without reloading successful scripts',async()=>{
 const h=behaviorLoader();await settleBehaviorLoader();
 h.scripts()[0].onload();await settleBehaviorLoader();
 const failed=h.scripts()[1],task=h.api.ready;
 failed.onerror();await assert.rejects(task,/segmented-buttons/);await settleBehaviorLoader();
 assert.equal(h.grid.inert,true);assert.equal(h.errors.length,1);
 assert.equal(h.hints().length,1,'the failed script hint must not block a retry');
 h.events['htmx:afterSwap']();await settleBehaviorLoader();
 assert.equal(h.hints().length,2);assert.equal(h.scripts().length,2);
 h.scripts()[1].onload();await h.api.ready;
 assert.equal(h.grid.inert,false);assert.equal(h.region.error,null);
});

test('superseded activations cannot enable a newer preview or report stale errors',async()=>{
 const h=behaviorLoader();await settleBehaviorLoader();
 const old=h.api.ready,next=h.replace(),latest=h.api.ready;
 h.declarations.add('dialogs.js');
 await settleBehaviorLoader();h.scripts()[0].onload();await settleBehaviorLoader();
 assert.equal(next.inert,true);assert.equal(h.dispatched.filter(name=>name==='tokens:assets-ready').length,0);
 h.scripts()[1].onload();await settleBehaviorLoader();
 assert.equal(next.inert,true);h.scripts()[2].onload();await Promise.all([old,latest]);
 assert.equal(next.inert,false);assert.equal(h.dispatched.filter(name=>name==='tokens:assets-ready').length,1);
 const failed=behaviorLoader();await settleBehaviorLoader();
 const outdated=failed.api.ready;failed.declarations.clear();failed.remove('.md-button-group');failed.remove('.md-segmented-button');
 const replacement=failed.replace();await failed.api.ready;
 failed.scripts()[0].onerror();await assert.rejects(outdated,/button-groups/);await settleBehaviorLoader();
 assert.equal(replacement.error,null);assert.equal(replacement.inert,false);assert.equal(failed.errors.length,0);
});

test('upstream HTMX preload reuses a response once and expires stale responses',async()=>{
 const h=prefetchHarness(),link=h.link('/md3/components/buttons');
 h.event('focusin',link);h.advance(65);
 const ctx={request:{action:'/md3/components/buttons'}};
 h.extension.htmx_before_request(link,{ctx});
 assert.equal(await ctx.fetch(),h.appended[0].response);
 assert.equal(h.appended.length,1,'navigation must reuse the prefetch rather than issue another fetch');
 const next={request:{action:'/md3/components/buttons'}};h.extension.htmx_before_request(link,{ctx:next});
 assert.equal(next.fetch,undefined,'a consumed response must not be reused');
 const expired=h.link('/md3/components/cards');h.event('focusin',expired);h.advance(65);h.advance(5001);
 const stale={request:{action:'/md3/components/cards'}};h.extension.htmx_before_request(expired,{ctx:stale});
 assert.equal(stale.fetch,undefined,'expired content falls back to a normal HTMX request');
});

test('failed prefetch falls back to navigation, and persistent links remain armed',async()=>{
 const failed=prefetchHarness({fail:true}),link=failed.link('/md3/components/buttons');
 failed.event('focusin',link);failed.advance(65);await new Promise(resolve=>setImmediate(resolve));
 const ctx={request:{action:'/md3/components/buttons'}};failed.extension.htmx_before_request(link,{ctx});
 assert.equal(ctx.fetch,undefined);
 const h=prefetchHarness(),home=h.link('/md3/'),unboosted=h.link('/md3/components/cards',{unboosted:true});
 assert.equal(home.attrs['hx-preload'],'tokens:prefetch');
 assert.equal(unboosted.attrs['hx-preload'],undefined);
 h.event('focusin',home);h.advance(65);assert.equal(h.appended.length,0);
 h.location.href='https://tokens.test/md3/components/buttons';
 h.event('focusin',home);h.advance(65);assert.equal(h.appended.length,1);
});

test('incoming HTML starts its script downloads before insertion without evaluating them',async()=>{
 const h=behaviorLoader({css:true});
 const incoming={querySelectorAll:selector=>selector==='script[data-component-behaviors]'?[{textContent:'["dialogs.js"]'}]:[]};
 await h.api.prepareStyles(incoming);
 assert(h.hints().some(node=>node.href.includes('/dialogs.js?')));
 assert.equal(h.scripts().length,0,'incoming behavior must not execute while current CSS is pending');
 const aborted=new AbortController();aborted.abort();
 await assert.rejects(h.api.prepareStyles(incoming,aborted.signal),{name:'AbortError'});
 h.style.sheet={};h.style.listeners.load();await settleBehaviorLoader();
 h.scripts()[0].onload();await settleBehaviorLoader();h.scripts()[1].onload();await h.api.ready;
 assert(!h.scripts().some(node=>node.src.includes('/dialogs.js?')),'preparing a response must not activate its absent dialog');
});


test('selectors cannot silently load undeclared behavior', async()=>{
 const h=behaviorLoader({css:true});h.declarations.clear();h.add('.md-dialog-demo');
 h.style.sheet={};h.style.listeners.load();await h.api.ready;
 assert.equal(h.scripts().length,0);
 assert.equal(h.api.audit().undeclared.includes('dialog-demo.js'),true);
});

test('unknown declared behavior fails before creating a script request', async()=>{
 const h=behaviorLoader({css:true});h.declarations.add('missing.js');
 h.style.sheet={};h.style.listeners.load();await assert.rejects(h.api.ready,/Unknown component behavior/);
 assert.equal(h.scripts().length,0);
});


test('declared delivery reuses host scripts and follows source order rather than markup order',async()=>{
 const h=behaviorLoader({declared:['runtime/segmented-buttons.js','runtime/navigation.js','runtime/button-groups.js'],eager:['runtime/navigation.js']});
 assert(h.hints()[0].href.includes('button-groups.js'));
 assert.equal(h.hints().length,2,'host navigation must not be requested again');
 await settleBehaviorLoader();h.scripts()[0].onload();await settleBehaviorLoader();h.scripts()[1].onload();await h.api.ready;
 assert(h.scripts()[0].src.includes('button-groups.js'));
});


test('export morph applies confirmed radio selection instead of preserving a disabled dirty choice',()=>{
 const events={};const live=[{id:'scope-view',checked:true},{id:'scope-collection',checked:false}];
 const confirmed=[{id:'scope-view',checked:false},{id:'scope-collection',checked:true}];
 const panel={querySelectorAll:()=>[]};const sourceElement={closest:()=>panel};
 runScript('export.js',{document:{readyState:'loading',addEventListener:(n,fn)=>events[n]=fn},location:{search:''}});
 events['htmx:before:swap']({detail:{ctx:{sourceElement},tasks:[{target:{id:'tokens-export-fields',querySelectorAll:()=>live},fragment:{querySelectorAll:selector=>selector.startsWith('input')?confirmed:[]}}]}});
 assert.deepEqual(live.map(i=>i.checked),[false,true]);
});

// Tests execute the same source or compiled entry that the asset server exposes.
function scriptPath(name) {
  for (const group of ['scripts', 'generated', 'vendor']) {
    const file = join(__dirname, '../../assets', group, name);
    if (existsSync(file)) return file;
  }
  throw new Error(`Unknown browser script: ${name}`);
}
