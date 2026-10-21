# Component implementation and export handover

Checked against the working tree and local review publication on 2026-10-20.
This is a description of the implementation, its limits, and proposed next work.
The library and adapter APIs below are proposals, not shipped features.

## Repository and runtime

The repository now contains three parts in one Go module:

- `source/` captures Material data, preserves evidence, verifies snapshots,
  publishes the database, and converts selected contexts to DTCG.
- `forge/` compiles tokens into CSS bindings and supplies Web component anatomy,
  state rendering, motion, and browser behavior.
- `catalog/` renders the site, component specimens, illustrations, API and downloads.

The server reads the publication and assets at startup. CSS and browser scripts
are also embedded in the Go binary. Editing a source stylesheet alone does not
update a running server. Rebuild generated assets when their inputs change,
rebuild the binary, then restart it.

The local review server runs as the temporary systemd user unit
`token-catalog-dev`, bound to `127.0.0.1:8081`. Its binary is `bin/tokens-dev`,
and its publication is `published-review`. This keeps testing independent of
chat command sessions. It is not installed or enabled to start at login.
The transient unit is collected when stopped; it can restart after a failure
while testing. Stop it with `systemctl --user stop token-catalog-dev`.

```sh
npm run build:tokens -- --published published-review
go build -o bin/tokens-dev.next ./cmd/tokens
mv bin/tokens-dev.next bin/tokens-dev
systemctl --user restart token-catalog-dev
journalctl --user -u token-catalog-dev -n 30 --no-pager
```

The temporary output name avoids overwriting an executable while it is running.
The temporary unit is testing infrastructure, not an installed development service.
No push or production deployment is part of this work.

## Work covered in this review

The detailed component contracts and source conflicts live in
[components.md](reference/components.md), with a separate
[toolbar specification](reference/toolbar-spec.md). These are more useful for
continuation than reconstructing the chat.

| Area | Result and decisions to preserve |
| --- | --- |
| Date pickers | Input spacing accounts for the floating label. Calendar, input and full-screen presentations transition between configurations; full-screen context must be retained when returning from input. |
| Dialogs | Explicit basic/full-screen anatomy and configuration controls; shared focus, dismissal and transition behavior. |
| FAB menu | One toggle surface changes color and shape together; a stale square must not remain behind the circular close surface. |
| Lists | Whole-row dragging, expansion choices for inset, surface and chevron treatment, and a single-pointer reorder alternative through the shared bottom sheet. The reorder sheet is standard/nonmodal, with no scrim. Move actions respect list boundaries. |
| Menus | Shared menu primitive, keyboard behavior and Material Symbols for shortcut glyphs; avoid depending on Unicode font fallback for the command symbol. |
| Drawer and rail | Drawer sections contain useful items without an artificial close section. Rail previews model a bounded app, collapse/expand behavior, horizontal expanded labels, and hug/fill active indicators with full-width targets. |
| Search | Contained/divided style is separate from docked/full-screen layout and elective shadow. Adaptive means flexible docked width with a matching view. Opening uses the same search surface. The responsive mail composition uses real primitives, no drawer, and a composition-wide scrim below the active surface. |
| Segmented buttons | Equal stable widths reserve selected content; density changes visual height while retaining the touch target. |
| Sheets and split buttons | Modal edge placement; split icon optical offsets, inward rotation, target areas and menu placement. |
| Cards and toolbar | Inter-card spacing, shared cards, toolbar variants/configurations, overflow and shared menu behavior. Preserve captured values where figures conflict with tokens; see the toolbar specification. |
| Tooltips | Plain: 24dp single-line height, 8dp horizontal and 4dp vertical padding. Rich: 12dp top, 16dp sides, 8dp bottom, 4dp heading gap, 12dp before the action. These supplement the component tokens. |
| Catalog indexes | Removed the preview review labels and their status registry. `/components` and `/styles` now have cards and illustrations, custom preview-card colors/press interaction and shared text-field search. Filtering uses `auto-fill` to preserve card width when only one result remains. |
| Illustrations | Use shared specimens or their actual anatomy/attribute helpers. Corrected segmented buttons to request the label-only configuration they render. Added elevation, motion, shape, spacing and type-scale illustrations, also used on the start page. Illustrations are inert and hidden from assistive technology; each card has one accessible link. |

Illustration framing, cropping, decorative documents and landscape artwork are
catalog presentation. They are not component specification measurements.
The type illustration uses actual type-scale tokens. The motion illustration
is explanatory artwork, not a measured plot of a captured easing function.

## What is generated, and what is authored?

The build path is:

