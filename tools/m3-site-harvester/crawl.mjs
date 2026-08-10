#!/usr/bin/env node

import { createHash } from "node:crypto";
import { gzip } from "node:zlib";
import { promisify } from "node:util";
import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { dirname, join, relative, resolve } from "node:path";
import process from "node:process";
import { pathToFileURL } from "node:url";
import { chromium } from "playwright";
import { buildSnapshot } from "./harvest.mjs";

const gzipAsync = promisify(gzip);
const DEFAULT_SEED = "https://m3.material.io/components/buttons/specs";
const DEFAULT_CACHE_DIR = ".cache/m3";
const DEFAULT_RUN_DIR = "sources/runs/m3";
const TOKEN_TABLE_PATTERN = /\/TOKEN_TABLE(?:\.[^/]+)?\.json(?:\?.*)?$/i;
const HTML_EXTENSION_PATTERN = /\.(?:css|gif|ico|jpe?g|js|json|mjs|png|svg|webp|xml|woff2?)(?:$|\?)/i;

function parseArgs(argv) {
  const args = {
    seed: DEFAULT_SEED,
    sitemap: null,
    pathPrefix: "/components/",
    maxPages: 500,
    concurrency: 2,
    pageTimeout: 10000,
    cacheDir: DEFAULT_CACHE_DIR,
    runDir: DEFAULT_RUN_DIR,
    refresh: false,
    dryRun: false,
    headed: false,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--") continue;
    if (argument === "--all") continue;
    if (argument === "--seed" || argument === "--url") args.seed = argv[++index];
    else if (argument === "--sitemap") args.sitemap = argv[++index];
    else if (argument === "--path-prefix") args.pathPrefix = argv[++index];
    else if (argument === "--max-pages") args.maxPages = Number(argv[++index]);
    else if (argument === "--concurrency") args.concurrency = Number(argv[++index]);
    else if (argument === "--page-timeout") args.pageTimeout = Number(argv[++index]);
    else if (argument === "--cache-dir") args.cacheDir = argv[++index];
    else if (argument === "--run-dir") args.runDir = argv[++index];
    else if (argument === "--refresh") args.refresh = true;
    else if (argument === "--dry-run") args.dryRun = true;
    else if (argument === "--headed") args.headed = true;
    else if (argument === "--help") {
      printHelp();
      process.exit(0);
    } else throw new Error(`Unknown argument: ${argument}`);
  }

  for (const [name, value] of [["max-pages", args.maxPages], ["concurrency", args.concurrency], ["page-timeout", args.pageTimeout]]) {
    if (!Number.isInteger(value) || value < 1) throw new Error(`--${name} must be a positive integer`);
  }
  return args;
}

function printHelp() {
  console.log(`Usage: pnpm crawl:m3 -- --all [options]

Options:
  --seed <url>             Starting page when sitemap discovery is unavailable
  --sitemap <url>          Sitemap or sitemap-index URL override
  --path-prefix <path>    Restrict discovered pages (default: /components/)
  --max-pages <number>    Maximum pages to inspect (default: 500)
  --concurrency <number>  Browser pages inspected in parallel (default: 2)
  --page-timeout <ms>     Page observation timeout (default: 10000)
  --cache-dir <path>      Content-addressed artifact cache (default: .cache/m3)
  --run-dir <path>        Run manifests (default: sources/runs/m3)
  --refresh               Refetch payloads already present in the cache
  --dry-run               Discover pages and payload URLs without fetching payloads
  --headed                Show Chromium while observing pages
  --help                  Show this help`);
}

async function fetchText(url) {
  const response = await fetch(url, { headers: { accept: "text/html,application/xml,text/xml,*/*" } });
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
  return response.text();
}

function extractLocs(xml) {
  return [...xml.matchAll(/<loc(?:\s[^>]*)?>([\s\S]*?)<\/loc>/gi)]
    .map((match) => match[1].trim().replaceAll("&amp;", "&"));
}

function extractRobotsSitemaps(robots, origin) {
  return robots.split(/\r?\n/)
    .map((line) => line.match(/^\s*sitemap\s*:\s*(\S+)/i)?.[1])
    .filter(Boolean)
    .map((url) => new URL(url, origin).href);
}

async function collectSitemapUrls(sitemapUrl, maxDepth = 3, seen = new Set()) {
  if (maxDepth < 0 || seen.has(sitemapUrl)) return [];
  seen.add(sitemapUrl);
  let xml;
  try {
    xml = await fetchText(sitemapUrl);
  } catch {
    return [];
  }
  const locs = extractLocs(xml);
  if (/<sitemapindex[\s>]/i.test(xml)) {
    const nested = await Promise.all(locs.map((loc) => collectSitemapUrls(loc, maxDepth - 1, seen)));
    return nested.flat();
  }
  return locs;
}

