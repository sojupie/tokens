# Site and scene layout

`forge/site/site.css` owns the shared layout measurements. Catalog pages and
scene windows use this contract. Forge component styles continue to own component
anatomy, typography, padding, states and target sizes.

## Window and content region

Material window breakpoints describe the complete application window:

| Window | Width | Reference columns |
| --- | --- | --- |
| Compact | Below 600dp | 4 |
| Medium | 600–839dp | 8 |
| Expanded | 840–1199dp | 8 |
| Large | 1200–1599dp | 12 |
| Extra-large | 1600dp and above | 12 |

The site's content region excludes the docked drawer. A scene window excludes
its illustration backdrop. `.tokens-layout` measures that region using a named
CSS container. Its descendants adapt to the region's available width, rather
than assuming the browser viewport is the width available to them. Region
thresholds reuse 600, 840 and 1200 CSS px; this does not change the application's
window breakpoint or automatically choose a navigation component.

The site drawer's docking threshold remains an explicit site behavior in
`forge/site/site.go`. It is not another Material window breakpoint.

## Shared measurements

Place `.tokens-layout-content` inside `.tokens-layout`. It supplies:

| Property | Value | Purpose |
| --- | --- | --- |
| `--layout-columns` | 4 below 600px; 8 from 600px; 12 from 1200px | Available-region grid |
| `--layout-margin` | `space-200` (16px), then `space-300` (24px) from 600px | Content inset |
| `--layout-gutter` | `space-100` (8px) | Local site/scene grid choice |
| `--layout-related-gap` | `space-100` (8px) | Related cards and items |
| `--layout-pane-gap` | `space-300` (24px) | Separation between content panes |
| `--layout-section-gap` | `space-400` (32px) | Default section separation |

Gutters and section spacing above are our defaults, not universal Material
requirements. A different context can use another captured spacing value.
State the reason beside that layout rule. Do not change component internals to
make a composition fit.

Use `.tokens-layout-grid` for a column grid. Direct children span
`--layout-span` columns, or the full grid when no span is supplied. Keep every
track shrinkable and each grid child at `min-inline-size: 0`.

Catalog card feeds use one card per row below 600px, two from 600px, and four
from 840px. Tool destinations use one below 840px and two above it. These are
content-specific span decisions in `catalog/web/static/app.css`; column count
does not dictate the number of cards or panes.

## Screen structure

1. Identify the task, visible state and information needed to perform it.
2. Choose the application window before choosing the exported image size.
3. Allocate navigation, app-bar and safety regions before content panes. Put
   rails and persistent drawers beside the content, not beneath its header.
4. Choose feed, list-detail, supporting-pane or another task-appropriate layout.
   Do not default every screen to two columns.
5. In app screens, put the screen title in an appropriate app-bar primitive.
   Flexible app bars own their larger title treatment. Website pages can use a
   content heading beneath site navigation, as Catalog does. In either case,
   give the page title one owner; a second heading must name different content.
6. Align content edges and related controls to shared rulers. Distinguish page
   margins, pane spacing, grid gutters and component padding.
7. Populate the screen with meaningful content and states. Counts, selections,
   menus and actions must agree. Use images where the task needs imagery.
8. Render actual primitives. Use their markup/slots and token bindings together;
   copying a class name onto partial markup does not implement the component.

A grid does not replace component measurements. Use the component's captured
anatomy and typography. Preserve accessible targets; change composition before
shrinking controls. Use restrained titles and labels that identify actual tasks.

## Scene framing

Keep the simple catalog thumbnail and the complete app scene as separate
compositions. Start large scenes with the intended phone, tablet or desktop
viewport. Preserve its geometry when framing it. The 1200×630 social-image export
is an output format, not the application's viewport.

Review the six applications at `/illustrations/review`. Mail, learning, artwork,
workspace and shop start at 1440×1000; museum starts at 412×892. The window-width
selector changes the actual iframe viewport. Outside scaling only fits that
viewport into the review page. Theme, contrast and column-overlay controls apply
to loaded frames and frames loaded later. The older component-by-component review
remains at `/illustrations/review/components`.

`catalog/web/application_scenes.go` owns the application/state inventory and
validated query selections. `application_scenes.templ` owns composition;
`static/application-scenes.css` allocates panes and content regions. Feeds use
`.tokens-layout-grid`, not a separately painted approximation of the overlay.
`document_screen.templ` provides the shared document composition. Forge surfaces
and controllers own controls, menu positioning, modal focus, sizing and behavior.
Do not place a FAB in a drawer or in a sibling area styled as part of that drawer.
Mail uses an expanded rail with its supported FAB slot; the workspace drawer has
navigation only. Keep playground wrappers and specimen controls out of component
surfaces. A successful render is not visual acceptance.

Review markup before regenerating illustration images. The existing published
images remain separate and are not rebuilt by visiting the review page.

## Verification

Check the real catalog and scene render routes, including widths on both sides
of 600, 840 and 1200px. Also check that a docked drawer reduces available content
width without causing overflow. Verify computed columns, margins and gaps;
inspect screenshots for hierarchy, clipping and density. Confirm keyboard focus
and modal/menu placement after changing the shell.

Regenerate affected scene images after adopting layout changes. Do not treat
successful rendering or an image count as visual acceptance.

Sources:

- [Material grids](https://m3.material.io/foundations/layout/grids-spacing/grids)
- [Material spacing](https://m3.material.io/foundations/layout/grids-spacing/spacing)
- [Material density](https://m3.material.io/foundations/layout/grids-spacing/density)
- [Material breakpoints](https://m3.material.io/foundations/layout/breakpoints/overview)
- [Canonical layouts](https://m3.material.io/foundations/layout/canonical-examples/overview)
- [App bars](https://m3.material.io/components/app-bars/overview)
