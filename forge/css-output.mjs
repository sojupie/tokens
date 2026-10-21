import postcss from "postcss";
import { transformSync } from "esbuild";

export function compactCSS(css) {
  return transformSync(css, {
    loader: "css",
    minifyWhitespace: true,
    legalComments: "inline",
  }).code;
}

// Move custom-property declarations into an earlier identical selector only if
// no intervening rule writes those properties. Preserve cascade precedence.
export function groupBindings(css) {
  const root = postcss.parse(css);
  const candidates = new Map(),
    lastWrite = new Map();
  for (const [index, rule] of [...root.nodes].entries()) {
    if (rule.type === "comment") continue;
    if (
      rule.type !== "rule" ||
      rule.nodes.some((n) => n.type !== "decl" || !n.prop.startsWith("--"))
    ) {
      candidates.clear();
      lastWrite.clear();
      continue;
    }
    const properties = rule.nodes.map((n) => n.prop);
    const earlier = candidates.get(rule.selector);
    if (
      earlier &&
      rule.nodes.every((n) => (lastWrite.get(n.prop) ?? -1) <= earlier.index)
    ) {
      for (const declaration of [...rule.nodes])
        earlier.rule.append(declaration);
      rule.remove();
    } else candidates.set(rule.selector, { rule, index });
    for (const property of properties) {
      // A conservative conflict barrier is preferable to moving a declaration
      // past another state/variant with the same custom property.
      lastWrite.set(property, index);
    }
  }
  return root.toString();
}

export function themeVariantsCSS(variants, names) {
  if (!variants.length) throw new Error("A default theme is required");
  const base = variants[0];
  const declarations = (values) =>
    [...names]
      .sort()
      .map((name) => {
        if (!(name in values))
          throw new Error(`Missing theme variable: ${name}`);
        return `--${name}:${values[name]};`;
      })
      .join("");
  let css =
    `:root{color-scheme:${base.theme};${declarations(base.values)}` +
    `--tokens-preview-on-light:${base.values["md-ref-palette-neutral0"]};` +
    `--tokens-preview-on-dark:${base.values["md-sys-color-on-primary"]};}\n`;
  for (const variant of variants) {
    const changed = new Set(
      [...names].filter((name) => {
        if (!(name in variant.values))
          throw new Error(`Missing theme variable: ${name}`);
        return variant.values[name] !== base.values[name];
      }),
    );
    css +=
      `${variant.selector}{color-scheme:${variant.theme};` +
      [...changed]
        .sort()
        .map((name) => `--${name}:${variant.values[name]};`)
        .join("") +
      "}\n";
  }
  return compactCSS(css);
}
