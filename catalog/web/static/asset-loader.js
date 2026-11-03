// The host page owns core controls. Load optional behavior once, including after
// HTMX swaps. Component styles remain in the head for the document lifetime.
(() => {
  const script = document.currentScript,
    sources = JSON.parse(script.dataset.sources),
    pending = new Map();
  const rules = [
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
  function load(name) {
    if (pending.has(name)) return pending.get(name);
    const promise = new Promise((resolve, reject) => {
      const node = document.createElement("script");
      node.src = new URL(sources[name], script.src);
      node.async = false;
      node.dataset.componentScript = "";
      if (name === "loading.js")
        node.dataset.shapesVersion = script.dataset.shapesVersion;
      node.onload = resolve;
      node.onerror = () => {
        node.remove();
        pending.delete(name);
        reject(new Error(`Could not load ${name}`));
      };
      document.head.append(node);
    });
    pending.set(name, promise);
    return promise;
  }
  const styles = new Map();
  function stylesheet(link) {
    const href = new URL(link.getAttribute('href'), document.baseURI).href;
    if (styles.has(href)) return styles.get(href);
    const ready = link.dataset.stylesheetError !== undefined ? Promise.reject(new Error(`Could not load ${href}`)) : link.sheet ? Promise.resolve() : new Promise((resolve, reject) => {
      link.addEventListener('load', resolve, {once:true});
      link.addEventListener('error', () => reject(new Error(`Could not load ${href}`)), {once:true});
    });
    styles.set(href, ready);
    ready.catch(() => { styles.delete(href); link.remove(); });
    return ready;
  }
  function stylesReady() {
    return Promise.all([...document.head.querySelectorAll('link[data-component-stylesheet]')].map(stylesheet));
  }
  function stylesChanged(incoming) {
    const existing = new Map([...document.head.querySelectorAll('link[data-component-stylesheet]')].map(link => {
      const url = new URL(link.getAttribute('href'), document.baseURI); return [url.pathname, url.href];
    }));
    return [...incoming.querySelectorAll('link[data-component-stylesheet]')].some(link => {
      const url = new URL(link.getAttribute('href'), document.baseURI);
      return existing.has(url.pathname) && existing.get(url.pathname) !== url.href;
    });
  }
  async function prepareStyles(incoming, signal) {
    const ready = [];
    for (const source of incoming.querySelectorAll('link[data-component-stylesheet]')) {
      const href = new URL(source.getAttribute('href'), document.baseURI).href;
      const existing = [...document.head.querySelectorAll('link[data-component-stylesheet]')].find(link => link.href === href);
      if (existing) { ready.push(stylesheet(existing)); continue; }
      const link = source.cloneNode(false);
      for (const attr of ['media','onload','onerror']) link.removeAttribute(attr);
      ready.push(stylesheet(link));
      // Manifest order controls the cascade, independently of navigation order.
      const next = [...document.head.querySelectorAll('link[data-component-stylesheet]')].find(node => node.dataset.componentOrder > link.dataset.componentOrder);
      document.head.insertBefore(link, next || null);
    }
    await Promise.all(ready);
    signal?.throwIfAborted();
  }
  const api = { ready: Promise.resolve(), prepareStyles, stylesChanged };
  globalThis.TokenAssets = api;
  function initialize() {
    const preview = document.getElementById("component-preview");
    // A newly inserted playground must not accept input before its behavior exists.
    if (preview) {
      preview.inert = true;
      preview.setAttribute("aria-busy", "true");
    }
    const task = (async () => {
      try {
        await stylesReady();
        document.dispatchEvent(new CustomEvent("tokens:styles-ready"));
      } catch (error) {
        document.dispatchEvent(new CustomEvent("tokens:styles-error"));
        throw error;
      }
      for (const [name, selector] of rules) {
        const needed = [...document.querySelectorAll(selector)].some(
          (node) =>
            name === "progress.js" ||
            !node.closest("[inert]") ||
            preview?.contains(node),
        );
        if (needed) await load(name);
      }
      if (preview?.isConnected) {
        preview.inert = false;
        preview.removeAttribute("aria-busy");
        preview.querySelector(".tokens-preview-grid").inert = false;
        preview.querySelector("[data-assets-error]")?.remove();
      }
      document.dispatchEvent(new CustomEvent("tokens:assets-ready"));
    })();
    api.ready = task;
    task.catch((error) => {
      console.error(error);
      if (preview?.isConnected) {
        preview.inert = false;
        preview.removeAttribute("aria-busy");
        preview.querySelector(".tokens-preview-grid").inert = true;
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
