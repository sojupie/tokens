// Optional browser-script checks: node --test catalog/web/ui_test.cjs
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

test('table modes preserve preview preference and folder state, with matching action icons', async () => {
  const classes = new Set();
  const listeners = {};
  const groups = [{open: true}, {open: false}];
  const controls = Object.fromEntries(['previews', 'view', 'expand'].map(action => {
    const attributes = {};
    const icons = {on: {hidden: false}, off: {hidden: true}};
    return [action, {
      dataset: {action}, attributes, icons,
      setAttribute: (key, value) => { attributes[key] = value; },
      getAttribute: key => attributes[key],
      querySelector: selector => icons[selector.includes('"off"') ? 'off' : 'on'],
    }];
  }));
  const menu = {addEventListener() {}, matches: () => false};
  const browser = {
    classList: {
      contains: value => classes.has(value),
      toggle: value => { if (classes.has(value)) { classes.delete(value); return false; } classes.add(value); return true; },
    },
    querySelector: selector => {
      if (selector === 'form') return {};
      if (selector === '.tokens-context-dialog') return null;
      if (selector === '.tokens-token-set') return {};
      if (selector === '.tokens-token-set-menu') return menu;
      return controls[selector.match(/data-action="([^"]+)"/)[1]];
    },
    querySelectorAll: selector => selector === '.tokens-token-folder' ? groups : [],
    addEventListener: (name, handler) => { listeners[name] = handler; },
  };
  let copied;
  runScript('token-table.js', {
    document: {querySelectorAll: selector => selector === '.tokens-token-browser' ? [browser] : []},
    addEventListener() {},
    navigator: {clipboard: {writeText: async value => { copied = value; }}},
  });
  const click = action => listeners.click({target: {closest: selector => selector === 'button[data-action]' ? controls[action] : null}});
  await click('previews');
  assert.equal(classes.has('tokens-no-previews'), true);
  assert.equal(controls.previews.icons.off.hidden, false);
  await click('view');
  assert.equal(classes.has('tokens-grid-view'), true);
  assert.equal(controls.previews.hidden, true);
  assert.equal(controls.view.icons.off.hidden, false);
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
  const copy = {dataset: {copyReference: 'md.sys.color.primary'}, closest: () => ({querySelector: () => status})};
  await listeners.click({target: {closest: selector => selector === '[data-copy-reference]' ? copy : null}});
  assert.equal(copied, 'md.sys.color.primary');
  assert.equal(status.textContent, 'Copied.');
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

test('export dialog retains collection URLs and binds restored history content', () => {
  const events = {};
  function panel() {
    const listeners = {}, link = {href:'https://catalog.example/md3/downloads/dtcg/light-default/tokens.tokens.json'}, report = {};
    const fields = {format:{value:'tokens.css'},context:{value:'dark-high'}};
    const node = {
      link, report, fields, opens:0,
      querySelector(selector) {
        if(selector==='[data-export-link]')return link;
        if(selector==='[data-export-report]')return report;
        if(selector==='[data-export-description]')return {};
        if(selector==='[role="combobox"],button')return {focus(){}};
        return fields[selector.includes('format')?'format':'context'];
      },
      addEventListener(type,handler){listeners[type]=handler;},
      showModal(){this.opens++;},
      close(){listeners.close();}
    };
    return node;
  }
  const original=panel();let current=original;
  runScript('site-ui.js',{
    document:{readyState:'complete',querySelector:selector=>selector==='svg[data-artwork]'?null:({src:'https://catalog.example/md3/static/site-ui.js'}),querySelectorAll:()=>[],addEventListener(type,handler){events[type]=handler;},getElementById(id){return id==='tokens-site-download'?current:null;}},
    window:{addEventListener(){}},matchMedia:()=>({matches:false,addEventListener(){}}),URL
  });
  assert.equal(current.link.href,'https://catalog.example/md3/downloads/dtcg/dark-high/tokens.css');
  assert.equal(current.report.href,'https://catalog.example/md3/downloads/dtcg/dark-high/report.json');
  current=panel();current.fields.format.value='report.json';events['htmx:historyRestore']();
  let focused=false;
  const trigger={isConnected:true,focus(){focused=true;}};
  events.click({target:{closest:selector=>selector==='[data-export-open]'?trigger:null}});
  assert.equal(current.opens,1);assert.equal(original.opens,0);
  assert.equal(current.link.href,'https://catalog.example/md3/downloads/dtcg/dark-high/report.json');
  current.close();assert.equal(focused,true);
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
  runInNewContext(readFileSync(join(__dirname,'static','artwork.js'),'utf8'),{module});
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
  runInNewContext(readFileSync(join(__dirname,'static','artwork.js'),'utf8'),{module});
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
  runInNewContext(readFileSync(join(__dirname,'static','artwork.js'),'utf8'),{module});
  const {artworkSpring,artworkFrames}=module.exports;
  const frames=artworkFrames([[28,28,28,28],[48,8,48,8]],'borderRadius',artworkSpring(.9,1400),24000,'px');
  assert.equal(frames[0].borderRadius,'28.000000px 28.000000px 28.000000px 28.000000px');
  assert.equal(frames.find(frame=>frame.offset===.5).borderRadius,'48.000000px 8.000000px 48.000000px 8.000000px');
  assert.equal(frames.at(-1).borderRadius,'28px 28px 28px 28px');
});

test('hero corners follow leading edges and flatten compressed sides', () => {
  const module={exports:{}};
  runInNewContext(readFileSync(join(__dirname,'static','artwork.js'),'utf8'),{module});
  const corners=module.exports.artworkDirectionalCorners,from={left:0,right:1,top:0,bottom:1},radii=[8,16,28,48];
  const values=to=>Array.from(corners(from,to,radii));
  assert.deepEqual(values({left:0,right:1,top:-1,bottom:0}),[48,48,8,8],'upwards rounds the top');
  assert.deepEqual(values({left:1,right:2,top:0,bottom:1}),[8,48,48,8],'rightwards rounds the right');
  assert.deepEqual(values({left:-1,right:0,top:0,bottom:1}),[48,8,8,48],'leftwards rounds the left');
  assert.deepEqual(values({left:0,right:1,top:1,bottom:2}),[8,8,48,48],'downwards rounds the bottom');
  assert.deepEqual(values({left:.25,right:.75,top:0,bottom:1}),[8,8,8,8],'compression flattens corners');
  assert.deepEqual(values(from),[28,28,28,28]);
});
