# DTCG exports and CSS

The exporter converts a saved publication into DTCG 2025.10 JSON. Each export
contains one selected context, its references and their dependencies. Google’s
captured payloads remain unchanged.

## Export tokens

```sh
go run ./cmd/tokens export --published published --output dist/tokens
```

The default context uses Light, Default contrast, Web, 3P, Static font,
Standard motion and Small language height. Override individual dimensions:

```sh
go run ./cmd/tokens export --published published \
  --context 'Theme=Dark' --context 'Contrast=High contrast' \
  --output dist/dark-high
```

`tokens.tokens.json` contains the DTCG document. `report.json` records the
publication, selected context, DTCG token count, mapped source names, aliases
and omitted values. `extensionOnly` lists circular shapes represented through
source metadata.
Unknown context choices and conflicting values fail the export. Tokens with
no defined value in the selected context appear in the report.

`--themes` exports the six Light/Dark and Default/Medium/High contrast
combinations used by the site, with an `index.json` listing their files.
Other context dimensions keep their selected values.

## Generate CSS locally

Install Node.js 22 or newer and Go, then run from the repository root:

```sh
npm ci
npm run build:tokens -- --published published
go run ./cmd/tokens serve --published published --listen 127.0.0.1:8081
```

The build uses pinned Style Dictionary 5.6.1. It writes DTCG documents, reports
and CSS for all six appearance contexts under `published/assets/`. The site
loads `assets/theme.css` on startup. Its manifest binds the generated theme to
the exact SQLite database and checks the CSS checksum.

The site's color palette is generated with Material Color Utilities from the
primary and secondary seeds in `forge/material/site-theme.json`. Change either
palette seed and rebuild to update
light and dark themes at default, medium, and high contrast. The build uses
Tonal Spot with contrast levels 0, 0.5, and 1. Component color aliases resolve
against these generated roles. Captured DTCG and downloadable CSS retain the
upstream palette.

Material Color Utilities is pinned to 0.3.0 because 0.4.0 contains extensionless
imports that fail in the Node.js build.

The build also generates [button styles and a coverage report](buttons.md)
from the current color and size sets. Rebuild the Go binary after this step.

The app stylesheet includes the token variables referenced by its static CSS
and generated button bindings. Downloadable CSS includes every converted token. The builder checks that
every required app token exists in all six themes before replacing the assets.
A failed build preserves the previous assets.

The server exposes these files when generated assets are present:

```text
/downloads/dtcg/light-default/tokens.tokens.json
/downloads/dtcg/light-default/tokens.css
/downloads/dtcg/light-default/report.json
```

The site's **Download tokens** dialog exposes these three formats and all six
appearance contexts. DTCG JSON is the default. Each file includes the whole
collection for that context; page, table and search filters do not apply. The
older custom JSON API/download URLs remain available for compatibility. They
have not been silently changed to DTCG.

Replace `light-default` with `light-medium`, `light-high`, `dark-default`,
`dark-medium` or `dark-high`. Include the collection mount, such as `/md3`,
when accessing them on the public site.

Rebuild local assets after publishing new data, then restart the server.
Keep these builds outside approved production generations, whose digests cover
every file.
Saved publications without generated assets continue to use the existing Go
stylesheet generator. The scheduled update pipeline currently builds the saved
publication without invoking Node.js; this CSS build is a separate step.
The existing API and JSON download formats remain unchanged.

## Material conversion

| Source value | DTCG representation |
| --- | --- |
| Color | `color` with sRGB components and alpha |
| Logical DIPS, type-scale POINTS, PIXELS | `dimension` in CSS pixels |
| Font family and weight | `fontFamily` and `fontWeight` |
| Typography | `typography`; absolute line height becomes a font-size multiplier |
| Duration | `duration` with milliseconds |
| Cubic Bézier | `cubicBezier` with four coordinates |
| Opacity, numeric value, font axis value | `number` |
| Rounded shape | A group of four corner dimensions |
| Elevation | A dimension retaining the source elevation distance |

An omitted numeric field in a typed source dimension represents zero. An
explicit null remains invalid. Named shape corners override the default;
unspecified corners inherit a published default. A shape lacking enough
information to resolve all four corners appears in the omission report.

The `dev.tokencatalog.material` extension preserves the original typed value,
source name, references and presentation metadata. A single primitive reference
chain can supply dependency values absent as standalone tokens. Conflicting
inferences are discarded. Aliases are emitted only for present targets with
matching types and values. Branching references remain source metadata.

DTCG has no percentage dimension or general shape, motion-path or text-transform
type. Percentage dimensions and unsupported value types appear in the report
with their source values. Circular shapes retain their family in a group
extension; the CSS builder emits an infinite border radius. Variable-font axes
inside typography remain in the source extension.

The downloadable CSS preserves elevation distances in pixels. The app theme
converts Material’s 0, 1, 3, 6, 8 and 12 dp distances into the levels expected by
its existing shadow renderer. Shadow geometry remains defined by that renderer.

The build includes a duration transform and adapts `$root` groups to Style
Dictionary’s input. The exported JSON retains the DTCG group structure.

## Checks

```sh
go test ./...
npm run test:tokens
TOKEN_ASSETS_DIR="$PWD/published/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published/tokens.sqlite" go test ./source/dtcg ./catalog/web
```

The last two commands exercise the real publication and its six generated
themes. The checks cover aliases, corner overrides, units, zero values,
typography line height, duration output, required theme tokens and stale assets.

Specifications: [DTCG format](https://www.designtokens.org/tr/2025.10/format/),
[Style Dictionary compatibility](https://styledictionary.com/info/dtcg/).

Percentage lengths remain in the Material source extension on groups without a
DTCG `$value`, following the existing handling of circular shapes. The CSS
compiler reads that extension and emits `%`. This preserves values such as
selected split-button corners without inventing a DTCG dimension unit.