function canonicalPageUrl(value, origin, pathPrefix) {
  try {
    const url = new URL(value, origin);
    url.hash = "";
    if (url.origin !== origin || !url.pathname.startsWith(pathPrefix)) return null;
    if (HTML_EXTENSION_PATTERN.test(url.pathname)) return null;
    return url.href;
  } catch {
    return null;
  }
}

function extractPageLinks(html, pageUrl, pathPrefix) {
  const origin = new URL(pageUrl).origin;
  return [...html.matchAll(/<a\b[^>]*href=["']([^"']+)["']/gi)]
    .map((match) => canonicalPageUrl(match[1], origin, pathPrefix))
    .filter(Boolean);
}

async function discoverPages({ seed, sitemap, pathPrefix, maxPages }) {
  const seedUrl = new URL(seed);
  const sitemapUrls = sitemap
    ? [new URL(sitemap, seedUrl.origin).href]
    : extractRobotsSitemaps(await fetchText(`${seedUrl.origin}/robots.txt`).catch(() => ""), seedUrl.origin);
  const discovered = new Set();

  for (const sitemapUrl of sitemapUrls.length ? sitemapUrls : [`${seedUrl.origin}/sitemap.xml`]) {
    for (const pageUrl of await collectSitemapUrls(sitemapUrl)) {
      const canonical = canonicalPageUrl(pageUrl, seedUrl.origin, pathPrefix);
      if (canonical) discovered.add(canonical);
      if (discovered.size >= maxPages) break;
    }
    if (discovered.size >= maxPages) break;
  }

  const seedPage = canonicalPageUrl(seed, seedUrl.origin, pathPrefix) ?? seed;
  if (!discovered.size) discovered.add(seedPage);

  if (discovered.size < maxPages) {
    const queue = [...discovered];
    const visited = new Set();
    while (queue.length && discovered.size < maxPages) {
      const pageUrl = queue.shift();
      if (visited.has(pageUrl)) continue;
      visited.add(pageUrl);
      const html = await fetchText(pageUrl).catch(() => "");
      for (const link of extractPageLinks(html, pageUrl, pathPrefix)) {
        if (!discovered.has(link)) {
          discovered.add(link);
          queue.push(link);
          if (discovered.size >= maxPages) break;
        }
      }
    }
  }

  return [...discovered].slice(0, maxPages).sort();
}

async function observePage(browser, pageUrl, timeout, headed) {
  const page = await browser.newPage();
  let tokenTableUrl = null;
  const onResponse = (response) => {
    if (TOKEN_TABLE_PATTERN.test(response.url())) tokenTableUrl = response.url();
  };
  page.on("response", onResponse);
  try {
    const response = await page.goto(pageUrl, { waitUntil: "domcontentloaded", timeout });
    const deadline = Date.now() + timeout;
    while (!tokenTableUrl && Date.now() < deadline) await page.waitForTimeout(200);
    return {
      pageUrl,
      status: response?.status() ?? null,
      tokenTableUrl,
      title: await page.title().catch(() => null),
    };
  } catch (error) {
    return {
      pageUrl,
      status: null,
      tokenTableUrl,
      title: null,
      error: error instanceof Error ? error.message : String(error),
    };
  } finally {
    page.off("response", onResponse);
    await page.close();
  }
}

async function mapLimit(values, concurrency, worker) {
  const results = new Array(values.length);
  let nextIndex = 0;
  async function consume() {
    while (true) {
      const index = nextIndex;
      nextIndex += 1;
      if (index >= values.length) return;
      results[index] = await worker(values[index], index);
    }
  }
  await Promise.all(Array.from({ length: Math.min(concurrency, values.length) }, consume));
  return results;
}

async function writeJsonAtomic(filePath, value) {
  await mkdir(dirname(filePath), { recursive: true });
  const temporaryPath = `${filePath}.${process.pid}.tmp`;
  await writeFile(temporaryPath, `${JSON.stringify(value, null, 2)}\n`, "utf8");
  await rename(temporaryPath, filePath);
}

async function readJsonOr(filePath, fallback) {
  try {
    return JSON.parse(await readFile(filePath, "utf8"));
  } catch {
    return fallback;
  }
}

