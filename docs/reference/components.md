# Generated Material Design 3 components

Forge generates component CSS from captured Material Design 3 token sets,
paired with handwritten HTML anatomy, layout, and browser behavior. Common
buttons are documented separately in [buttons](buttons.md). Curated
source-guidance checks live in each component's
[`checks.json`](guidance-curation.md).

Captured values and names stay unchanged in DTCG and source JSON exports.
Handwritten layout values and system-token references are tracked automatically
in `<publication>/assets/css-usage.json`, and excluded source tokens (deprecated
surface-tint layers, superseded baseline sets, composite typography duplicates)
are recorded in `components-coverage.json`.

## Supported families

| Family | Configurations |
| --- | --- |
| App bars | Current and baseline; small, centered, medium, large, flexible, bottom; subtitles, avatars, search, scroll states |
| Badges | Small dot and labeled |
| Bottom sheets | Standard and modal; expanded and minimized; drag, resize, dismissal |
| Button groups | Standard and connected; single/multiple selection; five sizes |
| Cards | Elevated, filled, outlined; enabled, hover, focus, press, drag, disabled |
| Carousels | Multi-browse, uncontained, hero, full-screen; optional outlines |
| Checkboxes | Unchecked, checked, indeterminate, error, disabled |
| Chips | Assist, suggestion, filter, input; flat/elevated; icons, avatars, removal |
| Date pickers | Docked, modal calendar, modal input, full-screen; single and range |
| Dialogs | Basic (`alertdialog`) and full-screen (`dialog`); icons, dividers, scroll states |
| Dividers | Full-width and inset |
| Extended FABs | Icon and label; four sizes; primary, secondary, tertiary, and container variants |
| FABs | Four sizes; primary, secondary, tertiary, and container variants |
| FAB menus | Primary, secondary, tertiary; action and close-button profiles |
| Icon buttons | Standard, filled, tonal, outlined; five sizes, two shapes, three widths, toggle |
| Lists | Standard and segmented; baseline/expressive; selection, expansion, reorder, reveal |
| Loading indicators | Contained and uncontained; determinate and indeterminate shape morphs |
| Menus | Baseline vertical; standard/vibrant vertical, horizontal, icon-only; submenus |
| Navigation bars | Horizontal and vertical items |
| Navigation drawers | Standard and modal |
| Navigation rails | Collapsed, expanded, hidden/immersive; standard or modal expansion |
| Progress indicators | Current and baseline; linear and circular; flat, wavy, four-color |
| Radio buttons | Selected, unselected, disabled |
| Search | Contained and divided; docked, full-screen, adaptive |
| Segmented buttons | Single/multiple selection; labels, icons, density steps |
| Side sheets | Standard and modal; docked and detached; start/end edges |
| Sliders | Continuous and discrete; single and range; horizontal and vertical; five sizes |
| Snackbars | One and two lines; action, dismiss, pausable timeout |
| Split buttons | Filled, tonal, outlined, elevated; five sizes; leading action + menu |
| Switches | Plain and icon handles |
| Tabs | Primary and secondary; labels and icons |
| Text fields, selects, autocomplete | Filled and outlined; empty/populated, icons, affixes, errors, listbox popups |
| Time pickers | Clock and keyboard entry; portrait/landscape; 12-hour and 24-hour |
| Toolbars | Docked and floating; standard/vibrant; horizontal/vertical; optional FAB |
| Tooltips | Plain and rich |

Component pages (`/md3/components/<slug>`) show interactive playgrounds above
their token tables, controlled via `preview.*` query parameters (for example,
`?preview.compare=sizes&preview.variant=tonal`) while preserving table filter
parameters.

## Build

Run from the repository root:

```sh
npm run build:tokens -- --published published
TOKEN_ASSETS_DIR="$PWD/published/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published/tokens.sqlite" go test ./...
npm run test:ui
```

- `forge/material/components/component-contract.json` pins the reviewed token
  identities; `generate-components.mjs` maps them into CSS. Missing tokens,
  unmapped properties, or unconverted values fail the build.
- `forge/components/styles/` holds handwritten anatomy;
  `npm run build:components` assembles `forge/components/generated/components-base.css`.
  `components.css` combines that anatomy, shared shadow rules, and generated
  token bindings. Always edit source files, not generated outputs.
- The Go binary embeds component CSS and verifies its checksum against the
  publication asset manifest, so rebuild the binary and assets together.

## Inspect CSS sources

