(() => {
  if (typeof document === "undefined") return;
  function positionTooltip(tip) {
    if (CSS.supports("top", "anchor(bottom)") || !tip.matches(":popover-open"))
      return;
    const trigger = tip.closest(".md-tooltip-demo")?.querySelector("button");
    if (!trigger) return;
    const anchor = trigger.getBoundingClientRect(),
      box = tip.getBoundingClientRect();
    const gap = Number.parseFloat(
      getComputedStyle(tip).getPropertyValue("--md-sys-measurement-space100"),
    );
    const width = document.documentElement.clientWidth,
      height = document.documentElement.clientHeight;
    tip.style.margin = "0";
    tip.style.left = `${Math.max(gap, Math.min(anchor.left, width - box.width - gap))}px`;
    tip.style.top = `${Math.max(gap, anchor.bottom + box.height + gap <= height ? anchor.bottom + gap : anchor.top - box.height - gap)}px`;
    tip.setAttribute("data-positioned", "");
  }
  document.addEventListener(
    "toggle",
    (event) => {
      if (event.target.matches?.(".md-tooltip[popover]"))
        positionTooltip(event.target);
    },
    true,
  );
  for (const type of ["resize", "scroll"])
    window.addEventListener(
      type,
      () => {
        for (const tip of document.querySelectorAll(".md-tooltip:popover-open"))
          positionTooltip(tip);
      },
      true,
    );
  document.addEventListener("click", (event) => {
    event.target
      .closest?.("[data-tooltip-close]")
      ?.closest(".md-tooltip[popover]")
      ?.hidePopover();
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && event.target instanceof Element)
      event.target
        .closest(".md-tooltip-demo")
        ?.setAttribute("data-tooltip-dismissed", "");
  });
  for (const type of ["pointerover", "focusin"])
    document.addEventListener(type, (event) => {
      if (!(event.target instanceof Element)) return;
      const demo = event.target.closest(".md-tooltip-demo");
      if (demo && !demo.contains(event.relatedTarget))
        demo.removeAttribute("data-tooltip-dismissed");
    });
})();
