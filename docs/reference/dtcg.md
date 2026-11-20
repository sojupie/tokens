# DTCG exports and CSS

The exporter converts a saved publication into DTCG 2025.10 JSON for one
selected context, including its references and required dependencies.

## Export tokens via CLI

```sh
go run ./cmd/tokens export --published published --output dist/tokens
```

Default context: Light, Default contrast, Web, 3P, Static font, Standard motion,
Small language height. Override individual dimensions with `--context`:

```sh
go run ./cmd/tokens export --published published \
  --context 'Theme=Dark' --context 'Contrast=High contrast' \
  --output dist/dark-high
```

Outputs:
- `tokens.tokens.json` — the DTCG document.
- `report.json` — publication metadata, selected context, token counts, aliases,
  omitted values, and `extensionOnly` entries (such as circular shapes and
  percentage lengths).

Pass `--themes` to export all six Light/Dark × Default/Medium/High contrast
combinations with an `index.json`. Unknown context choices and conflicting
values fail the export.

## Generate CSS locally

Requires Go and Node.js 22.18+:

```sh
npm ci
npm run build:tokens -- --published published
go run ./cmd/tokens serve --published published --listen 127.0.0.1:8081
```

The build uses pinned Style Dictionary 5.6.1 and writes DTCG documents, reports,
and CSS for all six appearances under `published/assets/`. It also regenerates
[button](buttons.md) and [component](components.md) CSS; rebuild the Go binary
afterwards so embedded bindings match the asset manifest. A failed build leaves
previous assets intact.

For an immutable production publication, build assets outside the publication
directory:

```sh
npm run build:tokens -- --published /path/to/publication --assets /tmp/site-assets --check-bindings
go run ./cmd/tokens serve --published /path/to/publication --assets /tmp/site-assets
```

`--check-bindings` verifies generated component CSS matches the checked-in files
without modifying them. In containers, pass `--tokens-binary /usr/local/bin/tokens`.
See [production asset builds](../operations/deployment.md#generated-publication-assets).

### Site theme

- Color palettes are generated with Material Color Utilities (pinned to 0.3.0;
  0.4.0 breaks Node ESM with extensionless imports) from the Tonal Spot seeds in
  `forge/material/site-theme.json` at contrast levels `0`, `0.5`, and `1`.
- `motionScheme` selects the site's spring context (currently `Expressive`);
  public downloads default to `Standard` and keep Google's captured palette.

## Download and export endpoints

Prebuilt whole-collection downloads (substitute `light-medium`, `light-high`,
`dark-default`, `dark-medium`, or `dark-high`; prefix with `/md3` in production):

```text
/downloads/dtcg/light-default/tokens.tokens.json
/downloads/dtcg/light-default/tokens.css
/downloads/dtcg/light-default/report.json
```

Scoped exports via `GET /downloads/export`:
- **Parameters:** `kind`, `slug`, `scope` (`view`, `set`, `subject`, or
  `collection`), `set`, `q`, `context.<dimension>`, and `format` (`dtcg`, `css`,
  `source`, `report`, or `summary`).
- **Batch ZIPs:** Repeat `format` or `context.<dimension>` (up to 64
  combinations) to download one ZIP with `selections.txt` and per-export
  conversion reports. Add `summary=1` to inspect a batch without downloading, or
  `options=1` to list applicable context values.
- **Scoped CSS:** Uses a build-time Style Dictionary value index so downloads
  never spawn Node.
- **Component implementations:** `content=implementation&format=zip` exports a
  component or collection archive with example HTML, required CSS modules,
  runtime scripts, catalog theme, fonts, and license notices (theme and contrast
  are selectable).

## Material conversion

| Source value | DTCG representation |
| --- | --- |
| Color | `color` (sRGB components + alpha) |
| Logical DIPS, type-scale POINTS, PIXELS | `dimension` in `px` |
| Font family and weight | `fontFamily`, `fontWeight` |
| Typography | `typography` (absolute line height converted to font-size multiplier) |
| Duration | `duration` in `ms` |
| Cubic Bézier | `cubicBezier` (4 coordinates) |
| Opacity, numeric value, font axis | `number` |
| Rounded shape | Group of four corner dimensions |
| Elevation | `dimension` retaining source distance (app theme maps 0/1/3/6/8/12 dp to shadow levels) |

- An omitted numeric field in a typed dimension means `0`; explicit `null` is
  invalid. Unspecified shape corners inherit a published default; unresolvable
  shapes are listed in `report.json`.
- `dev.tokencatalog.material` preserves the original typed value, source name,
  references, and variable-font axes. Single primitive reference chains can
  supply missing dependency values; aliases are emitted only when target types
  and values match.
- Values without a DTCG type—circular shapes, percentage lengths, motion paths,
  and text transforms—are recorded in `report.json`. Circular shapes and
  percentage lengths are kept in `dev.tokencatalog.material` on extension-only
  groups so the CSS compiler can emit `9999px` / `%`.

## Checks

```sh
go test ./...
npm run test:tokens
TOKEN_ASSETS_DIR="$PWD/published/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published/tokens.sqlite" go test ./source/dtcg ./catalog/web
```

Specifications: [DTCG 2025.10](https://www.designtokens.org/tr/2025.10/format/),
[Style Dictionary DTCG](https://styledictionary.com/info/dtcg/).
