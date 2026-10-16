// Run in the button gallery document to compare rendered properties with source variables.
export function checkButtonRendering() {
  const failures = [];
  const failureCounts = {};
  let checked = 0;
  const clean = (s) => String(s).trim().replace(/\s+/g, " ");
  const colors = (value) => {
    value = value.trim();
    if (value === "transparent") return [0, 0, 0, 0];
    if (value.startsWith("#"))
      return [1, 3, 5]
        .map((n) => Number.parseInt(value.slice(n, n + 2), 16) / 255)
        .concat(
          value.length === 9 ? Number.parseInt(value.slice(7, 9), 16) / 255 : 1,
        );
    const values = (value.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g) || []).map(
      Number,
    );
    return value.startsWith("rgb")
      ? values
          .slice(0, 3)
          .map((n) => n / 255)
          .concat(values[3] ?? 1)
      : values.slice(0, 3).concat(values[3] ?? 1);
  };
  const sameColor = (actual, expected, opacity = 1) => {
    const a = colors(actual),
      b = colors(expected);
    b[3] *= opacity;
    return (
      a.length === 4 &&
      b.length === 4 &&
      a.every((n, i) => Math.abs(n - b[i]) < 0.0001)
    );
  };
  for (const button of document.querySelectorAll("[data-button-case]")) {
    const [variant, size, shape, selection, forced] =
      button.dataset.buttonCase.split("-");
    const style = getComputedStyle(button),
      rect = button.getBoundingClientRect();
    const variable = (name) =>
      style.getPropertyValue("--" + name.replaceAll(".", "-")).trim();
    let state = forced === "enabled" ? "" : forced;
    if (state !== "disabled") {
      if (button.matches(":active") || button.hasAttribute("data-pressed"))
        state = "pressed";
      else if (button.matches(":focus-visible") && state !== "pressed")
        state = "focused";
      else if (
        button.matches(":hover") &&
        !["pressed", "focused"].includes(state)
      )
        state = "hovered";
    }
    const resolve = (property) => {
      const prefixes = [];
      if (selection && state) prefixes.push(`${selection}.${state}.`);
      if (state) prefixes.push(`${state}.`);
      if (selection) prefixes.push(`${selection}.`);
      prefixes.push("");
      for (const prefix of prefixes) {
        const value = variable(
          `md.comp.button.${variant}.${prefix}${property}`,
        );
        if (value) return value;
      }
      return "";
    };
    const expect = (test, property, actual, expected) => {
      checked++;
      if (!test) {
        failureCounts[property] = (failureCounts[property] || 0) + 1;
        if (failures.length < 12)
          failures.push({
            case: button.dataset.buttonCase,
            property,
            actual,
            expected,
          });
      }
    };
    for (const [role, property, fallback] of [
      ["--_button-container", "container.color", "transparent"],
      ["--_button-label", "label-text.color", ""],
      ["--_button-icon", "icon.color", ""],
      ["--_elevation-level", "container.elevation", "0"],
      ["--_button-container-opacity", "container.opacity", "1"],
      ["--_button-label-opacity", "label-text.opacity", "1"],
      ["--_button-icon-opacity", "icon.opacity", "1"],
      ["--_button-state-opacity", "state-layer.opacity", "0"],
      ["--_button-state-color", "state-layer.color", "currentColor"],
    ]) {
      const expected = resolve(property) || fallback,
        actual = style.getPropertyValue(role).trim();
      expect(clean(actual) === clean(expected), role, actual, expected);
    }
    const label = resolve("label-text.color"),
      labelOpacity = Number(resolve("label-text.opacity") || 1);
    expect(sameColor(style.color, label, labelOpacity), "color", style.color, [
      label,
      labelOpacity,
    ]);
    const container = resolve("container.color") || "transparent",
      containerOpacity = Number(resolve("container.opacity") || 1);
    // Fully transparent colors can retain arbitrary RGB channels in CSS serialization.
    if (container !== "transparent")
      expect(
        sameColor(style.backgroundColor, container, containerOpacity),
        "background",
        style.backgroundColor,
        [container, containerOpacity],
      );
    const icon = button.querySelector("svg");
    expect(!!icon, "icon", !!icon, true);
    if (icon) {
      const iconStyle = getComputedStyle(icon),
        iconSize = variable(`md.comp.button.${size}.icon.size`);
      expect(
        iconStyle.width === iconSize,
        "icon width",
        iconStyle.width,
        iconSize,
      );
      expect(
        sameColor(
          iconStyle.color,
          resolve("icon.color"),
          Number(resolve("icon.opacity") || 1),
        ),
        "icon color",
        iconStyle.color,
        resolve("icon.color"),
      );
    }
    const layerOpacity = Number(getComputedStyle(button, "::before").opacity);
    expect(
      Math.abs(layerOpacity - Number(resolve("state-layer.opacity") || 0)) <
        0.001,
      "state layer opacity",
      layerOpacity,
      resolve("state-layer.opacity") || 0,
    );
    const height = variable(`md.comp.button.${size}.container.height`);
    expect(style.minHeight === height, "min-height", style.minHeight, height);
    expect(
      style.paddingInlineStart ===
        variable(`md.comp.button.${size}.leading-space`),
      "leading space",
      style.paddingInlineStart,
      variable(`md.comp.button.${size}.leading-space`),
    );
    expect(
      style.paddingInlineEnd ===
        variable(`md.comp.button.${size}.trailing-space`),
      "trailing space",
      style.paddingInlineEnd,
      variable(`md.comp.button.${size}.trailing-space`),
    );
    const tracking = variable(
      `md.comp.button.${size}.label-text.letter-spacing`,
    );
    expect(
      (style.letterSpacing === "normal"
        ? 0
        : Number.parseFloat(style.letterSpacing)) ===
        Number.parseFloat(tracking),
      "letter spacing",
      style.letterSpacing,
      tracking,
    );
    const font = variable(`md.comp.button.${size}.label-text`),
      fontSize = font.match(/(\d+(?:\.\d+)?)px\//)?.[1];
    expect(
      Number.parseFloat(style.fontSize) === Number(fontSize),
      "font size",
      style.fontSize,
      font,
    );
    const fontWeight = font.match(/^(\d+) /)?.[1];
    expect(
      Number(style.fontWeight) === Number(fontWeight),
      "font weight",
      style.fontWeight,
      fontWeight,
    );
    const ratio = font.match(/px\/([\d.]+)/)?.[1];
    expect(
      Math.abs(
        Number.parseFloat(style.lineHeight) - Number(fontSize) * Number(ratio),
      ) < 0.1,
      "line height",
      style.lineHeight,
      Number(fontSize) * Number(ratio),
    );
    const width =
      variant === "outlined" && selection !== "selected"
        ? variable(`md.comp.button.${size}.outlined.outline.width`)
        : "0px";
    expect(
      style.borderTopWidth === width,
      "outline width",
      style.borderTopWidth,
      width,
    );
    const shapeToken =
      state === "pressed"
        ? `md.comp.button.${size}.pressed.container.shape`
        : `md.comp.button.${size}.${selection === "selected" ? "selected." : ""}container.shape.${shape}`;
    const radius = variable(shapeToken);
    const expectedRadius = radius.includes("infinity")
      ? Math.min(rect.width, rect.height) / 2
      : Math.min(
          Number.parseFloat(radius),
          Math.min(rect.width, rect.height) / 2,
        );
    const actualRadius = Math.min(
      Number.parseFloat(style.borderTopLeftRadius),
      Math.min(rect.width, rect.height) / 2,
    );
    expect(
      Math.abs(actualRadius - expectedRadius) < 0.1,
      "shape",
      actualRadius,
      expectedRadius,
    );
  }
  return {
    theme: document.documentElement.dataset.theme,
    contrast: document.documentElement.dataset.contrast,
    configurations: document.querySelectorAll("[data-button-case]").length,
    propertiesChecked: checked,
    failureCounts,
    failures,
  };
}
