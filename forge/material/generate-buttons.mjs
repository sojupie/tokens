// Material's current button model combines a color set and a size set.
// This is the reviewed mapping from those source concepts to browser properties.
const extension = "dev.tokencatalog.material";
const colors = ["elevated", "filled", "outlined", "text", "tonal"];
const sizes = ["xsmall", "small", "medium", "large", "xlarge"];
const legacy = new Set([
  "md.comp.button",
  ...["elevated", "filled", "filled-tonal", "outlined", "text"].map(
    (v) => `md.comp.${v}-button`,
  ),
]);
const colorProperties = {
  "container.color": "--_button-container",
  "container.opacity": "--_button-container-opacity",
  "container.elevation": "--_elevation-level",
  "container.shadow-color": "--_shadow-color",
  "label-text.color": "--_button-label",
  "label-text.opacity": "--_button-label-opacity",
  "icon.color": "--_button-icon",
  "icon.opacity": "--_button-icon-opacity",
  "outline.color": "--_button-outline",
  "state-layer.color": "--_button-state-color",
  "state-layer.opacity": "--_button-state-opacity",
};
const sizeProperties = {
  "container.height": "--_button-height",
  "icon.size": "--_button-icon-size",
  "label-text": "--_button-font",
  "leading-space": "--_button-leading",
  "trailing-space": "--_button-trailing",
  "icon-label-space": "--_button-gap",
  "outlined.outline.width": "--_button-size-outline-width",
  "container.shape.round": "--_button-round-shape",
  "container.shape.square": "--_button-square-shape",
  "selected.container.shape.round": "--_button-selected-round-shape",
  "selected.container.shape.square": "--_button-selected-square-shape",
  "pressed.container.shape": "--_button-pressed-shape",
  "pressed.container.corner-size.motion.spring.damping":
    "--_button-spring-damping",
  "pressed.container.corner-size.motion.spring.stiffness":
    "--_button-spring-stiffness",
};
const enabled = ':not(:disabled):not([aria-disabled="true"])';
const states = {
  hovered: `:is(:hover,[data-button-state="hovered"])${enabled}:not(:where([data-touch]))`,
  focused: `:is(:focus-visible,[data-button-state="focused"])${enabled}`,
  pressed: `:is(:active,[data-pressed],[data-button-state="pressed"])${enabled}:not([data-button-cancelled])`,
  disabled: ':is(:disabled,[aria-disabled="true"])',
};

export function buttonBinding(suffix, size = false) {
  if (size) {
    const property = sizeProperties[suffix];
    if (!property) throw new Error(`Unmapped button size token: ${suffix}`);
    return { property, state: "", selection: "" };
  }
  const parts = suffix.split(".");
  const selection = ["selected", "unselected"].includes(parts[0])
    ? parts.shift()
    : "";
  const state = Object.hasOwn(states, parts[0]) ? parts.shift() : "";
  const property = colorProperties[parts.join(".")];
  if (!property) throw new Error(`Unmapped button color token: ${suffix}`);
  return { property, state, selection };
}

