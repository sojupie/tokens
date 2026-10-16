import StyleDictionary from "style-dictionary";
import { generateComponents } from "./material/generate-components.mjs";
import { generateButtons } from "./material/generate-buttons.mjs";
import {
  realpath,
  readFile,
  writeFile,
  mkdtemp,
  rename,
  rm,
} from "node:fs/promises";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const extension = "dev.tokencatalog.material";

// Duration objects are still unsupported by time/seconds in this pinned release.
StyleDictionary.registerTransform({
  name: "catalog/duration",
  type: "value",
  filter: (token) => token.$type === "duration",
  transform: (token) => {
    const { value, unit } = token.$value;
    if (!Number.isFinite(value) || !["ms", "s"].includes(unit))
      throw new Error("Invalid duration");
    return `${value}${unit}`;
  },
});

// Flatten $root ourselves until Style Dictionary handles the complete 2025.10
// group model. This changes only its build input; the exported files stay DTCG.
export function flatten(document) {
  const tokens = {};
  const shapes = {};
  function walk(group, parts = [], inheritedType) {
    const type = group.$type ?? inheritedType;
    if ("$value" in group) {
      const name = parts.join("-");
      if (name in tokens) throw new Error(`CSS name collision: ${name}`);
      const node = structuredClone(group);
      node.$type = type;
      if (typeof node.$value === "string" && /^\{[^{}]+\}$/.test(node.$value)) {
        node.$value = `{${node.$value.slice(1, -1).split(".").join("-")}}`;
      }
      tokens[name] = node;
    }
    const source = group.$extensions?.[extension];
    if (source?.sourceType === "shape")
      shapes[parts.join("-")] = source.sourceValue.shape.family;
    for (const [key, child] of Object.entries(group)) {
      if (key === "$root") walk(child, parts, type);
      else if (!key.startsWith("$")) walk(child, [...parts, key], type);
    }
  }
  walk(document);
  return { tokens, shapes };
}

export async function compile(
  document,
  selector = ":root",
  { appTheme = false, names } = {},
) {
  const { tokens, shapes } = flatten(document);
  const dictionary = new StyleDictionary({
    tokens,
    usesDtcg: true,
    log: { verbosity: "silent", errors: { brokenReferences: "throw" } },
    platforms: {
      css: {
        transforms: [
          "color/css",
          "size/px",
          "fontFamily/css",
          "cubicBezier/css",
          "typography/css/shorthand",
          "catalog/duration",
        ],
      },
    },
  });
  const built = await dictionary.getPlatformTokens("css");
  const values = Object.fromEntries(
    built.allTokens.map((t) => [t.path.join("-"), t.$value]),
  );
  for (const t of built.allTokens) {
    if (t.$type === "typography") {
      const name = t.path.join("-");
      const spacing = tokens[name].$value.letterSpacing;
      if (
        !Number.isFinite(spacing?.value) ||
        !["px", "rem"].includes(spacing.unit)
      )
        throw new Error(`Invalid typography tracking: ${name}`);
      values[`${name}-letter-spacing`] = `${spacing.value}${spacing.unit}`;
    }
  }
  if (appTheme) {
    for (const t of built.allTokens) {
      const source = t.$extensions?.[extension];
      if (source?.sourceType === "elevation") {
        const dp = source.sourceValue.elevation.value ?? 0;
        const level = [0, 1, 3, 6, 8, 12].indexOf(dp);
        if (level < 0) throw new Error(`Unsupported Material elevation: ${dp}`);
        values[t.path.join("-")] = level;
      }
      if (source?.sourceType === "fontNames")
        values[t.path.join("-")] += ", sans-serif";
    }
  }
  for (const [name, family] of Object.entries(shapes)) {
    if (family === "SHAPE_FAMILY_CIRCULAR")
      values[name] = "calc(infinity * 1px)";
    else {
      const corners = ["topLeft", "topRight", "bottomRight", "bottomLeft"].map(
        (c) => values[`${name}-${c}`],
      );
      if (corners.some((c) => typeof c !== "string"))
        throw new Error(`Missing corners for ${name}`);
      values[name] = corners.every((c) => c === corners[0])
        ? corners[0]
        : corners.join(" ");
    }
  }
  for (const [name, value] of Object.entries(values)) {
    if (
      typeof value === "object" ||
      value == null ||
      /[{};\n\r]/.test(String(value))
    )
      throw new Error(`Invalid CSS value for ${name}`);
  }
  return { css: cssVariables(values, selector, names), values };
}

function cssVariables(values, selector, names) {
  const declarations = Object.entries(values)
    .filter(([name]) => !names || names.has(name))
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, value]) => `  --${name}: ${value};`);
  return `${selector} {\n${declarations.join("\n")}\n}\n`;
}

