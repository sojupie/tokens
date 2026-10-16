import { readFileSync } from "node:fs";

const contract = JSON.parse(
  readFileSync(new URL("./component-contract.json", import.meta.url)),
);
const extension = "dev.tokencatalog.material";
const selection =
  ':where(:checked,[aria-pressed="true"],[aria-selected="true"],[aria-current="page"],[data-selected="true"],:has(> input:checked))';
const enabled =
  ':not(:disabled):not([aria-disabled="true"]):not(:has(> input:disabled, > select:disabled))';
const stateNames = {
  hover: "hovered",
  hovered: "hovered",
  focus: "focused",
  focused: "focused",
  pressed: "pressed",
  disabled: "disabled",
  dragged: "dragged",
};
const states = {
  hovered: `:where(:hover,[data-component-state="hovered"])${enabled}:not(:where([data-touch]))`,
  focused: `:where(:focus-visible,:has(> input:focus-visible),[data-component-state="focused"])${enabled}`,
  pressed: `:where(:active,[data-pressed],[data-component-state="pressed"])${enabled}:not(:where([data-button-cancelled]))`,
  dragged: `:where([data-dragged="true"],[data-component-state="dragged"])${enabled}`,
  disabled:
    ':is(:disabled,[aria-disabled="true"],:has(> input:disabled, > select:disabled))',
};
const order = ["", "hovered", "focused", "pressed", "dragged", "disabled"];
const roleProperties = {
  "container.color": "container",
  "container.opacity": "container-opacity",
  "container.height": "height",
  "container.elevation": "elevation",
  "container.shadow-color": "shadow",
  "container.shape": "shape",
  "outline.color": "outline",
  "outline.opacity": "outline-opacity",
  "outline.width": "outline-width",
  "state-layer.color": "state-color",
  "state-layer.opacity": "state-opacity",
  "icon.color": "icon",
  "icon.opacity": "icon-opacity",
  "icon.size": "icon-size",
  "label-text.color": "label",
  "label-text.opacity": "label-opacity",
  "label-text.font": "font-family",
  "label-text.size": "font-size",
  "label-text.line-height": "line-height",
  "label-text.tracking": "tracking",
  "label-text.weight": "font-weight",
  "focus.indicator.color": "focus-color",
  "focus.indicator.outline.offset": "focus-offset",
  "focus.indicator.thickness": "focus-width",
  "focus.indicator.outline.color": "focus-color",
  "focus.indicator.outline.thickness": "focus-width",
};
const textProperty = {
  font: "font-family",
  size: "font-size",
  "line-height": "line-height",
  tracking: "tracking",
  weight: "font-weight",
  color: "color",
  opacity: "opacity",
};

