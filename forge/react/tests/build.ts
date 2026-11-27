import { build } from 'esbuild';
import { readFile, mkdir, writeFile, copyFile } from 'node:fs/promises';
import path from 'node:path';

const [packedDirectory, output] = process.argv.slice(2);
if (!packedDirectory || !output) throw new Error('Provide the extracted package directory and fixture output directory.');
const manifest: unknown = JSON.parse(await readFile(path.join(packedDirectory, 'package.json'), 'utf8'));
if (!manifest || typeof manifest !== 'object' || !('exports' in manifest) || !manifest.exports || typeof manifest.exports !== 'object') throw new Error('Package exports are missing.');
const exports = manifest.exports;
const main: unknown = Reflect.get(exports, '.');
const styles: unknown = Reflect.get(exports, './styles.css');
if (!main || typeof main !== 'object' || !('import' in main) || typeof main.import !== 'string' || typeof styles !== 'string') throw new Error('Package entry points are missing.');
const entry = (value: string) => {
  const resolved = path.resolve(packedDirectory, value);
  if (!resolved.startsWith(path.resolve(packedDirectory) + path.sep)) throw new Error('Package export escapes its directory.');
  return resolved;
};
await mkdir(output, { recursive: true });
await build({ entryPoints: ['forge/react/tests/fixture.tsx'], bundle: true, format: 'iife', outfile: path.join(output, 'fixture.js'),
  alias: { '@token-catalog/forge-react': entry(main.import), '@token-catalog/forge-react/styles.css': entry(styles) },
  nodePaths: [path.resolve('forge/react/node_modules')], jsx: 'automatic', loader: { '.woff2': 'dataurl' },
  define: { 'process.env.NODE_ENV': '"production"' } });
await writeFile(path.join(output, 'index.html'), '<!doctype html><html lang="en"><meta charset="utf-8"><title>Forge package checks</title><link rel="stylesheet" href="fixture.css"><div id="root"></div><script src="fixture.js"></script></html>');
await copyFile('forge/react/tests/rendered-checks.mjs', path.join(output, 'rendered-checks.mjs'));
await copyFile('forge/conformance/browser/check-button-rendering.mjs', path.join(output, 'check-button-rendering.mjs'));
