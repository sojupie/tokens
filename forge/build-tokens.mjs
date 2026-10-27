import motion from "./components/static/motion.cjs";
import { parseArgs } from "node:util";
import StyleDictionary from "style-dictionary";
import { siteTheme } from "./material/site-theme.mjs";
import { generateComponents } from "./material/generate-components.mjs";
import { generateButtons } from "./material/generate-buttons.mjs";
import { buildComponentAssets } from "./components/build-assets.mjs";
import { compactCSS, themeVariantsCSS } from './css-output.mjs';
import { componentBundles } from './component-bundles.mjs';
import { pageThemes } from './theme-subsets.mjs';
import {
  readdir,
  realpath,
  mkdir,
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
  const percentages = {};
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
    if (source?.sourceType === "length" && source.sourceValue.length.unit === "PERCENT")
      percentages[parts.join("-")] = source.sourceValue.length.value;
    if (source?.sourceType === "shape")
      shapes[parts.join("-")] = source.sourceValue.shape.family;
    for (const [key, child] of Object.entries(group)) {
      if (key === "$root") walk(child, parts, type);
      else if (!key.startsWith("$")) walk(child, [...parts, key], type);
    }
  }
  walk(document);
  return { tokens, shapes, percentages };
}

export function applyOverrides(document, overrides) {
  if (!Object.keys(overrides).length) return document;
  const result = structuredClone(document);
  for (const [name, value] of Object.entries(overrides)) {
    const group = name.split(".").reduce((node, part) => node?.[part], result);
    const token = group?.$root ?? group;
    if (!token || typeof token !== "object")
      throw new Error(`Unknown override token: ${name}`);
    if (group.$extensions?.[extension]?.sourceType === "shape") {
      if (
        !(typeof value === "string" && /^\{[^{}]+\}$/.test(value)) &&
        !(
          value &&
          Number.isFinite(value.value) &&
          ["px", "rem"].includes(value.unit)
        )
      )
        throw new Error(
          `Shape override needs a dimension or reference: ${name}`,
        );
      for (const corner of ["topLeft", "topRight", "bottomRight", "bottomLeft"])
        group[corner].$value =
          typeof value === "string"
            ? `{${value.slice(1, -1)}.${corner}}`
            : structuredClone(value);
      group.$extensions[extension].sourceValue.shape.family =
        "SHAPE_FAMILY_ROUNDED_CORNERS";
    } else {
      if (!("$value" in token))
        throw new Error(`Override is not a token: ${name}`);
      token.$value = structuredClone(value);
    }
  }
  return result;
}

export async function compile(
  document,
  selector = ":root",
  { appTheme = false, names, overrides = {} } = {},
) {
  document = applyOverrides(document, overrides);
  const { tokens, shapes, percentages } = flatten(document);
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
        const dp = Number.parseFloat(t.$value);
        const level = [0, 1, 3, 6, 8, 12].indexOf(dp);
        if (level < 0) throw new Error(`Unsupported Material elevation: ${dp}`);
        values[t.path.join("-")] = level;
      }
      if (source?.sourceType === "fontNames")
        values[t.path.join("-")] += ", sans-serif";
    }
  }
  for (const [name, value] of Object.entries(percentages)) {
    if (!Number.isFinite(value)) throw new Error(`Invalid percentage: ${name}`);
    values[name] = `${value}%`;
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
  const aliases = {};
  for (const [name, token] of Object.entries(tokens)) {
    const reference = typeof token.$value === "string" && /^\{([^{}]+)\}$/.exec(token.$value);
    if (!reference) continue;
    const target = reference[1].replaceAll(".", "-");
    // Preserve a captured alias only when app-specific conversions agree.
    if (target in values && String(values[name]) === String(values[target]))
      aliases[name] = target;
  }
  return { css: cssVariables(values, selector, names), values, aliases };
}

