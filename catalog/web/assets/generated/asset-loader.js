/* Generated from catalog/web/assets/scripts/asset-loader.ts. DO NOT EDIT. Run npm run build:components. */
(()=>{"use strict";

// catalog/web/assets/scripts/asset-loader.ts
(() => {
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
    ["component-demos.js", "#component-preview"]
  ];
  function behaviorNeeded(root, name, selector) {
    const preview = root.querySelector("#component-preview");
    return [...root.querySelectorAll(selector)].some(
      (node) => name === "progress.js" || !node.closest("[inert]") || preview?.contains(node)
    );
  }
  function discoverBehaviors(root) {
    return rules.filter(([name, selector]) => behaviorNeeded(root, name, selector)).map(([name]) => name);
  }
  function declaredBehaviors(root) {
    return [...new Set([...root.querySelectorAll("script[data-component-behaviors]")].flatMap((node) => {
      const names = JSON.parse(node.textContent || "null");
      if (!Array.isArray(names) || names.some((name) => typeof name !== "string"))
        throw new Error("Invalid component behavior declaration");
      return names;
    }))];
  }
  function auditBehaviors(root) {
    const declared = new Set(declaredBehaviors(root)), discovered = new Set(discoverBehaviors(root)), eager = [...root.querySelectorAll("script[src]:not([data-component-script])")].map((node) => (node.getAttribute("src") || "").split("?")[0]);
    for (const name of /* @__PURE__ */ new Set([...declared, ...discovered]))
      if (eager.some((url) => url.endsWith("/" + name))) {
        declared.delete(name);
        discovered.delete(name);
      }
    return {
      declared: [...declared].sort(),
      discovered: [...discovered].sort(),
      undeclared: [...discovered].filter((name) => !declared.has(name)).sort(),
      declarationOnly: [...declared].filter((name) => !discovered.has(name)).sort()
    };
  }
  if (typeof module !== "undefined") module.exports = { declaredBehaviors, discoverBehaviors, auditBehaviors };
  if (typeof document === "undefined") return;
  const script = document.currentScript, sources = JSON.parse(script.dataset.sources || "{}"), order = Object.keys(sources), pending = /* @__PURE__ */ new Map(), preloads = /* @__PURE__ */ new Map();
  function requirements(root) {
    const eager = [...document.querySelectorAll("script[src]:not([data-component-script])")].map((node) => new URL(node.getAttribute("src") || "", document.baseURI).pathname);
    const names = declaredBehaviors(root).filter((name) => !eager.some((path) => path.endsWith("/" + name)));
    for (const name of names)
      if (!Object.hasOwn(sources, name)) throw new Error(`Unknown component behavior: ${name}`);
    return names.sort((a, b) => order.indexOf(a) - order.indexOf(b));
  }
  function preload(root) {
    for (const name of requirements(root)) {
      if (pending.has(name) || preloads.has(name)) continue;
      const link = document.createElement("link");
      link.rel = "preload";
      link.as = "script";
      link.href = new URL(sources[name], script.src).href;
      link.onerror = () => {
        link.remove();
        if (preloads.get(name) === link) preloads.delete(name);
      };
      preloads.set(name, link);
      document.head.append(link);
    }
  }
  function load(name) {
    const existing = pending.get(name);
    if (existing) return existing;
    const promise = new Promise((resolve, reject) => {
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
  const styles = /* @__PURE__ */ new Map();
  function stylesheet(link) {
    const href = new URL(link.getAttribute("href") || "", document.baseURI).href;
    const existing = styles.get(href);
    if (existing) return existing;
    const ready = link.dataset.stylesheetError !== void 0 ? Promise.reject(new Error(`Could not load ${href}`)) : link.sheet ? Promise.resolve() : new Promise((resolve, reject) => {
      link.addEventListener("load", () => resolve(), { once: true });
      link.addEventListener("error", () => reject(new Error(`Could not load ${href}`)), { once: true });
    });
    styles.set(href, ready);
    ready.catch(() => {
      styles.delete(href);
      link.remove();
    });
    return ready;
  }
  function stylesReady() {
    return Promise.all([...document.head.querySelectorAll("link[data-component-stylesheet]")].map(stylesheet));
  }
  function stylesChanged(incoming) {
    const existing = new Map([...document.head.querySelectorAll("link[data-component-stylesheet]")].map((link) => {
      const url = new URL(link.getAttribute("href") || "", document.baseURI);
      return [url.pathname, url.href];
    }));
    return [...incoming.querySelectorAll("link[data-component-stylesheet]")].some((link) => {
      const url = new URL(link.getAttribute("href") || "", document.baseURI);
      return existing.has(url.pathname) && existing.get(url.pathname) !== url.href;
    });
  }
  async function prepareStyles(incoming, signal) {
    signal?.throwIfAborted();
    preload(incoming);
    const ready = [];
    for (const source of incoming.querySelectorAll("link[data-component-stylesheet]")) {
      const href = new URL(source.getAttribute("href") || "", document.baseURI).href;
      const existing = [...document.head.querySelectorAll("link[data-component-stylesheet]")].find((link2) => link2.href === href);
      if (existing) {
        ready.push(stylesheet(existing));
        continue;
      }
      const link = source.cloneNode(false);
      for (const attr of ["media", "onload", "onerror"]) link.removeAttribute(attr);
      ready.push(stylesheet(link));
      const next = [...document.head.querySelectorAll("link[data-component-stylesheet]")].find((node) => (node.dataset.componentOrder || "") > (link.dataset.componentOrder || ""));
      document.head.insertBefore(link, next || null);
    }
    await Promise.all(ready);
    signal?.throwIfAborted();
  }
  const api = { ready: Promise.resolve(), prepareStyles, stylesChanged, audit: (root = document) => auditBehaviors(root) };
  globalThis.TokenAssets = api;
  let generation = 0;
  function initialize() {
    const current = ++generation;
    const preview = document.getElementById("component-preview");
    if (preview) {
      preview.inert = true;
      preview.setAttribute("aria-busy", "true");
    }
    const task = (async () => {
      try {
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
        const grid = preview.querySelector(".tokens-preview-grid");
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
        const grid = preview.querySelector(".tokens-preview-grid");
        if (grid) grid.inert = true;
        if (!preview.querySelector("[data-assets-error]")) {
          const message = document.createElement("p");
          message.dataset.assetsError = "";
          message.setAttribute("role", "alert");
          message.textContent = "Some preview controls could not load. Reload this page to try again.";
          preview.append(message);
        }
      }
    });
  }
  document.addEventListener("htmx:afterSwap", initialize);
  document.addEventListener("htmx:historyRestore", initialize);
  initialize();
})();
})();
