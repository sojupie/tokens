// Read-only checks of actual rendered properties against captured source variables.
export function checkComponentRendering() {
  const failures = [];
  let checked = 0,
    cases = 0;
  const color = (value) => {
    if (value === "transparent") return [0, 0, 0, 0];
    if (value.startsWith("#"))
      return [1, 3, 5]
        .map((i) => Number.parseInt(value.slice(i, i + 2), 16) / 255)
        .concat(
          value.length === 9 ? Number.parseInt(value.slice(7, 9), 16) / 255 : 1,
        );
    const n = (value.match(/-?\d*\.?\d+(?:e[+-]?\d+)?/g) || []).map(Number);
    return value.startsWith("rgb")
      ? n
          .slice(0, 3)
          .map((x) => x / 255)
          .concat(n[3] ?? 1)
      : n.slice(0, 3).concat(n[3] ?? 1);
  };
  for (const el of document.querySelectorAll("[data-component-case]")) {
    cases++;
    const cs = getComputedStyle(el),
      f = el.dataset.family,
      state = el.dataset.componentState;
    const short = { hovered: "hover", focused: "focus" }[state] || state;
    const selected =
      el.checked ||
      el.getAttribute("aria-pressed") === "true" ||
      el.getAttribute("aria-current") === "page" ||
      el.querySelector("input")?.checked;
    const sel = selected ? "selected" : "unselected";
    const get = (name) =>
      cs.getPropertyValue("--" + name.replaceAll(".", "-")).trim();
    const first = (names, fallback) => names.map(get).find(Boolean) ?? fallback;
    function compare(name, actual, expected, colors = false, opacity = 1) {
      checked++;
      let okay = actual === expected;
      if (colors) {
        const a = color(actual),
          b = color(expected || "");
        b[3] *= opacity;
        okay =
          a.length === 4 &&
          b.length === 4 &&
          a.every((x, i) => Math.abs(x - b[i]) < 0.001);
      } else if (expected?.includes("infinity"))
        okay = Number.parseFloat(actual) > 1000000;
      else if (
        Number.isFinite(Number.parseFloat(actual)) &&
        Number.isFinite(Number.parseFloat(expected))
      )
        okay =
          Math.abs(Number.parseFloat(actual) - Number.parseFloat(expected)) <
          0.01;
      if (!okay)
        failures.push({
          id: el.dataset.componentCase,
          property: name,
          actual,
          expected,
          opacity,
        });
    }
    function c(name, expected, opacity = 1) {
      compare(name, cs[name], expected, true, Number(opacity));
    }
    function value(name, expected) {
      compare(name, cs[name], expected);
    }
    function layer(prefix, colorName, opacityName) {
      const st = getComputedStyle(el, "::before");
      compare("state layer color", st.backgroundColor, colorName, true);
      compare("state layer opacity", st.opacity, opacityName);
    }
    function shadow(element, level) {
      const clamp = (x, max = 1) => Math.max(0, Math.min(x, max));
      const kY = clamp(level) + clamp(level - 3) + 2 * clamp(level - 4);
      const kBlur = 2 * clamp(level) + clamp(level - 2) + clamp(level - 4);
      const aY = clamp(level) + clamp(level - 1) + 2 * clamp(level - 2, 3);
      const aBlur = 3 * clamp(level, 2) + 2 * clamp(level - 2, 3);
      const spread = clamp(level, 4) + 2 * clamp(level - 4);
      const offsets =
        getComputedStyle(element)
          .boxShadow.replace(/rgba?\([^)]*\)|color\([^)]*\)/g, "color")
          .match(/-?\d*\.?\d+px/g) || [];
      const expected = [0, kY, kBlur, 0, 0, aY, aBlur, spread];
      checked++;
      if (
        offsets.length !== 8 ||
        offsets.some(
          (x, i) => Math.abs(Number.parseFloat(x) - expected[i]) > 0.01,
        )
      )
        failures.push({
          id: el.dataset.componentCase,
          property: "shadow geometry",
          actual: offsets,
          expected,
        });
    }
    if (f === "card") {
      const prefix =
        "md.comp." +
        ["elevated", "filled", "outlined"].find((v) =>
          el.classList.contains("md-" + v + "-card"),
        ) +
        "-card.";
      const source = (key) => first([prefix + short + "." + key, prefix + key]);
      c(
        "backgroundColor",
        source("container.color"),
        source("container.opacity") || 1,
      );
      value("borderTopWidth", get(prefix + "outline.width") || "0px");
      if (cs.borderTopWidth !== "0px")
        c(
          "borderTopColor",
          source("outline.color"),
          source("outline.opacity") || 1,
        );
      value("borderTopLeftRadius", get(prefix + "container.shape"));
      compare(
        "elevation",
        cs.getPropertyValue("--_elevation-level").trim(),
        source("container.elevation"),
      );
      compare(
        "shadow color",
        cs.getPropertyValue("--_shadow-color").trim(),
        get(prefix + "container.shadow-color"),
        true,
      );
      shadow(el, Number(source("container.elevation")));
      const icon = el.querySelector("[data-slot=icon]");
      if (icon) {
        compare(
          "card icon width",
          getComputedStyle(icon).width,
          get(prefix + "icon.size"),
        );
        compare(
          "card icon color",
          getComputedStyle(icon).color,
          get(prefix + "icon.color"),
          true,
        );
      }
      if (["hovered", "focused", "pressed", "dragged"].includes(state))
        layer(
          prefix,
          get(prefix + short + ".state-layer.color"),
          get(prefix + short + ".state-layer.opacity"),
        );
    } else if (f === "icon") {
      const variant = ["standard", "filled", "tonal", "outlined"].find((v) =>
        el.classList.contains("md-" + v + "-icon-button"),
      );
      const p = "md.comp.icon-button." + variant + ".",
        size = "md.comp.icon-button." + el.dataset.size + ".";
      const source = (key) =>
        state === "disabled"
          ? first([
              p + sel + ".disabled." + key,
              p + "disabled." + key,
              p + sel + "." + key,
              p + key,
            ])
          : first([
              p + sel + "." + state + "." + key,
              p + state + "." + key,
              p + sel + "." + key,
              p + key,
            ]);
      c(
        "backgroundColor",
        source("container.color") || "transparent",
        source("container.opacity") || 1,
      );
      c("color", source("icon.color"), source("icon.opacity") || 1);
      value("height", get(size + "container.height"));
      value("paddingLeft", get(size + el.dataset.width + ".leading-space"));
      value("paddingRight", get(size + el.dataset.width + ".trailing-space"));
      const icon = el.querySelector("svg");
      if (icon)
        compare(
          "icon width",
          getComputedStyle(icon).width,
          get(size + "icon.size"),
        );
      value(
        "borderTopWidth",
        variant === "outlined" && !selected
          ? get(size + "outlined.outline.width")
          : "0px",
      );
      const shape =
        state === "pressed"
          ? "pressed.container.shape"
          : (selected ? "selected." : "") +
            "container.shape." +
            el.dataset.shape;
      value("borderTopLeftRadius", get(size + shape));
      if (["hovered", "focused", "pressed"].includes(state))
        layer(p, source("state-layer.color"), source("state-layer.opacity"));
    } else if (f === "chip") {
      const p = "md.comp.filter-chip.",
        elevated = el.dataset.elevated === "true",
        variant = elevated ? "elevated" : "flat";
      const source = (key) =>
        first([
          p +
            (state === "disabled" ? "disabled." : "") +
            sel +
            "." +
            (state === "disabled" ? "" : short + ".") +
            key,
          p + sel + "." + short + "." + key,
          p + sel + "." + key,
          p + (state === "disabled" ? "disabled." : "") + key,
          p + key,
        ]);
      const bg =
        state === "disabled"
          ? first(
              [
                p + variant + ".disabled." + sel + ".container.color",
                p + variant + ".disabled.container.color",
                p + variant + "." + sel + ".container.color",
              ],
              "transparent",
            )
          : first(
              [p + variant + "." + sel + ".container.color"],
              "transparent",
            );
      const opacity =
        state === "disabled"
          ? first(
              [
                p + variant + ".disabled." + sel + ".container.opacity",
                p + variant + ".disabled.container.opacity",
              ],
              "1",
            )
          : "1";
      c("backgroundColor", bg, opacity);
      c(
        "color",
        state === "disabled"
          ? get(p + "disabled.label-text.color")
          : source("label-text.color"),
        state === "disabled" ? get(p + "disabled.label-text.opacity") : 1,
      );
      const elevation =
        state === "dragged"
          ? get(p + "dragged.container.elevation")
          : first([
              p +
                variant +
                "." +
                (state === "disabled" ? "disabled." : "") +
                sel +
                "." +
                (state === "disabled" ? "" : short + ".") +
                "container.elevation",
              p + variant + "." + short + ".container.elevation",
              p + variant + "." + sel + "." + short + ".container.elevation",
              p + variant + ".container.elevation",
            ]);
      compare(
        "chip elevation",
        cs.getPropertyValue("--_elevation-level").trim(),
        elevation,
      );
      shadow(el, Number(elevation));
      value("minHeight", get(p + "container.height"));
      value("fontSize", get(p + "label-text.size"));
      value("fontWeight", get(p + "label-text.weight"));
      value("letterSpacing", get(p + "label-text.tracking"));
      if (["hovered", "focused", "pressed", "dragged"].includes(state))
        layer(p, source("state-layer.color"), source("state-layer.opacity"));
    } else if (f === "radio") {
      const p = "md.comp.radio-button.";
      const prefix =
        p +
        (state === "disabled" ? "disabled." : "") +
        sel +
        "." +
        (["hovered", "focused", "pressed"].includes(state) ? short + "." : "");
      c(
        "color",
        get(prefix + "icon.color"),
        state === "disabled" ? get(prefix + "icon.opacity") : 1,
      );
      value("width", get(p + "state-layer.size"));
      value("height", get(p + "state-layer.size"));
      compare(
        "radio icon width",
        getComputedStyle(el, "::before").width,
        get(p + "icon.size"),
      );
      c(
        "backgroundColor",
        ["hovered", "focused", "pressed"].includes(state)
          ? get(prefix + "state-layer.color")
          : "transparent",
        ["hovered", "focused", "pressed"].includes(state)
          ? get(prefix + "state-layer.opacity")
          : 1,
      );
    } else if (f === "field" || f === "select") {
      const p =
        f === "field"
          ? "md.comp.outlined-text-field."
          : "md.comp.outlined-select.text-field.";
      const invalid = el.getAttribute("aria-invalid") === "true";
      const prefixes =
        state === "disabled"
          ? ["disabled."]
          : invalid
            ? [
                "error." + (short === "enabled" ? "" : short + "."),
                "error.",
                short + ".",
                "",
              ]
            : [short + ".", ""];
      const source = (key) => first(prefixes.map((s) => p + s + key));
      c("color", source("input-text.color"), source("input-text.opacity") || 1);
      c(
        "borderTopColor",
        source("outline.color"),
        source("outline.opacity") || 1,
      );
      value("borderTopWidth", source("outline.width"));
      value("minHeight", get(p + "container.height"));
      value("height", get(p + "container.height"));
      const arrow = el.querySelector('[data-slot="trailing-icon"]');
      if (arrow) {
        compare(
          "select icon color",
          getComputedStyle(arrow).color,
          source("trailing-icon.color"),
          true,
          source("trailing-icon.opacity") || 1,
        );
        compare(
          "select icon size",
          getComputedStyle(arrow).width,
          get(p + "trailing-icon.size"),
        );
      }
      value("fontSize", get(p + "input-text.size"));
      value("fontWeight", get(p + "input-text.weight"));
    } else if (f === "menu") {
      const menu = el.closest(".md-menu"),
        p =
          "md.comp.menus." +
          (menu.dataset.variant || "standard") +
          ".menu-item.";
      shadow(menu, Number(get("md.comp.menus.container.elevation")));
      const prefixes =
        state === "disabled"
          ? [sel + ".disabled.", "disabled.", sel + ".", ""]
          : [sel + "." + state + ".", state + ".", sel + ".", ""];
      const source = (key) => first(prefixes.map((s) => p + s + key));
      c("color", source("label-text.color"), source("label-text.opacity") || 1);
      c(
        "backgroundColor",
        source("container.color") || first([p + "color"], "transparent"),
        source("container.opacity") || 1,
      );
      value("minHeight", get("md.comp.menus.menu-item.height"));
      value("fontSize", get("md.comp.menus.menu-item.label-text.size"));
      if (["hovered", "focused", "pressed"].includes(state))
        layer(p, source("state-layer.color"), source("state-layer.opacity"));
    } else if (f === "dialog") {
      shadow(el, Number(get("md.comp.dialog.container.elevation")));
      c("backgroundColor", get("md.comp.dialog.container.color"));
      value("borderTopLeftRadius", get("md.comp.dialog.container.shape"));
      compare(
        "headline color",
        getComputedStyle(el.querySelector("h2")).color,
        get("md.comp.dialog.headline.color"),
        true,
      );
      compare(
        "headline size",
        getComputedStyle(el.querySelector("h2")).fontSize,
        get("md.comp.dialog.headline.size"),
      );
    } else if (f === "divider") {
      value("borderTopWidth", get("md.comp.divider.thickness"));
      c("borderTopColor", get("md.comp.divider.color"));
    }
  }
  return {
    cases,
    checked,
    failureCount: failures.length,
    failures: failures.slice(0, 80),
  };
}
