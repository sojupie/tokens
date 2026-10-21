(() => {
  if (typeof document === "undefined") return;
  const { navigationIndex, splitMenuPosition } = globalThis.MaterialComponents;
  function positionSplitMenu(menu) {
    if (!menu.matches(":popover-open")) return;
    const trigger = menu
      .closest(".md-split-demo")
      .querySelector(".md-split-trailing");
    const r = trigger.getBoundingClientRect(),
      box = menu.getBoundingClientRect();
    const style = getComputedStyle(trigger),
      gap = Number.parseFloat(
        style.getPropertyValue("--md-sys-measurement-space50"),
      ),
      margin = Number.parseFloat(
        style.getPropertyValue("--md-sys-measurement-space100"),
      );
    const pos = splitMenuPosition(
      r,
      trigger.closest(".md-split-button").getBoundingClientRect(),
      box,
      {
        width: document.documentElement.clientWidth,
        height: window.innerHeight,
      },
      style.direction === "rtl",
      gap,
      margin,
    );
    menu.style.left = pos.left + "px";
    menu.style.top = pos.top + "px";
  }
  document.addEventListener(
    "toggle",
    (event) => {
      const menu = event.target;
      if (!menu.matches?.(".md-split-menu")) return;
      menu
        .closest(".md-split-demo")
        .querySelector(".md-split-trailing")
        .setAttribute("aria-expanded", String(menu.matches(":popover-open")));
      positionSplitMenu(menu);
      if (menu.matches(":popover-open")) {
        const items = menu.querySelectorAll("button");
        items[
          menu.dataset.initialFocus === "last" ? items.length - 1 : 0
        ]?.focus();
        delete menu.dataset.initialFocus;
      }
    },
    true,
  );
  document.addEventListener("keydown", (event) => {
    const trigger = event.target.closest?.(".md-split-trailing");
    if (
      trigger &&
      !trigger.disabled &&
      ["ArrowDown", "ArrowUp"].includes(event.key)
    ) {
      event.preventDefault();
      const menu = document.getElementById(
        trigger.getAttribute("popovertarget"),
      );
      menu.dataset.initialFocus = event.key === "ArrowUp" ? "last" : "first";
      menu.showPopover();
      return;
    }
    const item = event.target.closest?.(".md-split-menu button");
    if (!item) return;
    if (event.key === "Tab") {
      const demo = item.closest(".md-split-demo");
      item.closest(".md-split-menu").hidePopover();
      demo.querySelector(".md-split-trailing").focus();
      return;
    }
    const items = [
      ...item
        .closest(".md-split-menu")
        .querySelectorAll("button:not(:disabled)"),
    ];
    const next = navigationIndex(
      items.indexOf(item),
      items.length,
      event.key,
      false,
    );
    if (next !== null) {
      event.preventDefault();
      items[next]?.focus();
    }
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.(
      "[data-split-primary],[data-split-option]",
    );
    if (!button) return;
    const demo = button.closest(".md-split-demo");
    button.dispatchEvent(new CustomEvent("md-split-action", { bubbles: true }));
    const menu = button.closest(".md-split-menu");
    if (menu) {
      menu.hidePopover();
      demo.querySelector(".md-split-trailing").focus();
    }
  });
  for (const type of ["resize", "scroll"])
    window.addEventListener(
      type,
      () => {
        for (const menu of document.querySelectorAll(
          ".md-split-menu:popover-open",
        ))
          positionSplitMenu(menu);
      },
      true,
    );
})();
