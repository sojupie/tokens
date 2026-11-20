# Component preview scenes and thumbnails

How Catalog frames component index thumbnails (`PhoneScene`, `PhoneThumbnail`,
`ThumbnailWindow`) and planned 1200 × 630 banner scenes. Shared layout grids,
margins, spacing variables, and title ownership rules live in
[layout](layout.md); Go/templ primitive signatures live in
[components API](components-api.md).

## Catalog thumbnails (implemented)

Catalog index cards display captured 720 × 450 thumbnails (plus 360 × 225 and
inline 40 × 25 WebP placeholders) across all six appearances.

### Composition rules
- Make the featured Forge component recognizable at tile size on Catalog's
  tinted backdrop. Use a standalone specimen when appearance alone explains the
  component, or add minimal app context when placement against a screen edge or
  list matters.
- Use actual Forge primitives (`forge/components/ui`) in every slot; never fake
  a component by putting `md-*` classes on partial HTML or overriding primitive
  padding/targets to force a fit.
- Give any visible title one owner (`ui.AppBar`), never a duplicate body `<h1>`.

### `PhoneScene` and `PhoneThumbnail`
`PhoneScene` (`catalog/web/templates/previews/phone_scene.templ`) lays out a
reference 390 × 844dp phone (4dp bezel, 48dp top safety inset, 24dp bottom
safety inset, status marks 30dp below the top edge, 4 columns with 16px margins
and 8px gutters):

- **Slots:** Children fill the content pane; `TopBar` and `BottomBar` hold
  docked bars (`BottomBarClass` extends a bottom bar through the gesture inset);
  `BottomRail` holds floating/inset controls above `BottomBar`; `Overlay` holds
  surfaces like bottom sheets (use `--phone-gesture-inset` to clear the gesture
  bar); `PaneAtBottom` anchors pane content to the bottom bar/rail.
- **Ruler:** Expanded App Bar titles, outer small action glyph boxes, and pane
  edges share the 16dp inline ruler; use `--layout-span` for items narrower than
  4 columns.

Wrap `PhoneScene` in `PhoneThumbnail` to crop and scale the 360 × 225 CSS-pixel
artboard (rendered at 2× for 720 × 450):

| Crop | Focus | Scale |
| --- | --- | ---: |
| `PhoneCropTop`, `PhoneCropBottom` | Top or bottom half of the phone | 0.68 |
| `PhoneCropDetail` | Compact controls near the top | 0.80 |
| `PhoneCropAction` | Single action in trailing bottom corner | 0.92 |
| `PhoneCropActionMenu` | Expanded action stack in trailing bottom corner | 0.77 |
| `PhoneCropCenter`, `PhoneCropModal`, `PhoneCropModalUpper` | Centered dialogs and modal pickers | 0.68 |

Use `ThumbnailWindow` for wide-edge components (docked drawers, rails, side
sheets) in a 640 × 480dp frame with medium-width margins. Progress and Loading
render directly on the shared tonal backdrop.

### Capture pipeline
- `/illustrations/render/<subject-path>?composition=thumbnail` renders
  `catalog/web/templates/previews/illustrations.templ`.
- `illustration-ready.js` waits for fonts, images, loading geometry, and motion
  sampling before setting `data-scene-ready`.
- Run `go run ./cmd/tokens illustrations --published <dir> --chrome-path <path>`
  after building token assets to write `assets/illustrations/manifest.json` (see
  [development](../development.md#component-illustrations)).

## Rich banner scenes (planned, 1200 × 630)

`SubjectIllustration` and the manifest reserve space for 1200 × 630 banners,
though `tokens illustrations` currently captures thumbnails only. When adding
banner scenes:

- Target a real viewport first—Compact (<600dp, e.g. 412 × 915 with
  `ui.NavigationBar`), Medium (600–839dp, e.g. 700 × 800 with
  `ui.NavigationRail`), or Expanded (840dp+, e.g. 1024 × 640 or 1280 × 800 with
  dual panes)—then frame it inside the 1200 × 630 canvas without distorting
  component scale.
- Place `.tokens-layout` on the simulated device content region so container
  queries measure the app window rather than the outer banner canvas.
- Choose a task-appropriate archetype: **Feed** (1–4 card columns),
  **List-Detail** (side-by-side panes from 840px), or **Supporting Pane**
  (focus pane + inspector/sheet), anchored by a search bar, chips, or active
  rail indicator rather than generic "Home"/"Inbox" headings.
- Keep scene templates under `catalog/web/templates/previews/` inside
  `webview.ComponentAssetScope()` so `modules.json` resolves CSS and scripts
  automatically.

## Review and checks

```sh
npm run generate
npm run check:components
go test ./catalog/web/... ./forge/components/...
```

Inspect representative tiles side by side at desktop and mobile widths in light
and dark appearances before regenerating the full set.
