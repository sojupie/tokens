import postcss from "postcss";
import { compactCSS } from "./css-output.mjs";

// Follow CSS references in every appearance, including nested fallback values.
// DTCG aliases have already been resolved by Style Dictionary at this stage.
export function themeSubset(
  theme,
  sources,
  { scope = "", exclude = new Set() } = {},
) {
  const root = postcss.parse(theme),
    values = new Map(),
    wanted = new Set();
  root.walkDecls((declaration) => {
    if (!declaration.prop.startsWith("--")) return;
    const entries = values.get(declaration.prop) ?? [];
    entries.push(declaration.value);
    values.set(declaration.prop, entries);
  });
  function include(source) {
    for (const match of source.matchAll(/var\(\s*(--[\w-]+)/g)) {
      const name = match[1];
      if (wanted.has(name)) continue;
      if (!values.has(name)) {
        if (/^--(?:md|catalog)-/.test(name))
          throw new Error(`Missing theme dependency: ${name}`);
        continue; // Component-local properties are defined by component CSS.
      }
      wanted.add(name);
      for (const value of values.get(name)) include(String(value));
    }
  }
  for (const source of sources) include(source);
  for (const name of exclude) wanted.delete(name);
  root.walkDecls((declaration) => {
    if (declaration.prop.startsWith("--") && !wanted.has(declaration.prop))
      declaration.remove();
  });
  if (scope)
    root.walkRules((rule) => {
      rule.selector = rule.selectors
        .map((selector) => `${selector} ${scope}`)
        .join(",");
      rule.walkDecls("color-scheme", (declaration) => declaration.remove());
    });
  return { css: compactCSS(root.toString()), names: wanted };
}

// CSS and literal runtime reads form the dependency roots. Spring readers build
// their property names dynamically, so their twelve inputs stay explicit here.
export function pageThemes(theme, bundles, sharedSources, runtimeSources) {
  const inputs = [
    ...sharedSources,
    bundles.get("common"),
    "var(--tokens-preview-on-light) var(--tokens-preview-on-dark)",
  ];
  for (const source of runtimeSources)
    for (const match of source.matchAll(/["'`](--(?:md|catalog)-[\w-]+)["'`]/g))
      inputs.push(`var(${match[1]})`);
  for (const speed of ["fast", "default", "slow"])
    for (const type of ["spatial", "effects"])
      for (const property of ["damping", "stiffness"])
        inputs.push(`var(--md-sys-motion-spring-${speed}-${type}-${property})`);
  const shared = themeSubset(theme, inputs),
    themes = new Map([["shared", shared]]);
  for (const [name, css] of bundles) {
    if (name === "common") continue;
    themes.set(
      name,
      themeSubset(theme, [css], {
        scope: name === "all" ? "" : "#component-preview",
        exclude: shared.names,
      }),
    );
  }
  return themes;
}
