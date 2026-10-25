(() => {
  if (typeof document === "undefined") return;
  document.addEventListener("keydown", (event) => {
    const input = event.target;
    if (event.key !== "Enter" || event.defaultPrevented || event.isComposing ||
        !input.matches?.('.md-switch > input[type="checkbox"]') ||
        input.matches(":disabled") || input.closest('[inert],[aria-disabled="true"]')) return;
    event.preventDefault();
    if (!event.repeat) input.click();
  });
})();
