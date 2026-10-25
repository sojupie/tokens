(() => {
  if (typeof document === "undefined") return;
  document.addEventListener("keydown", (event) => {
    const input = event.target;
    if (event.key !== "Enter" || event.defaultPrevented || event.isComposing ||
        !input.matches?.('.md-segmented-button > input:is([type="radio"],[type="checkbox"])') ||
        input.matches(":disabled") || input.closest('[inert],[aria-disabled="true"]')) return;
    // Space and radio arrow navigation remain native. Suppress implicit form
    // submission and repeated Enter presses, then use native input activation.
    event.preventDefault();
    if (!event.repeat) input.click();
  });
})();
