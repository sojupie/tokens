// Catalog example actions and announcements. Component mechanics live in Forge.
(() => {
  document.addEventListener("click", (event) => {
    const proxy = event.target.closest?.("[data-toolbar-proxy]");
    if (proxy && !proxy.disabled)
      proxy
        .closest(".md-toolbar-demo")
        .querySelector(
          '[data-toolbar-item="' + proxy.dataset.toolbarProxy + '"]',
        )
        .click();
    const tool = event.target.closest?.('[data-toolbar-item="0"]');
    if (tool)
      queueMicrotask(() => {
        const peer = tool
          .closest(".md-toolbar-demo")
          .querySelector('[data-toolbar-proxy="0"]');
        if (!peer) return;
        const selected = tool.getAttribute("aria-pressed") === "true";
        peer.setAttribute("aria-checked", String(selected));
        peer.querySelector("[data-menu-checked-icon]").hidden = !selected;
        peer.querySelector("[data-menu-unchecked-icon]").hidden = selected;
      });
  });
  document.addEventListener("md-navigation-select", (event) => {
    const group = event.target,
      label = event.detail.label,
      output = group.querySelector("output");
    if (output) output.textContent = label + " selected";
    const frame = group.closest(".md-rail-frame");
    if (!frame) return;
    frame.querySelector("[data-navigation-title]").textContent = label;
    frame.querySelector("[data-navigation-description]").textContent =
      "Your " + label.toLowerCase() + " in this workspace.";
    const screen = frame.querySelector(".md-rail-screen");
    screen.scrollTop = 0;
    if (!matchMedia("(prefers-reduced-motion: reduce)").matches)
      screen.animate(
        [{ opacity: 0 }, { opacity: 1 }],
        MaterialMotion.timing(group, "fast-effects"),
      );
  });
  document.addEventListener("click", (event) => {
    const button = event.target.closest?.("[data-navigation-create]"),
      frame = button?.closest(".md-rail-frame");
    if (frame)
      frame.querySelector("[data-navigation-title]").textContent =
        "Create file";
  });
  for (const type of ["md-chip-select", "md-chip-action", "md-chip-remove"])
    document.addEventListener(type, (event) => {
      const chip = event.target,
        demo = chip.closest(".md-chip-demo"),
        output = demo?.querySelector("[data-chip-output]");
      if (!output) return;
      if (type === "md-chip-select")
        output.textContent = event.detail.selected ? "Selected" : "Unselected";
      if (type === "md-chip-action")
        output.textContent = event.detail.label + " activated";
      if (type === "md-chip-remove") {
        output.textContent =
          "Removed " + chip.querySelector(".md-chip-label").textContent;
        demo
          .closest("figure")
          ?.closest("#component-preview")
          ?.querySelector("select,button")
          ?.focus({ preventScroll: true });
      }
    });
  document.addEventListener("md-button-group-action", (event) => {
    const button = event.target,
      group = button.closest(".md-button-group"),
      output = group.closest(".md-button-group-demo")?.querySelector("output");
    if (!output) return;
    const selected = [
      ...group.querySelectorAll('button[aria-pressed="true"]'),
    ].map((item) => item.textContent.trim());
    output.textContent =
      group.dataset.control === "standard"
        ? button.textContent.trim() + " activated."
        : selected.length
          ? selected.join(", ") + " selected."
          : "No selection.";
  });
  document.addEventListener("md-snackbar-action", (event) => {
    const button = event.target,
      output = button.closest("[data-snackbar-demo]")?.querySelector("output");
    if (output)
      output.textContent = button.hasAttribute("data-snackbar-open")
        ? ""
        : button.hasAttribute("data-snackbar-action")
          ? "Message undone."
          : "Notification dismissed.";
  });
  document.addEventListener("md-fab-menu-select", (event) => {
    const button = event.target,
      output = button.closest(".md-fab-menu")?.querySelector("output");
    if (output) output.textContent = button.textContent.trim() + " selected.";
  });
  document.addEventListener("md-split-action", (event) => {
    const button = event.target,
      output = button.closest(".md-split-demo")?.querySelector("output");
    if (output)
      output.textContent = button.hasAttribute("data-split-primary")
        ? "Added to cart."
        : button.textContent + ".";
  });
})();
