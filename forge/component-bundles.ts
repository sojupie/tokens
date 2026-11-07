import { readFile } from "node:fs/promises";
import { compactCSS } from "./css-output.ts";

// Ownership is declared alongside component sources. No selectors are classified.
export async function componentBundles({
  bindings,
  motionCSS,
  elevation,
  allCSS,
}: { bindings: Map<string,string>; motionCSS: string; elevation: string; allCSS: string }) {
  const root = new URL("./components/", import.meta.url);
  const modules: Record<string, {styles:string[]; bindings:string[]}> = JSON.parse(
    await readFile(new URL("modules.json", root), "utf8"),
  );
  const sources: string[] = JSON.parse(
    await readFile(new URL("style-sources.json", root), "utf8"),
  );
  const files = new Map<string,string>(),
    owners = new Map<string,string>(),
    result = new Map<string,string>();
  for (const [id, module] of Object.entries(modules)) {
    for (const file of module.styles) {
      if (files.has(file)) throw Error(`Duplicate source owner: ${file}`);
      files.set(file, id);
    }
    for (const family of module.bindings) {
      if (owners.has(family)) throw Error(`Duplicate binding owner: ${family}`);
      owners.set(family, id);
    }
  }
  for (const file of sources)
    if (!files.has(file)) throw Error(`Unowned component source: ${file}`);
  for (const file of files.keys())
    if (!sources.includes(file))
      throw Error(`Unknown component source: ${file}`);
  for (const family of bindings.keys())
    if (!owners.has(family))
      throw Error(`Unowned generated binding: ${family}`);
  for (const [id, module] of Object.entries(modules)) {
    const anatomy = await Promise.all(
      sources
        .filter((file) => files.get(file) === id)
        .map((file) => readFile(new URL(file, root), "utf8")),
    );
    const generated = module.bindings.map((family) => {
      const css = bindings.get(family);
      if (typeof css !== "string") throw Error(`Missing binding family: ${family}`);
      return css;
    });
    result.set(
      id,
      compactCSS(
        (id === "common" ? motionCSS + elevation : "") +
          anatomy.join("\n") +
          "\n" +
          generated.join("\n"),
      ),
    );
  }
  // Retain the complete public artifact for existing consumers and comparisons.
  result.set("all", allCSS);
  return result;
}
