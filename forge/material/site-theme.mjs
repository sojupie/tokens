import { Blend, Hct, argbFromHex, hexFromArgb, SchemeTonalSpot, MaterialDynamicColors } from "@material/material-color-utilities";

// Customize only the app's build input. Public exports retain captured values.
export function siteTheme(document, appearance, { primary, secondary, extensions = {} }) {
  for (const seed of [primary, secondary, ...Object.values(extensions)])
    if (!/^#[0-9a-f]{6}$/i.test(seed)) throw new Error("Site palette seeds must be #RRGGBB");
  const contrast = { default: 0, medium: 0.5, high: 1 }[appearance.contrast];
  if (contrast === undefined) throw new Error("Unknown site theme contrast");
  const scheme = new SchemeTonalSpot(Hct.fromInt(argbFromHex(primary)), appearance.theme === "dark", contrast);
  // Keep the blue secondary independent of the purple primary and its neutrals.
  scheme.secondaryPalette = new SchemeTonalSpot(Hct.fromInt(argbFromHex(secondary)), scheme.isDark, contrast).secondaryPalette;
  const themed = structuredClone(document);
  const colorValue = argb => {
    const hex = hexFromArgb(argb);
    return { colorSpace: "srgb", components: [1, 3, 5].map(i => parseInt(hex.slice(i, i + 2), 16) / 255), alpha: 1 };
  };
  for (const [name, token] of Object.entries(themed.md.sys.color)) {
    const role = name.replace(/-([a-z])/g, (_, c) => c.toUpperCase());
    if (!MaterialDynamicColors[role]?.getArgb) throw new Error(`Unmapped site color: ${name}`);
    (token.$root ?? token).$value = colorValue(MaterialDynamicColors[role].getArgb(scheme));
  }
  for (const [name, token] of Object.entries(themed.md.ref.palette)) {
    const [, family, tone] = name.match(/^(.+?)(\d+)$/) ?? [];
    const palette = scheme[`${family?.replace(/-([a-z])/g, (_, c) => c.toUpperCase())}Palette`];
    if (!palette) throw new Error(`Unmapped site palette: ${name}`);
    (token.$root ?? token).$value = colorValue(palette.tone(Number(tone)));
  }
  // Decorative extensions belong to the catalog theme, never the source export.
  themed.catalog = { color: {} };
  for (const [name, seed] of Object.entries(extensions)) {
    if (!/^[a-z]+$/.test(name)) throw new Error("Invalid site color extension name");
    themed.catalog.color[name] = {
      $type: "color",
      $value: colorValue(Blend.harmonize(argbFromHex(seed), argbFromHex(primary))),
    };
  }
  return themed;
}