```text
verified snapshot -> published SQLite -> DTCG for one context
                                      -> Style Dictionary variables
                                      -> Material selector/state bindings
                                      -> browser CSS + authored DOM and behavior
```

`forge/build-tokens.mjs` orchestrates exports and the six appearance builds.
`forge/material/generate-buttons.mjs` and `generate-components.mjs` explicitly
map Material names, variants, selections and states into selectors and private
CSS properties. They do not infer an entire component from arbitrary tokens.
`forge/material/overrides.json` is the single place for deliberate token-value
overrides. Generated captures and publications stay out of commits.

| Layer | Ownership |
| --- | --- |
| Source values, references, contexts and provenance | Captured and preserved by Source. |
| Public `--md-*` CSS variables | Compiled from DTCG; app colors additionally use the configured site palette. |
| Component value/state bindings | Generated from explicit Material mappings. |
| DOM, slots, ARIA and native elements | Authored in Templ helpers and specimens. |
| Layout, hit regions and spacing omitted from tokens | Authored CSS, usually referring to system spacing tokens; supplemental measurements are documented. |
| Interaction and animation algorithms | Authored JavaScript; many durations, easing, spring parameters and shapes come from tokens. Algorithms, coordination and fallback policy do not. |
| Elevation shadows | Authored Material Web shadow renderer; captured elevation distances map to renderer levels. |
| SVG icons and loading geometry | Separate geometry assets; not created by color/size tokens. |
| Page layout, illustration framing and content | Catalog code. |

A declaration using a spacing token is token-backed, but the choice to apply it
at a particular point in the anatomy is still authored. Conversely,
`display:flex` is authored CSS without being an arbitrary visual value.
A single percentage cannot describe how much of the design is generated.

The current source-derived CSS inspection report is
`published-review/assets/css-usage.json`. It follows variable references through
the selected asset definitions and classifies declarations as Token, Mixed,
Manual or Unresolved. It scans app CSS, token-table CSS, generated button CSS,
component base CSS and elevation CSS; it does not measure JavaScript, DOM,
interaction completeness, or how many user-visible pixels match Material.
In the 2026-10-20 build, 5,569 scanned declarations comprise 2,451 Token
(44.0%), 307 Mixed (5.5%), 2,799 Manual (50.3%) and 12 Unresolved (0.2%).
Re-run the report after edits;
its categories are evidence for investigation, not a component fidelity score.

Site colors and downloads deliberately differ: the site generates a palette
from `forge/material/site-theme.json`, while downloadable token CSS retains the
captured palette. Document this when comparing a consumer to the live catalog.

## Are the components self-contained?

They are reusable Web primitives, but they are not yet a packaged component
library. The runtime uses ordinary DOM, CSS, SVG and browser APIs. Consumers
would not need Go or a live Token Catalog server if we shipped built assets and
markup/controllers separately.

The remaining dependencies matter:

- Public theme variables, generated bindings and authored base CSS must travel
  together. `tokens.css` alone contains variables, not component implementations.
- Markup must match the existing classes, slots, `data-*` attributes and ARIA
  relationships. Much of that markup currently comes from Catalog's Go helpers.
- Some behavior requires a surrounding demo/composition root and example content.
  Search uses `.md-search-demo`; sheets use `.md-sheet-demo`; dialogs use preview
  attributes. These are real dependencies despite the assets living in Forge.
- Scripts are self-running browser IIFEs with document listeners. Several own
  observers and state maps and use HTMX lifecycle events to initialize or clean
  up replaced content. There is no common public `mount/update/destroy` API.
- Utilities and fonts still live in app CSS. For example the segmented selector
  refers to `.tokens-visually-hidden`. Shipping only Forge CSS loses that utility.
- Icons are rendered by a Catalog helper; the SVG data, icon selection and
  licensing notices must become distributable assets.
- Modern native features include dialog, popover, `inert`, `:has`, observers and
  Web Animations. Some paths have fallbacks, but there is no declared and fully
  tested browser support policy for a released library.

These components can move to another site, provided the markup and dependencies
are carried with them. They cannot yet be imported as a stable npm API. A token
file also does not make this renderer support a different design system:
Material anatomy and mappings remain explicit.

## How to offer a library and adapters

Use the existing Web implementation as the starting point. Do not create a
second CSS implementation for each framework.

1. Separate the DOM contract from preview choices. Give each primitive a small
   typed options/events contract and document required anatomy. Keep mail demos,
   filtering controls and illustration framing in Catalog.
2. Extract controllers from the IIFEs. A controller receives its root and owns
   its listeners/observers; it returns cleanup and exposes only the operations
   it needs. HTMX integration becomes a Catalog adapter around these controllers.
   Avoid importing code that accesses `document` during server rendering.