function parse(suffix) {
  let parts = suffix.split("."),
    state = "",
    selected = "",
    variant = "";
  // Focus indicators describe the focus ring, not a transient focus-state value.
  const indicator = suffix.includes("focus.indicator.");
  parts = parts.filter((p, i) => {
    if (
      Object.hasOwn(stateNames, p) &&
      !(indicator && p === "focus" && parts[i + 1] === "indicator")
    ) {
      state = stateNames[p];
      return false;
    }
    if (p === "selected" || p === "unselected") {
      selected = p;
      return false;
    }
    if (p === "flat" || p === "elevated") {
      variant = p;
      return false;
    }
    return true;
  });
  return { sourceProperty: parts.join("."), state, selected, variant };
}
function condition({ state, selected }) {
  return (
    (selected === "selected"
      ? selection
      : selected === "unselected"
        ? `:not(${selection})`
        : "") + (state ? states[state] : "")
  );
}
function textRole(property, roles) {
  for (const role of roles) {
    const tail = property.slice(role.length + 1);
    if (property.startsWith(role + ".") && Object.hasOwn(textProperty, tail))
      return `${role}-${textProperty[tail]}`;
  }
}
function binding(set, suffix) {
  let parsed = parse(suffix),
    property = parsed.sourceProperty,
    selector = "",
    family = "",
    target = "";
  if (/^md.comp.(elevated|filled|outlined)-card$/.test(set)) {
    family = "card";
    selector = ".md-" + set.split(".")[2];
    if (parsed.state) selector += ".md-interactive-card";
  } else if (set.startsWith("md.comp.icon-button.")) {
    family = "icon";
    const name = set.slice("md.comp.icon-button.".length);
    selector = ["xsmall", "small", "medium", "large", "xlarge"].includes(name)
      ? `.md-icon-button[data-size="${name}"]${name === "small" ? ", .md-icon-button:not([data-size])" : ""}`
      : `.md-icon-button.md-${name}-icon-button`;
    const special = {
      "container.shape.round": "round-shape",
      "container.shape.square": "square-shape",
      "container.shape": "shape",
      "outlined.outline.width": "size-outline-width",
      "pressed.container.corner-size.motion.spring.damping": "spring-damping",
      "pressed.container.corner-size.motion.spring.stiffness":
        "spring-stiffness",
    };
    // Size-set pressed geometry and spring parameters are available at rest.
    if (suffix in special) {
      parsed = { state: "", selected: "" };
      return { family, selector, property: special[suffix], ...parsed };
    }
    if (
      property === "container.shape.round" ||
      property === "container.shape.square"
    )
      return {
        family,
        selector,
        property:
          "selected-" +
          (property.endsWith("round") ? "round" : "square") +
          "-shape",
        ...parsed,
      };
    if (property === "container.shape" && parsed.state === "pressed")
      return {
        family,
        selector,
        property: "pressed-shape",
        state: "",
        selected: "",
      };
    const spacing = property.match(
      /^(default|narrow|wide)\.(leading|trailing)-space$/,
    );
    if (spacing) {
      const width =
        spacing[1] === "default"
          ? ':is(:not([data-width]),[data-width="default"])'
          : `[data-width="${spacing[1]}"]`;
      selector = selector
        .split(", ")
        .map((s) => s + width)
        .join(", ");
      return {
        family,
        selector,
        property: spacing[2],
        state: "",
        selected: "",
      };
    }
  } else if (set === "md.comp.filter-chip") {
    family = "chip";
    selector = ".md-filter-chip";
    if (parsed.variant)
      selector +=
        parsed.variant === "flat"
          ? ':where(:not([data-elevated="true"]))'
          : ':where([data-elevated="true"])';
    for (const slot of ["with-leading-icon", "with-trailing-icon", "with-icon"])
      if (property.startsWith(slot + ".")) {
        property = property.slice(slot.length + 1);
        const name = slot === "with-icon" ? "icon" : slot.slice(5);
        property = property.replace(new RegExp("^" + name + "\\."), "icon.");
        target =
          slot === "with-icon"
            ? ""
            : ` > .md-chip-face > [data-slot="${name}"]`;
      }
  } else if (set === "md.comp.dialog") {
    family = "dialog";
    selector = ".md-dialog";
    if (property.startsWith("action.")) {
      selector += " .md-text-button";
      property = property.slice(7);
      family = "dialog-action";
    }
    const extra = {
      "with-divider.divider.color": "divider",
      "with-divider.divider.height": "divider-height",
      "with-icon.icon.color": "icon",
      "with-icon.icon.size": "icon-size",
    };
    if (extra[property])
      return { family, selector, property: extra[property], ...parsed };
  } else if (set === "md.comp.radio-button") {
    family = "radio";
    selector = ".md-radio";
  } else if (
    set === "md.comp.outlined-text-field" ||
    set === "md.comp.outlined-select"
  ) {
    family = "field";
    selector = set.endsWith("-select")
      ? ".md-outlined-select"
      : ".md-outlined-text-field";
    if (property.startsWith("menu."))
      return {
        excluded:
          "Native select popup rendering belongs to the browser; these menu tokens are not applied to the popup.",
      };
    if (property.startsWith("text-field.")) property = property.slice(11);
    if (property.startsWith("error.")) {
      selector += ':is([aria-invalid="true"],:has([aria-invalid="true"]))';
      property = property.slice(6);
    }
    const extras = {
      "leading-icon.size": "leading-icon-size",
      "trailing-icon.size": "trailing-icon-size",
      "caret.color": "caret",
      "input-text.placeholder.color": "placeholder",
      "input-text.prefix.color": "prefix",
      "input-text.suffix.color": "suffix",
      "label-text.populated.size": "populated-label-size",
      "label-text.populated.line-height": "populated-label-line-height",
    };
    if (extras[property])
      return { family, selector, property: extras[property], ...parsed };
  } else if (set.startsWith("md.comp.menus")) {
    family = "menu";
    selector = ".md-menu";
    if (set !== "md.comp.menus")
      selector += set.endsWith(".standard")
        ? ':not([data-variant="vibrant"])'
        : '[data-variant="vibrant"]';
    if (property.startsWith("horizontal."))
      return {
        excluded:
          "Horizontal menus are not implemented; the catalog uses vertical menus.",
      };
    if (
      property.startsWith("icon-button.") ||
      property === "disabled.icon-button.icon.color"
    )
      return {
        excluded:
          "Menu-embedded icon-button anatomy is not used; standalone icon buttons use the current icon-button sets.",
      };
    if (property.startsWith("menu-item.")) {
      property = property.slice(10);
      selector += " .md-menu-item";
      family = "menu-item";
    }
    if (
      property.startsWith("first-child.") ||
      property.startsWith("last-child.")
    ) {
      const part = property.split(".")[0];
      selector = selector
        .split(", ")
        .flatMap((s) => [
          s + `:where(:${part}):not(:where(li > *))`,
          s.replace(" .md-menu-item", ` :where(li:${part}) > .md-menu-item`),
        ])
        .join(", ");
      property = property.slice(part.length + 1);
      if (property === "inner-corner.corner-size")
        return {
          excluded: "First/last item shapes already encode the inner corners.",
        };
    }
    if (
      property === "active.container.shape" ||
      property === "inactive.container.shape"
    )
      return {
        excluded: "Active/inactive menu container morphing is not implemented.",
      };
    if (suffix.startsWith("menu-item.active."))
      return { excluded: "Active submenu trigger states are not implemented." };
    const extra = {
      gap: "gap",
      "group.padding": "group-padding",
      "group.shape": "group-shape",
      color: "container",
      shape: "shape",
      height: "height",
      "between-space": "gap",
      "leading-space": "leading",
      "trailing-space": "trailing",
      "top-space": "top",
      "bottom-space": "bottom",
      "leading-icon.size": "leading-icon-size",
      "trailing-icon.size": "trailing-icon-size",
    };
    if (extra[property])
      return { family, selector, property: extra[property], ...parsed };
  } else if (set === "md.comp.divider") {
    family = "divider";
    selector = ".md-divider";
    return {
      family,
      selector,
      property: { color: "color", thickness: "thickness" }[property],
      ...parsed,
    };
  } else return null;
  if (property === "container.surface-tint-layer.color")
    return {
      excluded:
        "Surface container colors include their tonal treatment; no second tint layer is applied.",
    };
  if (property.endsWith(".type"))
    return {
      excluded:
        "Composite typography is represented by the individual font, size, weight, line-height and tracking tokens.",
    };
  let prop =
    roleProperties[property] ||
    textRole(property, [
      "headline",
      "subhead",
      "supporting-text",
      "input-text",
      "leading-icon",
      "trailing-icon",
      "trailing-supporting-text",
      "section-label-text",
    ]);
  if (
    !prop &&
    [
      "leading-icon.color",
      "trailing-icon.color",
      "leading-icon.opacity",
      "trailing-icon.opacity",
    ].includes(property)
  )
    prop = property.replaceAll(".", "-");
  if (!prop && property === "state-layer.size") prop = "state-size";
  if (!prop)
    throw new Error(`Unmapped component token: ${set}.${suffix} (${property})`);
  return { family, selector, target, ...parsed, property: prop };
}

