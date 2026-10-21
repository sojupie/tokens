import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { buttonBinding, generateButtons } from "./generate-buttons.mjs";
import { compile } from "../build-tokens.mjs";

test("button schema maps known properties and rejects unknown source concepts", () => {
  assert.deepEqual(buttonBinding("selected.focused.icon.color"), {
    selection: "selected",
    state: "focused",
    property: "--_button-icon",
  });
  assert.equal(
    buttonBinding("pressed.container.corner-size.motion.spring.stiffness", true)
      .property,
    "--_button-spring-stiffness",
  );
  assert.throws(() => buttonBinding("hovered.mystery.color"), /Unmapped/);
  assert.throws(
    () => buttonBinding("dragged.container.height", true),
    /Unmapped/,
  );
});

test(
  "real button sets have complete mappings and generated configurations",
  { skip: !process.env.TOKEN_ASSETS_DIR },
  async () => {
    const file =
      process.env.TOKEN_ASSETS_DIR + "/dtcg/light-default/tokens.tokens.json";
    const document = JSON.parse(await readFile(file, "utf8"));
    const { values } = await compile(document, ":root", { appTheme: true });
    const generated = generateButtons(document, values);
    assert.equal(generated.coverage.mapped.length, 295);
    assert.match(
      generated.css,
      /--_button-pressed-state-opacity: var\(--md-comp-button-tonal-pressed-state-layer-opacity\)/,
    );
    assert.match(
      generated.css,
      /opacity: var\(--_button-pressed-state-opacity\)/,
    );
    assert.match(generated.css, /radial-gradient\(closest-side/);
    const pressRule=generated.css.split("\n").find(rule=>rule.includes('data-component-state="pressed"')&&rule.includes("--_button-background-state-color"));
    for(const chip of ["filter","assist","suggestion","input"])assert.ok(pressRule.includes(`.md-${chip}-chip`));
    assert.doesNotMatch(generated.css, /--_button-ink-opacity|md-ripple-enter/);
    assert.match(
      generated.css,
      /--_button-background-state-opacity: var\(--md-comp-button-tonal-hovered-state-layer-opacity\)/,
    );
    assert.doesNotMatch(
      generated.css,
      /--_button-background-state-opacity: var\(--md-comp-button-tonal-pressed/,
    );
    assert.equal(
      new Set(generated.coverage.mapped.map((m) => m.token)).size,
      295,
    );
    assert.equal(generated.coverage.configurations.length, 650);
    assert.doesNotMatch(
      generated.css,
      /var\(--md-comp-(filled|filled-tonal|outlined|text|elevated)-button-/,
    );
    assert.ok(
      generated.requiredNames.has(
        "md-comp-button-small-pressed-container-corner-size-motion-spring-stiffness",
      ),
    );
    assert.ok(
      generated.requiredNames.has(
        "md-comp-button-small-label-text-letter-spacing",
      ),
    );
    const missing = structuredClone(document);
    delete missing.md.comp.button.tonal.disabled["label-text"].opacity;
    assert.throws(
      () => generateButtons(missing, values),
      /Missing required button token/,
    );
    const unknown = structuredClone(document);
    unknown.md.comp.button.tonal.mystery = {
      $type: "color",
      $value: { colorSpace: "srgb", components: [0, 0, 0] },
      $extensions: {
        "dev.tokencatalog.material": {
          sourceName: "md.comp.button.tonal.mystery",
          presentation: { setName: "md.comp.button.tonal" },
        },
      },
    };
    assert.throws(
      () => generateButtons(unknown, values),
      /Unmapped button color token/,
    );
    const newSet = structuredClone(document);
    newSet.md.comp.button.newvariant = {
      mystery: {
        $extensions: {
          "dev.tokencatalog.material": {
            sourceName: "md.comp.button.newvariant.mystery",
            presentation: { setName: "md.comp.button.newvariant" },
          },
        },
      },
    };
    assert.throws(() => generateButtons(newSet, values), /Unmapped button set/);
  },
);
