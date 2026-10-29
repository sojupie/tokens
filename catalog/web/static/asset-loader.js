// The host page owns core controls. Load optional behavior once, including after
// HTMX swaps; native body stylesheets follow the lifetime of the page markup.
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
    ["runtime/app-bars.js", "[data-appbar-demo]"],
    ["runtime/snackbars.js", "[data-snackbar-demo]"],
    ["runtime/fab-menus.js", ".md-fab-menu"],
    ["runtime/split-buttons.js", ".md-split-demo"],
    ["runtime/tooltips.js", ".md-tooltip-demo"],
    ["progress.js", ".md-progress"],
    ["sheets.js", ".md-sheet"],
    ["lists.js", ".md-list"],
    ["carousel.js", ".md-carousel"],
    ["search.js", ".md-search-demo"],
    ["date-picker.js", ".md-date-demo"],
    ["time-picker.js", ".md-time-demo"],
    ["dialogs.js", ".md-dialog-demo"],
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
  function stylesReady() {
    return Promise.all(
      [...document.querySelectorAll("link[data-component-stylesheet]")].map(
        (link) => {
          if (link.dataset.stylesheetError !== undefined) return Promise.reject(new Error("Could not load component styles"));
          if (link.sheet) return Promise.resolve();
          return new Promise((resolve, reject) => {
            link.addEventListener("load", resolve, { once: true });
            link.addEventListener(
              "error",
              () => reject(new Error("Could not load component styles")),
              { once: true },
            );
          });
        },
      ),
    );
  }
  const api = { ready: Promise.resolve() };
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
