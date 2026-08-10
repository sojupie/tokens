#!/usr/bin/env node

import { readFile, writeFile, mkdir } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import process from "node:process";
import { pathToFileURL } from "node:url";

function parseArgs(argv) {
  const [input, outputFlag, output] = argv[0] === "--" ? argv.slice(1) : argv;
  if (!input || outputFlag !== "--output" || !output) {
    throw new Error("Usage: pnpm normalize:m3 -- <input.json> --output <output.json>");
  }
  return { input, output };
}

function normalizeName(tokenName, displayName) {
  if (tokenName?.startsWith("md.")) {
    return tokenName.split(".").slice(1).join(".");
  }

  return displayName
    ?.toLowerCase()
    .replace(/[^a-z0-9]+/g, ".")
    .replace(/^\\.|\\.$/g, "") ?? null;
}

export function normalizeSnapshot(snapshot) {
  if (snapshot.schemaVersion >= 2 && snapshot.source?.kind === "m3-token-table") {
    const tokens = (snapshot.tokenSets ?? []).flatMap((set) => (set.tokens ?? []).map((token) => ({
      name: normalizeName(token.name, token.displayName),
      displayName: token.displayName,
      tokenName: token.name,
      valueKind: token.valueType,
      directValues: token.directValues ?? [],
      contextualValues: token.contextualValues ?? [],
      sourceSet: set.name,
      sourceSetDisplayName: set.displayName,
      sourceTokenId: token.sourceId,
    })));

    return {
      schemaVersion: 2,
      source: snapshot.source,
      tokens,
    };
  }

  const tokens = [];

  for (const set of snapshot.tokenSets ?? []) {
    for (const list of set.lists ?? []) {
      for (const token of list.tokens ?? []) {
        tokens.push({
          name: normalizeName(token.tokenName, token.displayName),
          displayName: token.displayName,
          tokenName: token.tokenName,
          valueKind: token.valueKind,
          value: token.value,
          previewClass: token.previewClass,
          sourceSet: set.name,
          sourceSection: list.sectionIndex,
        });
      }
    }
  }

  return {
    schemaVersion: 1,
    source: snapshot.source,
    tokens,
  };
}

async function main() {
  const { input, output } = parseArgs(process.argv.slice(2));
  const snapshot = JSON.parse(await readFile(input, "utf8"));
  const normalized = normalizeSnapshot(snapshot);
  const outputPath = resolve(output);
  await mkdir(dirname(outputPath), { recursive: true });
  await writeFile(outputPath, `${JSON.stringify(normalized, null, 2)}\\n`, "utf8");
  console.log(`Normalized ${normalized.tokens.length} tokens to ${output}`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  await main();
}