function hashBytes(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

function artifactPath(cacheDir, hash) {
  return resolve(cacheDir, "payloads", hash.slice(0, 2), `${hash}.json.gz`);
}

async function persistPayload(cacheDir, tokenTableUrl, pageUrl, refresh, index) {
  const cached = index.payloads[tokenTableUrl];
  if (!refresh && cached?.hash) {
    const cachedPath = artifactPath(cacheDir, cached.hash);
    try {
      const compressed = await readFile(cachedPath);
      const raw = await new Promise((resolvePromise, reject) => {
        import("node:zlib").then(({ gunzip }) => gunzip(compressed, (error, bytes) => error ? reject(error) : resolvePromise(bytes)));
      });
      return { ...cached, pageUrl, reused: true, payload: JSON.parse(raw.toString("utf8")) };
    } catch {
      // Re-fetch a missing or corrupt cache entry.
    }
  }

  const response = await fetch(tokenTableUrl, { headers: { accept: "application/json" } });
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
  const raw = Buffer.from(await response.arrayBuffer());
  const hash = hashBytes(raw);
  const compressed = await gzipAsync(raw, { level: 9 });
  const path = artifactPath(cacheDir, hash);
  await mkdir(dirname(path), { recursive: true });
  try {
    await writeFile(path, compressed, { flag: "wx" });
  } catch (error) {
    if (error?.code !== "EEXIST") throw error;
  }
  const record = {
    pageUrl,
    tokenTableUrl,
    hash,
    artifact: relative(process.cwd(), path).replaceAll("\\", "/"),
    bytes: raw.byteLength,
    retrievedAt: new Date().toISOString(),
    etag: response.headers.get("etag"),
    lastModified: response.headers.get("last-modified"),
    reused: false,
    payload: JSON.parse(raw.toString("utf8")),
  };
  index.payloads[tokenTableUrl] = { ...record, payload: undefined };
  return record;
}

async function crawl(args) {
  const startedAt = new Date().toISOString();
  const pages = await discoverPages(args);
  console.log(`Discovered ${pages.length} candidate pages.`);

  const browser = await chromium.launch({ headless: !args.headed });
  let observations;
  try {
    observations = await mapLimit(pages, args.concurrency, (pageUrl) => observePage(browser, pageUrl, args.pageTimeout, args.headed));
  } finally {
    await browser.close();
  }

  const tokenSources = [...new Map(observations.filter((item) => item.tokenTableUrl).map((item) => [item.tokenTableUrl, item])).values()];
  console.log(`Discovered ${tokenSources.length} unique token-table payloads.`);
  const indexPath = resolve(args.cacheDir, "index.json");
  const index = await readJsonOr(indexPath, { schemaVersion: 1, payloads: {} });
  const payloads = [];

  if (!args.dryRun) {
    for (const source of tokenSources) {
      try {
        const record = await persistPayload(args.cacheDir, source.tokenTableUrl, source.pageUrl, args.refresh, index);
        const snapshot = buildSnapshot(record.payload, source.pageUrl, source.tokenTableUrl, Infinity);
        const derivedRaw = Buffer.from(JSON.stringify(snapshot));
        const derivedHash = hashBytes(derivedRaw);
        const derivedPath = resolve(args.cacheDir, "derived", derivedHash.slice(0, 2), `${derivedHash}.json.gz`);
        await mkdir(dirname(derivedPath), { recursive: true });
        await writeFile(derivedPath, await gzipAsync(derivedRaw, { level: 9 }), { flag: "wx" }).catch((error) => {
          if (error?.code !== "EEXIST") throw error;
        });
        payloads.push({
          ...record,
          payload: undefined,
          derivedHash,
          derivedArtifact: relative(process.cwd(), derivedPath).replaceAll("\\", "/"),
          tokenSetCount: snapshot.tokenSets.length,
          tokenCount: snapshot.tokenSets.reduce((count, set) => count + set.tokens.length, 0),
        });
      } catch (error) {
        payloads.push({
          pageUrl: source.pageUrl,
          tokenTableUrl: source.tokenTableUrl,
          error: error instanceof Error ? error.message : String(error),
        });
      }
    }
    await writeJsonAtomic(indexPath, index);
  }

  const runId = startedAt.replace(/[:.]/g, "-");
  const manifest = {
    schemaVersion: 1,
    runId,
    startedAt,
    completedAt: new Date().toISOString(),
    mode: "all",
    options: { ...args },
    pages,
    observations,
    payloads,
  };
  const manifestPath = resolve(args.runDir, `${runId}.json`);
  await writeJsonAtomic(manifestPath, manifest);
  console.log(`Wrote run manifest to ${manifestPath}`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    await crawl(parseArgs(process.argv.slice(2)));
  } catch (error) {
    console.error(error instanceof Error ? error.message : error);
    process.exitCode = 1;
  }
}

export { crawl, discoverPages };
