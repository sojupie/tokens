import test from "node:test";
import assert from "node:assert/strict";
import { themeSubset } from "./theme-subsets.mjs";

test("theme dependencies retain appearance aliases and nested fallbacks without leaking scoped inputs", () => {
  const css =
    ":root{color-scheme:light;--md-shared:red;--md-local:var(--md-shared,var(--md-fallback));--md-fallback:blue;--md-unused:0}:root[data-theme=dark]{color-scheme:dark;--md-local:var(--md-dark);--md-dark:black}";
  const result = themeSubset(css, ["color:var(--md-local)"], {
    scope: "#component-preview",
    exclude: new Set(["--md-shared"]),
  });
  assert.deepEqual([...result.names].sort(), [
    "--md-dark",
    "--md-fallback",
    "--md-local",
  ]);
  assert(result.css.includes(":root[data-theme=dark] #component-preview"));
  assert(
    !result.css.includes("--md-unused:") &&
      !result.css.includes("--md-shared:"),
  );
  assert(!result.css.includes("color-scheme:"));
  assert.throws(
    () => themeSubset(css, ["var(--md-missing)"]),
    /Missing theme dependency/,
  );
  const cycle = themeSubset(":root{--md-a:var(--md-b);--md-b:var(--md-a)}", [
    "var(--md-a)",
  ]);
  assert.equal(cycle.names.size, 2);
});
