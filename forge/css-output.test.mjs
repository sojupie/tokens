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

test("modules own source files and binding families rather than selectors", async () => {
  const { readFile } = await import("node:fs/promises");
  const modules = JSON.parse(
    await readFile(
      new URL("./components/modules.json", import.meta.url),
      "utf8",
    ),
  );
  const bindings = new Map(
    Object.values(modules)
      .flatMap((module) => module.bindings)
      .map((family) => [family, `.binding-${family}{--owned:1}`]),
  );
  const bundles = await componentBundles({
    bindings,
    motionCSS: ".motion{--x:1}",
    elevation: ".elevation{--x:2}",
    allCSS: "legacy",
  });
  assert.equal(bundles.get("all"), "legacy");
  assert.match(bundles.get("list"), /\.md-list/);
  assert.match(bundles.get("sheet"), /\.md-bottom-sheet/);
  assert(!bundles.get("list").includes(".md-bottom-sheet"));
  assert(!bundles.get("common").includes(".binding-list{"));
  assert.match(bundles.get("list"), /\.binding-list\{/);
  assert.match(bundles.get("time"), /@property/);
  for (const [id, selector] of [["checkbox", ".md-checkbox"], ["switch", ".md-switch"], ["messages", ".md-snackbar"], ["segmented", ".md-segmented-button"]]) {
    const root=postcss.parse(bundles.get(id));
    assert(root.nodes.some(rule=>rule.type==="rule"&&rule.selectors.includes(selector)&&rule.nodes.some(decl=>decl.prop==="display")),`${id} must own its layout`);
  }
  assert(!bundles.get("switch").includes(".md-checkbox"));
  assert(!bundles.get("badge").includes(".md-snackbar"));
  assert(bundles.get("common").includes(".md-menu-shortcut"));
  for (const [id, module] of Object.entries(modules)) {
    for (const family of module.bindings)
      assert(bundles.get(id).includes(`.binding-${family}{`));
  }
  bindings.set("unknown", "");
  await assert.rejects(
    componentBundles({ bindings, motionCSS: "", elevation: "", allCSS: "" }),
    /Unowned generated binding/,
  );
});
