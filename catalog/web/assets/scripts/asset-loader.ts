type BehaviorRoot = ParentNode;
type AssetAudit = { declared: string[]; discovered: string[]; undeclared: string[]; declarationOnly: string[] };
type AssetAPI = {
  ready: Promise<void>;
  prepareStyles(incoming: BehaviorRoot, signal?: AbortSignal): Promise<void>;
  stylesChanged(incoming: BehaviorRoot): boolean;
  audit(root?: BehaviorRoot): AssetAudit;
};
declare var TokenAssets: AssetAPI;
declare var TokenBenchmark: { mark(name: string): void } | undefined;

// The host page owns core controls. Load optional behavior once, including after
// HTMX swaps. Component styles remain in the head for the document lifetime.
(() => {
  const rules: [string, string][] = [
    ["runtime/chips.js", '[data-family="chip"]'],
    ["runtime/tabs.js", ".md-tabs-example"],
    ["runtime/button-groups.js", ".md-button-group"],
    ["runtime/segmented-buttons.js", ".md-segmented-button"],
    ["runtime/switches.js", ".md-switch"],
    // Site navigation loads with the core layout, before optional controls.
    ["runtime/sliders.js", ".md-slider"],
    ["runtime/toolbars.js", ".md-toolbar"],
    ["runtime/app-bars.js", "[data-appbar-demo],[data-appbar-scroll-target]"],
    ["runtime/snackbars.js", ".md-snackbar"],
    ["runtime/fab-menus.js", ".md-fab-menu"],
    ["runtime/split-buttons.js", ".md-split-demo"],
    ["runtime/tooltips.js", ".md-tooltip-demo"],
    ["progress.js", ".md-progress"],
    ["sheets.js", ".md-sheet"],
    ["lists.js", ".md-list"],
    ["carousel.js", ".md-carousel"],
    ["search.js", ".md-search-demo"],
    ["date-picker.js", ".md-date-demo"],
    ["time-picker.js", ".md-time-demo,[data-time-picker]"],
    ["dialogs.js", "dialog[data-dialog-surface],dialog[data-dialog-preview]"],
    ["dialog-demo.js", ".md-dialog-demo"],
    ["loading.js", ".md-loading"],
    ["artwork-canvas.js", '[data-artwork="canvas"]'],
    ["artwork.js", "[data-artwork]"],
    ["component-demos.js", "#component-preview"],
  ];
  function behaviorNeeded(root: BehaviorRoot, name: string, selector: string) {
    const preview = root.querySelector('#component-preview');
    return [...root.querySelectorAll(selector)].some(node =>
      name === 'progress.js' || !node.closest('[inert]') || preview?.contains(node),
    );
  }
  function discoverBehaviors(root: BehaviorRoot) {
    return rules.filter(([name, selector]) => behaviorNeeded(root, name, selector)).map(([name]) => name);
  }
  function declaredBehaviors(root: BehaviorRoot) {
    return [...new Set([...root.querySelectorAll('script[data-component-behaviors]')]
      .flatMap(node => {
        const names = JSON.parse(node.textContent || "null") as unknown;
        if (!Array.isArray(names) || names.some(name => typeof name !== 'string'))
          throw new Error('Invalid component behavior declaration');
        return names as string[];
      }))];
  }
  function auditBehaviors(root: BehaviorRoot) {
    const declared = new Set(declaredBehaviors(root)),
      discovered = new Set(discoverBehaviors(root)),
      eager = [...root.querySelectorAll('script[src]:not([data-component-script])')]
        .map(node => (node.getAttribute('src') || '').split('?')[0]);
    for (const name of new Set([...declared, ...discovered]))
      if (eager.some(url => url.endsWith('/' + name))) {
        declared.delete(name); discovered.delete(name);
      }
    return {
      declared: [...declared].sort(), discovered: [...discovered].sort(),
      undeclared: [...discovered].filter(name => !declared.has(name)).sort(),
      declarationOnly: [...declared].filter(name => !discovered.has(name)).sort(),
    };
  }
  if (typeof module !== 'undefined') module.exports = { declaredBehaviors, discoverBehaviors, auditBehaviors };
  if (typeof document === 'undefined') return;
  const script = document.currentScript as HTMLScriptElement,
    sources: Record<string, string> = JSON.parse(script.dataset.sources || "{}"),
    order = Object.keys(sources),
    pending = new Map<string, Promise<void>>(),
    preloads = new Map<string, HTMLLinkElement>();
  function requirements(root: BehaviorRoot) {
    const eager = [...document.querySelectorAll('script[src]:not([data-component-script])')]
      .map(node => new URL(node.getAttribute('src') || '', document.baseURI).pathname);
    const names = declaredBehaviors(root).filter(name =>
      !eager.some(path => path.endsWith('/' + name)));
    for (const name of names)
      if (!Object.hasOwn(sources, name)) throw new Error(`Unknown component behavior: ${name}`);
    return names.sort((a, b) => order.indexOf(a) - order.indexOf(b));
  }
  function preload(root: BehaviorRoot) {
    for (const name of requirements(root)) {
      if (pending.has(name) || preloads.has(name)) continue;
      const link = document.createElement('link');
      link.rel = 'preload';
      link.as = 'script';
      link.href = new URL(sources[name], script.src).href;
      link.onerror = () => { link.remove(); if (preloads.get(name) === link) preloads.delete(name); };
      preloads.set(name, link);
      document.head.append(link);
    }
  }
  function load(name: string) {
    const existing = pending.get(name);
    if (existing) return existing;
    const promise = new Promise<void>((resolve, reject) => {
      const node = document.createElement("script");
      node.src = new URL(sources[name], script.src).href;
      node.async = false;
      node.dataset.componentScript = "";
      if (name === "loading.js")
        node.dataset.shapesVersion = script.dataset.shapesVersion;
      node.onload = () => resolve();
      node.onerror = () => {
        node.remove();
        pending.delete(name);
        preloads.get(name)?.remove();
        preloads.delete(name);
        reject(new Error(`Could not load ${name}`));
      };
      document.head.append(node);
    });
    pending.set(name, promise);
    return promise;
  }
  const styles = new Map<string, Promise<void>>();
  function stylesheet(link: HTMLLinkElement) {
    const href = new URL(link.getAttribute('href') || '', document.baseURI).href;
    const existing = styles.get(href);
    if (existing) return existing;
    const ready = link.dataset.stylesheetError !== undefined ? Promise.reject(new Error(`Could not load ${href}`)) : link.sheet ? Promise.resolve() : new Promise<void>((resolve, reject) => {
      link.addEventListener('load', () => resolve(), {once:true});
      link.addEventListener('error', () => reject(new Error(`Could not load ${href}`)), {once:true});
    });
    styles.set(href, ready);
    ready.catch(() => { styles.delete(href); link.remove(); });
    return ready;
  }
  function stylesReady() {
    return Promise.all([...document.head.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')].map(stylesheet));
  }
  function stylesChanged(incoming: BehaviorRoot) {
    const existing = new Map([...document.head.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')].map(link => {
      const url = new URL(link.getAttribute('href') || '', document.baseURI); return [url.pathname, url.href];
    }));
    return [...incoming.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')].some(link => {
      const url = new URL(link.getAttribute('href') || '', document.baseURI);
      return existing.has(url.pathname) && existing.get(url.pathname) !== url.href;
    });
  }
  async function prepareStyles(incoming: BehaviorRoot, signal?: AbortSignal) {
    signal?.throwIfAborted();
    // The response identifies the next region's scripts before CSS is ready
    // and before insertion. Download them now; activation still follows the swap.
    preload(incoming);
    const ready = [];
    for (const source of incoming.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')) {
      const href = new URL(source.getAttribute('href') || '', document.baseURI).href;
      const existing = [...document.head.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')].find(link => link.href === href);
      if (existing) { ready.push(stylesheet(existing)); continue; }
      const link = source.cloneNode(false) as HTMLLinkElement;
      for (const attr of ['media','onload','onerror']) link.removeAttribute(attr);
      ready.push(stylesheet(link));
      // Manifest order controls the cascade, independently of navigation order.
      const next = [...document.head.querySelectorAll<HTMLLinkElement>('link[data-component-stylesheet]')].find(node => (node.dataset.componentOrder || "") > (link.dataset.componentOrder || ""));
      document.head.insertBefore(link, next || null);
    }
    await Promise.all(ready);
    signal?.throwIfAborted();
  }
  const api: AssetAPI = { ready: Promise.resolve(), prepareStyles, stylesChanged, audit: (root = document) => auditBehaviors(root) };
  globalThis.TokenAssets = api;
  let generation = 0;
  function initialize() {
    const current = ++generation;
    const preview = document.getElementById("component-preview");
    // A newly inserted playground must not accept input before its behavior exists.
    if (preview) {
      preview.inert = true;
      preview.setAttribute("aria-busy", "true");
    }
    const task = (async () => {
      try {
        // Fetch together without executing controllers before their CSS or
        // prerequisites. Rendered declarations also cover generated child markup.
        preload(document);
        globalThis.TokenBenchmark?.mark("styles-wait-start");
        await stylesReady();
        globalThis.TokenBenchmark?.mark("styles-wait-end");
        if (current !== generation) return;
        document.dispatchEvent(new CustomEvent("tokens:styles-ready"));
      } catch (error) {
        if (current === generation) document.dispatchEvent(new CustomEvent("tokens:styles-error"));
        throw error;
      }
      for (const name of requirements(document)) {
        if (current !== generation) return;
        globalThis.TokenBenchmark?.mark("script:" + name + ":start");
        await load(name);
        globalThis.TokenBenchmark?.mark("script:" + name + ":end");
      }
      if (current !== generation) return;
      if (preview?.isConnected) {
        preview.inert = false;
        preview.removeAttribute("aria-busy");
        const grid = preview.querySelector<HTMLElement>(".tokens-preview-grid");
        if (grid) grid.inert = false;
        preview.querySelector("[data-assets-error]")?.remove();
      }
      document.dispatchEvent(new CustomEvent("tokens:assets-ready"));
    })();
    api.ready = task;
    task.catch((error) => {
      if (current !== generation) return;
      console.error(error);
      if (preview?.isConnected) {
        preview.inert = false;
        preview.removeAttribute("aria-busy");
        const grid = preview.querySelector<HTMLElement>(".tokens-preview-grid");
        if (grid) grid.inert = true;
        if (!preview.querySelector("[data-assets-error]")) {
          const message = document.createElement("p");
          message.dataset.assetsError = "";
          message.setAttribute("role", "alert");
          message.textContent =
            "Some preview controls could not load. Reload this page to try again.";
          preview.append(message);
        }
      }
    });
  }
  document.addEventListener("htmx:afterSwap", initialize);
  document.addEventListener("htmx:historyRestore", initialize);
  initialize();
})();
