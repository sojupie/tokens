# Rich component preview scenes

Build complete application screens for component-page banners. Each scene
shows components collaborating around a realistic task: an inbox, an itinerary,
a project workspace or a media library. Give the featured component an authentic
role supported by credible data, selections, and interactive states. A row of
isolated specimens or one control floating in an empty frame is not an
application scene.

## Core scene philosophy: purpose over titles

Real Google products and Material 3 canonical applications (such as Gmail,
Drive, Photos, Keep and Maps) rarely display static page titles like "Inbox",
"Files" or "Home". The purpose of the screen is established naturally through:

- **The top anchor**: A functional search bar (`ui.SearchBar`), filter chips
  (`ui.Chip`), or tab strip anchors the top of the content area.
- **Navigation state**: The active indicator pill on the navigation rail
  (`ui.NavigationRail`) or navigation bar (`ui.NavigationBar`) shows the user's
  location without consuming vertical space for a text label.
- **Content semantics**: Card headlines, list items, maps, and media preview
  surfaces immediately communicate what the task is.

When an app screen or detail pane does present a title (for example, the subject
of an email, the name of a document, or an itinerary stop), that title belongs
in its bar primitive (`ui.AppBar`), never as a duplicate `<h1>` in the body.

## Scaffold architecture: bars, rails and panes

Material 3 organizes adaptive application layouts through three structural
regions:

```
┌────────────────────────────────────────────────────────────────────────┐
│ Status bar (time, Wi-Fi, battery)                                      │
├────┬───────────────────────────────────────────────────────────────────┤
│    │ Top bar (ui.SearchBar or filter chips; optional/pane-scoped)     │
│ R  ├─────────────────────────────┬─────────────────────────────────────┤
│ A  │ Pane 1 (e.g. List / Feed)   │ Pane 2 (e.g. Detail / Inspector)    │
│ I  │                             │                                     │
│ L  │ - Active items              │ - Content card / media              │
│    │ - Metadata & chips          │ - Pane action bar                   │
│    │                             │                       [ui.FAB]      │
├────┴─────────────────────────────┴─────────────────────────────────────┤
│ Navigation bar (Compact viewports) / Gesture bar                       │
└────────────────────────────────────────────────────────────────────────┘
```

### Rails
- Dock along the leading perimeter edge beside content panes (`ui.NavigationRail`).
- Include an anchor icon (such as `ui.RailToggle` for a menu) and navigation
  destinations with active indicator pills (`Current: true`).
- Rails always sit beside content panes, never underneath a header bar.

### Bars
- Bars are modular and optional; they are not required on every screen or pane.
- **Global search anchor**: A `ui.SearchBar` with leading menu/back icon and
  trailing avatar often serves as the single top anchor across panes.
- **Pane-scoped bars**: In multi-pane layouts, a top app bar (`ui.AppBar`) or
  action toolbar can belong exclusively to a detail pane, leaving the list or
  feed pane flush.
- If a bar carries a title or headline, it owns it exclusively.

### Panes
- Panes are content surfaces (`.tokens-pane`, `ui.CardSurface`).
- Use single-pane layouts for Compact viewports; use dual-pane structures for
  Medium and Expanded viewports.
- Panes use `--layout-pane-gap: var(--space-300)` for separation and contain
  real Forge primitives.

## Canonical layout archetypes

Choose a canonical pattern suited to the task rather than defaulting every
screen to identical columns:

| Pattern | Breakpoint behavior | Typical tasks | Key primitives |
| --- | --- | --- | --- |
| **Feed** | 1 column below 600px; 2–4 columns from 600px | Media library, card discovery, catalog | `ui.SearchBar`, `ui.Chip`, `ui.Card`, `ui.FAB` |
| **List-Detail** | Single pane with swap on compact; side-by-side panes from 840px | Inbox, itinerary planner, issue backlog | `ui.NavigationRail`, `ui.SearchBar`, `ui.List`, `ui.Card`, `ui.Action`, `ui.FAB` |
| **Supporting Pane** | Dominant focus pane with auxiliary inspector or side sheet | Document editor, media player with queue | `ui.NavigationRail`, `ui.AppBar`, `ui.Sheet`, `ui.SegmentedButtons` |

