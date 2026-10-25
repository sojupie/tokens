(() => {
  if (typeof document === "undefined") return;
  const { snackbarTimer } = globalThis.MaterialComponents;
  const timers = new WeakMap(), openers = new WeakMap();
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
      });
      bar.addEventListener("pointerenter", () => timer.pause("pointer"));
      bar.addEventListener("pointerleave", () => timer.resume("pointer"));
      bar.addEventListener("focusin", () => timer.pause("focus"));
      bar.addEventListener("focusout", (event) => {
        if (!bar.contains(event.relatedTarget)) timer.resume("focus");
      });
    }
  }
  initialize();
  document.addEventListener("htmx:afterSwap", () => initialize());
  document.addEventListener("htmx:beforeCleanupElement", (event) => {
    const root = event.detail.elt;
    if (root.matches?.(".md-snackbar[popover]")) timers.get(root)?.stop();
    for (const bar of root.querySelectorAll?.(".md-snackbar[popover]") || [])
      timers.get(bar)?.stop();
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("button"),
      demo = button?.closest("[data-snackbar-demo]");
    if (!demo) return;
    const bar = demo.querySelector(".md-snackbar");
    if (button.hasAttribute("data-snackbar-open")) {
      openers.set(bar, button);
      bar.showPopover();
      const inline = demo.querySelector("[data-snackbar-inline]");
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
