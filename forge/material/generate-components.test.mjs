import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { generateComponents } from "./generate-components.mjs";
import { compile } from "../build-tokens.mjs";

const contract = JSON.parse(
  await readFile(new URL("./component-contract.json", import.meta.url)),
);
function fixture() {
  const document = {},
    values = {};
  let i = 0;
  for (const [set, suffixes] of Object.entries(contract))
    for (const suffix of suffixes) {
      const name = set + "." + suffix;
      document["token" + i++] = {
        $type: "number",
        $value: 1,
        $extensions: {
          "dev.tokencatalog.material": {
            sourceName: name,
            presentation: { setName: set },
          },
        },
      };
      values[name.replaceAll(".", "-")] = "1";
    }
  return { document, values };
}
test("reviewed contracts map every token or report its explicit exclusion", () => {
  const { document, values } = fixture();
  const { coverage, css } = generateComponents(document, values);
  assert.equal(
    coverage.mapped.length + coverage.excluded.length,
    Object.keys(document).length,
  );
  assert.ok(coverage.mapped.length > 750);
  assert.match(
    css,
    /--_card-pressed-state-opacity: var\(--md-comp-elevated-card-pressed-state-layer-opacity\)/,
  );
  assert.match(
    css,
    /--_card-elevation: var\(--md-comp-elevated-card-focus-container-elevation\)/,
  );
  assert.match(
    css,
    /--_icon-pressed-shape: var\(--md-comp-icon-button-small-pressed-container-shape\)/,
  );
  assert.match(
    css,
    /--_field-focus-color: var\(--md-comp-outlined-text-field-focus-indicator-outline-color\)/,
  );
  assert.match(
    css,
    /--_field-trailing-icon-size: var\(--md-comp-outlined-select-text-field-trailing-icon-size\)/,
  );
  assert.match(
    css,
    /--_card-background-state-opacity: var\(--md-comp-elevated-card-hover-state-layer-opacity\)/,
  );
  assert.doesNotMatch(
    css,
    /--_card-background-state-opacity: var\(--md-comp-elevated-card-pressed/,
  );
  const focus = coverage.mapped.find(
    (r) => r.token === "md.comp.outlined-card.focus.indicator.color",
  );
  assert.doesNotMatch(focus.selector, /data-component-state/);
  for (const row of coverage.mapped.filter((r) =>
    r.token.includes(".narrow.leading-space"),
  )) {
    for (const selector of row.selector.split(", "))
      assert.ok(selector.includes('[data-width="narrow"]'), selector);
  }
  assert.equal(
    coverage.configurations.filter((c) => c.family === "icon").length,
    1200,
  );
  assert.equal(
    coverage.configurations.filter((c) => c.family === "card").length,
    18,
  );
});
test("missing, unknown and unconverted component tokens stop generation", () => {
  const { document, values } = fixture();
  const broken = structuredClone(document);
  delete broken.token0;
  assert.throws(() => generateComponents(broken, values), /Missing required/);
  const unknown = structuredClone(document);
  unknown.new = {
    $type: "number",
    $value: 1,
    $extensions: {
      "dev.tokencatalog.material": {
        sourceName: "md.comp.filled-card.unreviewed.color",
        presentation: { setName: "md.comp.filled-card" },
      },
    },
  };
  assert.throws(
    () => generateComponents(unknown, values),
    /Unknown component token/,
  );
  const invalidSpring = {
    ...values,
    "md-comp-icon-button-small-pressed-container-corner-size-motion-spring-stiffness":
      "0",
  };
  assert.throws(
    () => generateComponents(document, invalidSpring),
    /Invalid component spring/,
  );
  const missing = { ...values };
  delete missing["md-comp-filled-card-container-color"];
  assert.throws(
    () => generateComponents(document, missing),
    /Missing converted/,
  );
  unknown.new.$extensions["dev.tokencatalog.material"].presentation.setName =
    "md.comp.icon-button.new";
  assert.throws(
    () => generateComponents(unknown, values),
    /Unknown component set/,
  );
});
test(
  "captured components build the same bindings in all six appearances",
  { skip: !process.env.TOKEN_ASSETS_DIR },
  async () => {
    let css;
    for (const appearance of [
      "light-default",
      "light-medium",
      "light-high",
      "dark-default",
      "dark-medium",
      "dark-high",
    ]) {
      const document = JSON.parse(
        await readFile(
          `${process.env.TOKEN_ASSETS_DIR}/dtcg/${appearance}/tokens.tokens.json`,
        ),
      );
      const compiled = await compile(document, ":root", { appTheme: true });
      const generated = generateComponents(document, compiled.values);
      assert.equal(generated.coverage.mapped.length, 777);
      if (css !== undefined) assert.equal(generated.css, css);
      css = generated.css;
      assert.doesNotMatch(css, /undefined|NaN|\[object Object\]/);
    }
  },
);
