/* Generated from catalog/web/assets/scripts/thumbnail-images.ts. DO NOT EDIT. Run npm run build:components. */
(()=>{"use strict";

// catalog/web/assets/scripts/thumbnail-images.ts
(() => {
  const pending = /* @__PURE__ */ new WeakSet();
  async function reveal(image) {
    const picture = image.closest(".tokens-thumbnail-image");
    if (!picture) return;
    try {
      await image.decode();
    } catch {
      return;
    }
    picture.removeAttribute("data-image-loading");
  }
  function initializeImages() {
    for (const image of document.querySelectorAll(".tokens-thumbnail-image img")) {
      if (pending.has(image)) continue;
      pending.add(image);
      if (!image.complete || !image.naturalWidth) image.closest("picture")?.setAttribute("data-image-loading", "");
      if (image.complete && image.naturalWidth) void reveal(image);
    }
  }
  document.addEventListener("load", (event) => {
    if (event.target instanceof HTMLImageElement) void reveal(event.target);
  }, true);
  document.addEventListener("htmx:afterSwap", initializeImages);
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initializeImages);
  else initializeImages();
})();
})();
