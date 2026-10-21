(() => {
  if (typeof document === "undefined") return;
  const { navigationIndex } = globalThis.MaterialComponents;
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("[data-group-button]");
    if (!button || button.disabled) return;
    const group = button.closest(".md-button-group");
    if (
      group.dataset.control === "connected" &&
      group.dataset.shape === "single"
    )
      for (const item of group.querySelectorAll("button"))
        item.setAttribute("aria-pressed", String(item === button));
    button.dispatchEvent(
      new CustomEvent("md-button-group-action", { bubbles: true }),
    );
  });
  document.addEventListener("keydown", (event) => {
    const button = event.target.closest?.("[data-group-button]");
    if (!button) return;
    const group = button.closest(".md-button-group"),
      items = [...group.querySelectorAll("button:not(:disabled)")];
    const next = navigationIndex(
      items.indexOf(button),
      items.length,
      event.key,
      true,
      getComputedStyle(group).direction === "rtl",
    );
    if (next !== null) {
      event.preventDefault();
      items[next]?.focus();
    }
  });
})();
