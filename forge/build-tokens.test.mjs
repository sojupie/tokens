import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { compile, flatten } from "./build-tokens.mjs";

const meta = (kind, value) => ({
  "dev.tokencatalog.material": {
    sourceType: kind,
    sourceValue: { [kind]: value },
  },
});

test("v5 resolves aliases, object dimensions/durations, colors, typography and $root", async () => {
  const document = {
    size: {
      $root: { $type: "dimension", $value: { value: 16, unit: "px" } },
      small: { $type: "dimension", $value: { value: 4, unit: "px" } },
    },
    alias: { $type: "dimension", $value: "{size}" },
    duration: { $type: "duration", $value: { value: 250, unit: "ms" } },
    color: {
      $type: "color",
      $value: { colorSpace: "srgb", components: [1, 0, 0], alpha: 0.5 },
    },
    typography: {
      $type: "typography",
      $value: {
        fontFamily: ["Roboto"],
        fontSize: { value: 16, unit: "px" },
        fontWeight: 400,
        letterSpacing: { value: 0, unit: "px" },
        lineHeight: 1.5,
      },
    },
  };
  const { values } = await compile(document);
  assert.equal(values.size, "16px");
  assert.equal(values["size-small"], "4px");
  assert.equal(values.alias, "16px");
  assert.equal(values.duration, "250ms");
  assert.match(values.color, /rgba\(255, 0, 0, 0\.5\)/);
  assert.match(values.typography, /16px\/1.5/);
});

test("shape corners, circular shapes and elevation have explicit CSS policies", async () => {
  const corner = (value) => ({
    $type: "dimension",
    $value: { value, unit: "px" },
  });
  const shape = {
    topLeft: corner(0),
    topRight: corner(16),
    bottomRight: corner(16),
    bottomLeft: corner(0),
    $extensions: meta("shape", { family: "SHAPE_FAMILY_ROUNDED_CORNERS" }),
  };
  const document = {
    shape,
    full: { $extensions: meta("shape", { family: "SHAPE_FAMILY_CIRCULAR" }) },
    elevation: {
      $type: "dimension",
      $value: { value: 3, unit: "px" },
      $extensions: meta("elevation", { value: 3, unit: "DIPS" }),
    },
  };
  const standard = await compile(document);
  assert.equal(standard.values.shape, "0px 16px 16px 0px");
  assert.equal(standard.values.full, "calc(infinity * 1px)");
  assert.equal(standard.values.elevation, "3px");
  const app = await compile(document, ":root", { appTheme: true });
  assert.equal(app.values.elevation, 2);
});

test("missing aliases and CSS name collisions fail the build", async () => {
  await assert.rejects(
    compile({ alias: { $type: "dimension", $value: "{missing}" } }),
  );
  assert.throws(
    () =>
      flatten({
        "a-b": { $type: "number", $value: 1 },
        a: { b: { $type: "number", $value: 2 } },
      }),
    /collision/,
  );
});

test(
  "six real themes generate CSS for every exported leaf",
  { skip: !process.env.TOKEN_ASSETS_DIR },
  async () => {
    const root = process.env.TOKEN_ASSETS_DIR;
    const index = JSON.parse(await readFile(`${root}/dtcg/index.json`, "utf8"));
    assert.equal(index.length, 6);
    const primary = new Set();
    for (const variant of index) {
      const document = JSON.parse(
        await readFile(`${root}/dtcg/${variant.file}`, "utf8"),
      );
      const { values, css } = await compile(document);
      const { tokens } = flatten(document);
      let checked = 0;
      for (const [name, token] of Object.entries(tokens)) {
        const source = token.$extensions?.["dev.tokencatalog.material"];
        if (!source) continue;
        const raw = source.sourceValue;
        for (const field of [
          "length",
          "fontSize",
          "lineHeight",
          "fontTracking",
          "elevation",
        ]) {
          if (raw[field]) {
            assert.equal(
              values[name],
              `${raw[field].value ?? 0}px`,
              `${variant.name}: ${name}`,
            );
            checked++;
          }
        }
        for (const field of ["fontWeight", "opacity", "numeric"]) {
          if (field in raw) {
            assert.equal(values[name], raw[field], `${variant.name}: ${name}`);
            checked++;
          }
        }
        if ("durationMs" in raw) {
          assert.equal(values[name], `${raw.durationMs}ms`);
          checked++;
        }
        if (raw.type) {
          const t = raw.type;
          const expected = `${t.fontSize.value ?? 0}px/${(t.lineHeight.value ?? 0) / t.fontSize.value}`;
          assert.ok(
            values[name].includes(expected),
            `${variant.name}: ${name}`,
          );
          checked++;
        }
        if (raw.color) {
          const c = raw.color;
          const expected = [c.red ?? 0, c.green ?? 0, c.blue ?? 0].map((n) =>
            Math.round(n * 255),
          );
          const cssColor = String(values[name]);
          if (cssColor.startsWith("#")) {
            assert.equal(
              cssColor.slice(1, 7).toLowerCase(),
              expected.map((n) => n.toString(16).padStart(2, "0")).join(""),
              `${variant.name}: ${name}`,
            );
          } else {
            const channels = cssColor.match(/[0-9.]+/g).map(Number);
            assert.deepEqual(
              channels.slice(0, 3).map(Math.round),
              expected,
              `${variant.name}: ${name}`,
            );
            assert.equal(channels[3] ?? 1, c.alpha);
          }
          checked++;
        }
      }
      assert.ok(checked > 5000);
      assert.ok(Object.keys(values).length > 6000);
      assert.equal(
        values["md-sys-shape-corner-large-end"],
        "0px 16px 16px 0px",
      );
      assert.equal(values["md-sys-motion-duration-long4"], "600ms");
      primary.add(values["md-sys-color-primary"]);
      assert.doesNotMatch(css, /\[object Object\]|undefined|NaN/);
    }
    assert.equal(primary.size, 6);
  },
);
