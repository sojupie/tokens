import { readFile, writeFile, readdir, mkdir, unlink } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { compileBehavior } from '../../forge/components/build-assets.ts';

const root = new URL('./assets/scripts/', import.meta.url);

export async function catalogAssets() {
  const outputs = new Map<string, string>();
  // Discover maintained entries so a new source cannot omit its served output.
  for (const file of (await readdir(root)).sort()) {
    if (!file.endsWith('.ts') || file.endsWith('.d.ts')) continue;
    const { code } = await compileBehavior(`../../catalog/web/assets/scripts/${file}`, false);
    outputs.set(file.replace(/\.ts$/, '.js'), `/* Generated from catalog/web/assets/scripts/${file}. DO NOT EDIT. Run npm run build:components. */\n` + code);
  }
  return outputs;
}

export async function buildCatalogAssets({ check = false } = {}) {
  const outputs = await catalogAssets();
  const directory = new URL("./assets/generated/", import.meta.url);
  await mkdir(directory, {recursive:true});
  for (const file of await readdir(directory)) {
    if (!file.endsWith(".js") || outputs.has(file)) continue;
    const destination = new URL(file, directory);
    if (check) throw Error(`Orphaned generated script: ${file}`);
    if (!(await readFile(destination, "utf8")).startsWith("/* Generated from catalog/web/assets/scripts/"))
      throw Error(`Refusing to remove unrecognized script: ${file}`);
    await unlink(destination);
  }
  for (const [file, code] of outputs) {
    const destination = new URL(`./assets/generated/${file}`, import.meta.url);
    if (check) {
      if (await readFile(destination, 'utf8') !== code)
        throw Error(`${file} is stale; run npm run build:components`);
    } else { await mkdir(new URL('./assets/generated/', import.meta.url), {recursive:true}); await writeFile(destination, code); }
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url))
  await buildCatalogAssets({ check: process.argv.includes('--check') });
