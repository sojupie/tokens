# Rich component preview scenes

Build a few complete application screens for component-page banners. Each scene
should show components working together around a real task: a schedule, an inbox,
a project workspace or a media library. Give the featured component a clear role.
Use enough supporting content and states to make the screen credible. A row of
specimens or one control in an empty frame is not an application scene.

## Use Forge primitives

Import `github.com/sojupie/tokens/forge/components/ui` in Go/templ. Use typed props,
children and slots; do not imitate a component by putting its class on partial
HTML. Ordinary HTML and scene CSS are appropriate for application content and
layout. Forge owns control anatomy, state, typography, token bindings and motion.

This illustrates the API, not the finished scene composition:

```templ
package render

import "github.com/sojupie/tokens/forge/components/ui"

templ ReviewActions() {
    @ui.Action(ui.ButtonProps{Variant: ui.Filled, Color: ui.ColorTertiary}) {
        Review plan
    }
}

templ PlanScene() {
    @ui.AppBar(ui.AppBarProps{Title: "Weekend plans", Variant: "small"})
    @ui.Card(ui.CardProps{
        Variant: "outlined", Headline: "Saturday by the river",
        Supporting: "Three stops · 10:00–14:00", Actions: ReviewActions(),
    }) {
        <p>Meet at the station, then walk to the exhibition.</p>
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

## Layout and appearance

Follow [the layout contract](layout.md). Start with the application's intended
phone, tablet or desktop viewport, then frame or crop it for the banner.
1200 × 630 is the banner output size, not necessarily the app viewport. Preserve
component proportions and target sizes. Put the screen title in its app bar.

Use shared region grids and spacing from `forge/site/site.css`. Scene-specific CSS
should arrange panes and content, without overriding primitive internals. Use
semantic color, shape and typography tokens. Check light/dark themes and all
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