Every token build writes `<publication>/assets/css-usage.json`. To run the
scanner directly:

```sh
go run ./cmd/tokens forge inspect --assets published-review/assets --output /tmp/css-usage.json
```

Omit `--assets` to scan checked-in CSS without a publication. Each declaration
is classified as `Token` (all values from tokens), `Mixed` (tokens plus
handwritten values), `Manual` (handwritten only), or `Unresolved`, with
`manualValues`, `tokenValues`, `unknownVariables`, and active `overrides`.

## Override a token

Edit `forge/material/overrides.json` using source token names and DTCG values:

```json
{
  "md.comp.menus.group.padding": { "value": 4, "unit": "px" }
}
```

Then rebuild CSS, rebuild the binary, and restart:

```sh
npm run build:tokens -- --published published-review
go build -o ./bin/tokens ./cmd/tokens
./bin/tokens serve --published published-review --listen 127.0.0.1:8081 --public-origin http://127.0.0.1:8081/md3
```

Overrides apply to all six appearances and dependent aliases in generated site
CSS and CSS downloads without requiring a new crawl. A removed or renamed
upstream token referenced in `overrides.json` fails the build. Delete the entry
and rebuild to restore the upstream value.

## Shared markup, ripple, and renderer rules

```html
<a class="md-card md-outlined-card md-interactive-card" href="/components">
  Browse components
</a>
<button class="md-icon-button md-tonal-icon-button" type="button"
        data-size="small" data-shape="round" data-width="default"
        aria-label="Bookmark" aria-pressed="false" data-button-toggle>
  <!-- icon -->
</button>
<label class="md-filter-chip">
  <input class="tokens-visually-hidden" type="radio" name="theme" value="light">
  <span class="md-chip-face">Light</span>
</label>
```

