import { navigationIndex } from "../helpers/navigation.ts";

(() => {
  if (typeof document === "undefined") return;
  document.addEventListener("click", (event) => {
    const button = (event.target as Element | null)?.closest?.<HTMLButtonElement>("[data-group-button]");
    if (!button || button.disabled) return;
    const group = button.closest<HTMLElement>(".md-button-group");
    if (!group) return;
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
  document.addEventListener("focusin", event => {
    const button = (event.target as Element | null)?.closest?.<HTMLButtonElement>("[data-group-button]");
    if (button) for (const item of button.closest<HTMLElement>(".md-button-group")?.querySelectorAll("button") || []) item.tabIndex = item === button ? 0 : -1;
  });
  document.addEventListener("keydown", (event) => {
    const button = (event.target as Element | null)?.closest?.<HTMLButtonElement>("[data-group-button]");
    if (!button) return;
    const group = button.closest<HTMLElement>(".md-button-group");
    if (!group) return;
    const items = [...group.querySelectorAll<HTMLButtonElement>("button:not(:disabled):not([aria-disabled=\"true\"])")].filter(item => item.getClientRects().length && !item.closest("[inert]") && getComputedStyle(item).visibility === "visible");
    if (!items.includes(button)) return;
    const next = navigationIndex(
      items.indexOf(button),
      items.length,
      event.key,
      true,
      getComputedStyle(group).direction === "rtl",
    );
    if (next !== null) {
      event.preventDefault();
      for (const item of group.querySelectorAll("button")) item.tabIndex = item === items[next] ? 0 : -1;
      items[next]?.focus();
    }
  });
})();
