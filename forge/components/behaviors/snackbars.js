import { snackbarTimer } from "../helpers/controls.ts";

(() => {
  if (typeof document === "undefined") return;
  const timers = new WeakMap(), openers = new WeakMap();
  const obstacleSelector = ".md-fab, .md-navigation-bar, .md-toolbar[data-variant='docked'], [data-snackbar-avoid]";
  const observed = new Set();
  let layoutFrame = 0, watching = false;
  const resizeObserver = typeof ResizeObserver === "function" ? new ResizeObserver(queueLayout) : undefined;
  const mutationObserver = typeof MutationObserver === "function" ? new MutationObserver(records => {
    // Placement writes only this inline style; observing that write would keep
    // scheduling frames after the layout has settled.
    if (records.some(record => record.type !== "attributes" || record.attributeName !== "style" || !record.target.matches?.(".md-snackbar"))) queueLayout();
  }) : undefined;
  function openBars() { return [...document.querySelectorAll(".md-snackbar:popover-open")].filter(bar => bar.matches(":popover-open")); }
  function obstacles() {
    const targets = new Set(document.querySelectorAll(obstacleSelector));
    if (document.activeElement && document.activeElement !== document.body && document.activeElement !== document.documentElement) targets.add(document.activeElement);
    return targets;
  }
  function watchLayout(bars) {
    if (bars.length && !watching && document.body) {
      mutationObserver?.observe(document.body, { subtree: true, childList: true, characterData: true, attributes: true,
        attributeFilter: ["hidden", "open", "class", "style", "data-snackbar-avoid", "data-variant"] });
      watching = true;
    } else if (!bars.length) {
      mutationObserver?.disconnect(); watching = false;
      if (layoutFrame) { cancelAnimationFrame(layoutFrame); layoutFrame = 0; }
    }
    const targets = new Set(bars.length ? [...bars, ...obstacles()].filter(target => typeof target.getBoundingClientRect === "function") : []);
    for (const target of observed) if (!targets.has(target)) { resizeObserver?.unobserve(target); observed.delete(target); }
    for (const target of targets) if (!observed.has(target)) { resizeObserver?.observe(target); observed.add(target); }
  }
  function queueLayout() {
    if (!layoutFrame) layoutFrame = requestAnimationFrame(() => { layoutFrame = 0; reposition(); });
  }
  function position(bar) {
    if (!bar.matches(":popover-open")) return;
    bar.style.removeProperty("--_snackbar-avoid-offset");
    const base = bar.getBoundingClientRect(), gap = parseFloat(getComputedStyle(bar).getPropertyValue("--md-sys-measurement-space200")) || 16;
    const boxes = [...obstacles()].filter(element => element !== bar && !bar.contains(element) && element.getClientRects().length && getComputedStyle(element).visibility === "visible").map(element => element.getBoundingClientRect()).sort((a,b) => b.top-a.top);
    let bottom = base.bottom;
    for (const box of boxes) {
      const top = bottom-base.height;
      if (box.right>base.left && box.left<base.right && box.top<bottom && box.bottom>top) bottom=box.top-gap;
    }
    bar.style.setProperty("--_snackbar-avoid-offset", Math.max(0,base.bottom-bottom)+"px");
  }
  function reposition() { const bars = openBars(); watchLayout(bars); for (const bar of bars) position(bar); }
  function dismiss(bar) {
    const restore = bar.contains(document.activeElement);
    timers.get(bar)?.stop();
    bar.hidePopover();
    if (restore) openers.get(bar)?.focus({ preventScroll: true });
  }
  const clock = {
    now: () => performance.now(),
    setTimeout: (fn, delay) => setTimeout(fn, delay),
    clearTimeout: (id) => clearTimeout(id),
  };
  function initialize(root = document) {
    for (const bar of root.querySelectorAll(".md-snackbar[popover]")) {
      if (timers.has(bar)) continue;
      const timer = snackbarTimer(() => {
        if (bar.isConnected && bar.matches(":popover-open")) dismiss(bar);
      }, clock);
      timers.set(bar, timer);
      bar.addEventListener("toggle", () => {
        if (!bar.matches(":popover-open")) timer.stop();
        reposition();
      });
      bar.addEventListener("pointerenter", () => timer.pause("pointer"));
      bar.addEventListener("pointerleave", () => timer.resume("pointer"));
      bar.addEventListener("focusin", () => timer.pause("focus"));
      bar.addEventListener("focusout", (event) => {
        if (!bar.contains(event.relatedTarget)) timer.resume("focus");
      });
    }
  }
  // Applications can publish status without fabricating a hidden open button.
  function showSnackbar(bar, opener = document.activeElement, duration = 0) {
    initialize();
    openers.set(bar, opener);
    bar.showPopover();
    reposition();
    timers.get(bar).start(Number.isFinite(duration) && duration > 0 ? duration : 0);
  }
  globalThis.MaterialSnackbars = Object.freeze({ show: showSnackbar, dismiss });
  initialize();
  window.addEventListener("resize", reposition);
  window.addEventListener("scroll", reposition, true);
  document.addEventListener("focusin", reposition);
  document.addEventListener("htmx:afterSwap", () => initialize());
  document.addEventListener("htmx:beforeCleanupElement", (event) => {
    const root = (event.detail?.elt || event.target);
    if (root.matches?.(".md-snackbar[popover]")) timers.get(root)?.stop();
    for (const bar of root.querySelectorAll?.(".md-snackbar[popover]") || [])
      timers.get(bar)?.stop();
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("button"),
      demo = button?.closest("[data-snackbar-demo]");
    const bar = button?.closest(".md-snackbar") || demo?.querySelector(".md-snackbar");
    if (!bar) return;
    if (button.hasAttribute("data-snackbar-open")) {
      showSnackbar(bar, button);
      const inline = demo?.querySelector("[data-snackbar-inline]");
      if (inline) inline.textContent = bar.querySelector(".md-feedback-text").textContent.trim();
      // Web notifications must remain available either in the actionable bar
      // itself or as inline feedback after a non-actionable bar disappears.
      timers.get(bar).start(bar.querySelector("[data-snackbar-action]") || !inline ? 0 : Number(bar.dataset.snackbarDuration));
    } else if (
      button.hasAttribute("data-snackbar-close") ||
      button.hasAttribute("data-snackbar-action")
    ) {
      dismiss(bar);
    } else return;
    button.dispatchEvent(
      new CustomEvent("md-snackbar-action", { bubbles: true }),
    );
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "F6" && !event.altKey && !event.ctrlKey && !event.metaKey && !event.shiftKey) {
      const action = document.querySelector('.md-snackbar:popover-open [data-snackbar-action]:not(:disabled)');
      if (action) { event.preventDefault(); action.focus({ preventScroll: true }); }
      return;
    }
    if (event.key === "Escape")
      for (const bar of document.querySelectorAll(
        ".md-snackbar:popover-open",
      )) {
        dismiss(bar);
      }
  });
  document.addEventListener('tokens:assets-ready',()=>initialize());
})();
