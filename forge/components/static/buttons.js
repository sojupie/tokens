(() => {
  // Material supplies damping ratios and stiffness. The renderer uses unit mass.
  function springFrames(damping, stiffness) {
    if (!(damping > 0) || !(stiffness > 0))
      throw new Error("Invalid button spring");
    const omega = Math.sqrt(stiffness);
    const progress = (time) => {
      if (damping < 1) {
        const wd = omega * Math.sqrt(1 - damping * damping);
        return (
          1 -
          Math.exp(-damping * omega * time) *
            (Math.cos(wd * time) +
              ((damping * omega) / wd) * Math.sin(wd * time))
        );
      }
      if (damping === 1)
        return 1 - Math.exp(-omega * time) * (1 + omega * time);
      const root = Math.sqrt(damping * damping - 1);
      const a = -omega * (damping - root),
        b = -omega * (damping + root);
      return 1 - (b * Math.exp(a * time) - a * Math.exp(b * time)) / (b - a);
    };
    const samples = [0];
    for (let i = 1; i <= 600; i++) {
      samples.push(progress(i / 60));
      if (
        i > 2 &&
        Math.abs(1 - samples[i]) < 0.001 &&
        Math.abs(samples[i] - samples[i - 1]) < 0.001
      )
        break;
    }
    samples[samples.length - 1] = 1;
    return { samples, duration: ((samples.length - 1) * 1000) / 60 };
  }
  function corners(value, width, height) {
    const parts = value.trim().split(/\s+/);
    const radius = parts.map((part) =>
      part.includes("infinity")
        ? Math.min(width, height) / 2
        : part.endsWith("%")
          ? (Math.min(width, height) * parseFloat(part)) / 100
          : Math.min(parseFloat(part), Math.min(width, height) / 2),
    );
    if (value.includes("infinity"))
      return Array(4).fill(Math.min(width, height) / 2);
    if (radius.some((n) => !Number.isFinite(n)))
      throw new Error("Invalid button shape");
    if (radius.length === 1) return Array(4).fill(radius[0]);
    if (radius.length === 2)
      return [radius[0], radius[1], radius[0], radius[1]];
    if (radius.length === 3) return [...radius, radius[1]];
    return radius;
  }
  // Geometry and lifecycle adapted from Material Web's Apache-2.0 ripple.
  // Copyright 2022 Google LLC, Apache-2.0. See THIRD_PARTY_NOTICES.md.
  function rippleGeometry(width, height, point, zoom = 1) {
    const max = Math.max(width, height);
    const size = Math.max(1, Math.floor((max * 0.2) / zoom));
    const scale =
      (Math.hypot(width, height) + 10 + Math.max(max * 0.35, 75)) / size / zoom;
    const x = point ? point.x / zoom : width / zoom / 2;
    const y = point ? point.y / zoom : height / zoom / 2;
    return {
      size,
      start: `translate(${x - size / 2}px, ${y - size / 2}px) scale(1)`,
      end: `translate(${(width / zoom - size) / 2}px, ${(height / zoom - size) / 2}px) scale(${scale})`,
    };
  }
  if (typeof module !== "undefined")
    module.exports = { springFrames, corners, rippleGeometry };
  if (typeof document === "undefined") return;

  const reduced = matchMedia("(prefers-reduced-motion: reduce)");
  const forced = matchMedia("(forced-colors: active)");
  const active = new Map(),
    shapes = new Map(),
    ripples = new Set(),
    cancelled = new Map();
  const disabled = (button) =>
    button.matches(':disabled,[aria-disabled="true"],:has(> input:disabled)');
  const buttonFor = (target) =>
    target instanceof Element
      ? target.closest(
          ".md-button,.md-icon-button,.md-filter-chip,.md-interactive-card,.md-menu-item",
        )
      : null;
  function shapeTo(button, target, initial) {
    const old = shapes.get(button);
    const style = getComputedStyle(button),
      rect = button.getBoundingClientRect();
    const start =
      initial ||
      ["TopLeft", "TopRight", "BottomRight", "BottomLeft"].map((c) =>
        Math.min(
          parseFloat(style[`border${c}Radius`]),
          Math.min(rect.width, rect.height) / 2,
        ),
      );
    const end = corners(
      style.getPropertyValue(target),
      rect.width,
      rect.height,
    );
    old?.cancel();
    if (!button.isConnected) {
      shapes.delete(button);
      return;
    }
    if (
      reduced.matches ||
      !button.animate ||
      start.some((n) => !Number.isFinite(n))
    ) {
      shapes.delete(button);
      return;
    }
    // Only components with source spring parameters morph their corners.
    if (!style.getPropertyValue("--_button-spring-damping").trim()) return;
    const { samples, duration } = springFrames(
      parseFloat(style.getPropertyValue("--_button-spring-damping")),
      parseFloat(style.getPropertyValue("--_button-spring-stiffness")),
    );
    const frames = samples.map((value, i) => ({
      borderRadius: start
        .map((n, c) => `${Math.max(0, n + (end[c] - n) * value)}px`)
        .join(" "),
      offset: i / (samples.length - 1),
    }));
    const animation = button.animate(frames, { duration, easing: "linear" });
    shapes.set(button, animation);
    animation.finished.then(
      () => {
        if (shapes.get(button) === animation) shapes.delete(button);
      },
      () => {},
    );
  }
  function begin(button, point, id) {
    if (disabled(button) || forced.matches || active.has(button)) return null;
    button.removeAttribute("data-button-cancelled");
    cancelled.delete(button);
    const entry = [...ripples].find((r) => r.button === button) || {
      button,
      revision: 0,
    };
    clearTimeout(entry.touchTimer);
    clearTimeout(entry.releaseTimer);
    entry.revision++;
    entry.id = id;
    entry.point = point;
    entry.phase = "waiting-click";
    active.set(button, entry);
    const start = shapes.has(button)
      ? null
      : corners(
          getComputedStyle(button).getPropertyValue("--_button-rest-shape"),
          button.getBoundingClientRect().width,
          button.getBoundingClientRect().height,
        );
    button.setAttribute("data-pressed", "");
    shapeTo(button, "--_button-pressed-shape", start);
    return entry;
  }
  function startInk(entry) {
    const { button } = entry;
    if (
      reduced.matches ||
      forced.matches ||
      !button.animate ||
      disabled(button)
    )
      return;
    let fromOpacity = "0";
    if (entry.layer) {
      fromOpacity = getComputedStyle(entry.layer).opacity;
      entry.growth?.cancel();
      entry.fade?.cancel();
    } else {
      entry.layer = document.createElement("span");
      entry.layer.className = "md-button-ripple";
      entry.layer.setAttribute("aria-hidden", "true");
      entry.ink = document.createElement("span");
      entry.layer.append(entry.ink);
      button.append(entry.layer);
      ripples.add(entry);
    }
    const rect = entry.layer.getBoundingClientRect();
    const control = button.getBoundingClientRect();
    const point = entry.point && {
      x: entry.point.x + control.left - rect.left,
      y: entry.point.y + control.top - rect.top,
    };
    const geometry = rippleGeometry(
      rect.width,
      rect.height,
      point,
      button.currentCSSZoom || 1,
    );
    Object.assign(entry.ink.style, {
      width: `${geometry.size}px`,
      height: `${geometry.size}px`,
      left: "0",
      top: "0",
    });
    button.setAttribute("data-rippling", "");
    entry.started = performance.now();
    entry.growth = entry.ink.animate(
      [{ transform: geometry.start }, { transform: geometry.end }],
      {
        duration: 450,
        easing: "cubic-bezier(0.2, 0, 0, 1)",
        fill: "forwards",
      },
    );
    entry.fade = entry.layer.animate(
      [{ opacity: fromOpacity }, { opacity: 1 }],
      {
        duration: 105,
        easing: "linear",
        fill: "forwards",
      },
    );
    entry.growth.finished.catch(() => {});
    entry.fade.finished.catch(() => {});
  }
  function release(button, abort = false) {
    const entry = active.get(button);
    if (!entry) return;
    active.delete(button);
    clearTimeout(entry.touchTimer);
    button.removeAttribute("data-pressed");
    if (abort) {
      button.setAttribute("data-button-cancelled", "");
      cancelled.set(button, entry.id);
    }
    shapeTo(button, "--_button-rest-shape");
    if (!entry.layer) return;
    if (
      reduced.matches ||
      forced.matches ||
      !button.isConnected ||
      disabled(button)
    ) {
      finishRipple(entry);
      return;
    }
    const revision = ++entry.revision;
    const fadeOut = () => {
      if (
        entry.revision !== revision ||
        active.has(button) ||
        !ripples.has(entry)
      )
        return;
      const opacity = getComputedStyle(entry.layer).opacity;
      entry.fade?.cancel();
      const fade = (entry.fade = entry.layer.animate(
        [{ opacity }, { opacity: 0 }],
        {
          duration: 375,
          easing: "linear",
          fill: "forwards",
        },
      ));
      fade.finished.then(
        () => {
          if (entry.revision === revision && entry.fade === fade)
            finishRipple(entry);
        },
        () => {},
      );
    };
    const time = entry.growth.currentTime;
    const elapsed =
      typeof time === "number"
        ? time
        : time?.to
          ? time.to("ms").value
          : performance.now() - entry.started;
    const remaining = Math.max(0, 225 - elapsed);
    if (remaining) entry.releaseTimer = setTimeout(fadeOut, remaining);
    else fadeOut();
  }
  function finishRipple(entry) {
    clearTimeout(entry.releaseTimer);
    clearTimeout(entry.touchTimer);
    entry.revision++;
    entry.growth?.cancel();
    entry.fade?.cancel();
    entry.layer?.remove();
    ripples.delete(entry);
    entry.layer = null;
    entry.button.removeAttribute("data-rippling");
  }
  function clearCancelled(id) {
    for (const [button, pointer] of cancelled)
      if (id === undefined || id === pointer) {
        button.removeAttribute("data-button-cancelled");
        cancelled.delete(button);
      }
  }
  function cancelAll() {
    for (const button of active.keys()) release(button, true);
    for (const entry of ripples) finishRipple(entry);
    for (const animation of shapes.values()) animation.cancel();
    shapes.clear();
    clearCancelled();
  }
  document.addEventListener(
    "pointerdown",
    (event) => {
      const button = buttonFor(event.target);
      if (!button || event.button !== 0 || event.isPrimary === false) return;
      const touch = event.pointerType === "touch";
      if (touch) button.setAttribute("data-touch", "");
      else button.removeAttribute("data-touch");
      const rect = button.getBoundingClientRect();
      const entry = begin(
        button,
        { x: event.clientX - rect.left, y: event.clientY - rect.top },
        event.pointerId,
      );
      if (!entry) return;
      if (!touch) {
        startInk(entry);
        return;
      }
      entry.phase = "touch-delay";
      const revision = entry.revision;
      entry.touchTimer = setTimeout(() => {
        if (
          active.get(button) !== entry ||
          entry.revision !== revision ||
          entry.phase !== "touch-delay"
        )
          return;
        entry.phase = "holding";
        startInk(entry);
      }, 150);
    },
    true,
  );
  document.addEventListener(
    "pointerup",
    (event) => {
      if (event.isPrimary === false) return;
      for (const [button, entry] of active) {
        if (entry.id !== event.pointerId) continue;
        if (entry.phase === "touch-delay") {
          clearTimeout(entry.touchTimer);
          startInk(entry);
        }
        entry.phase = "waiting-click";
        button.removeAttribute("data-pressed");
        shapeTo(button, "--_button-rest-shape");
      }
      clearCancelled(event.pointerId);
    },
    true,
  );
  document.addEventListener(
    "pointercancel",
    (event) => {
      for (const [button, entry] of active)
        if (entry.id === event.pointerId) release(button, true);
      clearCancelled(event.pointerId);
    },
    true,
  );
  document.addEventListener(
    "pointerover",
    (event) => {
      if (event.pointerType !== "touch")
        buttonFor(event.target)?.removeAttribute("data-touch");
    },
    true,
  );
  document.addEventListener(
    "pointerout",
    (event) => {
      if (event.pointerType === "touch") return;
      const button = buttonFor(event.target);
      if (
        button &&
        active.has(button) &&
        !(
          event.relatedTarget instanceof Node &&
          button.contains(event.relatedTarget)
        )
      )
        release(button, true);
    },
    true,
  );
  document.addEventListener(
    "contextmenu",
    (event) => {
      const button = buttonFor(event.target);
      if (button) release(button, true);
    },
    true,
  );
  document.addEventListener(
    "keydown",
    (event) => {
      const button = buttonFor(event.target);
      if (
        !button ||
        event.repeat ||
        !(
          event.key === "Enter" ||
          (event.key === " " &&
            (button.tagName === "BUTTON" || button.matches(".md-filter-chip")))
        )
      )
        return;
      begin(button, null, "keyboard");
    },
    true,
  );
  document.addEventListener(
    "keyup",
    (event) => {
      const button = buttonFor(event.target);
      if (["Enter", " "].includes(event.key)) {
        if (button && active.get(button)?.id === "keyboard") release(button);
        clearCancelled("keyboard");
      }
    },
    true,
  );
  document.addEventListener(
    "focusout",
    (event) => {
      const button = buttonFor(event.target);
      if (
        button &&
        !(
          event.relatedTarget instanceof Node &&
          button.contains(event.relatedTarget)
        )
      )
        release(button, true);
    },
    true,
  );
  document.addEventListener(
    "click",
    (event) => {
      const button = buttonFor(event.target);
      if (!button) return;
      if (disabled(button)) {
        event.preventDefault();
        event.stopImmediatePropagation();
        return;
      }
      // A label's native input owns activation, including forwarded and keyboard clicks.
      if (button.tagName === "LABEL" && event.target.tagName !== "INPUT")
        return;
      let entry = active.get(button);
      if (!entry) entry = begin(button, null, "keyboard");
      if (entry && (!entry.layer || entry.id === "keyboard")) startInk(entry);
      if (entry) release(button);
      if (button.hasAttribute("data-button-toggle")) {
        button.setAttribute(
          "aria-pressed",
          String(button.getAttribute("aria-pressed") !== "true"),
        );
        shapeTo(button, "--_button-rest-shape");
      }
    },
    true,
  );
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) cancelAll();
  });
  window.addEventListener("blur", cancelAll);
  reduced.addEventListener("change", cancelAll);
  forced.addEventListener("change", cancelAll);
  new MutationObserver(() => {
    for (const button of cancelled.keys())
      if (!button.isConnected) cancelled.delete(button);
    for (const [button] of active)
      if (!button.isConnected || disabled(button)) release(button, true);
    for (const entry of ripples)
      if (!entry.button.isConnected || disabled(entry.button))
        finishRipple(entry);
    for (const [button, animation] of shapes)
      if (!button.isConnected || disabled(button)) {
        animation.cancel();
        shapes.delete(button);
      }
  }).observe(document.documentElement, {
    subtree: true,
    childList: true,
    attributes: true,
    attributeFilter: ["disabled", "aria-disabled"],
  });
  document.documentElement.setAttribute("data-feedback-ready", "");
})();
