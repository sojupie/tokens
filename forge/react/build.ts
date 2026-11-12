import { build } from 'esbuild';
import { execFileSync } from 'node:child_process';
import { cp, mkdir, readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const assets = process.argv[2];
if (!assets) throw new Error('Usage: node forge/react/build.ts <original-name publication assets directory>');

const directory = fileURLToPath(new URL('.', import.meta.url));
const dist = path.join(directory, 'dist');
const generated = path.join(directory, '../components/generated');
const catalogAssets = path.join(directory, '../../catalog/web/assets');
const read = (root: string, file: string) => readFile(path.join(root, file), 'utf8');

const [shared, buttons, common, dialogTokens, dialog, checkboxTokens, checkbox, motion, press, modal, catalogCSS] = await Promise.all([
  read(assets, 'themes/shared.css'), read(assets, 'buttons.css'),
  read(generated, 'bundles/common.css'), read(assets, 'themes/dialog.css'),
  read(generated, 'bundles/dialog.css'), read(assets, 'themes/checkbox.css'),
  read(generated, 'bundles/checkbox.css'), read(generated, 'motion.cjs'),
  read(generated, 'buttons.js'), read(generated, 'dialogs.js'), read(catalogAssets, 'styles/app.css'),
]);
if (!shared.includes('--md-sys-color-primary:') || !buttons.includes('--_button-pressed-shape:')) {
  throw new Error('Build publication assets with --css-names original for this package.');
}

// Keep the same token and component cascade as the catalog. The final rule
// restores native dialog centering when an application reset removes its margin.
const fontFaces = catalogCSS.match(/^@font-face \{.*roboto-.*\}$/gm);
if (fontFaces?.length !== 2) throw new Error('Expected Catalog’s two self-hosted Roboto subsets.');
const styles = [...fontFaces, shared, buttons, common, checkboxTokens, checkbox, dialogTokens, dialog,
  '.md-dialog[data-dialog-surface]:modal { margin: auto; }'].join('\n');
const runtime = ['if (typeof document !== "undefined") {', motion, press, modal, '}'].join('\n');

await mkdir(path.join(dist, 'fonts'), { recursive: true });
await Promise.all([
  ...['roboto-latin.woff2', 'roboto-latin-ext.woff2', 'OFL.txt'].map(file =>
    cp(path.join(catalogAssets, 'fonts', file), path.join(dist, 'fonts', file))),
  cp(path.join(generated, 'NOTICE.txt'), path.join(dist, 'NOTICE.txt')),
  writeFile(path.join(dist, 'runtime.js'), runtime),
  writeFile(path.join(dist, 'styles.css'), styles),
  build({ entryPoints: [path.join(directory, 'index.tsx')], outdir: dist,
    banner: { js: "import './runtime.js';" }, format: 'esm', jsx: 'automatic', target: 'es2022' }),
]);
execFileSync(process.execPath, [path.join(directory, '../../node_modules/typescript/bin/tsc'), '-p', directory], { stdio: 'inherit' });
