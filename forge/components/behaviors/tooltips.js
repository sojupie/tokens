(() => {
  function tooltipPosition(anchor, box, viewport, gap = 8) {
    const fit = (value, size, edge) => {
      if (value + size > edge - gap) value -= Math.ceil((value + size - edge + gap) / gap) * gap;
      if (value < gap) value += Math.ceil((gap - value) / gap) * gap;
      return value;
    };
    const candidates = [[anchor.right, anchor.bottom + gap], [anchor.right, anchor.top - box.height - gap],
      [anchor.left - box.width, anchor.bottom + gap], [anchor.left - box.width, anchor.top - box.height - gap],
      [anchor.right + gap, anchor.top], [anchor.left - box.width - gap, anchor.top]];
    for (const [x,y] of candidates) {
      const left = fit(x, box.width, viewport.width), top = fit(y, box.height, viewport.height);
      const overlap = left < anchor.right && left + box.width > anchor.left && top < anchor.bottom && top + box.height > anchor.top;
      if (!overlap && left >= gap && top >= gap && left + box.width <= viewport.width - gap && top + box.height <= viewport.height - gap) return {left, top};
    }
    return null;
  }
  if (typeof module !== "undefined") Object.assign(module.exports, {tooltipPosition});
  if (typeof document === "undefined") return;
  let active, exitTimer, hold, pointerFocus = false;
  function hide() {
    clearTimeout(exitTimer);
    active?.removeAttribute("data-tooltip-visible");
    active = null;
  }
  function show(demo) {
    if (!demo.querySelector('.md-tooltip[data-variant="plain"]')) return;
    hide();
    for (const tip of document.querySelectorAll('.md-tooltip[popover]:popover-open')) tip.hidePopover();
    demo.removeAttribute("data-tooltip-dismissed");
    demo.setAttribute("data-tooltip-visible", "");
    active = demo;
  }
  function leave(demo) {
    if (active === demo) {
      if (!pointerFocus && demo.matches(":hover, :focus-within")) return;
      clearTimeout(exitTimer);
      exitTimer = setTimeout(hide, 1500);
    }
  }
  function cancelHold() {
    if (hold) clearTimeout(hold.timer);
    hold = null;
  }
  function positionTooltip(tip) {
    if (!tip.matches(":popover-open")) return;
    const trigger = tip.closest(".md-tooltip-demo")?.querySelector("button");
    if (!trigger) return;
    const anchor = trigger.getBoundingClientRect(),
      gap = Number.parseFloat(getComputedStyle(tip).getPropertyValue("--md-sys-measurement-space100")) || 8,
      viewport = {width: document.documentElement.clientWidth, height: document.documentElement.clientHeight};
    tip.style.maxHeight = "";
    tip.style.overflowY = "";
    let box = tip.getBoundingClientRect(), position = tooltipPosition(anchor, box, viewport, gap);
    if (!position) {
      // Preserve all content in a short window instead of covering the trigger.
      tip.style.maxHeight = Math.max(2 * gap, anchor.top - 2 * gap, viewport.height - anchor.bottom - 2 * gap) + "px";
      tip.style.overflowY = "auto";
      box = tip.getBoundingClientRect();
      position = tooltipPosition(anchor, box, viewport, gap);
    }
    if (!position) return;
    tip.style.margin = "0";
    tip.style.left = `${position.left}px`;
    tip.style.top = `${position.top}px`;
    tip.setAttribute("data-positioned", "");
  }
  function init() {
    for (const demo of document.querySelectorAll('.md-tooltip-demo')) demo.setAttribute('data-tooltip-managed', '');
  }
  document.addEventListener("toggle", event => {
    if (event.target.matches?.('.md-tooltip[popover]:popover-open')) { hide(); positionTooltip(event.target); }
  }, true);
  for (const type of ["resize", "scroll"]) window.addEventListener(type, event => {
    if (type === "scroll" && event.target.closest?.(".md-tooltip:popover-open")) return;
    for (const tip of document.querySelectorAll(".md-tooltip:popover-open")) positionTooltip(tip);
  }, true);
  document.addEventListener("click", event => {
    event.target.closest?.("[data-tooltip-close]")?.closest(".md-tooltip[popover]")?.hidePopover();
  });
  document.addEventListener("keydown", event => {
    pointerFocus = false;
    if (event.key === "Escape") { hide(); cancelHold(); }
  });
  for (const type of ["pointerover", "focusin"]) document.addEventListener(type, event => {
    const demo = event.target.closest?.(".md-tooltip-demo");
    if (!demo || demo.contains(event.relatedTarget) || event.pointerType === 'touch' || type === 'focusin' && pointerFocus) return;
    if (active === demo) clearTimeout(exitTimer); else show(demo);
  });
  for (const type of ["pointerout", "focusout"]) document.addEventListener(type, event => {
    const demo = event.target.closest?.(".md-tooltip-demo");
    if (demo && !demo.contains(event.relatedTarget)) leave(demo);
  });
  document.addEventListener("pointerdown", event => {
    pointerFocus = true;
    cancelHold();
    const demo = event.target.closest?.('.md-tooltip-demo');
    if (!demo || event.pointerType !== 'touch' || !event.isPrimary) return;
    // Material specifies tap-and-hold, without a web hold duration. Use 500ms.
    hold = { demo, id: event.pointerId, x: event.clientX, y: event.clientY,
      timer: setTimeout(() => show(demo), 500) };
  });
  document.addEventListener("pointermove", event => {
    if (hold && hold.id === event.pointerId && Math.hypot(event.clientX - hold.x, event.clientY - hold.y) > 8) cancelHold();
  });
  for (const type of ["pointerup", "pointercancel"]) document.addEventListener(type, event => {
    if (!hold || hold.id !== event.pointerId) return;
    const demo = hold.demo;
    cancelHold();
    if (type === 'pointercancel') hide(); else leave(demo);
  });
  document.addEventListener('contextmenu', event => {
    if (active && active === event.target.closest?.('.md-tooltip-demo') && pointerFocus) event.preventDefault();
  });
  document.addEventListener("htmx:beforeCleanupElement", event => {
    const root = event.detail.elt;
    if (active && root.contains(active)) hide();
    if (hold && root.contains(hold.demo)) cancelHold();
  });
  document.addEventListener('htmx:afterSwap', init);
  document.addEventListener('tokens:assets-ready', init);
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init); else init();
})();
