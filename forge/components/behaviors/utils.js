import { navigationIndex } from "../helpers/navigation.ts";
import { snackbarTimer, toolbarGap, sliderValues, splitMenuPosition } from "../helpers/controls.ts";

const entryInputs = new WeakSet();
const pendingEntryReplacement = new WeakSet();

function focusForEntry(input) {
  if (!entryInputs.has(input)) {
    entryInputs.add(input);
    const clear = () => pendingEntryReplacement.delete(input);
    input.addEventListener("blur", clear);
    input.addEventListener("pointerdown", clear);
    input.addEventListener("keydown", event => {
      if (/^(Arrow|Home|End|Backspace|Delete)/.test(event.key) || event.ctrlKey || event.metaKey) clear();
    });
    input.addEventListener("beforeinput", event => {
      if (!pendingEntryReplacement.delete(input) || !event.cancelable ||
          !/^(insertText|insertFromPaste)$/.test(event.inputType) || !event.data) return;
      event.preventDefault();
      const text = input.maxLength < 0 ? event.data : event.data.slice(0, input.maxLength);
      input.setRangeText(text, 0, input.value.length, "end");
      input.dispatchEvent(new Event("input", {bubbles: true}));
    });
  }
  input.focus({preventScroll: true});
  const end = input.value.length;
  input.setSelectionRange(end, end);
  pendingEntryReplacement.add(input);
}

(() => {
  const helpers = {
    snackbarTimer,
    navigationIndex,
    sliderValues,
    toolbarGap,
    splitMenuPosition,
    focusForEntry,
  };
  if (typeof module !== "undefined") module.exports = helpers;
  if (typeof document !== "undefined")
    globalThis.MaterialComponents = Object.freeze(helpers);
})();