export function generateComponents(document, values) {
  const source = new Map();
  function walk(node) {
    const meta = node.$extensions?.[extension];
    if (meta?.sourceName) source.set(meta.sourceName, meta);
    for (const [key, child] of Object.entries(node))
      if (
        (!key.startsWith("$") || key === "$root") &&
        child &&
        typeof child === "object"
      )
        walk(child);
  }
  walk(document);
  const mapped = [],
    excluded = [],
    rules = [],
    requiredNames = new Set();
  for (const [set, suffixes] of Object.entries(contract)) {
    for (const suffix of suffixes)
      if (!source.has(set + "." + suffix))
        throw new Error(`Missing required component token: ${set}.${suffix}`);
  }
  const records = [];
  for (const [name, meta] of source) {
    const set = meta.presentation?.setName;
    if (
      !set &&
      /^md\.comp\.(?:elevated-card|filled-card|outlined-card|filter-chip|dialog|radio-button|outlined-text-field|outlined-select|icon-button|menus)\./.test(
        name,
      )
    )
      throw new Error(`Missing component source set: ${name}`);
    if (!(set in contract)) {
      if (
        set === "md.comp.icon-button" ||
        /^md.comp.(filled|filled-tonal|outlined)-icon-button$/.test(
          set || "",
        ) ||
        set === "md.comp.menu"
      ) {
        excluded.push({
          token: name,
          reason:
            "Superseded baseline or deprecated set; current icon-button and menus sets are used.",
        });
        continue;
      }
      if (
        set?.startsWith("md.comp.icon-button.") ||
        set?.startsWith("md.comp.menus.")
      )
        throw new Error(`Unknown component set: ${set}`);
      continue;
    }
    const suffix = name.slice(set.length + 1);
    if (!contract[set].includes(suffix))
      throw new Error(`Unknown component token: ${name}`);
    const b = binding(set, suffix);
    if (!b) throw new Error(`Missing component mapping: ${name}`);
    if (b.excluded) {
      excluded.push({ token: name, reason: b.excluded });
      continue;
    }
    const variable = name.replaceAll(".", "-");
    if (
      name.includes(".corner-size.motion.spring.") &&
      !(Number(values[variable]) > 0)
    )
      throw new Error(`Invalid component spring: ${name}`);
    if (!(variable in values))
      throw new Error(`Missing converted component value: ${name}`);
    if (!b.property) throw new Error(`Missing component property: ${name}`);
    // Indicators retain their normal property selector; interaction states are emitted in precedence order.
    records.push({ ...b, variable, token: name });
  }
  records.sort(
    (a, b) =>
      order.indexOf(a.state) - order.indexOf(b.state) ||
      Number(!!a.selected) - Number(!!b.selected),
  );
  for (const b of records) {
    const fieldCondition =
      b.family === "field" && b.state === "focused"
        ? condition({ ...b, state: "" }) +
          `:where(:focus,:focus-within,[data-component-state="focused"])${enabled}`
        : condition(b);
    const selectors = b.selector
      .split(", ")
      .map((s) => s + fieldCondition + (b.target || ""))
      .join(", ");
    const property =
      b.family === "dialog-action" &&
      ["label", "state-color", "state-opacity"].includes(b.property)
        ? `--_button-${b.property}`
        : `--_${b.family}-${b.property}`;
    rules.push(`${selectors} { ${property}: var(--${b.variable}); }`);
    if (
      !["pressed", "disabled"].includes(b.state) &&
      /state-(color|opacity)$/.test(b.property)
    )
      rules.push(
        `${selectors} { ${property.replace("state-", "background-state-")}: var(--${b.variable}); }`,
      );
    if (b.state === "pressed" && /state-(color|opacity)$/.test(b.property)) {
      const persistent = b.selector
        .split(", ")
        .map((s) => s + condition({ ...b, state: "" }) + (b.target || ""))
        .join(", ");
      rules.push(
        `${persistent} { ${property.replace("state-", "pressed-state-")}: var(--${b.variable}); }`,
      );
    }
    mapped.push({ token: b.token, selector: selectors, property });
    requiredNames.add(b.variable);
  }
  const configurations = [];
  for (const variant of ["elevated", "filled", "outlined"])
    for (const state of [
      "enabled",
      "hovered",
      "focused",
      "pressed",
      "dragged",
      "disabled",
    ])
      configurations.push({ family: "card", variant, state });
  for (const variant of ["standard", "filled", "tonal", "outlined"])
    for (const size of ["xsmall", "small", "medium", "large", "xlarge"])
      for (const shape of ["round", "square"])
        for (const width of ["default", "narrow", "wide"])
          for (const selected of [false, true])
            for (const state of [
              "enabled",
              "hovered",
              "focused",
              "pressed",
              "disabled",
            ])
              configurations.push({
                family: "icon",
                variant,
                size,
                shape,
                width,
                selected,
                state,
              });
  for (const elevated of [false, true])
    for (const selected of [false, true])
      for (const state of [
        "enabled",
        "hovered",
        "focused",
        "pressed",
        "dragged",
        "disabled",
      ])
        configurations.push({ family: "chip", elevated, selected, state });
  for (const selected of [false, true])
    for (const state of [
      "enabled",
      "hovered",
      "focused",
      "pressed",
      "disabled",
    ])
      configurations.push({ family: "radio", selected, state });
  for (const family of ["field", "select"])
    for (const invalid of [false, true])
      for (const state of ["enabled", "hovered", "focused", "disabled"])
        configurations.push({ family, invalid, state });
  for (const variant of ["standard", "vibrant"])
    for (const selected of [false, true])
      for (const state of [
        "enabled",
        "hovered",
        "focused",
        "pressed",
        "disabled",
      ])
        configurations.push({ family: "menu", variant, selected, state });
  configurations.push(
    { family: "dialog", state: "enabled" },
    { family: "divider", state: "enabled" },
  );
  return {
    css:
      "/* Generated from the reviewed Material component contract. */\n" +
      rules.join("\n") +
      "\n",
    coverage: { mapped, excluded, configurations },
    requiredNames,
  };
}
