(() => {
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
    const anchor = trigger.getBoundingClientRect(), box = tip.getBoundingClientRect(),
      gap = Number.parseFloat(getComputedStyle(tip).getPropertyValue("--md-sys-measurement-space100")) || 8,
      width = document.documentElement.clientWidth, height = document.documentElement.clientHeight;
    // Snap the correction, rather than the anchor, to Material's 8dp increments.
    let left = anchor.left, top = anchor.bottom + gap;
    if (left + box.width > width - gap) left -= Math.ceil((left + box.width - width + gap) / gap) * gap;
    if (left < gap) left += Math.ceil((gap - left) / gap) * gap;
    if (top + box.height > height - gap) top = anchor.top - box.height - gap;
    tip.style.margin = "0";
    tip.style.left = `${left}px`;
    tip.style.top = `${top}px`;
    tip.setAttribute("data-positioned", "");
  }
  function init() {
    for (const demo of document.querySelectorAll('.md-tooltip-demo')) demo.setAttribute('data-tooltip-managed', '');
  }
  document.addEventListener("toggle", event => {
    if (event.target.matches?.('.md-tooltip[popover]:popover-open')) { hide(); positionTooltip(event.target); }
  }, true);
  for (const type of ["resize", "scroll"]) window.addEventListener(type, () => {
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
    pointerFocus = event.pointerType === 'touch';
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