## Viewport selection and device framing

1. **Target the application viewport first**:
   - **Compact (<600dp)**: e.g. 412 × 915 phone with `ui.NavigationBar`.
   - **Medium (600–839dp)**: e.g. 700 × 800 foldable with `ui.NavigationRail` and single wide pane.
   - **Expanded (840dp+)**: e.g. 1024 × 640 or 1280 × 800 tablet/desktop with `ui.NavigationRail` and dual panes.
2. **Frame for the 1200 × 630 export banner**:
   - 1200 × 630 is the image delivery size, not the app's browser window.
   - Frame the device viewport against an illustration backdrop with realistic
     bezels, device corners, status indicators (9:30, network, battery), and
     system gesture bars.
   - Preserve component proportions, touch targets, and typography tokens;
     never scale or distort components to fill the frame.

## Use Forge primitives

Import `github.com/sojupie/tokens/forge/components/ui` in Go/templ. Use typed
props, children and slots; do not imitate a component by putting its class on
partial HTML. Ordinary HTML and scene CSS are appropriate for application
content and layout. Forge owns control anatomy, state, typography, token
bindings and motion.

This example illustrates a canonical List-Detail scene in an Expanded viewport:

```templ
package render

import "github.com/sojupie/tokens/forge/components/ui"

templ TripPlannerScene() {
    <div class="scene-device-frame" data-viewport="expanded">
        <aside class="scene-rail">
            @ui.NavigationRail(ui.NavigationRailProps{Variant: "standard"}) {
                @ui.RailToggle(false, templ.Attributes{"aria-label": "Menu"})
                @ui.NavigationItem(ui.NavigationItemProps{Label: "Trips", Current: true})
                @ui.NavigationItem(ui.NavigationItemProps{Label: "Saved", Current: false})
                @ui.NavigationItem(ui.NavigationItemProps{Label: "Settings", Current: false})
            }
        </aside>
        <div class="scene-content-region tokens-layout">
            <header class="scene-search-bar">
                @ui.SearchBar(ui.SearchProps{Placeholder: "Search itinerary..."}, "Search", "", false)
            </header>
            <main class="scene-panes tokens-layout-content">
                <section class="scene-pane-list">
                    <div class="scene-chips">
                        @ui.Chip(ui.ChipProps{Label: "All stops", Selected: true})
                        @ui.Chip(ui.ChipProps{Label: "Transit"})
                    </div>
                    <!-- Stop items with active selection state -->
                </section>
                <section class="scene-pane-detail">
                    @ui.Card(ui.CardProps{
                        Variant: "outlined",
                        Headline: "Riverwalk & Station",
                        Supporting: "Stop 1 · 10:00–11:30 · Guide provided",
                        Actions: TripDetailActions(),
                    }) {
                        <p>Meet at North Gate, then proceed along the river trail.</p>
                    }
                    @ui.FAB(ui.FABProps{Icon: "add", Label: "Add stop"})
                </section>
            </main>
        </div>
    </div>
}

templ TripDetailActions() {
    @ui.Action(ui.ButtonProps{Variant: ui.Filled, Color: ui.ColorPrimary}) {
        Review stop
    }
}
```

Use `ButtonLink` and `CardLink` for navigation, `Action` for commands, and `Card`
for content with independent actions. Never nest buttons inside a link card.
Fields and panels need unique IDs, including when two scenes share a page.
Props such as `Leading`, `Actions` and `Search` accept `templ.Component` slots.
Use the component's real selection, checked, disabled and open-state props.

Read [the component API](components-api.md) for fields, lists, dialogs, menus,
sheets and asset loading. Inspect these source locations for exact signatures:

