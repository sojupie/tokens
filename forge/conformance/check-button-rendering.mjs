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
    const [variant, size, shape, configuredSelection, forced] =
      button.dataset.buttonCase.split("-");
    const selection = button.hasAttribute("aria-pressed")
      ? button.getAttribute("aria-pressed") === "true"
        ? "selected"
        : "unselected"
      : configuredSelection;
    const split = button.closest(".md-split-button"),
      group = button.closest(".md-button-group"),
      trailing = button.classList.contains("md-split-trailing");
    const style = getComputedStyle(button),
      rect = button.getBoundingClientRect();
    const variable = (name) =>
      style.getPropertyValue("--" + name.replaceAll(".", "-")).trim();
    let state = ["enabled", "interactive"].includes(forced) ? "" : forced;
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
        iconSize = variable(
          split && trailing
            ? `md.comp.split-button.${size}.trailing-button.icon.size`
            : `md.comp.button.${size}.icon.size`,
        );
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
    const height = variable(
      split
        ? `md.comp.split-button.${size}.container.height`
        : group
          ? `md.comp.button-group.${group.dataset.control}.${size}.container.height`
          : `md.comp.button.${size}.container.height`,
    );
    expect(style.minHeight === height, "min-height", style.minHeight, height);
    expect(
      style.paddingInlineStart ===
        variable(
          split
            ? `md.comp.split-button.${size}.${trailing ? "trailing" : "leading"}-button.leading-space`
            : `md.comp.button.${size}.leading-space`,
        ),
      "leading space",
      style.paddingInlineStart,
      variable(
        split
          ? `md.comp.split-button.${size}.${trailing ? "trailing" : "leading"}-button.leading-space`
          : `md.comp.button.${size}.leading-space`,
      ),
    );
    expect(
      style.paddingInlineEnd ===
        variable(
          split
            ? `md.comp.split-button.${size}.${trailing ? "trailing" : "leading"}-button.trailing-space`
            : `md.comp.button.${size}.trailing-space`,
        ),
      "trailing space",
      style.paddingInlineEnd,
      variable(
        split
          ? `md.comp.split-button.${size}.${trailing ? "trailing" : "leading"}-button.trailing-space`
          : `md.comp.button.${size}.trailing-space`,
      ),
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
    if (!split && group?.dataset.control !== "connected") {
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

// Convert the bounded corners used by the component styles to physical radii.
export function componentRadius(value, width, height) {
  const half = Math.min(width, height) / 2,
    text = String(value).trim();
  if (text.includes("infinity")) return half;
  const length = (part) =>
    /^\d*\.?\d+(?:e[+-]?\d+)?(?:px|%)?$/i.test(part.trim())
      ? part.trim().endsWith("%")
        ? (Number.parseFloat(part) * half) / 50
        : Number.parseFloat(part)
      : NaN;
  return Math.min(
    half,
    text.startsWith("min(") && text.endsWith(")")
      ? Math.min(...text.slice(4, -1).split(",").map(length))
      : length(text),
  );
}

export function checkCompositeButtonRendering() {
  const failures = [];
  let checked = 0,
    cases = 0;
  const roots = [
    ...document.querySelectorAll(".md-split-button,.md-button-group"),
  ];
  const expect = (root, property, actual, expected) => {
    checked++;
    if (
      !Number.isFinite(actual) ||
      !Number.isFinite(expected) ||
      Math.abs(actual - expected) > 0.12
    )
      failures.push({
        id: root.dataset.componentCase,
        property,
        actual,
        expected,
      });
  };
  for (const root of roots) {
    if (!root.getClientRects().length) continue;
    cases++;
    const split = root.classList.contains("md-split-button"),
      size = root.dataset.size,
      control = root.dataset.control,
      p =
        "md.comp." +
        (split ? "split-button." : "button-group." + control + ".") +
        size +
        ".",
      style = getComputedStyle(root),
      rtl = style.direction === "rtl";
    const source = (name) =>
        style.getPropertyValue("--" + name.replaceAll(".", "-")).trim(),
      number = (name) => Number.parseFloat(source(name));
    expect(
      root,
      "gap",
      Number.parseFloat(style.columnGap),
      number(p + "between-space"),
    );
    const buttons = [...root.querySelectorAll(":scope > .md-button")];
    buttons.forEach((button, index) => {
      const cs = getComputedStyle(button),
        rect = button.getBoundingClientRect(),
        state =
          button.dataset.buttonState ||
          (button.disabled
            ? "disabled"
            : button.hasAttribute("data-pressed")
              ? "pressed"
              : button.matches(":focus-visible")
                ? "focused"
                : button.matches(":hover")
                  ? "hovered"
                  : "enabled");
      expect(
        root,
        "height " + index,
        rect.height,
        number(p + "container.height"),
      );
      if (split) {
        const target=getComputedStyle(button,"::after");
        expect(root,"touch target width "+index,Math.max(rect.width,Number.parseFloat(target.width),Number.parseFloat(target.minWidth)),Math.max(rect.width,48));
        expect(root,"touch target height "+index,Math.max(rect.height,Number.parseFloat(target.height),Number.parseFloat(target.minHeight)),Math.max(rect.height,48));
        if(index===1){
          const icon=getComputedStyle(button.querySelector("svg")),active=state==="pressed"||button.matches(":active")||button.getAttribute("aria-expanded")==="true";
          const offset={xsmall:-1,small:-1,medium:-2,large:-3,xlarge:-6}[size];
          expect(root,"menu icon optical offset",Number.parseFloat(icon.translate),active?0:offset*(rtl?-1:1));
          expect(root,"menu icon rotation",Number.parseFloat(icon.rotate),active?(rtl?-180:180):0);
        }
      }
      if (split || control === "connected") {
        const first = index === 0,
          last = index === buttons.length - 1,
          selected = button.getAttribute("aria-pressed") === "true",
          pressed = state === "pressed" || button.hasAttribute("data-pressed");
        const shape = (name) =>
          componentRadius(source(name), rect.width, rect.height);
        const outer = shape(
          source(p + "outer-corner.corner-size")
            ? p + "outer-corner.corner-size"
            : p + "container.shape",
        );
        let inner;
        if (split)
          inner = shape(
            p +
              (pressed
                ? "inner-corner.pressed.corner-size"
                : last && button.getAttribute("aria-expanded") === "true"
                  ? "trailing-button.inner-corner.selected.corner-size"
                  : state === "hovered" || button.matches(":hover")
                    ? "inner-corner.hovered.corner-size"
                    : "inner-corner.corner-size"),
          );
        else
          inner = shape(
            p +
              (pressed ? "pressed." : selected ? "selected." : "") +
              "inner-corner.corner-size",
          );
        const left = (first && !rtl) || (last && rtl),
          right = (last && !rtl) || (first && rtl);
        for (const [corner, outside] of [
          ["TopLeft", left],
          ["BottomLeft", left],
          ["TopRight", right],
          ["BottomRight", right],
        ])
          expect(
            root,
            corner + " " + index,
            componentRadius(
              cs["border" + corner + "Radius"],
              rect.width,
              rect.height,
            ),
            outside ? outer : inner,
          );
      }
      const mark = button.querySelector(".md-group-check");
      if (mark) {
        const label = mark.nextElementSibling.getBoundingClientRect(),
          selected = button.getAttribute("aria-pressed") === "true",
          icon = mark.getBoundingClientRect(),
          contentLeft = selected ? Math.min(icon.left,label.left) : label.left,
          contentRight = selected ? Math.max(icon.right,label.right) : label.right;
        expect(root,"visible content center " + index,(contentLeft+contentRight)/2,rect.left+rect.width/2);
        checked++;
        if (
          (getComputedStyle(mark).visibility === "visible") !==
          (button.getAttribute("aria-pressed") === "true")
        )
          failures.push({
            id: root.dataset.componentCase,
            property: "selection mark " + index,
          });
      }
    });
    if (
      !split &&
      control === "standard" &&
      root.dataset.componentState === "pressed"
    ) {
      const peer = roots.find(
        (x) =>
          x !== root &&
          x.dataset.componentState === "enabled" &&
          x.dataset.control === control &&
          x.dataset.size === size &&
          x.dataset.variant === root.dataset.variant,
      );
      if (peer) {
        const base = [...peer.querySelectorAll(":scope > .md-button")].map(
            (x) => x.getBoundingClientRect().width,
          ),
          actual = buttons.map((x) => x.getBoundingClientRect().width),
          limits = buttons.map((x) => {
            const s = getComputedStyle(x);
            return Math.min(
              Number.parseFloat(s.paddingInlineStart),
              Number.parseFloat(s.paddingInlineEnd),
            );
          }),
          growth = Math.min(
            (base[1] * number(p + "pressed.item.width.multiplier")) / 200,
            limits[0],
            limits[2],
          );
        expect(root, "pressed width", actual[1], base[1] + 2 * growth);
        expect(root, "previous width", actual[0], base[0] - growth);
        expect(root, "next width", actual[2], base[2] - growth);
        expect(
          root,
          "constant group width",
          actual.reduce((a, b) => a + b, 0),
          base.reduce((a, b) => a + b, 0),
        );
      }
    }
  }
  return { cases, checked, failureCount: failures.length, failures };
}
