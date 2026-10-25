(() => {
  if (typeof document === "undefined") return;
  document.addEventListener("keydown", (event) => {
    const input = event.target;
    if (event.defaultPrevented || event.isComposing ||
        !input.matches?.('.md-segmented-button > input:is([type="radio"],[type="checkbox"])') ||
        input.matches(":disabled") || input.closest('[inert],[aria-disabled="true"]')) return;
    if (input.type === "checkbox" && ["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) {
      const group = input.closest(".md-segmented-buttons");
      const items = [...group.querySelectorAll('input[type="checkbox"]')].filter(item =>
        !item.matches(":disabled") && !item.closest('[inert],[aria-disabled="true"]') &&
        item.getClientRects().length && getComputedStyle(item).visibility === "visible");
      if (!items.includes(input)) return;
      const next = globalThis.MaterialComponents.navigationIndex(items.indexOf(input), items.length,
        event.key, true, getComputedStyle(group).direction === "rtl");
      event.preventDefault();
      items[next].focus();
      return;
    }
    if (event.key !== "Enter") return;
    // Space and radio arrow navigation remain native. Suppress implicit form
    // submission and repeated Enter presses, then use native input activation.
    event.preventDefault();
    if (!event.repeat) input.click();
  });
})();