- **Shared press & ripple:** One controller handles pointer/keyboard feedback,
  cancellation, and disabled suppression for buttons, icon buttons, interactive
  cards, chips, and menu items. Ripple geometry follows
  [Material Web](https://github.com/material-components/material-web/tree/5a95667bb70ca5be8785b762fb0303737c3b4ac1/ripple/internal)
  (450 ms expansion from 20%, 105 ms fade-in, 225 ms min press, 375 ms fade-out,
  15 ms hover, 150 ms touch delay) with captured colors and opacities. Reduced
  motion keeps static state feedback.
- **Shadows & scrims:** Elevation shadows follow Material Web's key/ambient
  formulas. Modal backdrops use neutral-black (`palette neutral0`) at 32%
  opacity.
- **Cards:** Add `md-interactive-card` only when the whole card is an action
  (`aria-disabled="true"` suppresses activation; `data-dragged="true"` applies
  dragged state). Layout uses 16dp horizontal padding and an 8dp grid gap.
- **Icon buttons:** `data-size` (`xsmall`–`xlarge`, default `small`),
  `data-shape` (`round`/`square`, default `round`), `data-width`
  (`default`/`narrow`/`wide`, default `default`). Pressed shapes use captured
  size-set springs.

## Rendered checks

- `forge/conformance/browser/check-component-rendering.mjs` compares computed
  styles against source CSS variables across the gallery (`checkBarRendering()`
  covers app bars and toolbars; `checkHandoffRendering()` checks `/md3/for-agents`
  contrast ≥ 4.5:1 across all six appearances).
- Start Catalog on port 8081 and run `go run ./forge/conformance`, then open:
  - `http://127.0.0.1:8082/feedback-transitions.html` — press, release, exit,
    and rapid-press ripple checks across all six appearances.
  - `http://127.0.0.1:8082/motion.html` — spring, reduced-motion, and popover
    fallback checks.

---

## Component notes

### Navigation bars, rails, and drawers
- Bars and rails use `md.comp.nav-bar` and `md.comp.nav-rail`; drawers use
  `md.comp.navigation-drawer`. Active items use `aria-current="page"`, Arrow /
  Home / End navigation, and filled vs. outlined Material Symbols.
- **Rails:** Collapsed, Expanded, or Hidden (immersive, with one persistent menu
  toggle). Standard expansion shifts body content; Modal expansion overlays a
  scoped scrim and traps focus inside the preview frame. Optional header slots
  hold menu/FAB controls; badges follow vertical or horizontal label layout.
  Missing section-header spacing uses 12px above / 8px below; inactive state
  opacity and focus rings fall back to system tokens.
- **Drawers:** Standard and Modal (native `<dialog>` with Escape, backdrop, and
  destination dismissal). Bottom drawers are excluded from the renderer.

### Checkboxes, switches, and radio buttons
- All three use native inputs (`role="switch"` on switches) activated with Space
  (Enter does not trigger pressed feedback). Forced-colors mode reveals native
  controls; reduced motion disables transitions.
- **Checkboxes & switches:** `buttons.js` initializes indeterminate checkboxes
  on load and HTMX swaps. Switch handles center half a track-height from each
  end and animate with the fast spatial spring.
- **Radio buttons:** 2px ring stroke and half-size center dot; center dot
  animates with fast spatial (geometry) and fast effects (opacity).

### FABs and extended FABs
- Use the current separate size and color token sets (default/small/medium/large;
  primary/secondary/tertiary and containers). No disabled variant (absent from
  current sets). Older combined, surface/branded, and lowered sets are excluded.

### Badges, snackbars, and tooltips
- **Badges:** Dot and labeled forms driven by captured badge tokens.
- **Snackbars:** Single- and two-line native popovers with a 6 s timeout
  (pauses on hover/focus; `data-snackbar-duration="0"` disables auto-dismiss;
  Escape closes).
- **Tooltips:** Plain tooltips open on hover/focus (Escape closes); rich
  tooltips use non-modal popovers with dialog semantics. Supplemental layout
  measurements:

| Measurement | Plain | Rich |
| --- | --- | --- |
| Container height | 24dp min (grows on wrap) | Fits content |
| Top / bottom padding | 4dp / 4dp | 12dp / 8dp |
| Left / right padding | 8dp | 16dp |
| Subhead → supporting / → action | — | 4dp / 12dp |

### Tabs and segmented buttons
- **Tabs:** Primary (content-width indicator) and secondary (full-width
  indicator), WAI-ARIA tabs pattern, fast spatial indicator motion.
- **Segmented buttons:** Native radios (single) or checkboxes (multi) with 48dp
  touch targets, equal-width segments, 12dp horizontal padding, 8dp gaps, and a
  reserved checkmark/icon slot so selection never resizes segments. Density
  steps (`default` 40dp, `−1` 36dp, `−2` 32dp, `−3` 28dp) subtract 4dp per step
  from visible height while keeping 48dp targets.

### Sliders
- Five sizes (XS–XL), horizontal/vertical, continuous/discrete, single/range,
  with optional inset icons on M/L/XL single sliders.
- Captured token dimensions take precedence over M3 diagram numbers where they
  differ (medium handle 44px vs 52dp, label height 28px vs 44dp, label gap 12px
  vs 4dp, stop inset 4px vs 6dp). Shared CSS supplies the missing 48px label
  width and 16px horizontal padding.
- Range sliders use native `<input type="range">` clamped against the neighbor
  with `aria-valuemin`/`aria-valuemax`. Focus/press shrink handles with fast
  spatial motion.

### Progress indicators
- Linear and circular (normal/thick, flat/wavy) plus baseline four-color
  indeterminate mode. Cycles follow AndroidX Material 3 (1,750 ms linear,
  6,000 ms circular) and pause offscreen or in hidden tabs.
- `ui.Progress` supports `CSSOnly: true` for thin, flat, indeterminate linear
  bars rendered in the common stylesheet.

### Loading indicators
- Contained and uncontained (`transparent` container following AndroidX),
  determinate (circle → Soft Burst) and indeterminate (7-shape AndroidX morph,
  650 ms interval, 4,666 ms rotation).
- `forge/components/data/loading-shapes.json` stores matched cubic curves
  exported from AndroidX Graphics Shapes 1.1.0. First frame renders in HTML;
  live previews fetch geometry once. To regenerate from
  `forge/material/reference/LoadingShapes.java` (requires JDK with records):

```sh
mkdir -p /tmp/loading-shapes
curl -fsSL https://dl.google.com/dl/android/maven2/androidx/graphics/graphics-shapes-android/1.1.0/graphics-shapes-android-1.1.0.aar -o /tmp/loading-shapes/shapes.aar
unzip -p /tmp/loading-shapes/shapes.aar classes.jar > /tmp/loading-shapes/shapes.jar
curl -fsSL https://repo.maven.apache.org/maven2/org/jetbrains/kotlin/kotlin-stdlib/2.0.21/kotlin-stdlib-2.0.21.jar -o /tmp/loading-shapes/kotlin.jar
curl -fsSL https://dl.google.com/dl/android/maven2/androidx/collection/collection-jvm/1.5.0/collection-jvm-1.5.0.jar -o /tmp/loading-shapes/collection.jar
java --class-path '/tmp/loading-shapes/*' forge/material/reference/LoadingShapes.java > forge/components/data/loading-shapes.json
npm run test:ui
```

### App bars and toolbars
- **App bars:** Small, centered, medium, large, flexible, bottom, and baseline
  variants with optional subtitle, avatar, and search. Avatars sit in a centered
  slot matching the small icon-button height. Scrollable previews apply on-scroll
  color and elevation.
- **Toolbars:** Docked and floating, standard/vibrant, horizontal/vertical,
  optional separate FAB (two sizes). Overflowing actions collapse into the
  shared menu; floating FAB hover/focus/press elevations use AndroidX system
  roles. See [toolbar spec](toolbar-spec.md) for detailed measurements.

### Bottom and side sheets
- **Bottom sheets:** Standard and modal (`<dialog>`), expanded (`min(24rem, 85svh)`,
  max modal width `40rem`) and minimized (`56px`), 56px drag threshold and
  125px/s fling threshold following AndroidX `SheetDefaults`. Dragging below
  half the minimized height dismisses.
- **Side sheets:** Standard (divider on page edge) and modal, docked and
  detached, start/end edges with RTL corner mirroring.

### Split buttons, button groups, and FAB menus
- **Split buttons:** Five sizes, four color variants. Leading action and
  trailing chevron respond independently; menu opens 4dp from the button in a
  native popover. Supplemental chevron offset toward leading button:
  1 / 1 / 2 / 3 / 6dp (XS–XL), rotating 180° inward with Standard fast spatial.
  Each half keeps a 48 × 48dp touch target.
- **Button groups:** Standard (pressed item expands by captured percentage,
  compressing immediate neighbors up to their smaller horizontal padding) and
  connected (single/multi selection, reserved checkmark width, transparent
  border on selected outlined buttons, 48px min width/target on XS/S).
- **FAB menus:** Primary/secondary/tertiary. One persistent toggle transitions
  between FAB and close-button profiles while actions open in a native popover,
  revealing bottom-up with fast spatial width, fast effects opacity, and slow
  effects stagger.

### Lists
- Standard and segmented, baseline and expressive, 1–3 lines, leading
  icon/avatar/image/video/checkbox/radio, optional dividers, and
  `data-surface="transparent"` for transparent unselected rows.
- Rows ≥ 88px top-align content (tracked via `ResizeObserver`).
- Interactive behaviors: single/multi selection, animated expansion (optional
  inset and tonal surface), drag or bottom-sheet reordering (Move up / Move
  down), and horizontal swipe or More-button action reveal.

### Carousels
- Multi-browse, uncontained, hero, and full-screen layouts; optional outlines
  (the source labels outline width as an elevation; rendered as pixels).
- `carousel.js` owns keylines and sticky visual tracks (preferred widths
  160/240/320px, heights 128/192/256px, 56px small window, 8px drag threshold).
- Native touch/trackpad momentum; mouse drag coasts with 250 ms exponential
  decay; `material:press` widens item masks by 5px per side (matching a 5px
  image overscan) over 350 ms (`medium3` + standard easing). Reduced motion
  disables mask expansion, parallax, coasting, and smooth scroll.

### Search
- Contained and divided styles; docked (360–720px wide, 240px min height,
  max 2/3 canvas height), full-screen, and adaptive (fills available width
  without switching to full-screen).
- Flat by default; optional Shadow control opts into captured elevation.
- Uses one persistent header surface; scrim covers only the bounded app canvas
  using modal-drawer scrim tokens. Motion uses fast spatial (docked), slow
  spatial (full-screen expand), default spatial (return), and fast effects
  (opacity).

### Date pickers
- Docked (native popover), modal calendar (`<dialog>`), modal text input, and
  full-screen range picker (Gregorian 1900–2100).
- Three upstream fields misspelled `opcaity` are checked against canonical
  `opacity` values during build.
- Docked calendar starts at the 6-week 456px height and subtracts 48px per
  unused week; separate month/year menus sit outside the scroll region.
- Text entry accepts ISO or US month-first dates and normalizes to `YYYY-MM-DD`;
  calendar icon or `Alt+ArrowDown` opens the picker. Keyboard navigation:
  Arrows (day/week), Home/End (week/month), PageUp/Down (month, +Shift for
  year), `Shift+M`/`Shift+Y` (month/year menus), Enter (commit), Escape (cancel).
- Modal input layout: 24px sides, 16px top, 36px supporting-text gap, 10px
  around header divider, 16px below field, 8px/12px around actions.

### Time pickers
- Portrait and landscape clocks plus keyboard entry (`HH:mm`), 12-hour and
  24-hour (outer 00–11, inner 12–23).
- Dial radii (outer 101, inner 69, ring boundary 74) scale to the 256px dial;
  clock hand takes the shortest rotation path. Arrows step by 1, PageUp/Down by
  5, Home/End to bounds. Disclosure uses fast effects; dial uses fast spatial.

### Chips
- Assist, suggestion, filter, and input (flat and elevated; `data-elevated="true"`
  for elevated filter chips). Outlines overlay the container so selection
  doesn't shift layout.
- Padding: 12px text, 8px icons/gaps, 4px avatar. Arrow keys move between
  primary and remove actions; Delete/Backspace removes.

### Text fields, selects, and autocomplete
- Filled and outlined profiles for all three controls. Filled fields use the
  captured 3px focused thickness for keyboard focus rings alongside the 2px
  active indicator; outlined fields use `<fieldset>` notches.
- Selects use read-only combobox inputs with Home/End and 500 ms prefix typing;
  autocomplete filters case-insensitively and allows free text. Popups use
  native popovers flipping above when space is tighter below.
- Note: The 15 Oct 2026 capture maps error-hover label/outline/icon colors to
  `md.sys.color.on-error-container`, which has low contrast in some medium/high
  contrast schemes—override in `forge/material/overrides.json` if needed.

### Dialogs
- Basic (`alertdialog`, 280–560px wide, 24px outer padding) and full-screen
  (`dialog`, scroll elevation on header and bottom action bar).
- `MaterialDialogs.open(panel, opener)` attaches scroll/resize listeners on open
  and releases them on close. `MaterialDialogs.initialize(root)` adopts
  already-open markup; `MaterialDialogs.dispose(root)` cleans up before DOM
  removal without emitting `md-dialog-closed`.
- Motion: fast effects for basic dialogs; slow spatial (enter) and default
  spatial (exit) for full-screen.

### Menus
- **Standard & Vibrant (`Common` set):** Vertical (supporting text, submenus)
  and horizontal (labeled and icon-only). Native popovers with Arrow / Home /
  End / Escape / Tab navigation; submenus open after 200 ms hover or Arrow Right
  (Arrow Left in RTL), flipping sides or replacing the parent with a Back item
  when space is tight. Main menu max width `20rem`; submenu min width `12rem`.
- **Baseline (`md.comp.menu`):** 4px radius, 48px rows, 112–280px width, 8px
  vertical padding, 12px row padding/gaps. Shares popover and keyboard behavior.

### Dividers
- Full-width and inset (16px start, optional 16px end for middle-inset;
  supporting text 4px below). Optional 8px right/bottom outer margins are off by
  default.

---

## Motion schemes and lifecycle

Catalog selects `Expressive` in `forge/material/site-theme.json` for the site
theme; public DTCG/CSS downloads default to `Standard`.

`forge/components/generated/motion.cjs` (served as `motion.js` before component
scripts) provides the shared unit-mass spring model. The CSS build samples it
into `--_motion-*-duration` and `--_motion-*-easing` variables used by both CSS
and Web Animations.

| Property or component | Assignment |
| --- | --- |
| Menu reveal, submenu corners, switch, radio dot, slider bubble, field label, tab indicator, split-button shape | Fast spatial |
| Split-button chevron rotation | Standard fast spatial |
| Small opacity/color changes, basic dialogs, date/time disclosure | Fast effects |
| Sheet movement/settling, list expansion/reorder, expanded rail geometry | Default spatial |
| Rail content/scrim effects, sheet scrim | Default effects |
| Full-screen search expand, full-screen dialog entry | Slow spatial |
| Full-screen search return, full-screen dialog exit | Default spatial |
| FAB-menu item width / opacity / stagger | Fast spatial / fast effects / slow effects |

- **Interruption:** Buttons, button groups, and sheets carry velocity into
  retargeted springs; closing pickers/dialogs start from their current rendered
  frame and become inert.
- **Popover exit fallback:** Browsers with native overlay transitions animate
  menus and FAB menus in place; browsers without overlay support use an inert,
  `aria-hidden` visual clone reading a fixed set of geometry/paint properties
  and cleaned up on completion, reopen, reduced motion, or timeout.
