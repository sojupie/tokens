(() => {
  if (typeof document === "undefined") return;
  const { navigationIndex } = globalThis.MaterialComponents;
  function selectNavigation(button) {
    const group = button.closest("[data-navigation-group]");
    if (!group) return;
    for (const item of group.querySelectorAll("[data-navigation-item]"))
      item.removeAttribute("aria-current");
    button.setAttribute("aria-current", "page");
    const label = button.querySelector(".md-nav-label").textContent;
    group.dispatchEvent(
      new CustomEvent("md-navigation-select", {
        bubbles: true,
        detail: { label },
      }),
    );
    const frame = group.closest(".md-rail-frame");
    if (frame?.dataset.modalOpen === "true") setRailExpanded(frame, false);
    group.closest("dialog[data-navigation-modal]")?.close();
  }
  document.addEventListener("keydown", (event) => {
    const button = event.target.closest?.("[data-navigation-item]");
    if (!button) return;
    const group = button.closest("[data-navigation-group]"),
      items = [...group.querySelectorAll("[data-navigation-item]")];
    const next = navigationIndex(
      items.indexOf(button),
      items.length,
      event.key,
      group.classList.contains("md-navigation-bar"),
      getComputedStyle(group).direction === "rtl",
    );
    if (next === null) return;
    event.preventDefault();
    items[next].focus();
  });
  const railFrames = new WeakMap(),
    railLabels = new WeakSet();
  function measureRailLabels(labels) {
    const sizes = Array.from(labels, (label) => ({
      item: label.closest(".md-nav-item"),
      height: label.getBoundingClientRect().height,
    }));
    for (const { item, height } of sizes)
      if (item && height)
        item.style.setProperty("--_nav-label-height", height + "px");
  }
  const railLabelSizes = new ResizeObserver((entries) =>
    measureRailLabels(entries.map((entry) => entry.target)),
  );
  function syncRail(frame) {
    const rail = frame.querySelector(".md-navigation-rail"),
      expanded = rail.dataset.variant === "expanded",
      modal = expanded && rail.dataset.layout === "modal",
      hidden = !expanded && frame.dataset.collapse === "hidden";
    frame.dataset.modalOpen = String(modal);
    frame.dataset.railHidden = String(hidden);
    frame.querySelector(".md-navigation-surface").hidden = hidden;
    const surface = frame.querySelector(".md-navigation-surface");
    if (modal) {
      surface.setAttribute("role", "dialog");
      surface.setAttribute("aria-label", "Expanded navigation");
    } else {
      surface.removeAttribute("role");
      surface.removeAttribute("aria-label");
    }
    frame.querySelector(".md-navigation-scrim").hidden = !modal;
    frame.querySelector(".md-rail-body").inert = modal;
    for (const menu of frame.querySelectorAll("[data-navigation-expand]")) {
      menu.setAttribute("aria-expanded", String(expanded));
      menu.setAttribute(
        "aria-label",
        expanded ? "Collapse navigation" : "Expand navigation",
      );
    }
    const external = frame.querySelector(".md-rail-external-menu");
    external.hidden =
      !hidden && !["fab", "none"].includes(frame.dataset.header);
    const fab = rail.querySelector(".md-nav-fab");
    if (fab) {
      fab.classList.toggle("md-fab", !expanded);
      fab.classList.toggle("md-extended-fab", expanded);
    }
  }
  function setRailExpanded(frame, expanded, focus = false) {
    const rail = frame.querySelector(".md-navigation-rail"),
      state = railFrames.get(frame),
      start = rail.getBoundingClientRect().width,
      wasModal = frame.dataset.modalOpen === "true";
    const slot = frame.querySelector(".md-navigation-slot"),
      body = frame.querySelector(".md-rail-body"),
      bodyBefore = body.getBoundingClientRect();
    const textTargets =
      ".md-nav-label,.md-nav-section,.md-nav-fab > span:not(.md-button-ripple)";
    const outgoing = [...rail.querySelectorAll(textTargets)]
      .filter((e) => e.getClientRects().length)
      .map((element) => {
        const style = getComputedStyle(element);
        return {
          text: element.textContent,
          rect: element.getBoundingClientRect(),
          font: style.font,
          color: style.color,
          spacing: style.letterSpacing,
          align: style.textAlign,
          opacity: style.opacity,
          direction: style.direction,
          section: element.matches(".md-nav-section"),
        };
      });
    const motionTargets =
      ".md-nav-indicator,.md-nav-icon,.md-nav-label,.md-rail-menu,.md-nav-fab,.md-nav-fab > span:not(.md-button-ripple)";
    const before = new Map(
      [...rail.querySelectorAll(motionTargets)]
        .filter((e) => e.getClientRects().length)
        .map((e) => [e, e.getBoundingClientRect()]),
    );
    for (const a of state?.animations || []) a.cancel();
    for (const copy of state?.copies || []) copy.remove();
    delete frame.dataset.transitioning;
    const fab = rail.querySelector(".md-nav-fab"),
      fabStart = before.get(fab)?.width;
    rail.dataset.variant = expanded ? "expanded" : "collapsed";
    rail.dataset.orientation = expanded ? "horizontal" : "vertical";
    syncRail(frame);
    // Resolve the new label metrics before capturing the FLIP destination.
    // ResizeObserver otherwise updates the CSS base transform mid-animation.
    measureRailLabels(rail.querySelectorAll(".md-nav-label"));
    const slotEnd = slot.getBoundingClientRect().width;
    const targetHidden = frame.dataset.railHidden === "true",
      end = targetHidden ? 0 : rail.getBoundingClientRect().width,
      timing = MaterialMotion.timing(rail, "default-spatial"),
      effects = MaterialMotion.timing(rail, "default-effects");
    const restoreFocus = () => {
      if (focus) {
        const target = targetHidden
          ? frame.querySelector(".md-rail-external-menu")
          : rail.querySelector(
              "[data-navigation-expand],.md-nav-fab,[data-navigation-item]",
            );
        target?.focus({ preventScroll: true });
      }
    };
    if (matchMedia("(prefers-reduced-motion: reduce)").matches) {
      restoreFocus();
      return;
    }
    if (targetHidden) {
      frame.querySelector(".md-navigation-surface").hidden = false;
      frame.dataset.railHidden = "false";
    }
    if (wasModal && !expanded) {
      frame.dataset.modalOpen = "true";
      frame.querySelector(".md-rail-body").inert = true;
      frame.querySelector(".md-navigation-scrim").hidden = false;
    }
    frame.style.setProperty(
      "--_rail-content-width",
      (targetHidden ? start : end) + "px",
    );
    frame.style.setProperty(
      "--_rail-content-align",
      getComputedStyle(rail).direction === "rtl" ? "flex-end" : "flex-start",
    );
    frame.style.setProperty("--_rail-slot-width", slotEnd + "px");
    frame.dataset.transitioning = "";
    const current = { animations: [], copies: [] };
    railFrames.set(frame, current);
    // Labels crossfade between layouts. Retain only the outgoing text, at its
    // current position and typography, while the real label takes its new slot.
    // Label phase follows the spatial transition, adapted from AndroidX
    // NavigationItem.kt, Copyright 2024 The Android Open Source Project,
    // Apache-2.0. See THIRD_PARTY_NOTICES.md.
    const style = getComputedStyle(rail),
      spring = MaterialMotion.springFrames(
        Number(
          style.getPropertyValue(
            "--md-sys-motion-spring-default-spatial-damping",
          ),
        ),
        Number(
          style.getPropertyValue(
            "--md-sys-motion-spring-default-spatial-stiffness",
          ),
        ),
      );
    const phases = spring.samples.map((progress, i) => ({
        progress,
        offset: i / (spring.samples.length - 1),
      })),
      textTiming = { duration: timing.duration, easing: "linear" },
      frameRect = frame.getBoundingClientRect(),
      sign = style.direction === "rtl" ? -1 : 1;
    const middle = phases.findIndex(({ progress }) => progress >= 0.5),
      a = phases[middle - 1],
      b = phases[middle];
    phases.splice(middle, 0, {
      progress: 0.5,
      offset:
        a.offset +
        ((0.5 - a.progress) / (b.progress - a.progress)) *
          (b.offset - a.offset),
    });
    for (const old of outgoing) {
      const copy = document.createElement("span");
      copy.textContent = old.text;
      copy.dataset.railTextExit = "";
      copy.setAttribute("aria-hidden", "true");
      copy.inert = true;
      if (old.section) copy.dataset.railSectionExit = "";
      Object.assign(copy.style, {
        position: "absolute",
        pointerEvents: "none",
        zIndex: "4",
        left: old.rect.x - frameRect.x - frame.clientLeft + "px",
        top: old.rect.y - frameRect.y - frame.clientTop + "px",
        width: old.rect.width + "px",
        height: old.rect.height + "px",
        font: old.font,
        color: old.color,
        letterSpacing: old.spacing,
        textAlign: old.align,
        direction: old.direction,
      });
      frame.append(copy);
      current.copies.push(copy);
      current.animations.push(
        copy.animate(
          old.section
            ? [{ opacity: old.opacity }, { opacity: 0 }]
            : phases.map(({ progress, offset }) => ({
                offset,
                opacity:
                  Number(old.opacity) *
                  Math.max(0, 1 - 2 * Math.max(0, progress)) ** 2,
              })),
          { ...(old.section ? effects : textTiming), fill: "forwards" },
        ),
      );
    }
    // Fix the final inner layout while the outer rail resizes. Move existing
    // contents from their current positions instead of reflowing every frame.
    // Capture every destination before starting any animation. An animated
    // parent must not contaminate the measured destination of its children.
    const destinations = new Map(
      [...rail.querySelectorAll(motionTargets)].map((element) => {
        const style = getComputedStyle(element);
        return [
          element,
          {
            rect: element.getBoundingClientRect(),
            left: parseFloat(style.left) || 0,
            top: parseFloat(style.top) || 0,
          },
        ];
      }),
    );
    const heading = rail.querySelector(".md-nav-section");
    if (heading?.getClientRects().length)
      current.animations.push(
        heading.animate([{ opacity: 0 }, { opacity: 1 }], effects),
      );
    const shifts = new Map();
    for (const [element, { rect: next, left, top }] of destinations) {
      const prior = before.get(element);
      if (!next.width || !next.height) continue;
      const parent = element.parentElement.closest(motionTargets),
        parentShift = shifts.get(parent) || { x: 0, y: 0 };
      const shift = prior
        ? { x: prior.x - next.x, y: prior.y - next.y }
        : { x: 0, y: 0 };
      shifts.set(element, shift);
      if (element.matches(textTargets)) {
        // Exit text stays fixed. On entry, the horizontal label approaches
        // from the icon while nearly transparent; its vertical slot is fixed.
        const parentRect = destinations.get(parent)?.rect,
          approach =
            expanded && !element.matches(".md-nav-section") && parentRect
              ? sign === 1
                ? next.left - parentRect.left
                : parentRect.right - next.right
              : 0;
        current.animations.push(
          element.animate(
            phases.map(({ progress: p, offset }) => ({
              offset,
              left:
                left -
                parentShift.x * (1 - p) +
                sign * approach * (p - 1) +
                "px",
              top: top - parentShift.y * (1 - p) + "px",
              opacity: Math.max(0, Math.min(1, 2 * p - 1)) ** 2,
            })),
            textTiming,
          ),
        );
        continue;
      }
      // Paint text at its moving position rather than promoting it into a
      // transformed texture that changes rasterization when the effect ends.
      current.animations.push(
        element.animate(
          prior
            ? [
                {
                  left: left + shift.x - parentShift.x + "px",
                  top: top + shift.y - parentShift.y + "px",
                },
                { left: left + "px", top: top + "px" },
              ]
            : [{ opacity: 0 }, { opacity: 1 }],
          prior ? timing : effects,
        ),
      );
    }
    const bodyEnd = body.getBoundingClientRect(),
      bodyShift = bodyBefore.x - bodyEnd.x;
    if (Math.abs(bodyShift) > 0.01)
      current.animations.push(
        body.animate([{ left: bodyShift + "px" }, { left: "0px" }], timing),
      );
    const animation = rail.animate(
      [
        { width: start + "px", minWidth: start + "px" },
        { width: end + "px", minWidth: end + "px" },
      ],
      { ...timing, fill: "both" },
    );
    current.animations.push(animation);
    if (fab)
      current.animations.push(
        fab.animate(
          [
            { width: fabStart + "px", minWidth: fabStart + "px" },
            {
              width: fab.getBoundingClientRect().width + "px",
              minWidth: fab.getBoundingClientRect().width + "px",
            },
          ],
          timing,
        ),
      );
    if (wasModal && !expanded)
      current.animations.push(
        frame
          .querySelector(".md-navigation-scrim")
          .animate([{ opacity: 1 }, { opacity: 0 }], {
            ...effects,
            fill: "forwards",
          }),
      );
    if (expanded) restoreFocus();
    animation.finished.then(
      () => {
        if (railFrames.get(frame) !== current) return;
        delete frame.dataset.transitioning;
        syncRail(frame);
        for (const a of current.animations) a.cancel();
        for (const copy of current.copies) copy.remove();
        if (!expanded) restoreFocus();
      },
      () => {},
    );
  }
  function initRailFrames() {
    for (const frame of document.querySelectorAll(".md-rail-frame")) {
      if (!railFrames.has(frame)) {
        railFrames.set(frame, {});
        syncRail(frame);
      }
      for (const label of frame.querySelectorAll(".md-nav-label"))
        if (!railLabels.has(label)) {
          railLabels.add(label);
          railLabelSizes.observe(label);
        }
    }
  }
  document.addEventListener("keydown", (event) => {
    const frame = event.target.closest?.(".md-rail-frame");
    if (!frame || frame.dataset.modalOpen !== "true") return;
    if (event.key === "Escape") {
      event.preventDefault();
      setRailExpanded(frame, false, true);
      return;
    }
    if (event.key === "Tab") {
      const controls = [
          ...frame
            .querySelector(".md-navigation-rail")
            .querySelectorAll("button"),
        ].filter((e) => e.getClientRects().length),
        first = controls[0],
        last = controls.at(-1);
      if (event.shiftKey && event.target === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && event.target === last) {
        event.preventDefault();
        first.focus();
      }
    }
  });
  document.addEventListener("htmx:beforeCleanupElement", (event) => {
    const root = event.detail.elt;
    for (const label of root.querySelectorAll?.(
      ".md-rail-frame .md-nav-label",
    ) || [])
      railLabelSizes.unobserve(label);
  });
  document.addEventListener("htmx:afterSwap", initRailFrames);
  document.addEventListener("htmx:historyRestore", initRailFrames);
  if (document.readyState === "loading")
    document.addEventListener("DOMContentLoaded", initRailFrames);
  else initRailFrames();
  document.addEventListener("click", (event) => {
    if (!(event.target instanceof Element)) return;
    const modal = event.target.closest("[data-navigation-modal]");
    if (modal && event.target === modal) {
      const rect = modal.getBoundingClientRect();
      if (
        event.clientX < rect.left ||
        event.clientX > rect.right ||
        event.clientY < rect.top ||
        event.clientY > rect.bottom
      )
        modal.close();
    }
    const button = event.target.closest("button");
    if (!button) return;
    if (button.hasAttribute("data-navigation-item")) selectNavigation(button);
    if (button.hasAttribute("data-navigation-open"))
      button
        .closest("[data-navigation-demo]")
        .querySelector("dialog")
        .showModal();
    if (button.hasAttribute("data-navigation-expand")) {
      const frame = button.closest(".md-rail-frame");
      if (frame && frame.dataset.interactive === "true")
        setRailExpanded(
          frame,
          frame.querySelector(".md-navigation-rail").dataset.variant !==
            "expanded",
          true,
        );
    }
    if (button.hasAttribute("data-navigation-dismiss"))
      setRailExpanded(button.closest(".md-rail-frame"), false, true);
  });
})();
