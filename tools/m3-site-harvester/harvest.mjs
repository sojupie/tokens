#!/usr/bin/env node

import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import process from "node:process";
import { pathToFileURL } from "node:url";
import { chromium } from "playwright";

const DEFAULT_URL = "https://m3.material.io/components/buttons/specs";
const DEFAULT_OUTPUT = "sources/raw/m3-material/buttons.json";
const TOKEN_TABLE_PATTERN = /\/TOKEN_TABLE(?:\.[^/]+)?\.json(?:\?.*)?$/i;
const BUTTON_SET_PREFIX = "md.comp.button";
const VALUE_FIELDS = [
  "color",
  "length",
  "opacity",
  "shape",
  "elevation",
  "numeric",
  "axisValue",
  "fontTracking",
  "fontNames",
  "fontSize",
  "lineHeight",
  "fontWeight",
  "type",
];

function readArgs(argv) {
  const args = {
    url: DEFAULT_URL,
    output: DEFAULT_OUTPUT,
    tokenTableUrl: null,
    limit: Infinity,
    discoveryTimeout: 30000,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--") continue;
    if (argument === "--url") args.url = argv[++index];
    else if (argument === "--output") args.output = argv[++index];
    else if (argument === "--token-table-url") args.tokenTableUrl = argv[++index];
    else if (argument === "--limit") args.limit = Number(argv[++index]);
    else if (argument === "--discovery-timeout") args.discoveryTimeout = Number(argv[++index]);
    else if (argument === "--help") {
      printHelp();
      process.exit(0);
    } else throw new Error(`Unknown argument: ${argument}`);
  }

  if (args.limit !== Infinity && (!Number.isInteger(args.limit) || args.limit < 1)) {
    throw new Error("--limit must be a positive integer");
  }
  if (!Number.isInteger(args.discoveryTimeout) || args.discoveryTimeout < 1000) {
    throw new Error("--discovery-timeout must be at least 1000 milliseconds");
  }
  return args;
}

function printHelp() {
  console.log(`Usage: pnpm harvest:m3 -- [options]

Options:
  --url <url>                   Material specification URL used for discovery
  --token-table-url <url>      Fetch a known TOKEN_TABLE JSON URL directly
  --output <path>              Raw JSON output path
  --limit <number>             Harvest only the first N button token sets
  --discovery-timeout <ms>     Maximum browser discovery time (default: 30000)
  --help                       Show this help`);
}

async function discoverTokenTableUrl(url, timeout) {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  let tokenTableUrl = null;

  const onResponse = (response) => {
    if (TOKEN_TABLE_PATTERN.test(response.url())) tokenTableUrl = response.url();
  };
  page.on("response", onResponse);

  try {
    await page.goto(url, { waitUntil: "domcontentloaded", timeout });
    const deadline = Date.now() + timeout;
    while (!tokenTableUrl && Date.now() < deadline) {
      await page.waitForTimeout(250);
    }
  } finally {
    page.off("response", onResponse);
    await browser.close();
  }

  if (!tokenTableUrl) {
    throw new Error(`Could not discover a TOKEN_TABLE response from ${url}. Pass --token-table-url to use a captured URL.`);
  }
  return tokenTableUrl;
}

async function fetchJson(url) {
  const response = await fetch(url, { headers: { accept: "application/json" } });
  if (!response.ok) throw new Error(`Token table request failed (${response.status} ${response.statusText}): ${url}`);
  return response.json();
}

function tokenSetMetadata(tokenSet) {
  return {
    name: tokenSet.tokenSetName,
    displayName: tokenSet.displayName ?? tokenSet.tokenSetName,
    description: tokenSet.description ?? null,
    state: tokenSet.state,
    tokenType: tokenSet.tokenType ?? null,
    revisionId: tokenSet.revisionId ?? null,
    revisionCreateTime: tokenSet.revisionCreateTime ?? null,
    sourceId: tokenSet.name,
  };
}

function extractLiteral(value) {
  if (!value || typeof value !== "object") return null;
  const field = VALUE_FIELDS.find((key) => Object.hasOwn(value, key));
  return field ? { kind: field, value: value[field] } : null;
}

