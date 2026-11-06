import { readFile, writeFile, readdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { compileBehavior } from '../../forge/components/build-assets.ts';

const root = new URL('./static/', import.meta.url);

export async function catalogAssets() {
  const outputs = new Map<string, string>();
  // Discover maintained entries so a new source cannot omit its served output.
  for (const file of (await readdir(root)).sort()) {
    if (!file.endsWith('.ts') || file.endsWith('.d.ts')) continue;
    const { code } = await compileBehavior(`../../catalog/web/static/${file}`, false);
    outputs.set(file.replace(/\.ts$/, '.js'), `/* Generated from ${file}. Do not edit. */\n` + code);
  }
  return outputs;
}

export async function buildCatalogAssets({ check = false } = {}) {
  for (const [file, code] of await catalogAssets()) {
    const destination = new URL(file, root);
    if (check) {
      if (await readFile(destination, 'utf8') !== code)
        throw Error(`${file} is stale; run npm run build:components`);
    } else await writeFile(destination, code);
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url))
  await buildCatalogAssets({ check: process.argv.includes('--check') });
