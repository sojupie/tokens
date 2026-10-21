import test from "node:test";
import assert from "node:assert/strict";
import postcss from "postcss";
import { groupBindings, themeVariantsCSS } from "./css-output.mjs";
import { componentBundles } from "./component-bundles.mjs";

test("binding grouping preserves intervening state and variant precedence", () => {
  const css = groupBindings(".a{--x:1}.b{--z:2}.a{--y:3}.b{--x:4}.a{--x:5}");
  const rules = postcss.parse(css).nodes;
  assert.equal(rules.length, 3);
  assert.deepEqual(
    rules[0].nodes.map((n) => n.prop),
    ["--x", "--y"],
  );
  assert.equal(rules.at(-1).nodes[0].value, "5");
});

test("theme deltas restore defaults when switching in any direction", () => {
  const variants = ["light", "dark"].flatMap((theme) =>
    ["default", "medium", "high"].map((contrast, i) => ({
      selector: `:root[data-theme="${theme}"][data-contrast="${contrast}"]`,
      theme,
      values: {
        shared: "10px",
        color: theme === "light" ? "white" : "black",
        outline: String(i),
        "md-ref-palette-neutral0": "black",
        "md-sys-color-on-primary": "white",
      },
    })),
  );
  const names = new Set(Object.keys(variants[0].values));
  const css = themeVariantsCSS(variants, names),
    rules = postcss.parse(css).nodes;
  assert.equal((css.match(/--shared:/g) || []).length, 1);
  for (const variant of [...variants, ...variants].reverse()) {
    const effective = {};
    for (const rule of rules)
      if (
        rule.selector === ":root" ||
        rule.selector.replaceAll('"', "") ===
          variant.selector.replaceAll('"', "")
      )
        for (const declaration of rule.nodes)
          effective[declaration.prop] = declaration.value;
    for (const name of names)
      assert.equal(effective["--" + name], variant.values[name]);
  }
});

test("page bundles preserve source order, conditional rules and shared controls", () => {
  const css =
    '.md-icon-button{--x:1}.md-date-panel{--x:2}@media(a){.md-date-panel{--x:3}.md-slider{--x:4}}.md-date-panel{--x:5}@property --_time-angle{syntax:"<angle>";inherits:true;initial-value:0deg}';
  const bundles = componentBundles(css);
  assert(!bundles.get("common").includes(".md-date-panel"));
  assert(bundles.get("date").includes("@media"));
  assert(!bundles.get("date").includes(".md-slider"));
  assert(!bundles.get("date").includes("@property"));
  assert(bundles.get("time").includes("@property"));
  const values = [];
  postcss
    .parse(bundles.get("date"))
    .walkDecls("--x", (n) => values.push(n.value));
  assert.deepEqual(values, ["2", "3", "5"]);
});

test("cross-family reduced-motion overrides follow each family's transitions", () => {
  const bundles = componentBundles('.md-date-menu-button>svg{transition:transform 1s}.md-search-leading-slot>*{transition:transform 1s}@media(prefers-reduced-motion:reduce){.md-date-menu-button>svg,.md-search-leading-slot>*{transition:none}}');
  for (const name of ['date','search']) {
    const css = bundles.get(name);
    assert(css.indexOf('transition:none') > css.indexOf('transition:transform'));
  }
  assert(!bundles.get('common').includes('transition:none'));
});
