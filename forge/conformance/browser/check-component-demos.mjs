// Verify the boundary between Forge mechanics and Catalog example responses.
export async function checkComponentDemos(load, assert) {
  {
    const { d, w } = await load("navigation-rail");
    const saved = [...d.querySelectorAll("[data-navigation-item]")].find(
      (item) => item.textContent.trim() === "Saved",
    );
    saved.click();
    assert(
      saved.getAttribute("aria-current") === "page",
      "Navigation behavior selects a destination",
    );
    assert(
      d.querySelector("[data-navigation-title]").textContent === "Saved",
      "Catalog updates workspace content from the navigation event",
    );
    d.querySelector("[data-navigation-create]").click();
    assert(
      d.querySelector("[data-navigation-title]").textContent === "Create file",
      "Catalog owns the example create action",
    );
    assert(
      !w.motionErrors.length,
      "Extracted navigation initializes without script errors",
    );
  }
  {
    const { d, w } = await load("chips", { query: "&preview.variant=input" });
    const chip = d.querySelector('[data-family="chip"]'),
      primary = chip.querySelector("[data-chip-select]");
    const previous = primary.getAttribute("aria-pressed");
    primary.click();
    const selected = primary.getAttribute("aria-pressed") === "true";
    assert(
      primary.getAttribute("aria-pressed") !== previous,
      "Chip behavior toggles selection",
    );
    assert(
      d.querySelector("[data-chip-output]").textContent ===
        (selected ? "Selected" : "Unselected"),
      "Catalog announces chip selection",
    );
    chip.querySelector("[data-chip-remove]").click();
    assert(
      chip.hidden &&
        d.querySelector("[data-chip-output]").textContent === "Removed Alex",
      "Chip removal retains the Catalog announcement",
    );
    assert(
      !w.motionErrors.length,
      "Extracted chips initialize without script errors",
    );
  }
  {
    const { d, w } = await load("button-groups");
    d.querySelector("[data-group-button]").click();
    assert(
      d.querySelector(".md-button-group-demo output").textContent.length > 0,
      "Catalog reports button-group actions",
    );
    assert(
      !w.motionErrors.length,
      "Extracted button groups initialize without script errors",
    );
  }
  {
    const { d, w } = await load("snackbar");
    d.querySelector("[data-snackbar-open]").click();
    assert(
      d.querySelector(".md-snackbar").matches(":popover-open"),
      "Extracted snackbar opens",
    );
    d.querySelector("[data-snackbar-action]").click();
    assert(
      !d.querySelector(".md-snackbar").matches(":popover-open") &&
        d.querySelector("[data-snackbar-demo] output").textContent ===
          "Message undone.",
      "Snackbar dismissal and Catalog announcement remain connected",
    );
    assert(
      !w.motionErrors.length,
      "Extracted snackbar initializes without script errors",
    );
  }
  {
    const { d, w } = await load("split-button");
    d.querySelector("[data-split-primary]").click();
    assert(
      d.querySelector(".md-split-demo output").textContent === "Added to cart.",
      "Catalog owns the split-button example action",
    );
    assert(
      !w.motionErrors.length,
      "Extracted split buttons initialize without script errors",
    );
  }
  {
    const { d, w } = await load("tabs");
    const tab = d.querySelectorAll(".md-tab")[1];
    tab.click();
    assert(
      tab.getAttribute("aria-selected") === "true" &&
        !d.getElementById(tab.getAttribute("aria-controls")).hidden,
      "Extracted tabs select the corresponding panel",
    );
    assert(
      !w.motionErrors.length,
      "Extracted tabs initialize without script errors",
    );
  }
  {
    const { d, w } = await load("sliders");
    const slider = d.querySelector(".md-slider"),
      input = slider.querySelector('input[type="range"]');
    input.value = "60";
    input.dispatchEvent(new w.Event("input", { bubbles: true }));
    assert(
      slider.hasAttribute("data-slider-ready") &&
        slider.querySelector(".md-slider-value").textContent === "60",
      "Extracted sliders initialize and update geometry and labels",
    );
    assert(
      !w.motionErrors.length,
      "Extracted sliders initialize without script errors",
    );
  }
  {
    const { d, w } = await load("toolbars");
    assert(
      d.querySelector(".md-toolbar").hasAttribute("data-toolbar-ready"),
      "Extracted toolbar initializes its layout observer",
    );
    assert(
      !w.motionErrors.length,
      "Extracted toolbars initialize without script errors",
    );
  }
}