function buildTokenRecord(token, tokenSet, system, valuesByName) {
  const contexts = system.contextualReferenceTrees?.[token.name]?.contextualReferenceTree ?? [];
  const directValues = system.values
    .filter((value) => value.name.startsWith(`${token.name}/values/`))
    .map((value) => ({ ...value, literal: extractLiteral(value) }));

  const contextualValues = contexts.map((context) => {
    const references = [];
    const visit = (node) => {
      if (!node || typeof node !== "object") return;
      if (node.value?.name) {
        const value = valuesByName.get(node.value.name);
        references.push({
          tokenName: node.tokenName ?? null,
          valueName: node.value.name,
          valueRevisionId: node.value.revisionId ?? null,
          literal: extractLiteral(value),
          valueTokenName: value?.tokenName ?? null,
        });
      }
      for (const child of node.childNodes ?? []) visit(child);
    };
    visit(context.referenceTree);
    return {
      contextTags: context.contextTags ?? [],
      resolvedValue: context.resolvedValue ?? null,
      references,
    };
  });

  return {
    name: token.tokenName,
    displayName: token.displayName ?? null,
    description: token.description ?? null,
    valueType: token.tokenValueType ?? null,
    revisionId: token.revisionId ?? null,
    revisionCreateTime: token.revisionCreateTime ?? null,
    sourceId: token.name,
    tokenSet: tokenSet.tokenSetName,
    directValues,
    contextualValues,
  };
}

function buildSnapshot(payload, sourceUrl, tokenTableUrl, limit) {
  const system = payload.system;
  if (!system || !Array.isArray(system.tokenSets) || !Array.isArray(system.tokens)) {
    throw new Error("The TOKEN_TABLE payload did not contain the expected system/tokenSets/tokens structure.");
  }

  const valuesByName = new Map((system.values ?? []).map((value) => [value.name, value]));
  const tokenSets = system.tokenSets
    .filter((tokenSet) => tokenSet.state === "ACTIVE")
    .filter((tokenSet) => tokenSet.tokenSetName === BUTTON_SET_PREFIX || tokenSet.tokenSetName?.startsWith(`${BUTTON_SET_PREFIX}.`))
    .filter((tokenSet) => !tokenSet.displayName?.startsWith("[Deprecated]"))
    .sort((left, right) => (left.order ?? 0) - (right.order ?? 0) || left.tokenSetName.localeCompare(right.tokenSetName))
    .slice(0, limit);

  const sets = tokenSets.map((tokenSet) => ({
    ...tokenSetMetadata(tokenSet),
    tokens: system.tokens
      .filter((token) => token.name.startsWith(`${tokenSet.name}/tokens/`))
      .sort((left, right) => (left.orderInDisplayGroup ?? 0) - (right.orderInDisplayGroup ?? 0) || left.tokenName.localeCompare(right.tokenName))
      .map((token) => buildTokenRecord(token, tokenSet, system, valuesByName)),
  }));

  return {
    schemaVersion: 2,
    source: {
      kind: "m3-token-table",
      pageUrl: sourceUrl,
      tokenTableUrl,
      retrievedAt: new Date().toISOString(),
      upstream: {
        name: system.name ?? null,
        displayName: system.displayName ?? null,
        dsdbVersion: system.dsdbVersion ?? null,
        revisionId: system.revisionId ?? null,
        revisionCreateTime: system.revisionCreateTime ?? null,
        version: system.version ?? null,
        tokenNamePrefix: system.tokenNamePrefix ?? null,
      },
    },
    tokenSets: sets,
  };
}

async function harvest({ url, output, tokenTableUrl, limit, discoveryTimeout }) {
  const resolvedTokenTableUrl = tokenTableUrl ?? await discoverTokenTableUrl(url, discoveryTimeout);
  const payload = await fetchJson(resolvedTokenTableUrl);
  const snapshot = buildSnapshot(payload, url, resolvedTokenTableUrl, limit);
  const outputPath = resolve(output);
  await mkdir(dirname(outputPath), { recursive: true });
  await writeFile(outputPath, `${JSON.stringify(snapshot, null, 2)}\n`, "utf8");
  const tokenCount = snapshot.tokenSets.reduce((count, set) => count + set.tokens.length, 0);
  console.log(`Harvested ${snapshot.tokenSets.length} token sets and ${tokenCount} tokens to ${output}`);
}

export { buildSnapshot, discoverTokenTableUrl, fetchJson };

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    await harvest(readArgs(process.argv.slice(2)));
  } catch (error) {
    console.error(error instanceof Error ? error.message : error);
    process.exitCode = 1;
  }
}