function expectedColorTokens(variant) {
  const tokens = [];
  const add = (prefix, properties) =>
    tokens.push(...properties.map((p) => prefix + p));
  const interactions = ["hovered", "focused", "pressed"];
  if (variant === "text") {
    add("", ["label-text.color", "icon.color"]);
    add("disabled.", [
      "container.color",
      "container.opacity",
      "label-text.color",
      "label-text.opacity",
      "icon.color",
      "icon.opacity",
    ]);
    for (const state of interactions)
      add(state + ".", [
        "label-text.color",
        "icon.color",
        "state-layer.color",
        "state-layer.opacity",
      ]);
  } else if (variant === "outlined") {
    add("", ["label-text.color", "icon.color", "outline.color"]);
    add("disabled.", [
      "container.opacity",
      "icon.color",
      "icon.opacity",
      "label-text.color",
      "label-text.opacity",
      "outline.color",
    ]);
    for (const state of interactions)
      add(state + ".", [
        "icon.color",
        "label-text.color",
        "outline.color",
        "state-layer.color",
        "state-layer.opacity",
      ]);
    add("selected.", ["container.color", "icon.color", "label-text.color"]);
    add("selected.disabled.", ["container.color"]);
    for (const state of interactions)
      add("selected." + state + ".", [
        "icon.color",
        "label-text.color",
        "state-layer.color",
      ]);
    add("unselected.", ["icon.color", "label-text.color"]);
    add("unselected.disabled.", ["outline.color"]);
    for (const state of interactions)
      add("unselected." + state + ".", [
        "icon.color",
        "label-text.color",
        "outline.color",
        "state-layer.color",
      ]);
  } else {
    add("", [
      "container.color",
      "container.elevation",
      "container.shadow-color",
      "label-text.color",
      "icon.color",
    ]);
    add("disabled.", [
      "container.color",
      "container.opacity",
      "container.elevation",
      "label-text.color",
      "label-text.opacity",
      "icon.color",
      "icon.opacity",
    ]);
    for (const state of interactions)
      add(state + ".", [
        "container.elevation",
        "label-text.color",
        "icon.color",
        "state-layer.color",
        "state-layer.opacity",
      ]);
    for (const selection of ["selected", "unselected"]) {
      add(selection + ".", [
        "container.color",
        "label-text.color",
        "icon.color",
      ]);
      for (const state of interactions)
        add(selection + "." + state + ".", [
          "label-text.color",
          "icon.color",
          "state-layer.color",
        ]);
    }
  }
  return tokens;
}

function sourceNodes(document) {
  const nodes = new Map();
  function walk(node) {
    const meta = node.$extensions?.[extension];
    if (meta?.sourceName) nodes.set(meta.sourceName, { node, meta });
    for (const [key, value] of Object.entries(node))
      if (
        (!key.startsWith("$") || key === "$root") &&
        value &&
        typeof value === "object"
      )
        walk(value);
  }
  walk(document);
  return nodes;
}

