import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { Hct, argbFromHex } from "@material/material-color-utilities";
import { siteTheme } from "./site-theme.mjs";
import { compile } from "../build-tokens.mjs";

test("site palette follows appearance and aliases without modifying captured tokens", async () => {
  const token = value => ({ $type: "color", $value: value });
  const source = { md: { sys: { color: { primary: token("{md.ref.palette.secondary90}") } }, ref: { palette: { secondary90: token({ colorSpace: "srgb", components: [1, 0, 0], alpha: 1 }) } }, comp: { example: token("{md.sys.color.primary}") } } };
  const before = structuredClone(source);
  const config = JSON.parse(await readFile("forge/material/site-theme.json", "utf8"));
  const values = [];
  for (const theme of ["light", "dark"]) for (const contrast of ["default", "medium", "high"]) {
    const { values: css } = await compile(siteTheme(source, { theme, contrast }, config));
    for (const name of Object.keys(config.extensions)) {
      assert.match(css[`catalog-color-${name}`], /^#[0-9a-f]{6}$/i, "extensions must reach the compiled CSS");
    }
    assert.equal(css["md-ref-palette-secondary90"], "#e2e0f9");
    assert.equal(css["md-comp-example"], css["md-sys-color-primary"]);
    const primaryHue = Hct.fromInt(argbFromHex(css["md-sys-color-primary"])).hue;
    assert.ok(Math.abs(primaryHue - Hct.fromInt(argbFromHex(config.primary)).hue) < 5, "primary stays purple at every contrast");
    values.push(css["md-sys-color-primary"]);
  }
  assert.equal(new Set(values).size, 6);
  assert.deepEqual(source, before);
});