async function main() {
  const args = process.argv.slice(2);
  if (args.length !== 2 || args[0] !== "--published")
    throw new Error(
      "Usage: npm run build:tokens -- --published <publication-directory>",
    );
  const published = await realpath(args[1]);
  const database = await readFile(path.join(published, "tokens.sqlite"));
  const databaseSha256 = createHash("sha256").update(database).digest("hex");
  const output = await mkdtemp(path.join(published, ".token-assets-"));
  try {
    const result = spawnSync(
      "go",
      [
        "run",
        "./cmd/tokens",
        "export",
        "--published",
        published,
        "--output",
        path.join(output, "dtcg"),
        "--themes",
      ],
      { stdio: "inherit" },
    );
    if (result.status !== 0) throw new Error("DTCG export failed");
    const index = JSON.parse(
      await readFile(path.join(output, "dtcg/index.json"), "utf8"),
    );
    const required = new Set([
      "md-ref-palette-neutral0",
      "md-sys-color-on-primary",
    ]);
    for (const file of ["app.css", "token-table.css"]) {
      const css = await readFile(path.join("catalog/web/static", file), "utf8");
      for (const match of css.matchAll(/var\(\s*--(md-[a-z0-9-]+)\s*[,)]/g))
        required.add(match[1]);
    }
    let themeCSS = "";
    let identity;
    let buttonCSS;
    let buttonCoverage;
    let componentCSS;
    let componentCoverage;
    const anatomy = await readFile(
      "forge/components/static/components-base.css",
      "utf8",
    );
    const elevation = await readFile(
      "forge/components/static/elevation.css",
      "utf8",
    );
    for (const match of (anatomy + elevation).matchAll(
      /var\(\s*--(md-[a-z0-9-]+)\s*[,)]/g,
    ))
      required.add(match[1]);
    const totals = [];
    for (const variant of index) {
      const document = JSON.parse(
        await readFile(path.join(output, "dtcg", variant.file), "utf8"),
      );
      identity = document.$extensions["dev.tokencatalog.export"];
      const selector = `:root[data-theme="${variant.theme}"][data-contrast="${variant.contrast}"]`;
      const compiled = await compile(document, selector, {
        appTheme: true,
        names: required,
      });
      const report = JSON.parse(
        await readFile(
          path.join(output, "dtcg", variant.name, "report.json"),
          "utf8",
        ),
      );
      const missingButtons = report.omitted.filter((token) =>
        token.name.startsWith("md.comp.button."),
      );
      if (missingButtons.length)
        throw new Error(
          `Button tokens omitted from DTCG: ${missingButtons.map((t) => t.name).join(", ")}`,
        );
      const buttons = generateButtons(document, compiled.values);
      if (buttonCSS !== undefined && buttonCSS !== buttons.css)
        throw new Error("Button structure differs across appearance contexts");
      buttonCSS = buttons.css;
      buttonCoverage = buttons.coverage;
      for (const name of buttons.requiredNames) required.add(name);
      const components = generateComponents(document, compiled.values);
      const generated = anatomy + "\n" + elevation + "\n" + components.css;
      if (componentCSS !== undefined && componentCSS !== generated)
        throw new Error("Component bindings differ across appearances");
      componentCSS = generated;
      componentCoverage = components.coverage;
      for (const name of components.requiredNames) required.add(name);
      compiled.css = cssVariables(compiled.values, selector, required);
      for (const name of required)
        if (!(name in compiled.values))
          throw new Error(`${variant.name} is missing --${name}`);
      const downloadable = await compile(document);
      await writeFile(
        path.join(output, "dtcg", variant.name, "tokens.css"),
        downloadable.css,
      );
      const defaultLight = variant.name === "light-default";
      const vars = defaultLight
        ? `  --tokens-preview-on-light: ${compiled.values["md-ref-palette-neutral0"]};\n  --tokens-preview-on-dark: ${compiled.values["md-sys-color-on-primary"]};\n`
        : "";
      if (defaultLight && vars.includes("undefined"))
        throw new Error("Missing preview text colors");
      themeCSS +=
        (defaultLight ? ":root, " : "") +
        compiled.css.replace(
          "{\n",
          `{\n  color-scheme: ${variant.theme};\n${vars}`,
        );
      totals.push({ context: variant.name, cssVariables: required.size });
    }
    if (
      createHash("sha256")
        .update(await readFile(path.join(published, "tokens.sqlite")))
        .digest("hex") !== databaseSha256
    )
      throw new Error("Publication changed during the build");
    await writeFile(path.join(output, "theme.css"), themeCSS);
    await writeFile(path.join(output, "buttons.css"), buttonCSS);
    await writeFile(path.join(output, "components.css"), componentCSS);
    await writeFile(
      path.join(output, "components-coverage.json"),
      JSON.stringify(componentCoverage, null, 2) + "\n",
    );
    await writeFile(
      path.join(output, "buttons-coverage.json"),
      JSON.stringify(buttonCoverage, null, 2) + "\n",
    );
    await writeFile(
      path.join(output, "manifest.json"),
      JSON.stringify(
        {
          formatVersion: "2025.10",
          styleDictionary: "5.6.1",
          publicationId: identity.publicationId,
          snapshotId: identity.snapshotId,
          databaseSha256,
          componentsSha256: createHash("sha256")
            .update(componentCSS)
            .digest("hex"),
          buttonsSha256: createHash("sha256").update(buttonCSS).digest("hex"),
          cssSha256: createHash("sha256").update(themeCSS).digest("hex"),
          variants: totals,
        },
        null,
        2,
      ) + "\n",
    );
    const target = path.join(published, "assets");
    const backup = `${output}-previous`;
    let previous = false;
    try {
      await rename(target, backup);
      previous = true;
    } catch (e) {
      if (e.code !== "ENOENT") throw e;
    }
    try {
      await rename(output, target);
    } catch (e) {
      if (previous) await rename(backup, target);
      throw e;
    }
    if (previous) await rm(backup, { recursive: true });
    await writeFile("forge/components/static/buttons.css", buttonCSS);
    await writeFile("forge/components/static/components.css", componentCSS);
    console.log(`Built six themes and DTCG/CSS downloads in ${target}`);
  } finally {
    await rm(output, { recursive: true, force: true });
  }
}

if (
  process.argv[1] &&
  path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
  });
}
