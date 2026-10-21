(() => {
  if (typeof document === "undefined") return;
  const { navigationIndex } = globalThis.MaterialComponents;
  function positionFABMenu(panel) {
    if (!panel.matches(":popover-open")) return;
    const trigger = panel
      .closest(".md-fab-menu")
      .querySelector(".md-fab-menu-opener");
    const anchor = trigger.getBoundingClientRect(),
      box = panel.getBoundingClientRect();
    const gap = Number.parseFloat(
        getComputedStyle(panel).getPropertyValue(
          "--md-sys-measurement-space200",
        ),
      ),
      closeGap = Number.parseFloat(
        getComputedStyle(panel).getPropertyValue("--_fab-menu-close-gap"),
      );
    const left =
      getComputedStyle(panel).direction === "rtl"
        ? anchor.left
        : anchor.right - box.width;
    panel.style.left =
      Math.max(
        gap,
        Math.min(left, document.documentElement.clientWidth - box.width - gap),
      ) + "px";
    panel.style.top = Math.max(gap, anchor.top - closeGap - box.height) + "px";
  }
  document.addEventListener(
    "toggle",
    (event) => {
      const panel = event.target;
      if (!panel.matches?.(".md-fab-menu-panel[popover]")) return;
      const trigger = panel
        .closest(".md-fab-menu")
        .querySelector(".md-fab-menu-opener");
      const opening = panel.matches(":popover-open");
      trigger.classList.toggle("md-fab-menu-close", opening);
      trigger.setAttribute("aria-expanded", String(opening));
      trigger.setAttribute("aria-label", opening ? "Close menu" : "Create");
      if (panel.matches(":popover-open")) {
        positionFABMenu(panel);
        const items = panel.querySelectorAll('[role="menuitem"]');
        if (panel.dataset.initialFocus)
          items[
            panel.dataset.initialFocus === "last" ? items.length - 1 : 0
          ]?.focus();
        else trigger.focus({ preventScroll: true });
        delete panel.dataset.initialFocus;
      } else if (
        panel.contains(document.activeElement) ||
        document.activeElement === document.body
      )
        trigger.focus({ preventScroll: true });
    },
    true,
  );
  document.addEventListener("keydown", (event) => {
    const trigger = event.target.closest?.(".md-fab-menu-opener");
    if (
      trigger &&
      event.key === "Tab" &&
      trigger.getAttribute("aria-expanded") === "true"
    )
      document
        .getElementById(trigger.getAttribute("popovertarget"))
        .hidePopover();
    if (trigger && ["ArrowDown", "ArrowUp"].includes(event.key)) {
      event.preventDefault();
      const panel = document.getElementById(
        trigger.getAttribute("popovertarget"),
      );
      const edge = event.key === "ArrowUp" ? "last" : "first";
      if (panel.matches(":popover-open")) {
        const items = panel.querySelectorAll('[role="menuitem"]');
        items[edge === "last" ? items.length - 1 : 0]?.focus();
      } else {
        panel.dataset.initialFocus = edge;
        panel.showPopover();
      }
      return;
    }
    const panel = event.target.closest?.(".md-fab-menu-panel[popover]");
    if (!panel) return;
    const items = [...panel.querySelectorAll('[role="menuitem"]')],
      index = items.indexOf(event.target.closest('[role="menuitem"]'));
    if (event.key === "Tab") {
      panel.hidePopover();
      panel
        .closest(".md-fab-menu")
        .querySelector(".md-fab-menu-opener")
        .focus();
      return;
    }
    const next = navigationIndex(index, items.length, event.key, false);
    if (next !== null) {
      event.preventDefault();
      items[next]?.focus();
    }
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("[data-fab-menu-action]");
    if (!button) return;
    const root = button.closest(".md-fab-menu"),
      panel = button.closest(".md-fab-menu-panel");
    button.dispatchEvent(
      new CustomEvent("md-fab-menu-select", { bubbles: true }),
    );
    if (panel.matches("[popover]")) {
      panel.hidePopover();
      root.querySelector(".md-fab-menu-opener").focus();
    }
  });
  for (const type of ["resize", "scroll"])
    window.addEventListener(
      type,
      () => {
        for (const panel of document.querySelectorAll(
          ".md-fab-menu-panel:popover-open",
        ))
          positionFABMenu(panel);
      },
      true,
    );
})();
