import { navigationIndex } from "../helpers/navigation.ts";
import { snackbarTimer, toolbarGap, sliderValues, splitMenuPosition } from "../helpers/controls.ts";

(() => {
  const helpers = {
    snackbarTimer,
    navigationIndex,
    sliderValues,
    toolbarGap,
    splitMenuPosition,
  };
  if (typeof module !== "undefined") module.exports = helpers;
  if (typeof document !== "undefined")
    globalThis.MaterialComponents = Object.freeze(helpers);
})();