export function generateButtons(document, values) {
  const groups = new Map();
  const excluded = [];
  const source = sourceNodes(document);
  for (const [name, { meta }] of source) {
    const set = meta.presentation?.setName;
    if (!set) {
      if (name.startsWith("md.comp.button."))
        throw new Error(`Button source set is missing: ${name}`);
      continue;
    }
    if (legacy.has(set)) {
      excluded.push({
        name,
        reason:
          set === "md.comp.button"
            ? "Superseded baseline set"
            : "Deprecated source set",
      });
      continue;
    }
    if (!set.startsWith("md.comp.button.")) continue;
    const variant = set.slice("md.comp.button.".length);
    const size = sizes.includes(variant);
    if (!size && !colors.includes(variant))
      throw new Error(`Unmapped button set: ${set}`);
    const suffix = name.slice(set.length + 1);
    const binding = buttonBinding(suffix, size);
    const key = size
      ? `size:${variant}`
      : `color:${variant}:${binding.state}:${binding.selection}`;
    if (!groups.has(key)) groups.set(key, []);
    const variable = name.replaceAll(".", "-");
    if (!(variable in values))
      throw new Error(`Missing converted button value: ${name}`);
    groups.get(key).push({ name, variable, ...binding });
  }
  const requiredNames = new Set();
  const mapped = [];
  const rules = [];
  function emit(selector, bindings) {
    rules.push(
      `${selector} {\n${bindings
        .flatMap((b) => [
          `  ${b.property}: var(--${b.variable});`,
          ...(!["pressed", "disabled"].includes(b.state) &&
          /^--_button-state-(color|opacity)$/.test(b.property)
            ? [
                `  ${b.property.replace("state-", "background-state-")}: var(--${b.variable});`,
              ]
            : []),
        ])
        .join("\n")}\n}`,
    );
    for (const b of bindings) {
      if (
        b.state === "pressed" &&
        /^--_button-state-(color|opacity)$/.test(b.property)
      )
        rules.push(
          `${selector.replace(states.pressed, "")} { ${b.property.replace("state-", "pressed-state-")}: var(--${b.variable}); }`,
        );
      requiredNames.add(b.variable);
      mapped.push({ token: b.name, selector, property: b.property });
    }
  }
  for (const size of sizes) {
    const bindings = groups.get(`size:${size}`);
    if (!bindings || bindings.length !== Object.keys(sizeProperties).length)
      throw new Error(`Incomplete button size: ${size}`);
    for (const property of ["damping", "stiffness"]) {
      const name = `md-comp-button-${size}-pressed-container-corner-size-motion-spring-${property}`;
      if (!(Number(values[name]) > 0))
        throw new Error(`Invalid button spring: ${name}`);
    }
    const selector =
      size === "small"
        ? `.md-button:not([data-size]), .md-button[data-size="small"]`
        : `.md-button[data-size="${size}"]`;
    emit(selector, bindings);
    const tracking = `md-comp-button-${size}-label-text-letter-spacing`;
    if (!(tracking in values))
      throw new Error(`Missing typography letter spacing: ${size}`);
    requiredNames.add(tracking);
    rules.push(`${selector} { --_button-letter-spacing: var(--${tracking}); }`);
  }
  const configurations = [];
  for (const variant of colors) {
    for (const suffix of expectedColorTokens(variant)) {
      const name = `md.comp.button.${variant}.${suffix}`;
      if (!source.has(name))
        throw new Error(`Missing required button token: ${name}`);
    }

    const base = groups.get(`color:${variant}::`);
    if (
      !base?.some((b) => b.property === "--_button-label") ||
      !base?.some((b) => b.property === "--_button-icon")
    )
      throw new Error(`Incomplete button variant: ${variant}`);
    const selections = groups.has(`color:${variant}::selected`)
      ? ["", "unselected", "selected"]
      : [""];
    for (const state of ["", "hovered", "focused", "pressed", "disabled"]) {
      if (state && !groups.has(`color:${variant}:${state}:`))
        throw new Error(`Missing ${variant} ${state} state`);
      for (const selection of selections) {
        const bindings = groups.get(`color:${variant}:${state}:${selection}`);
        if (!bindings) continue; // Source state overrides inherit the reviewed base properties.
        const selector = `.md-${variant}-button${selection ? `[aria-pressed="${selection === "selected"}"]` : ""}${state ? states[state] : ""}`;
        emit(selector, bindings);
      }
    }
    for (const size of sizes)
      for (const shape of ["round", "square"])
        for (const selection of selections)
          for (const state of [
            "enabled",
            "hovered",
            "focused",
            "pressed",
            "disabled",
          ]) {
            configurations.push({ variant, size, shape, selection, state });
          }
  }
  const css = `/* Generated by forge/material/generate-buttons.mjs from current Material button sets. */
.md-button {
  --_button-container: transparent; --_button-container-opacity: 1;
  --_button-label-opacity: 1; --_button-icon-opacity: 1;
  --_button-state-color: currentColor; --_button-state-opacity: 0;
  --_button-background-state-color: currentColor; --_button-background-state-opacity: 0;
  --_button-outline: transparent; --_button-outline-width: 0px;
  --_elevation-level: 0; --_shadow-color: transparent;
  --_button-rest-shape: var(--_button-round-shape);
  --_button-shape: var(--_button-rest-shape);
  --_button-effects-duration: var(--_motion-fast-effects-duration);
  --_button-easing: var(--_motion-fast-effects-easing);
  position: relative; isolation: isolate; display: inline-flex; touch-action: manipulation; -webkit-tap-highlight-color: transparent; user-select: none;
  align-items: center; justify-content: center; flex: none;
  min-height: var(--_button-height); min-inline-size: max-content; white-space: nowrap;
  padding-block: 0; padding-inline: var(--_button-leading) var(--_button-trailing);
  gap: var(--_button-gap); font: var(--_button-font);
  letter-spacing: var(--_button-letter-spacing); text-decoration: none;
  border: var(--_button-outline-width) solid var(--_button-outline);
  border-radius: var(--_button-shape); vertical-align: middle;
  color: color-mix(in srgb,var(--_button-label) calc(var(--_button-label-opacity) * 100%),transparent);
  background: color-mix(in srgb,var(--_button-container) calc(var(--_button-container-opacity) * 100%),transparent);
  transition: background-color var(--_button-effects-duration) var(--_button-easing), color var(--_button-effects-duration) var(--_button-easing), border-color var(--_button-effects-duration) var(--_button-easing), box-shadow var(--_button-effects-duration) var(--_button-easing);
}
.md-button[hidden] { display: none; }
.md-button[data-shape="square"] { --_button-rest-shape: var(--_button-square-shape); }
.md-button[aria-pressed="true"] { --_button-rest-shape: var(--_button-selected-round-shape); }
.md-button[aria-pressed="true"][data-shape="square"] { --_button-rest-shape: var(--_button-selected-square-shape); }
.md-outlined-button { --_button-outline-width: var(--_button-size-outline-width); }
.md-outlined-button[aria-pressed="true"] { --_button-outline-width: 0px; --_button-outline: transparent; }
.md-button:is(:active,[data-pressed],[data-button-state="pressed"])${enabled}:not([data-button-cancelled]) { --_button-shape: var(--_button-pressed-shape); }
.md-button::before { content: ""; position: absolute; inset: 0; z-index: -1; pointer-events: none; border-radius: inherit; background: var(--_button-background-state-color); opacity: var(--_button-background-state-opacity); transition: opacity 15ms linear, background-color 15ms linear; }
.md-button > svg { flex: none; width: var(--_button-icon-size); height: var(--_button-icon-size); color: color-mix(in srgb,var(--_button-icon) calc(var(--_button-icon-opacity) * 100%),transparent); }
.md-button:is(:focus-visible,[data-button-state="focused"])${enabled} { outline: var(--md-sys-state-focus-indicator-thickness) solid var(--md-sys-color-secondary); outline-offset: var(--md-sys-state-focus-indicator-outer-offset); }
.md-button:is(:disabled,[aria-disabled="true"]) { cursor: default; }
:root:not([data-feedback-ready]) :is(.md-button,.md-icon-button,.md-interactive-card,.md-filter-chip,.md-assist-chip,.md-suggestion-chip,.md-input-chip,.md-menu-item)::before { background: var(--_button-state-color); opacity: var(--_button-state-opacity); }
/* A forced gallery state is already the resting appearance of that specimen. */
:root :is(.md-button,.md-icon-button,.md-interactive-card,.md-filter-chip,.md-assist-chip,.md-suggestion-chip,.md-input-chip,.md-menu-item):is([data-button-state="pressed"],[data-component-state="pressed"]) { --_button-background-state-color: var(--_button-state-color); --_button-background-state-opacity: var(--_button-state-opacity); }
.md-button-ripple { position: absolute; inset: 0; z-index: -1; overflow: hidden; border-radius: inherit; pointer-events: none; opacity: 0; }
.md-button-ripple > span { position: absolute; border-radius: 50%; transform-origin: center center; background: radial-gradient(closest-side, var(--_button-pressed-state-color) max(calc(100% - 70px), 65%), transparent 100%); opacity: var(--_button-pressed-state-opacity); }
@media (forced-colors: active) { .md-button-ripple, :is(.md-button,.md-icon-button,.md-interactive-card,.md-filter-chip,.md-assist-chip,.md-suggestion-chip,.md-input-chip,.md-menu-item)::before { display: none; } }

@media (prefers-reduced-motion: reduce) { .md-button, .md-button::before { transition: none; } .md-button::before { background: var(--_button-state-color); opacity: var(--_button-state-opacity); } }
${rules.join("\n")}
`;
  for (const match of css.matchAll(/var\(\s*--(md-[a-z0-9-]+)\s*[,)]/g))
    requiredNames.add(match[1]);
  for (const name of requiredNames)
    if (!(name in values)) throw new Error(`Missing button binding: ${name}`);
  return {
    css,
    requiredNames,
    coverage: { mapped, excluded, variants: colors, sizes, configurations },
  };
}