3. Package browser CSS, tokens/themes, fonts/utilities, SVG assets and controllers
   through explicit entry points, with declarations and license notices. Keep
   public theme hooks stable; private `--_*` implementation variables need not
   become an API. Publish built ESM and `.d.ts`, rather than requiring consumers
   to run Go and the full capture/compiler pipeline.
4. Add a plain TypeScript adapter that constructs the documented DOM or mounts a
   controller on consumer markup. CSS, focus and behavior should be the same as
   the catalog. No Web Component layer is required to achieve this.
5. Add React components that render the same DOM and consume the same CSS.
   React owns children and application state; controllers own explicitly
   delegated imperative behavior such as measurement and native dialog motion.
   Initialize and clean them up with effects. Do not let both independently
   change selected state or replace the same subtree. Define controlled and
   uncontrolled props, refs and event payloads before wrapping complex inputs.
6. Test adapters with server rendering, remounts, cleanup, multiple instances,
   keyboard and pointer input, reduced motion, RTL and supported browsers. An
   adapter that only recreates appearance is incomplete.

Start with Button, Card and SegmentedButton, then a stateful Menu or Dialog.
That tests the CSS-only and controller paths before extracting search and
pickers. Keep one implementation of geometry and behavior. This is a proposed
sequence, not code that was added in this change.