// Reference palette tones are shared across appearances. Context exports only
// include tones used by their aliases, so retain captured base tones for site art.
export function inheritReferencePalette(values, baseValues) {
  for (const [name, value] of Object.entries(baseValues)) {
    if (name.startsWith("md-ref-palette-") && !(name in values)) values[name] = value;
  }
}

function cssVariables(values, selector, names) {
  const declarations = Object.entries(values)
    .filter(([name]) => !names || names.has(name))
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, value]) => `  --${name}: ${value};`);
  return `${selector} {\n${declarations.join("\n")}\n}\n`;
}

async function main() {
  const { values: args } = parseArgs({
    options: {
      published: { type: "string" },
      assets: { type: "string" },
      "tokens-binary": { type: "string" },
      "check-bindings": { type: "boolean", default: false },
    },
  });
  if (!args.published)
    throw new Error("Usage: npm run build:tokens -- --published <directory> [--assets <directory>] [--check-bindings]");
  await buildComponentAssets({ check: args["check-bindings"] });
  const executable = args["tokens-binary"] ?? "go";
  const prefix = args["tokens-binary"] ? [] : ["run", "./cmd/tokens"];
  const published = await realpath(args.published);
  const target = path.resolve(args.assets ?? path.join(published, "assets"));
  await mkdir(path.dirname(target), { recursive: true });
  const overrides = JSON.parse(
    await readFile("forge/material/overrides.json", "utf8"),
  );
  if (!overrides || Array.isArray(overrides) || typeof overrides !== "object")
    throw new Error("Token overrides must be an object keyed by token name");
  const overrideValues = Object.fromEntries(
    Object.keys(overrides).map((token) => [
      token,
      { token, value: overrides[token], upstream: {}, effective: {} },
    ]),
  );
  const database = await readFile(path.join(published, "tokens.sqlite"));
  const databaseSha256 = createHash("sha256").update(database).digest("hex");
  const output = await mkdtemp(path.join(path.dirname(target), ".token-assets-"));
  try {
    const result = spawnSync(
      executable,
      [
        ...prefix,
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
    for (const file of ["catalog/web/static/app.css", "catalog/web/static/token-table.css", "forge/site/site.css"]) {
      const css = await readFile(file, "utf8");
      for (const match of css.matchAll(/var\(\s*--((?:md|catalog)-[a-z0-9-]+)\s*[,)]/g))
        required.add(match[1]);
    }
    const themeConfig = JSON.parse(await readFile("forge/material/site-theme.json", "utf8"));
    if (!["Standard", "Expressive"].includes(themeConfig.motionScheme)) throw new Error("Unknown site motion scheme");
    const motionOutput = path.join(output, "site-motion");
    const motionExport = spawnSync(executable, [...prefix, "export", "--published", published, "--context", `Motion scheme=${themeConfig.motionScheme}`, "--output", motionOutput], { stdio: "inherit" });
    if (motionExport.status !== 0) throw new Error("Site motion export failed");
    const motionDocument = JSON.parse(await readFile(path.join(motionOutput, "tokens.tokens.json"), "utf8"));
    await rm(motionOutput, { recursive: true });
    let themeCSS = "";
    const themeVariants = [];
    let referencePalette = {};
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
      /var\(\s*--((?:md|catalog)-[a-z0-9-]+)\s*[,)]/g,
    ))
      required.add(match[1]);
    const totals = [];
    const bindingValues = new Map();
    let componentModules, moduleMotionCSS;
    for (const variant of index) {
      const document = JSON.parse(
        await readFile(path.join(output, "dtcg", variant.file), "utf8"),
      );
      identity = document.$extensions["dev.tokencatalog.export"];
      const selector = `:root[data-theme="${variant.theme}"][data-contrast="${variant.contrast}"]`;
      const compiled = await compile(siteTheme(document, variant, themeConfig, motionDocument), selector, {
        appTheme: true,
        names: required,
        overrides,
      });
      if (variant.name === "light-default") referencePalette = { ...compiled.values };
      inheritReferencePalette(compiled.values, referencePalette);
      if (Object.keys(overrides).length) {
        const original = await compile(document, selector, { appTheme: true });
        for (const [token, entry] of Object.entries(overrideValues)) {
          const name = token.replaceAll(".", "-");
          entry.upstream[variant.name] = String(original.values[name]);
          entry.effective[variant.name] = String(compiled.values[name]);
        }
      }
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
      const motionProperties = [];
      for (const speed of ["fast", "default", "slow"]) for (const type of ["spatial", "effects"]) {
        const name = `md-sys-motion-spring-${speed}-${type}`;
        for (const property of ["damping", "stiffness"]) required.add(`${name}-${property}`);
        const timing = motion.cssTiming(Number(compiled.values[`${name}-damping`]), Number(compiled.values[`${name}-stiffness`]));
        for (const property of ["duration", "easing"]) motionProperties.push(`  --_motion-${speed}-${type}-${property}: ${timing[property]};`);
      }
      const motionCSS = `:root {\n${motionProperties.join("\n")}\n}\n`;
      const generated = compactCSS(motionCSS + anatomy + "\n" + elevation + "\n" + components.css);
      if (componentCSS !== undefined && componentCSS !== generated)
        throw new Error("Component bindings differ across appearances");
      componentCSS = generated;
      componentCoverage = components.coverage;
      componentModules = components.modules;
      moduleMotionCSS = motionCSS;
      for (const binding of [
        ...buttons.coverage.mapped,
        ...components.coverage.mapped,
      ]) {
        if (!bindingValues.has(binding.token))
          bindingValues.set(binding.token, {});
        bindingValues.get(binding.token)[variant.name] = String(
          compiled.values[binding.token.replaceAll(".", "-")],
        );
      }
      for (const name of components.requiredNames) required.add(name);
      compiled.css = cssVariables(compiled.values, selector, required);
      for (const name of required)
        if (!(name in compiled.values))
          throw new Error(`${variant.name} is missing --${name}`);
      const downloadable = await compile(document, ":root", { overrides });
      await writeFile(
        path.join(output, "dtcg", variant.name, "tokens.css"),
        downloadable.css,
      );
      themeVariants.push({ selector, theme: variant.theme, values: compiled.values, aliases: compiled.aliases });
      totals.push({ context: variant.name, cssVariables: required.size });
    }
    if (
      createHash("sha256")
        .update(await readFile(path.join(published, "tokens.sqlite")))
        .digest("hex") !== databaseSha256
    )
      throw new Error("Publication changed during the build");
    const bundles = await componentBundles({
      bindings: componentModules, motionCSS: moduleMotionCSS, elevation, allCSS: componentCSS,
    });
    const sharedSources = await Promise.all(
      [
        "forge/site/site.css",
        "forge/site/site.templ",
        "catalog/web/static/app.css",
        "catalog/web/static/token-table.css",
        ...(await readdir("catalog/web"))
          .filter((file) => file.endsWith(".templ"))
          .map((file) => "catalog/web/" + file),
      ].map((file) => readFile(file, "utf8")),
    );
    sharedSources.push(buttonCSS);
    const runtimeFiles = (
      await Promise.all(
        ["forge/components/static", "catalog/web/static"].map(
          async (directory) =>
            (await readdir(directory))
              .filter((file) => file.endsWith(".js"))
              .map((file) => path.join(directory, file)),
        ),
      )
    ).flat();
    const runtimeSources = await Promise.all(
      runtimeFiles.map((file) => readFile(file, "utf8")),
    );
    for (const source of sharedSources)
      for (const match of source.matchAll(/var\(\s*--((?:md|catalog)-[\w-]+)/g))
        required.add(match[1]);
    for (const source of runtimeSources)
      for (const match of source.matchAll(
        /["'`]--((?:md|catalog)-[\w-]+)["'`]/g,
      ))
        required.add(match[1]);
    for (const variant of themeVariants) {
      variant.values = { ...variant.values };
      for (const [name, target] of Object.entries(variant.aliases))
        variant.values[name] = `var(--${target})`;
    }
    // Set iteration also visits newly discovered targets, following alias chains.
    for (const name of required)
      for (const variant of themeVariants)
        for (const match of String(variant.values[name]).matchAll(/var\(--([\w-]+)/g))
          required.add(match[1]);
    for (const entry of totals) entry.cssVariables = required.size;
    themeCSS = themeVariantsCSS(themeVariants, required);
    const themes = pageThemes(themeCSS, bundles, sharedSources, runtimeSources),
      themeBundles = {};
    await mkdir(path.join(output, "themes"));
    for (const [name, { css }] of themes) {
      await writeFile(path.join(output, "themes", name + ".css"), css);
      themeBundles[name] = createHash("sha256").update(css).digest("hex");
    }
    await writeFile(path.join(output, "theme.css"), themeCSS);
    await writeFile(path.join(output, "buttons.css"), buttonCSS);
    await writeFile(path.join(output, "components.css"), componentCSS);
    for (const binding of [
      ...buttonCoverage.mapped,
      ...componentCoverage.mapped,
    ])
      binding.values = bindingValues.get(binding.token);
    await writeFile(
      path.join(output, "components-coverage.json"),
      JSON.stringify(componentCoverage, null, 2) + "\n",
    );
    await writeFile(
      path.join(output, "buttons-coverage.json"),
      JSON.stringify(buttonCoverage, null, 2) + "\n",
    );
    await writeFile(
      path.join(output, "overrides.json"),
      JSON.stringify(Object.values(overrideValues), null, 2) + "\n",
    );
    const audit = spawnSync(
      executable,
      [
        ...prefix,
        "forge",
        "inspect",
        "--assets",
        output,
        "--output",
        path.join(output, "css-usage.json"),
      ],
      { stdio: "inherit" },
    );
    if (audit.status !== 0) throw new Error("CSS usage report failed");
    await writeFile(
      path.join(output, "manifest.json"),
      JSON.stringify(
        {
          motionScheme: themeConfig.motionScheme,
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
          themeBundles,
        },
        null,
        2,
      ) + "\n",
    );
    if (args["check-bindings"]) {
      for (const [file, css] of [["buttons.css", buttonCSS], ["components.css", componentCSS]]) {
        if (await readFile(path.join("forge/components/static", file), "utf8") !== css)
          throw new Error(`${file} differs from the committed server bindings; regenerate and rebuild the application`);
      }
    }
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
    if (!args["check-bindings"]) {
      await writeFile("forge/components/static/buttons.css", buttonCSS);
      await writeFile("forge/components/static/components.css", componentCSS);
    }
    const moduleManifest = JSON.stringify([...bundles.keys()].filter(name => name !== "all")) + "\n";
    const manifestFile = "forge/components/static/bundles/modules.json";
    if (args["check-bindings"]) {
      if (await readFile(manifestFile, "utf8") !== moduleManifest)
        throw new Error("Stale component module manifest");
    } else await writeFile(manifestFile, moduleManifest);
    for (const file of await readdir("forge/components/static/bundles")) {
      if (!file.endsWith(".css") || bundles.has(file.slice(0, -4))) continue;
      if (args["check-bindings"]) throw new Error(`Obsolete component bundle: ${file}`);
      await rm(path.join("forge/components/static/bundles", file));
    }
    for (const [name, css] of bundles) {
      const file = `forge/components/static/bundles/${name}.css`;
      if (args['check-bindings']) {
        if (await readFile(file, 'utf8') !== css) throw new Error(`Stale component bundle: ${name}`);
      } else {
        await mkdir('forge/components/static/bundles', { recursive: true });
        await writeFile(file, css);
      }
    }
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
