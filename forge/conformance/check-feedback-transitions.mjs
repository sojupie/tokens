// Run from a local browser fixture with real component markup, CSS and buttons.js.
// The fixture's Run button calls this function; no browser instrumentation is required.
export async function checkFeedbackTransitions(elements) {
  const results = [];
  const wait = (duration) =>
    new Promise((resolve) => setTimeout(resolve, duration));
  for (const element of elements) {
    const frames = [],
      failures = [];
    let phase = "hover",
      sampling = true;
    const state = (value) => {
      element.dataset.componentState = value;
      element.dataset.buttonState = value;
    };
    const pointer = (type, extra = {}) =>
      element.dispatchEvent(
        new PointerEvent(type, {
          bubbles: true,
          button: 0,
          buttons: type === "pointerdown" ? 1 : 0,
          pointerType: "mouse",
          pointerId: 1,
          isPrimary: true,
          clientX: element.getBoundingClientRect().left + 12,
          clientY: element.getBoundingClientRect().top + 12,
          ...extra,
        }),
      );
    const preventActivation = (event) => event.preventDefault();
    element.addEventListener("click", preventActivation);
    const click = () =>
      (element.querySelector("input") || element).dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true, detail: 1 }),
      );
    const sample = (time) => {
      if (!sampling) return;
      const base = Number(getComputedStyle(element, "::before").opacity);
      const layers = [...element.querySelectorAll(".md-button-ripple")];
      const ink = layers.reduce((alpha, layer) => {
        const a =
          Number(getComputedStyle(layer).opacity) *
          Number(getComputedStyle(layer.firstElementChild).opacity);
        return a + alpha * (1 - a);
      }, 0);
      frames.push({
        time,
        phase,
        radius: Number.parseFloat(
          getComputedStyle(element).borderTopLeftRadius,
        ),
        base,
        ink,
        composite: base + ink * (1 - base),
        layers: layers.length,
      });
      requestAnimationFrame(sample);
    };
    state("hovered");
    await wait(220);
    const hover = Number(getComputedStyle(element, "::before").opacity);
    const restingRadius = Number.parseFloat(
      getComputedStyle(element).borderTopLeftRadius,
    );
    requestAnimationFrame(sample);
    phase = "held-press";
    pointer("pointerdown");
    await wait(650);
    const pressedRadius = Number.parseFloat(
      getComputedStyle(element).borderTopLeftRadius,
    );
    phase = "held-release";
    pointer("pointerup");
    click();
    await wait(650);
    phase = "press";
    pointer("pointerdown");
    await wait(40);
    phase = "leave";
    state("enabled");
    pointer("pointerout");
    await wait(30);
    phase = "reenter";
    state("hovered");
    await wait(30);
    pointer("pointerup");
    phase = "repeat";
    pointer("pointerdown");
    await wait(40);
    phase = "release-hover";
    pointer("pointerup");
    click();
    await wait(650);
    phase = "leave";
    state("enabled");
    pointer("pointerout");
    await wait(220);
    sampling = false;
    if (
      Math.abs(restingRadius - pressedRadius) > 1 &&
      !frames.some(
        (frame) =>
          frame.phase === "held-release" &&
          frame.radius > Math.min(restingRadius, pressedRadius) + 0.1 &&
          frame.radius < Math.max(restingRadius, pressedRadius) - 0.1,
      )
    )
      failures.push("Shape snaps back after a held press");
    if (frames.length < 20)
      failures.push("Insufficient rendered animation frames");
    for (const [i, frame] of frames.entries()) {
      if (!Number.isFinite(frame.composite))
        failures.push("Invalid layer opacity");
      if (frame.layers > 1) failures.push("Stacked ripples");
      if (
        ["press", "release-hover"].includes(frame.phase) &&
        frame.base < hover - 0.005
      )
        failures.push("Hover disappears during press or release");
      const previous = frames[i - 1];
      if (previous) {
        const elapsed = frame.time - previous.time;
        // Bound per-frame jumps at a 60 Hz equivalent; exclude stalled rendering.
        if (
          frame.phase === previous.phase &&
          ["press", "release-hover"].includes(frame.phase) &&
          elapsed > 0 &&
          elapsed < 50 &&
          (Math.abs(frame.composite - previous.composite) * 16.67) / elapsed >
            0.04
        )
          failures.push("Abrupt opacity change");
      }
    }
    if (
      element.querySelector(".md-button-ripple") ||
      element.hasAttribute("data-rippling")
    )
      failures.push("Ripple remains after release");
    results.push({
      component: element.className,
      frameCount: frames.length,
      failures: [...new Set(failures)],
      frames,
    });
    state("enabled");
    element.removeEventListener("click", preventActivation);
  }
  return results;
}