References: [TypeScript package declarations](https://www.typescriptlang.org/docs/handbook/declaration-files/publishing.html)
and [React effect setup and cleanup](https://react.dev/reference/react/useEffect).

## DTCG and existing downloads

DTCG 2025.10 export is implemented in `source/dtcg` and the `tokens export` CLI.
It chooses one context and includes referenced dependencies. Six Light/Dark ×
Default/Medium/High contrast documents are built for hosted downloads. Other
context dimensions can be selected through the CLI; there is no equivalent
arbitrary-context DTCG HTTP export form today.

The local review publication exports 6,585 typed DTCG tokens in Light/Default,
with 4,163 aliases and 6,149 mapped source names. Counts differ across themes.
There are 116 extension-only names and 47 omission records in that context.
These counts have different meanings: shape groups and dependency expansion
mean exported token count is not source-name count.

The `dev.tokencatalog.material` extension preserves source values, context,
references and metadata. Shapes or percentages without a standard typed
representation use group extensions; unsupported values are reported. This is
DTCG plus documented source metadata, not a replacement custom token schema.
Tools that ignore extensions can consume the supported typed values but cannot
reproduce every Material value or this renderer from them. The exported report
is part of the delivery and must remain visible.

With generated assets loaded, these routes are available below `/md3`:

| Download | Contents |
| --- | --- |
| `/downloads/dtcg/light-default/tokens.tokens.json` | DTCG values, aliases, groups and source extensions for one context. |
| `/downloads/dtcg/light-default/report.json` | Context, exported/mapped counts, omissions and extension-only names. |
| `/downloads/dtcg/light-default/tokens.css` | Compiled CSS custom properties for that context. |

Replace `light-default` with `light-medium`, `light-high`, `dark-default`,
`dark-medium` or `dark-high`. The server returns 404 for these routes when the
publication has no generated assets. That build is still a separate step from
publishing; it must be part of any release that promises these downloads.

The visible “Download tokens” controls now open one shared dialog. DTCG JSON is
the default, with CSS variables and the conversion report as alternatives. The
appearance selector lists all six generated contexts. Every choice explicitly
covers all components and styles; page, table and search filters do not apply.
The dialog disables unavailable downloads when no generated assets are loaded.

Existing collection/subject JSON endpoints and API/MCP responses retain the
catalog model for compatibility. Their URLs and filtering behavior have not
changed. Consumers no longer need that schema for the default token download.
A component-only or arbitrarily filtered DTCG export still requires work; it must
include alias dependencies and report its actual context and scope.

Captured evidence, all-context history and the source graph remain useful
internally and through the data API. They are not interchangeable with a
one-context DTCG document. Do not silently change the schema of existing URLs.

[DTCG's format](https://www.designtokens.org/tr/2025.10/format/) defines token
exchange, not component markup, events or behavior. See
[the export reference](reference/dtcg.md) for conversion rules and commands.

## Is CSS export still useful?

Yes. Token CSS saves browser consumers from writing their own DTCG-to-CSS build.
It supplies reusable colors, dimensions, typography, references and motion
values even if a consumer writes its own components. Its purpose is narrower
than a component stylesheet.

Offer three distinct outputs:

| Output | Consumer need |
| --- | --- |
| DTCG + report | Portable token data for tools and multiple platforms. |
| Token CSS | Browser variables compiled from that data. |
| Component library | Authored anatomy, generated bindings, CSS, assets, behavior and framework adapters. |

Handwritten component layout does not make token CSS pointless. It means we must
avoid presenting the variable download as the complete library. A usable
component CSS distribution also needs the compiled base/binding rules and their
theme dependencies; current hosted `tokens.css` does not include them.

## Validation and remaining work

The current generated reports map 3,655 component token bindings and 295 button
bindings. They list excluded source tokens separately; exclusion is not proof
that an entire component lacks support. The 7,329 component and 650 button
configuration entries describe generated cases, not individually verified
browser interactions.

- Focused Catalog rendering/generated-asset tests and Source DTCG tests pass.
  The browser-script suite passes all 92 tests. The token suite passes 37 tests
  and skips three optional checks without captured-data environment settings.
- Checked all 36 component illustrations and five style illustrations in the
  rendered indexes. A live HTMX search reduced the component grid to one card;
  it stayed 270px wide in the same 1104px grid. The style grid also fits a 360px
  viewport without horizontal overflow. Light/dark theme review exposed and
  corrected a fixed light illustration backdrop. Type samples resolve to 36px,
  22px and 14px from their roles.
- All 18 local DTCG routes (six contexts times JSON, CSS and report) returned
  HTTP 200 with the expected MIME types; both JSON files parse in each context.
  This verifies the local review server, not the deployed production server.
- The shared generator suite has a known captured-data test failure: one
  appearance has 3,655 bindings while the test expects the first appearance's
  3,656. This predates the card changes and is not caused by authored card CSS.
  Determine which context legitimately lacks a value before changing the test
  or exporter. Do not claim that all theme parity checks pass.
- The CSS report still contains 12 unresolved declarations. Trace them before
  publishing a claim of complete token coverage.
- Visual review in this change uses the in-app browser. Firefox, screen-reader
  behavior and a packaged consumer installation need separate validation.
- TS/React adapters, a published package and a browser support policy have not
  been implemented here.

The implementation is a substantial base for a library. Its strongest reusable
parts are the token compiler, generated bindings and shared Web primitives.
The next boundary to establish is the public DOM/controller API, followed by
adapter and consumer tests.

## Site controls follow-up

The site navigation now uses shared text/icon button primitives with a native
popover disclosure on compact windows. Desktop and mobile share the same links
and appearance controls. Preview filters use the shared select/listbox primitive,
with separate submitted values and display labels. The token table controls and
their stylesheet/behavior were left intact.

Markdown negotiation explicitly excludes the component preview region and
ARIA-hidden illustrations. It retains content navigation, token tables and
references; example email content and empty preview list bullets are omitted.

Copyable examples are feasible now for a small supported set such as Button,
Card and SegmentedButton. A useful example must state its CSS/theme/icon
requirements and be tested in an otherwise empty page. Complex interactive
examples also need a clear initialization and cleanup contract. Copying the
current specimen DOM alone does not provide that contract.

The export-dialog regression check also replaces the dialog with restored history
content and verifies that downloads retain the collection mount, the active
context and correct file format. Native mobile popover dismissal and dialog
focus return were checked in the browser.

Preview-card styling remains deliberately custom: its colored hover surface and
pressed corner change are part of the catalog design. Reusing primitives does
not require replacing this treatment with a stock filled card.

## Performance follow-up

The index regression came from reading and decoding the 3.6MB component coverage
file once per thumbnail, on every request. Coverage files are now decoded once
per running server, and fixed thumbnail specimen choices are reused across
filtered requests. Restart after changing generated coverage, as with the loaded
publication. Icon paths also use one immutable lookup instead of rebuilding the
same lookup for each SVG.

HTML/Markdown responses and shared CSS/JS assets now negotiate gzip. Asset ETags
identify the encoded representation; conditional requests, HEAD and byte ranges
have regression checks. Existing content-negotiation Vary fields are preserved.
API/MCP streaming and download response formats are unchanged.

Local HTTP measurements on 2026-10-20, after warming the preview cache: the
unfiltered component-search fragment fell from roughly 460ms to 4ms, with a
144KB decoded body transferred as about 17KB. The Switch document retains its
504KB decoded body and transfers about 29KB. These are local response timings
and body sizes, not Firefox paint or interaction measurements.

The rendered Switch page has 4,154 DOM elements, of which 3,841 belong to the
token browser and 116 to the component-preview section. Compression does not
reduce that DOM. Reducing its browser-side cost further would require a separate
review of token-table rendering, which this change leaves intact.