| Location | Purpose |
| --- | --- |
| `forge/components/ui/api_generated.go` | Public functions, props aliases and constants; read only |
| `forge/components/ui/internal/model/` | Prop definitions, defaults and attribute rules |
| `forge/components/ui/templates/` | Component anatomy and slot placement |
| `catalog/web/templates/playground/` | Existing compositions and interaction wiring |

Primitives declare module dependencies during rendering. In Catalog, keep scene
markup inside `webview.ComponentAssetScope()` under the normal server render
context. The host resolves CSS and behavior through `modules.json`. Do not hand
maintain a second list of component scripts. A standalone host must collect
`ui.Assets` through `ui.WithAssets` and resolve those modules as described in the
API guide. Collection alone does not load styles or initialize controllers.

## Layout, grids and spacing

Follow [the layout contract](layout.md). Start with the application's intended
phone, tablet or desktop viewport, then frame or crop it for the banner. Place
`.tokens-layout` on the device frame's content region so container queries
measure the simulated application window, not the outer 1200 × 630 banner canvas.

### Spacing tiers in scenes

Apply layout variables from `forge/site/site.css` to arrange panes and content:

- **Pane separation**: Use `--layout-pane-gap: var(--space-300)` (24px) between
  side-by-side list and detail panes.
- **Content insets**: Use `--layout-margin` (16px compact / 24px expanded) for
  outer pane padding.
- **Related elements**: Use `--layout-related-gap: var(--space-100)` (8px)
  between filter chips, adjacent cards, and list rows.
- **Grid columns**: Use `.tokens-layout-grid` with `--layout-gutter: var(--space-100)` (8px)
  when arranging card feeds.

### Component integrity and density

- Align elements to the 8dp baseline grid (4dp for compact controls).
- Preserve 48 × 48dp minimum accessible touch targets.
- Never override primitive padding, typography, or touch bounds to force a
  scene fit; adjust the screen composition or pane distribution instead.
- Use semantic color, shape and typography tokens. Check light/dark themes and all
  three contrasts. Keep sample data, counts and selected states consistent.

## Files and capture integration

Put maintained scene templates in a dedicated directory under
`catalog/web/templates/previews/`, with scene layout CSS in
`catalog/web/assets/styles/`. Edit templates and regenerate; never edit
`catalog/web/internal/render/` or `forge/components/generated/` directly.

The current pipeline captures **simple card thumbnails only**:

- `catalog/web/internal/server/illustrations.go` accepts
  `/illustrations/render/<subject-path>?composition=thumbnail`.
- `catalog/web/templates/previews/illustrations.templ` renders that document.
- `cmd/tokens/illustrations.go` captures 720 × 450 thumbnails in six appearances,
  plus smaller images and placeholders. It does not currently capture banners.
- `SubjectIllustration` and the manifest's `Files` map retain support for
  1200 × 630 images, but the capture command does not populate them.

Add a separate scene composition and banner capture path when integrating the
new scenes. Preserve the thumbnail composition and existing public URLs. Keep
catalog cards image-based; do not replace their images with live scene markup.
Use the manifest and content-addressed image delivery rather than hardcoded
image URLs. Captures and publication assets remain generated, uncommitted output.

Reuse `illustration-ready.js` for capture readiness: it waits for fonts and images,
waits for loading geometry, samples motion and sets `data-scene-ready`. Do not
replace readiness with an arbitrary screenshot delay.

## Review and checks

Start with one rich scene and inspect it before multiplying compositions. Reuse
port **8085** for local review. Check the real render route, framing, text,
clipping and component behavior; successful capture is not visual acceptance.

After source changes, run from the repository root:

```sh
npm run generate
npm run check:components
go test ./catalog/web/... ./forge/components/...
```

Rebuild/restart the server after changing embedded assets or templates. Run
`npm run test:ui` if controller wiring changes. Inspect the intended banner and
app viewport, plus relevant responsive widths. Generate final images only after
reviewing the HTML composition. See [development commands](../development.md#component-illustrations)
for the existing thumbnail command; banner generation needs the integration above.
