import { readFile } from "node:fs/promises";
import { compactCSS } from "./css-output.mjs";

// Ownership is declared alongside component sources. No selectors are classified.
export async function componentBundles({
  bindings,
  motionCSS,
  elevation,
  allCSS,
}) {
  const root = new URL("./components/", import.meta.url);
  const modules = JSON.parse(
    await readFile(new URL("modules.json", root), "utf8"),
  );
  const sources = JSON.parse(
    await readFile(new URL("style-sources.json", root), "utf8"),
  );
  const files = new Map(),
    owners = new Map(),
    result = new Map();
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
      if (!bindings.has(family))
        throw Error(`Missing binding family: ${family}`);
      return bindings.get(family);
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
