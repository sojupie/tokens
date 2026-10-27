# Generated Material Design 3 components

Forge generates the remaining component families used by the catalog from the
captured Material token sets. Common buttons have their own
[reference](buttons.md).

| Family | Implemented configurations |
| --- | --- |
| Cards | Elevated, filled, and outlined; enabled, hover, focus, press, drag, and disabled states |
| Icon buttons | Standard, filled, tonal, and outlined; five sizes, two shapes, three widths, selection, and five interaction states |
| Chips | Assist, suggestion, filter and input; icons and avatars; selection, removal, flat/elevated profiles and six interaction states |
| Radio buttons | Selected and unselected; enabled, hover, focus, press, and disabled states |
| Text fields | Filled and outlined; empty and populated; icons and affixes; normal, error, hover, focus and disabled states |
| Selects and autocomplete | Filled and outlined fields with token-styled listbox popups, filtering and keyboard selection |
| Menus | Baseline vertical menus; standard/vibrant vertical, horizontal and icon-only layouts; supporting text, submenus and interaction states |
| Dialogs | Basic and full-screen; headlines, supporting text, icons, dividers, actions and scroll states |
| Dividers | Color and thickness |
| Navigation bars | Horizontal and vertical items; selection, hover, focus, press, and keyboard navigation |
| Navigation rails | Collapsed and expanded; standard or modal expansion, bounded app layout, optional header and badges, top or center destinations |
| Navigation drawers | Standard and modal; two sections, selection, states, keyboard navigation, and native dialog dismissal |
| Checkboxes | Selected, unselected and indeterminate; pointer and keyboard feedback; disabled states |
| Switches | Plain and icon handles; selection, pointer and keyboard feedback; disabled states |
| Floating action buttons | Color variants, sizes, selection and interaction states |
| Extended FABs | Icon and label layouts, color variants, sizes and interaction states |
| Split buttons | Leading action, trailing menu, sizes, colors and interaction states |
| Button groups | Standard and connected; single and multiple selection; sizes and interaction states |
| FAB menus | Primary, secondary and tertiary profiles; action and close-button states; keyboard navigation |
| Badges | Small and large; label and color tokens |
| Snackbars | One and two lines; actions, dismissal and pausable timeout |
| Tooltips | Plain and rich; focus, hover, dismissal and rich-tooltip actions |
| Tabs | Primary and secondary; icon and label layouts, selection and keyboard navigation |
| Segmented buttons | Single and multiple selection, icons, labels and interaction states |
| Sliders | Continuous and discrete; single and range; horizontal and vertical; sizes and value indicators |
| Progress indicators | Current and baseline; linear and circular; determinate and indeterminate; flat, wave and four-color profiles |
| Loading indicators | Contained and uncontained; determinate and indeterminate shape morphs |
| App bars | Current and baseline; sizes, subtitles, actions and scroll states |
| Toolbars | Docked and floating; standard and vibrant; horizontal and vertical layouts |
| Bottom sheets | Standard and modal; expanded and minimized; drag, resize and dismissal |
| Side sheets | Standard and modal; docked and detached; start and end edges |
| Lists | Standard and segmented; selection, expansion, reorder and reveal; text and media variants |
| Carousels | Multi-browse, uncontained, hero and full-screen; keyboard navigation and outlined items |
| Search | Baseline and contained; docked and full-screen; filtering and keyboard selection |
| Date pickers | Modal, docked and range; calendar and text entry; month and year selection |
| Time pickers | Clock and keyboard entry; portrait and landscape; 12-hour and 24-hour formats |

The generated `components-coverage.json` lists source bindings, exclusions, and supported configurations.
Component pages such as `/md3/components/cards`, `/md3/components/menus`, and
`/md3/components/text-fields` contain live previews above their token tables.

The homepage thumbnails have separate compositions in
`catalog/web/component_thumbnails.templ`. Their fixed specimen choices are in
`catalog/web/component_thumbnails.go`; app frames and illustration layout are in
`catalog/web/static/app.css`. Reuse the shared component anatomy and generated
styles. Scale the whole scene or crop the app frame to fit, keeping component
proportions intact. Thumbnails are inert and share one secondary-palette backdrop. An SVG viewBox
scales each HTML scene uniformly through a foreignObject, without CSS zoom or
browser-specific sizing calculations.

Component pages open with Variants selected when several variants are available,
and States otherwise. Progress and loading indicators default to Indeterminate.

Choose States, Sizes, or Variants to vary one dimension across the preview grid.
The remaining controls fix the other dimensions. Interactive renders one specimen
without a forced interaction state. Only supported dimensions appear. Text fields
include a control selector for text fields, selects and autocomplete.

Preview settings use `preview.*` query parameters, for example
`/md3/components/icon-buttons?preview.compare=sizes&preview.variant=tonal`.
The form preserves token-table query parameters. Theme and contrast follow the
site controls. Unsupported component families retain their token tables without
a preview. New renderers use the same configuration list and preview container.

Navigation bars and rails use the current separate `md.comp.nav-bar` and
`md.comp.nav-rail` sets. Their older combined sets remain available as data.
The drawer uses `md.comp.navigation-drawer`.

Navigation item selection uses `aria-current="page"`. Arrow keys, Home and End
move focus; Space or Enter activates the destination. Outlined and filled
Material Symbols distinguish inactive and active destinations. Menu keyboard
shortcuts use the Material Symbols command-key SVG beside a normal text letter,
sized by the captured trailing-supporting-text typography.

Rail variants are Collapsed and Expanded. Standard or Modal is an independent
expanded-layout configuration. Interactive previews use a bounded app frame:
standard expansion moves body content, while modal expansion overlays it with a
scrim. The simulated app body becomes inert, focus stays within the expanded
surface, and Escape, the scrim, the menu button or destination activation
collapses it. Controls elsewhere on the catalog page remain available; this is
a scoped preview, not a document-wide native modal. Collapsing can restore the
collapsed rail or hide the expanded rail until its app menu is used again.

Rail headers optionally contain a menu, a FAB, both or neither. Menus change to
the collapse icon when expanded. The nested FAB has level-zero elevation and
becomes an extended FAB. Four primary destinations remain visible when
collapsed; expanded rails reveal a secondary section. Destinations can be top
or center aligned while header controls stay at the top. Badges are optional;
their position follows vertical or horizontal label layout. The rail itself
always remains vertical and sits outside the simulated page pane. Page content
scrolls independently. Compact-app navigation-bar substitution is application
layout behavior and is not simulated by these explicit rail specimens.

Expanded destination targets span the full rail width; their indicators hug
contents by default. The Expanded indicator control can fill the available
container instead, as an explicit renderer configuration. Labels sit below icons
when collapsed and beside icons when expanded. Expansion keeps the inner layout fixed and moves
contents between their measured positions while the outer width animates; rail
scrollbars are suppressed during that transition. Menu glyphs rotate and fade
within one fixed icon slot. Captured rail widths, bounds, header gap and top spacing remain in
use. The 6px container vertical-space token supplies vertical padding; the
separate 4px vertical icon-label-space token supplies that gap. The default and
short height tokens remain minimum target heights; measured label height can
increase the target to avoid clipping scaled text. Missing section-header
spacing follows the measurement diagram's 12px above and 8px below. These layout
rules and the app frame appear in the generated CSS usage report.

Drawers retain Standard and Modal configurations. Bottom is no longer offered;
its captured tokens remain in exports and are marked excluded from the renderer.
The sample has eight destinations across Library and Workspace sections, one
section divider and no footer close button. Modal drawers use a native dialog;
Escape, outside click or selecting a destination dismisses it.

The source has no inactive bar/rail state opacity, so those states use system
hover, focus, and press opacity. The scoped modal scrim uses the existing drawer
scrim tokens. Drawer content padding uses system spacing because its source set
supplies no item content padding. Bar/rail focus rings use system thickness and
offset where component fields are absent.

Rail guidance: [specs](https://m3.material.io/components/navigation-rail/specs),
[guidelines](https://m3.material.io/components/navigation-rail/guidelines) and
[accessibility](https://m3.material.io/components/navigation-rail/accessibility).

## Inspect CSS sources

Every token build runs the Go scanner and writes
`<publication>/assets/css-usage.json`. To generate the JSON report separately,
run from the repository root:

```sh
go run ./cmd/tokens forge inspect --assets published-review/assets --output /tmp/css-usage.json
```

Replace `published-review` with your publication directory. Omit `--assets` to
scan the checked-in CSS without a publication. The command prints category
counts; the JSON contains each declaration and its definitions:

- `manualValues`: handwritten values, with their defining selector, file and line.
- `tokenValues`: source token values and the theme selectors that use them.
- `unknownVariables`: variables whose definitions could not be traced.

`kind` is `Manual` when no token reference was found, `Mixed` when both token
references and handwritten values were found, and `Token` when all traced
values came from tokens. Unresolved variables produce `Unresolved`.

The scanner follows CSS variable chains across definitions for all variants and
states. Check the selector against the rendered element to identify the active
rule. The report does not inspect JavaScript behavior or component anatomy.

## Override a token

Edit `forge/material/overrides.json`. Keys are source token
names and values use DTCG types. For example, to set menu group padding to 4 px:

```json
{
  "md.comp.menus.group.padding": { "value": 4, "unit": "px" }
}
```

Rebuild the CSS and server from the repository root:

```sh
npm run build:tokens -- --published published-review
go build -o ./bin/tokens ./cmd/tokens
```

Stop the running local server, then start the rebuilt binary:

```sh
./bin/tokens serve --published published-review --listen 127.0.0.1:8081 --public-origin http://127.0.0.1:8081/md3
```

Refresh the lab after the server starts. The JSON report lists active overrides
under `overrides`, with upstream and effective values for each appearance. The
binary embeds component CSS, so CSS changes require both a rebuild and a restart.

Overrides apply to all six appearances and aliases that depend on the overridden
value. Generated site CSS and CSS downloads use the override. Captures and
exported source DTCG retain upstream values. A removed or renamed override token
fails the build.

Delete an entry and repeat the build and restart to restore its upstream value.
Routine token updates continue to use new upstream values for all tokens without
an override. No new crawl is needed to apply or remove an override.

## Build

Run commands from the repository root:

```sh
npm run build:tokens -- --published published
TOKEN_ASSETS_DIR="$PWD/published/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published/tokens.sqlite" go test ./...
npm run test:ui
```

`forge/material/component-contract.json` records the reviewed source-token
identities. `generate-components.mjs` maps their anatomy and states into CSS
properties. A missing token, new property, unknown icon-button or menu set, or
unconverted mapped value stops the build. Review new upstream concepts before
updating the contract and renderer together.

`forge/components/styles/` defines browser anatomy by family and concern;
`forge/components/static/components-base.css` is its assembled output.
`components.css` combines that anatomy, the shared shadow renderer, and generated
bindings. The server embeds it. Its checksum must match the publication asset
manifest, so regenerate assets and rebuild the server together.

The coverage report at `published/assets/components-coverage.json` records each
mapped token, selector, property, excluded token, and supported configuration.
The same selector structure is checked across all six appearances.

## What a rebuild changes

`components/styles/` and `static/elevation.css` contain handwritten CSS. The Material
JavaScript generators contain explicit token-to-property mappings and handwritten
CSS templates. Each build combines these sources into generated `components.css`
and `buttons.css`. Edit the source files rather than the generated output.

A token value change updates the generated CSS variables. Handwritten values stay
as written, including defaults inside declarations classified as Mixed. New tokens
need a mapping before the renderer can use them. Unknown tokens in supported
component sets fail the build; excluded sets remain listed in the coverage report.
The scanner reports these choices and does not rewrite them.

## Markup and behavior

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

Add `md-interactive-card` only when the entire card is an action. Static cards
containing buttons retain their container appearance while their buttons handle
interaction states. An interactive card with `aria-disabled="true"` suppresses
activation. Applications that implement dragging can set `data-dragged="true"`.

Card container shape is token-backed at 12dp. Supplemental measurements use
16dp left/right content padding, start-aligned labels, and at most 8dp between
adjacent cards. The card comparison grid uses an 8dp gap without extra horizontal
specimen padding. Captions remain outside the cards. Card previews demonstrate
appearance and interaction states; they do not implement a reorder workflow.

The shared press controller handles pointer and keyboard feedback, ripples,
cancellation, removal, and disabled controls for buttons, icon buttons,
interactive cards, chips, and menu items. Icon-button pressed shapes use the
captured size-set spring parameters. Cards, chips, and menu items retain their
source shape and receive ripple feedback without an invented shape spring.
Hover and focus keep their own state layer during a press. The press ripple uses
the full captured pressed opacity independently, following Material Web. Its soft
edge expands from 20% of the largest dimension toward the center over 450 ms.
Ink fades in over 105 ms, remains pressed for at least 225 ms, and fades out over
375 ms while expansion continues. Hover transitions take 15 ms. Touch waits
150 ms to distinguish a press from scrolling; a quick tap starts ink on touch-up.
Repeated presses restart expansion from the new position. These timings and
geometry follow [Material Web's ripple implementation](https://github.com/material-components/material-web/tree/5a95667bb70ca5be8785b762fb0303737c3b4ac1/ripple/internal), while colors and opacity come from the
captured component tokens. Reduced motion keeps static state feedback.
Native inputs own their checked state and radio-group keyboard behavior.
Outlined fields keep the captured container height, float their label on focus
or a populated value, and place supporting text outside the outline. Outlined
selects wrap a native select with a token-colored trailing icon; the browser
continues to own the popup and keyboard selection.

Use `data-elevated="true"` for an elevated filter chip. Icon buttons support
`data-size` values `xsmall`, `small`, `medium`, `large`, and `xlarge`;
`data-shape` values `round` and `square`; and `data-width` values `default`,
`narrow`, and `wide`. Defaults are small, round, and default width.

Menu previews use native buttons and submenu popovers. Arrow keys move within
each menu, with Home, End and initial-letter navigation. Tab leaves the menu.
Native dialogs retain browser focus management and Escape handling.
Reduced-motion preferences suppress animation while preserving state feedback.

## Renderer decisions and exclusions

Component tokens supply values. The web renderer supplies markup, graphic
geometry, event handling, and layout where the source has no matching token.
The renderer uses system measurement tokens for content spacing. Radio rings
use a 2 px stroke and a half-size center dot. Dialog backdrops use the captured
neutral-black reference at 32% opacity. Shared shadow geometry follows Material
Web's key and ambient formulas, with the original attribution retained.

The coverage report identifies these exclusions:

- Superseded baseline and deprecated icon-button sets.
- Cascading indicators in single-choice listbox popups.
- Composite typography tokens where the renderer uses the corresponding
  individual font, size, weight, line-height, and tracking tokens.
- Surface-tint layers, since the renderer uses the supplied surface colors
  directly.

The catalog's token-table layout, folders, preview tiles, and compact search
remain application UI.

## Rendered checks

`forge/conformance/check-component-rendering.mjs` compares computed colors,
opacity, dimensions, typography, shapes, outlines, elevation bindings, and
state layers against source variables in the gallery. Run it in the browser
through the available browser-testing tool. It reads the DOM without changing
page state. Also check real catalog pages, including download and context
dialogs, menus, theme controls, and card navigation.

Sheet checks cover container geometry, handles, headlines, action states and
focus indicators. FAB menu checks cover both action and close-button profiles.
Progress checks cover current and baseline profiles, wave and thick variants,
track gaps, end stops, and four-color strokes.
Loading checks cover the container, indicator size, colors, and closed SVG path.
Carousel checks cover visible items, including disabled media, outlines, focus
indicators, and state layers.
List, search, date and time checks compare their container, content and control
states against captured source variables. Test selection, cancellation and input
validation in Interactive mode as well.

Run `checkHandoffRendering()` on `/md3/for-agents` in all six appearance settings.
It checks each nested card's foreground role and a minimum text contrast of 4.5:1.

`forge/conformance/feedback-transitions.html` exercises press, release, pointer
exit, re-entry, and rapid repeated presses across all six appearances. It loads
the running catalog's component CSS and shared controller, samples rendered
layers on animation frames, and checks continuity, ripple overlap, and cleanup.
With the catalog running on port 8081, serve the fixture:

```sh
go run ./forge/conformance
```

Open `http://127.0.0.1:8082/feedback-transitions.html` and click **Run transitions**.
The fixture contains representative markup copied from the component gallery.


## Checkboxes and switches

`/md3/components/checkbox` and `/md3/components/switch` show the generated
selection controls. Checkbox previews include unchecked, checked, indeterminate,
and error states. Switch previews include plain and icon handles. Both use native
checkbox inputs for keyboard selection and disabled behavior; switches expose
`role="switch"`.

The build maps every captured checkbox and switch token. Source dimensions,
shapes, colors, opacities, focus rings, and interaction layers feed their CSS.
The renderer supplies SVG checkmarks, indeterminate marks, and switch icons.
It centers the switch handle one half-track-height from either end and moves it
between those positions using the fast spatial spring from the site motion scheme.
These positioning and motion choices are handwritten renderer rules. Checkboxes and switches use the shared ripple lifecycle with their captured
pressed-state colors and opacities. Checkbox ink is clipped to its circular
state layer. Switch ink follows the handle inside a circular state layer.
Hover and focus layers remain behind the ink.

Indeterminate inputs are initialized by `buttons.js` on load and after an HTMX
swap. Native activation clears the indeterminate state. Forced-color mode shows
the native inputs, and reduced-motion mode disables handle transitions.


## FABs and extended FABs

`/md3/components/floating-action-button` and `/md3/components/extended-fab`
show the current separate size and color sets. Each supports default, small,
medium, and large sizes, with primary, secondary, tertiary, and their container
variants. Compare one dimension or select Interactive to use the native button.

Size tokens supply dimensions, corner radii, icon size, and extended-FAB
padding, gap, and typography. Color tokens supply container and icon colors,
label colors, state layers, and elevation. FABs reuse the shared button ripple,
keyboard, pointer, cancellation, and reduced-motion behavior. The renderer uses
the system focus indicator and the shared Material Web shadow geometry.

The coverage report excludes older combined size/color sets, surface/branded
sets, and lowered tokens. These previews have no disabled variant because the
current sets have no disabled-state tokens. The size and color selectors remain
independent, and the preview varies one dimension at a time.

## Badges, snackbars, and tooltips

Badges support dot and labeled forms. Their size, shape, fill, label color, and
font metrics use the captured badge tokens.

Snackbar previews support single-line and two-line containers, Undo, and
Dismiss. The Control selector chooses which button receives the displayed state. Interactive mode uses a native popover and a six-second timeout. The
timeout pauses while hovered or focused and resumes with the remaining time.
Escape closes the notification. The timeout and layout spacing are renderer
choices because the captured tokens provide neither. Set
`data-snackbar-duration="0"` to leave a notification open until dismissed.
Action and close buttons reuse common button anatomy with snackbar-specific
colors, typography, state layers, and ripples.

Plain tooltips appear on hover or keyboard focus and close with Escape. Rich
tooltips use a non-modal native popover with a Got it action. Their interactive
content has dialog semantics. Both use their source surface, shape, and text
tokens; rich tooltips also use source elevation and action tokens. Placement
and trigger behavior are renderer rules.

The supplied Material tooltip measurement figures supplement the captured
component tokens, which do not include padding or height:

| Measurement | Plain | Rich |
| --- | --- | --- |
| Container height | 24dp minimum; grows for wrapped text | Fits content |
| Top padding | 4dp (16dp text line within 24dp height) | 12dp |
| Bottom padding | 4dp | 8dp |
| Left/right padding | 8dp | 16dp |
| Subhead to supporting text | Not applicable | 4dp |
| Supporting text to action | Not applicable | 12dp |

The plain tooltip's illustrated 8dp padding is horizontal; applying it to all
four sides would make a single-line tooltip 32dp tall. Rich tooltip action text
aligns with the subhead and supporting text. These layout rules reuse system
measurement variables without changing captured token values.

## Tabs and segmented buttons

Tabs include primary and secondary styles with labels, icons, or both. Each tab
controls a linked panel using the [WAI-ARIA tabs pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/). Arrow keys, Home, and End move focus and activate the
matching panel. Only the selected tab participates in the page tab sequence.
Primary indicators follow the content width; secondary indicators fill the tab.
Indicator movement uses the fast spatial spring from the site motion scheme, with
motion disabled by the reduced-motion preference. Indicator width, content
padding, icon/label spacing, and panel behavior are renderer rules.

Segmented buttons include single selection through native radios and multiple
selection through native checkboxes. Each supports labels, icons, or both,
selected checkmarks, disabled states, and the shared ripple. Source tokens supply
the container height, shape, outline, typography, colors, and state layers.
System spacing supplies content padding and gaps. The renderer joins segment
borders and rounds the outer corners of the group.

The supplementary measurements below describe layout rather than additional
captured component tokens. CSS applies them without replacing source values.

| Attribute | Value | Source |
| --- | --- | --- |
| Container width | Dynamic, based on labels | Supplementary measurement |
| Segment width | Equal share of the container | Supplementary measurement |
| Height | 40dp | `container.height` token |
| Outline width | 1dp | `outline.width` token |
| Label alignment | Center | Supplementary measurement |
| Left/right padding | Minimum 12dp | Supplementary measurement; system spacing |
| Padding between elements | 8dp | Supplementary measurement; system spacing |
| Target size | 48dp | Supplementary measurement; system spacing |

Checkmarks and leading icons share one layout slot and remain in intrinsic sizing
when hidden, so selecting a segment does not resize the group or its segments.
Unselected labels remain centered without a visible icon.

Density changes only visible height: the preview offers default (40dp), −1
(36dp), −2 (32dp), and −3 (28dp). Each step subtracts the 4dp system spacing
value from the captured height; it leaves widths, typography and 48dp targets
unchanged. This is supplementary [Material density guidance](https://m3.material.io/components/segmented-buttons/specs), not a replacement height token.

The native inputs extend beyond the visible segments to provide the 48dp
touch target. The group reserves that height so targets do not overlap nearby rows.

## Sliders

Sliders support five captured sizes, horizontal and vertical layouts, continuous
and stepped values, and single or range selection. The preview configures value
visibility, end stops, initial values, discrete steps, and optional volume or
brightness inset icons. Inset icons apply only to standard single sliders in
M/L/XL sizes. Only one dimension varies across a preview grid.

The [M3 measurements section](https://m3.material.io/components/sliders/specs#16868fdb-4ed2-4e3f-99ad-24328cbfd2f4)
provides the reference geometry. Existing captured measurements take precedence;
slider measurements are not overridden to match the diagrams. Source tokens
supply track heights of 16/24/40/56/96px and outer corner radii of 8/8/12/16/28px
for XS/S/M/L/XL, 4px resting handles, 24px inset icons in M/L and 32px in XL,
and 6px handle-to-track gaps.

Four captured dimensions differ from the diagrams: the medium handle is 44px
rather than 52dp, the label is 28px high rather than 44dp, the label-to-handle
gap is 12px rather than 4dp, and the stop inset is 4px rather than 6dp. Those
source values remain in force. Shared slider CSS complements missing tokens
with the documented 48px label width and 16px horizontal padding. Vertical label
padding is derived from the existing height and line-height tokens so text fits.
The published `label` and `value-indicator` token groups describe the same
anatomy and are not offered as separate preview variants.

Native range inputs handle pointer, touch, arrow keys, Home, End, and disabled
behavior. Range inputs keep their full native min/max for positioning and expose
the neighboring value through `aria-valuemin` or `aria-valuemax`. Moving a handle
past its neighbor clamps that handle at the neighbor's value. Both remain
keyboard accessible when their values overlap.

Source tokens supply handle and track dimensions, shapes, colors, disabled
opacity, tick marks, end stops, value-label typography, and overlap outlines.
The shared ripple uses each handle's source pressed-state color and opacity.
Pointer hover applies within the handle's interaction area. Focus and press
shrink the handle using their source width and the fast spatial spring.

The renderer supplies axis layout, gaps around the handles, native-input clipping,
vertical orientation, icon color roles, label padding and shape, and a system
focus ring. Vertical length grows with track thickness, and extra-large horizontal
previews get wider cells. Label appearance and width changes follow reduced-motion preferences.
These choices remain in the handwritten CSS and JavaScript. Their source-token
references and literal CSS values appear in `css-usage.json`; the scanner does
not evaluate JavaScript or template positioning calculations.

## Progress indicators

Progress previews include current linear and circular indicators, normal and thick
tracks, and flat or wavy shapes. The baseline linear and circular sets remain
separate variants and include their four-color indeterminate mode. Interactive
mode adds a range input for determinate progress.

Source tokens supply colors, thickness, circular size, wave amplitude and
wavelength, track gaps, and linear end stops. The renderer samples SVG paths and
rounds the circular wave count to make the seam continuous. Thick circular
previews use the thick size when the wave option is selected.

Animation uses timing adapted from AndroidX Material 3: a 1,750 ms linear cycle
and a 6,000 ms circular cycle. Wave travel, SVG geometry, and the preview's linear
length are renderer rules. Motion pauses outside the viewport and in hidden tabs.
Reduced-motion previews show a stationary segment. Catalog thumbnails remain
stationary. These animation rules live in `forge/components/static/progress.js`;
the CSS usage report covers CSS declarations only.

## Loading indicators

Loading previews include determinate and indeterminate modes. Contained uses the
contained color pair, and Uncontained has a transparent container following
AndroidX. The separate base `container.color` remains in exports and is listed
as unused in component coverage. Source tokens supply
the 38 px indicator, 48 px container, container shape, and colors in each theme.

The browser morphs between the seven shapes defined by
[AndroidX Material 3](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/LoadingIndicator.kt).
Determinate progress morphs from the circle to Soft Burst and rotates
counterclockwise. Indeterminate progress uses the upstream 650 ms morph interval,
4,666 ms global rotation, and spring damping/stiffness. The browser evaluates the
spring directly between intervals. Shape matching, centering, rotation, and
animation scheduling are renderer behavior.

`forge/components/static/loading-shapes.json` holds matched cubic curves exported
with AndroidX Graphics Shapes 1.1.0. The first frame is rendered in HTML. Live
previews fetch the geometry once and animate while visible. Homepage thumbnails
remain stationary and need no geometry request. Reduced motion keeps indeterminate
previews stationary; determinate controls still update their shape.

The normal asset build reads the committed geometry. To update the upstream shape
definitions, edit `forge/material/reference/LoadingShapes.java` and regenerate:

```sh
mkdir -p /tmp/loading-shapes
curl -fsSL https://dl.google.com/dl/android/maven2/androidx/graphics/graphics-shapes-android/1.1.0/graphics-shapes-android-1.1.0.aar -o /tmp/loading-shapes/shapes.aar
unzip -p /tmp/loading-shapes/shapes.aar classes.jar > /tmp/loading-shapes/shapes.jar
curl -fsSL https://repo.maven.apache.org/maven2/org/jetbrains/kotlin/kotlin-stdlib/2.0.21/kotlin-stdlib-2.0.21.jar -o /tmp/loading-shapes/kotlin.jar
curl -fsSL https://dl.google.com/dl/android/maven2/androidx/collection/collection-jvm/1.5.0/collection-jvm-1.5.0.jar -o /tmp/loading-shapes/collection.jar
java --class-path '/tmp/loading-shapes/*' forge/material/reference/LoadingShapes.java > forge/components/static/loading-shapes.json
npm run test:ui
```

The exporter uses Java's source-file launcher and requires a JDK with record
support. Java is only used for this geometry update.

## App bars and toolbars

App bar previews include small, centered, medium, large, flexible, bottom, and
baseline top-app-bar variants. Current sets include subtitle, avatar, and search
options where their source tokens apply. Flexible subtitle layouts use their
separate captured heights. Interactive previews pin the bar inside a scrollable
container and apply the source on-scroll colors and elevation. Expanded bars keep
their captured height while scrolling.

Source tokens supply surfaces, elevation, shapes, typography, icon/avatar sizes,
and available content insets. App-bar row structure, expanded-title placement,
avatar colors, and search-field anatomy are renderer rules. Actions reuse the
existing icon buttons and FAB implementation. Baseline composite typography and
surface-tint fields have explicit exclusions because their individual metrics
and surface colors already supply those values.

Toolbars include docked and floating variants, standard and vibrant colors,
horizontal and vertical floating orientation, and an optional separate FAB in
two sizes. The [toolbar specification](toolbar-spec.md) combines the captured
values, supplemental measurements, source guidance, conflicts and verification.
Horizontal toolbars support icons, labels, or both. Each preview has a bounded
page: docked bars span its bottom edge, while floating bars keep their outside
clearance. Actions that do not fit move into the shared menu. Small buttons keep
their visible dimensions inside 48dp targets. Vertical toolbars use icons
within their captured width. The State selector applies to the first toolbar
button. Edit toggles its selected state.

Toolbar tokens supply container geometry, spacing limits, FAB geometry, colors,
disabled opacity, and state layers. Buttons reuse common button or icon-button
anatomy for their size, typography, shape, focus ring, and ripple. Docked spacing
uses the available width between the captured minimum and maximum gaps. Arrow
keys follow orientation, Home and End select the first or last enabled control,
and Tab moves through the toolbar as one group. Floating FABs remain separate
controls in the tab order. Their hover, focus, and press elevations use AndroidX
system elevation roles where the toolbar-FAB capture only supplies resting
elevation.

`checkBarRendering()` in `forge/conformance/check-component-rendering.mjs` checks
rendered bar dimensions, icon colors, toolbar states, and FAB geometry against
the active token bindings. Run it on the component page after switching variants
or appearance.

## Bottom and side sheets

Bottom sheets include expanded and minimized states with standard or modal
behavior. Side sheets include docked and detached shapes on either logical edge.
Detached sheets use their captured shape with the selected standard or modal
surface and elevation. The State selector applies to the bottom drag handle or
the side sheet's Done action. Disabled side-sheet actions use common text-button
colors and opacity.

Source tokens supply surfaces, elevation, shapes, side width, headline metrics,
handle dimensions and opacity, focus indicators, and side-sheet action states.
The standard docked side sheet uses the captured divider color at its page edge.
The renderer mirrors asymmetric corners for the leading edge and RTL layouts.
Side sheets explicitly clear the unused logical inset so native modal-dialog
defaults cannot pin an end-aligned sheet to the start edge.

Sizing and gestures use explicit renderer defaults in `forge/components/styles/sheets.css`:
56 px minimized height, a `min(24rem, 85svh)` expanded height, a 56 px drag threshold,
and a 125 px/s fling threshold. The maximum modal bottom-sheet width is 40 rem.
The minimized height, drag thresholds, and scrim opacity follow
[AndroidX SheetDefaults](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/SheetDefaults.kt).
Scrims use palette neutral0 at 32% opacity. Body spacing, the demo frame, content
layout, and animation scheduling are renderer rules. These CSS values appear in
`css-usage.json`; the scanner covers CSS declarations rather than JavaScript.

Interactive modal sheets use native dialogs for focus trapping and background
inertness. Escape, the close button, and a completed backdrop gesture dismiss them.
Exit animation finishes before the dialog closes, then focus returns to the opener.
Standard sheets remain inside the preview and keep the surrounding page active.
Open/close and resize animations use system motion tokens and honor reduced motion.

The bottom handle supports dragging, clicking, Arrow Up/Down, and Home/End. A drag
below half the minimized height dismisses the sheet. Animation interruptions start
from the current rendered position. Cancel and dismissal restore the name field;
Done retains its value for the next opening. This demo state lasts until the page
reloads.

## Split buttons

Split buttons combine a primary action and a menu trigger. The preview supports
filled, tonal, outlined, and elevated buttons at all five sizes, with an icon,
label, or both in the leading action. The trailing action always uses the menu
chevron. Split-button
tokens supply the height, gap, outer shape, inner corners, padding, and trailing
icon size. Common button tokens supply colors, typography, focus, elevation,
state layers, ripples, and shape spring parameters.

Each half responds independently. Hover and press change its inner corners;
opening the menu applies the trailing button's selected inner-corner token.
The layout mirrors in RTL. The menu uses the existing standard menu styles and
a native popover with Escape dismissal, arrow keys, Home, End, and focus return.
The demo actions update an inline message. Menu positioning and viewport margins
are renderer behavior. The menu sits 4dp from the button and prefers a trailing
button edge, with group-edge and viewport fallbacks. Each half has a minimum
48×48dp target without changing its captured visible height. Supplemental specs
shift the closed menu icon toward the leading button by 1, 1, 2, 3 and 6dp for
XS through XL. Pressing or opening the menu centers it and rotates it 180°
inward; RTL mirrors the offset and rotation. These are CSS behavior and
measurements absent from the captured split-button tokens.

## Button groups

The preview includes standard action groups and connected selection groups at
all five sizes. Each item reuses common button colors, typography, focus, ripple,
and shape springs. The State and Selection controls apply to the middle item.
Connected groups support single or multiple selection; single selection keeps
one item selected. Arrow keys, Home, and End move focus between enabled items.

Group tokens supply height, spacing, connected outer and inner corners, selected
corners, and pressed corners. Standard groups grow the pressed item's width by
the captured percentage and compress its immediate neighbors. The renderer caps
compression at each neighbor's smaller horizontal padding, following AndroidX
ButtonGroup. Width changes use the captured damping and stiffness and start from
the current rendered width when interrupted. Reduced motion applies the target
width immediately.

The three-item example, reserved check-icon space, demo messages, and horizontal
scrolling for large specimens are renderer rules. Large specimens keep their
captured dimensions rather than shrinking to fit the preview.

## FAB menus

FAB menus include primary, secondary, and tertiary colors. The State target
selector applies the chosen state to the first item or the close button. The
capture provides enabled, hover, focus, and press states. Tokens supply item and
close-button geometry, typography, colors, elevation, shadows, and state layers.
Each action uses the shared ripple and focus behavior.

Interactive menus use a native popover. Create opens the menu, Escape or an
outside click dismisses it, and choosing an item returns focus to Create.
Arrow keys, Home, and End navigate actions. One persistent button uses the FAB
profile while closed and the menu close-button profile while open. Its shape
and color transition; the plus rotates into a close icon. Only the actions live
in the popover, so no second surface covers the original FAB. Focus stays on
the toggle on opening; Arrow Down or Arrow Up enters the actions. Tab exits
and dismisses the menu.

Following AndroidX FloatingActionButtonMenu, items reveal from the bottom upward
and collapse in reverse order. Width uses the system fast spatial spring, opacity
uses fast effects, and stagger timing uses slow effects. The stagger follows the
sampled slow effects spring as its animated item count crosses each step. Interrupted animations
start from their current width and opacity. Reduced motion applies the open or
closed state immediately. Native display and overlay transitions keep the exit
animation visible until it finishes.

The three example actions, preview frame, viewport margins, scroll clearance for
focus rings and shadows, popover positioning, and stagger scheduling are renderer rules. Source tokens remain unchanged.

## Lists

The preview has separate choices for video size, text layout, trailing content,
and inset, full-width, or omitted dividers. Checkbox, radio, and switch
compositions use the existing control renderers and native inputs. Leading
checkboxes and radios are available for selection rows.

Baseline rows at least 88px tall top-align their content, following the Material
list measurements. A ResizeObserver also handles rows that grow through text
wrapping or media. Captured padding remains in effect. The preview dot grid uses
the theme's surface and outline roles; its dot geometry is application artwork.

Lists include standard and segmented layouts, baseline and expressive shapes,
and one-, two-, and three-line content. Leading content can be an icon, avatar,
image, video, checkbox, or radio. Text layouts include supporting text, wrapped
supporting text, and an overline. Heights are minimums so larger media keeps its
captured size and padding. Video size is selected independently of text lines.

The preview includes single and multiple selection, expansion, reordering, and
action reveal. State and Selection apply to the first row. Expansion animates
both opening and closing. Expanded children retain the parent row width and
leading-content alignment by default. Child inset and expanded surface are
independent preview choices: inset or flush, and tonal or no additional fill.
Tonal uses the system surface-container role as one continuous parent and
child surface. This is an explicit composition choice: the captured normal and
expanded container colors are identical in the current scheme. None preserves normal row colors instead of making the rows
transparent. Disclosure icons independently offer contained and plain styles;
plain omits the icon background while retaining its size and state colors. These are
renderer compositions, not token overrides. Reordering starts from anywhere
on a row. Activating its handle opens the shared standard bottom sheet with Move up
and Move down action rows as a single-pointer alternative to dragging. Boundary
actions are disabled. The sheet has no scrim and leaves the page interactive. Escape dismisses it
and restores focus to the handle; an outside click dismisses it without moving
focus away from the clicked control. The sheet stays open while moving the chosen item, and
announces its new position. Reordering does not select the row.

Reveal works by swiping horizontally or activating the More button. The primary
action remains separate. A decorative trailing navigation chevron is omitted
for this composition. Arrow keys move between primary and secondary actions;
Escape hides the revealed action. Its direction follows the writing direction. The
example action marks the row done. The reveal distance combines the common
small icon-button height and the captured row insets. A swipe settles open
when it crosses half that distance. Pointer gestures use system space100 as
the movement threshold. These thresholds, example content,
and gesture handling are renderer rules. Utility controls use common icon-button
geometry and system state opacity where list-specific fields are absent.

Single-selection rows offer radio controls; multiple-selection rows offer
checkboxes or switches. Selection rows do not contain nested secondary actions.
The [guidelines](https://m3.material.io/components/lists/guidelines) and
[accessibility guidance](https://m3.material.io/components/lists/accessibility)
describe these composition and keyboard rules separately from visual tokens.

The compiler maps 198 list fields. Five composite typography fields duplicate
their individual metrics, and three repeated height names duplicate the normal
height fields. These eight exclusions appear in `components-coverage.json`.
A disagreement between repeated heights fails the build. Captured values and
names remain unchanged in the exports.

### Carousel

The carousel preview includes multi-browse, uncontained, hero and full-screen
layouts. Items support the captured hover, focus, pressed and disabled states,
with optional outlines. The build maps 26 item fields and excludes the deprecated
surface-tint field. The source labels outline width as an elevation; the renderer
uses its numeric value as pixels while preserving the source type in the export.

The captured item tokens do not define carousel layout. `carousel.js` owns the
keylines, clipping windows and scrolling. Items keep their full media size while
their visible windows expand and contract. Multi-browse fits large, medium and
small windows into the available width. Hero centers the focal item between small
windows and aligns the first and last items to the corresponding edge. Uncontained
uses equal item widths; full-screen uses the viewport width. Layouts use the
[Material carousel guidance](https://m3.material.io/components/carousel/guidelines)
and the [AndroidX carousel anatomy](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/carousel/Carousel.kt).

The following values are renderer inputs rather than captured carousel fields:
preferred item widths of 160, 240 and 320px; preview heights of 128, 192 and 256px;
a 56px small window; and an 8px mouse-drag threshold. The gap uses system space100.
The web keyline calculation fits these inputs to the viewport and interpolates
between them during scrolling. Narrow viewports fall back to equal-width items.
The SVG media and numbered labels are sample content using system color roles.

Touch and trackpad scrolling use the browser's native scroll container. Mouse
dragging, arrow keys, Home, End and previous/next buttons navigate the same track.
The current item is announced after scrolling stops. Reduced motion disables
programmatic smooth scrolling. There is no autoplay.

### Search

Search uses one persistent header, input and set of actions. The same surface
changes position, width and height when focused. CSS transitions reverse from
the current geometry when interrupted; text and icons are not scaled. The
contained bar keeps its rounded shape. Divided search adds a divider between
its header and results.

Compare **Layouts** shows focused Docked and Full-screen search. Compare
**Styles** shows Contained and Divided (baseline). Focus selects Unfocused or
Focused independently. Adaptive keeps the docked layout and lets the bar and
focused view fill the available width, including beyond the fixed Docked width.
It does not switch to full-screen at narrow widths. Adaptive uses a responsive
mail composition built from the shared navigation rail, FAB,
list (including avatar and text slots), filled card and button primitives. It shows
a navigation rail, inbox and message at wide and medium widths, and inbox or
message at compact widths.
The Composition width slider resizes that application within the available
preview space. Search stays attached to the inbox column as it changes width.
The canvas is
the application screen for every preview, including interactive search: its
scrim covers the whole application composition and makes its background panes
inert, while the active search surface stays above it. It never covers the site's
filters or neighbouring previews. Full-screen
fills that canvas; docked stays near the bar.

Leading content can be a search icon, navigation button, or none. Trailing
configurations offer clear, avatar, options, clear plus options, options plus
avatar, or none. Query content is independent of actions. Options demonstrate
exact-name matching. Enter submits the typed query and moves focus to the
results heading without replacing the query. Selecting a result updates the
sample destination. Back, Escape and the canvas scrim dismiss search. Tab,
Shift+Tab and arrows navigate the controls and results. A visually hidden live
status announces suggestions and results. Physical screen-reader testing remains
separate validation work.

Captured tokens supply colors, typography, heights, shapes, spacing and states.
Contained margins change from 24px to 12px. Its docked results have their own
captured shape and gap. Full-screen contained search uses its captured background
role. Result rows inherit their host surface instead of painting a different
surface behind each row.

Search has no shadow by default in either style, following the M2-to-M3 guidance.
The Shadow control offers the captured token elevation as an explicit opt-in.
The unfocused bar uses the bar elevation field; focused docked search uses the
view elevation field. Contained docked search paints both surface shadows behind
all search content, so neither surface shades the other. Divided search applies
its shadow to the enclosing surface.
Full-screen remains flat. Captured elevation values are preserved.

Clipping belongs to the results area and the bounded app screen, not the shared
header's wrapper. This preserves rounded surface corners and the complete focus
indicator when focus returns to the closed search bar.

Docked width is 360–720px where space allows, with a 240px minimum height and a
maximum of two-thirds of the canvas height. These are manual sizing rules from
the measurement table, absent from the captured token sets. Narrow forced-docked
examples fit their available width. Search has no captured scrim fields, so its
scrim explicitly composes the existing modal drawer scrim color and opacity.
Contained actions reserve 48px targets around the captured 40px icon buttons.

The captured contained motion spring has no value. Geometry transitions use
the site motion scheme: fast spatial for docked geometry, slow spatial for
full-screen expansion and default spatial for its return. Fast effects controls
opacity; reduced motion disables these transitions, including icon rotation.
Android predictive back is platform-specific and is not simulated on the web.

Source guidance: [specifications](https://m3.material.io/components/search/specs),
[guidelines](https://m3.material.io/components/search/guidelines),
[accessibility](https://m3.material.io/components/search/accessibility).

### Date pickers

Date pickers include a docked calendar, modal calendar, modal text input and
full-screen range picker. Modal calendars have a year grid. Docked calendars have
a month list with year navigation. The preview compares one content type and date
appearance at a time, including selected dates, today and range interiors.

The build maps 210 fields and excludes 19 duplicate or deprecated fields. Three
source fields spell opacity as `opcaity`. They remain in the export, and the build
checks that they equal their canonical `opacity` counterparts. Composite typography
uses its individual metrics. Deprecated surface-tint layers are not drawn.

Captured fields control container sizes, header sizes and typography, day and year
geometry, state layers, selected colors, today's outline, docked month rows and
range highlighting. The active range's captured shape supplies the endpoint
shape. Focus rings use the system indicator tokens. Input fields, action buttons
and navigation buttons use the existing component implementations. The scrim uses
the captured navigation-drawer scrim color and opacity.

The renderer supplies calendar arithmetic, Sunday-first weeks, 12px calendar
padding, 24px header padding, navigation layout and the sample date of 18 October
2026. The example starts with 21 October selected, or 21–26 October for ranges.
It supports Gregorian dates from 1900 to 2100. Text input uses `YYYY-MM-DD` and
rejects impossible dates and reversed ranges. Range selection starts again after
a complete range or when the next date precedes the start.

Arrow keys move by day or week. Home and End move within a week. Page Up and Page
Down move by month; Shift moves by year. Month changes preserve the day where
possible and clamp to the new month's last day. Confirm commits the pending
selection. Cancel and Escape restore it. Docked selection commits immediately.
Native dialogs provide modal focus containment. The docked view uses a native
popover above the page and remains nonmodal. Live views cap their minimum height
to the available viewport and scroll their contents. Resizing dismisses a docked
view so it cannot remain detached from its opener.
Opening and closing use the fast effects spring from the site motion scheme, with no fade
when reduced motion is enabled.

Source guidance: [Material date pickers](https://m3.material.io/components/date-pickers/overview),
[AndroidX DatePicker](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/DatePicker.kt)
and [DateRangePicker](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/DateRangePicker.kt).

### Time pickers

Time pickers include portrait and landscape clocks, keyboard entry, 12-hour and
24-hour formats. The preview compares the active value, dial, period selector or
input field. Value and period controls support enabled, hover, focus and press
states. Input fields support enabled, hover and focus states.

The build maps 160 fields from `md.comp.time-picker` and `md.comp.time-input`.
It excludes ten duplicate typography composites, two deprecated tint layers and
two deprecated names ending in `.ignore`. Source values remain in the exports.
Captured fields supply container shape and color, typography, dial and handle
geometry, value sizes, period layout, state layers and input focus outlines.
The source has no dedicated time-picker motion fields, so transitions use the system
fast effects spring for disclosure and fast spatial for the dial. Reduced motion
removes the transitions.

The renderer supplies clock arithmetic, pointer and keyboard handling, layout and
validation. Dial labels use the Material implementation's outer radius of 101,
inner radius of 69 and ring boundary of 74, scaled to the captured dial size of
256. System spacing supplies 24px outer padding, a 24px separator, a 36px gap
between the display and clock, and a 12px period gap. Supporting labels use the
Material implementation's 7px gap. These layout rules appear as handwritten
values or system-token references in `css-usage.json`.

Selecting an hour advances to minutes. The 24-hour clock has an outer 00–11 ring
and an inner 12–23 ring. Arrow keys change the active value by one; Page Up and
Page Down change it by five. Home and End select its bounds. Enter or Space on
the hour dial advances to minutes. Period buttons support arrow keys, Home and
End. Keyboard entry rejects hours outside the chosen format and minutes outside
0–59. Invalid fields use the system error color with the captured focus outline
width. Save commits the time as `HH:mm`. Cancel and Escape restore it.

The native dialog contains focus and restores it to the opener on close.
Clock-hand transitions follow the shortest rotation from the current displayed
angle. Pointer dragging follows the pointer directly. The selected label color
is clipped to the moving handle, keeping the text readable during rotation.

Source guidance: [Material time pickers](https://m3.material.io/components/time-pickers/overview)
and [AndroidX TimePicker](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/TimePicker.kt).


### Chips

Assist and suggestion chips activate actions. Filter and input chips support
selection, and input chips have a separate remove action. Filter chips offer an
optional remove action. The preview supports icons and text-only chips, plus
avatars for input chips. Assist, suggestion and filter chips have flat and
elevated profiles. Input chips use their captured flat profile.

The four chip sets map 309 source fields. Each set supplies geometry, typography,
colors, state layers, focus rings and disabled opacity. Leading and trailing icons
have separate color bindings. The input avatar uses its captured size and shape.
Four duplicate typography fields and three tint layers are excluded from rendering.
Source values remain available in exports.

System spacing supplies 12px text padding, 8px icon padding and gaps, and 4px avatar
padding. Outlines sit over the container so selection cannot shift the content
when the outline width changes. Labels stay on one line and truncate to fit.
The example avatar uses an initial with system tertiary-container colors.
These renderer choices appear in `css-usage.json`.

Interactive filter and input chips toggle selection through a native button.
Arrow keys move between the primary and remove actions, accounting for text
direction. Delete or Backspace removes the chip. The preview announces selection,
action and removal, and restores focus to its controls after removal. Disabled
chips disable both native actions. The shared ripple implementation handles
pointer and keyboard activation, cancellation and reduced motion.

Source guidance: [Material chips](https://m3.material.io/components/chips/overview)
and [Material Web chip actions](https://github.com/material-components/material-web/blob/main/chips/internal/multi-action-chip.ts).


### Text fields, selects and autocomplete

Text fields, selects and autocomplete have filled and outlined profiles. Previews
compare interaction states or field styles, with controls for empty and populated
values, icons, affixes and errors. All six profiles support enabled, hover, focus
and disabled states.

The six sets map 555 fields. Captured values supply container geometry, typography,
label metrics, input and icon colors, state layers, outlines, underlines and popup
styles. Popup fields include their elevation, shape, divider, option height,
selected background and typography. Select popups also use their captured leading
and trailing icon fields. Composite typography duplicates and deprecated popup
tint layers are excluded. Cascading indicators are excluded from the single-choice
listbox anatomy. Raw source values remain in exports.

Filled fields use the captured active-indicator height for their underline. The
filled text-field source also supplies a separate focused thickness of 3px; the
renderer uses it for the keyboard focus ring alongside the 2px focused underline.
Other profiles use their captured focus-ring width or the system focus indicator.
Filled labels use 8px top spacing and keep their alignment with the input text.
Outlined fields and their focus rings use fieldset notches, keeping the label
cutout clear on any surface.
System spacing supplies 16px content padding, 8px gaps and 4px notch padding.
Label transitions use the fast spatial spring; focus-ring opacity uses fast
effects. Both stop when
reduced motion is enabled. These layout and motion rules appear in `css-usage.json`.

Selects use read-only native inputs with combobox semantics. Autocomplete inputs
remain editable and filter suggestions case-insensitively. Arrow keys open the
popup and move its active option. Enter commits an active option. Escape closes
the popup without committing navigation. Tab closes it and leaves the field.
Selects also support Home, End and prefix typing. Repeated letters cycle matching
options, with a 500ms typing interval. Autocomplete accepts free text, including
text with no matching suggestion.

Popups use native popovers above the page. Their width follows the field and their
height fits the available viewport. They open above the field when there is more
room there. Resize and page scrolling dismiss them. Native inputs retain editing,
selection and form events; choosing an option emits input and change events.

The 15 October 2026 capture’s error-hover label, outline and trailing-icon colors reference
`md.sys.color.on-error-container`. Some medium and high contrast schemes give
that role poor contrast on the field background. The renderer keeps
the source reference. Review those states before reusing them, and use
`forge/material/overrides.json` for a deliberate replacement.

Source guidance: [Material field layout](https://github.com/material-components/material-web/tree/main/field/internal)
and the [WAI-ARIA combobox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/combobox/).

### Dialogs

Basic dialogs use native `alertdialog` elements; fullscreen dialogs use native
`dialog` elements. Both contain keyboard focus, handle Escape and restore focus
to their opener. Optional headlines retain an accessible name when hidden.
Basic previews combine icons and dividers independently and offer short or long
content. Fullscreen previews offer header or bottom actions and three scroll
positions. The coverage file contains 256 configurations across these options
and action states.

Interactive previews validate a required name and retain saved edits. Basic
confirming actions are disabled until the field is valid; dismissal stays
available. Fullscreen Save remains available and reports errors beside the
field. Closing a fullscreen dialog with unsaved edits opens a discard
confirmation. Keep editing or Escape returns to the edits; Discard cancels them.
Initial focus goes to the first control: the basic field or fullscreen Close.

Captured tokens supply fullscreen container shape and color, header and
action-bar heights, typography, icon geometry, action states, dividers and scroll
elevation. Three composite typography fields and the header tint layer are
excluded from rendering. Raw values remain in exports. The header gains
elevation when content is above the scroll position; the bottom action bar gains
elevation while content remains below it.

Renderer layout supplies basic widths of 280–560px, 24px outer spacing, 16px
headline-to-content spacing, 24px content-to-actions spacing and 8px between
actions. Body padding is applied once. Floating labels reserve their overflow
above fields. Fullscreen header controls use an 8px gap; titles can wrap. The
captured header height is preserved even though the source measurement diagram
and table disagree. No token override is added for that discrepancy.

Fullscreen dialogs fill the viewport; inline specimens have a bounded height.
The explicit variant preview keeps fullscreen visible at desktop widths so its
layout can be inspected. Applications must choose fullscreen for compact
layouts and basic dialogs for medium or expanded layouts. This adaptive choice
is not implemented by the preview. The required-name form is sample content;
acknowledgment-only actions, list selection and application-specific save errors
are application compositions. Material calls for at most two actions, confirming
actions trailing (with RTL ordering), and stacked long labels with confirmation
above dismissal. These are composition requirements, not additional tokens.

Fullscreen entry uses the slow spatial spring; exit uses default spatial.
Basic dialogs fade with fast effects. Closing during entry starts from the current rendered frame.
Reduced motion skips these animations. Native keyboard behavior and computed
styles are checked locally; physical screen-reader behavior needs separate
assistive-technology testing.

Source guidance: [specifications](https://m3.material.io/components/dialogs/specs),
[guidelines](https://m3.material.io/components/dialogs/guidelines) and
[accessibility](https://m3.material.io/components/dialogs/accessibility).

### Menus

Menus use the captured Common set with Standard or Vibrant colors. The renderer
maps all 213 fields across those sets. Vertical previews include supporting text
and submenus. Horizontal previews include labeled and icon-only actions.
Each layout supports selection, hover, focus, press and disabled states.

Horizontal spacing and shapes use their dedicated tokens. Labeled actions retain
the common leading-icon and text order in a row. Icon-only actions use
the menu icon-button colors with menu-item geometry and feedback opacity.
The Standard token `icon-button.disabled.icon.color` belongs to the source's
“Disabled, Selected” group; the mapping applies it to selected disabled icons.
The separate `disabled.icon-button.icon.color` supplies unselected disabled icons.

Interactive menus open from a button into native popovers. Arrow Down enters
the first action, and Arrow Up enters the last action. Actions dismiss the menu.
Submenus also use native popovers. Arrow Right opens a submenu in LTR, and Arrow Left
returns to its trigger. Those directions reverse in RTL. Escape closes a submenu
and restores focus; Tab leaves it. Hover opens a submenu after 200ms. Pointer and
keyboard focus update active and inactive container shapes. The expanded trigger
retains its captured active state layer while the submenu is open.

Renderer layout limits the demonstration's main menu to 20rem. Submenus have a
12rem minimum width and align with the parent's outer edge, flipping sides when
needed. If neither side fits, the child replaces the visible parent surface and
provides a Back item. Back and Escape restore focus to the parent item. Disabled
actions remain focusable with `aria-disabled` and cannot be activated.

Material recommends considering bottom sheets on compact screens. This renderer
uses Back navigation when adjacent menus cannot fit.
Static submenu comparisons stack below their triggers on narrow screens. Container shape changes
use the fast spatial spring. These layout and motion choices
appear in `css-usage.json`.

Source guidance: [Material menu behavior and focus](https://m3.material.io/components/menus/guidelines).

### Baseline menus

Baseline menus use the separate `md.comp.menu` set. Its 47 rendered fields supply
container color, 4px shape, elevation, 48px row height, label typography, icon
colors and sizes, selection, feedback, dividers and the inset keyboard focus ring.
Composite typography and the deprecated tint layer are excluded. Modern menu
bindings do not apply to this variant.

The preview supports plain, divided and cascading menus. Native popover behavior,
selection and keyboard controls are shared with Standard and Vibrant menus.
Baseline disabled colors apply to selected and unselected rows.

The measurements table supplies 112–280px width, 8px vertical container padding,
12px row padding and gaps, and 8px spacing around dividers. Rectangular rows use
zero corner radius. These renderer layout values appear in `css-usage.json`.

Source guidance: [Material baseline menu specifications](https://m3.material.io/components/menus/specs).

### Radio buttons

Interactive radio previews contain three native inputs sharing one name. The
browser manages single selection, arrow navigation and Space activation. Static
previews keep each selection and interaction state separate. The radio input sits
inside a token-sized state-layer host, which carries pointer and keyboard
ripples without adding children to the native input.

Captured radio tokens supply the state-layer and icon sizes, icon colors, disabled
opacity and hover, focus and pressed feedback. The ring's 2px stroke and half-size
center dot remain renderer geometry. The focus ring uses system indicator values.
The center dot uses fast spatial for geometry and fast effects for opacity, with animation
removed for reduced motion. Forced colors show the native radio control.

Native checkbox, switch and radio controls use Space for activation. Enter does
not start their pressed feedback. Radio previews are marked new for review.

Source guidance: [Material radio accessibility](https://m3.material.io/components/radio-button/accessibility)
and the [WAI-ARIA radio group pattern](https://www.w3.org/WAI/ARIA/apg/patterns/radio/).

## Divider measurements

Standalone dividers support full-width and inset layouts. Insets use 16px at the
start edge and optionally 16px at the end edge for middle-inset. Supporting text
sits 4px below the divider. These layout measurements use system spacing tokens;
color and thickness use the captured divider tokens.

The source table also lists 8px right and bottom margins without defining their
context. The preview exposes them as optional outer spacing, separate from the
insets. They are off by default.

## App bar and connected button spacing

App bar avatars keep their captured size inside a centered control slot matching
the small icon button's height. This slot is renderer layout; the captured
icon-button gap and app bar edge padding remain unchanged.

Connected buttons reserve the checkmark's width in both selection states. The
label alone is centered when unselected; the checkmark and label are centered
together when selected. Outlined buttons retain a transparent border when
selected so their footprint also stays fixed. Extra-small and small
connected buttons have a 48px minimum width and a transparent 48px-high target,
as specified in the button group guidance. Their visible heights still use the
captured container tokens. The target measurement uses `space600`.

## Carousel motion

Carousel scrolling uses an empty track with snap positions. The visible items
share a stationary sticky layer, where their masks and positions update together.
Native scrolling never translates the visible items separately from those
updates. Item shapes, colors and states continue to use the captured tokens.
Mouse release scrolls smoothly to the nearest stop before restoring native
snapping. Reduced-motion mode settles immediately.

## Date picker review

Interactive examples use outlined date fields with a calendar icon. Typed values
use `YYYY-MM-DD`, with validation for real dates, the supported 1900–2100 range,
and range ordering. Calendar selection commits with OK, and Cancel restores the
previous selection. Selection announcements are
screen-reader feedback, with committed values displayed in the entry fields.

The captured date-picker tokens supply colors, typography, shapes, day and year
targets, state layers, and panel dimensions. Field composition, parsing, calendar
arithmetic, focus, popup positioning, announcements, and save/cancel behavior are
renderer code.

The docked header contains separate month and year menus, each with previous and
next buttons. Both scrollable menus use the captured docked menu-list tokens.
The header stays outside the scroll regions, the other selector is disabled while
a menu is open, and Escape returns from a menu to the calendar before closing the
picker. The menu divider uses the standalone divider tokens. Flat menu rows and
the two header groups follow the anatomy diagrams.

Docked calendar height starts with the captured six-week 456px layout and subtracts
one captured 48px date-row height for each unused week. Month and year menus use
the captured full container height. This is renderer layout, with source values
preserved.

Picker presentation and date selection are independent. The preview offers docked,
modal calendar, modal input, and full-screen layouts, each with single-date or range
selection. Full-screen is the compact layout described under anatomy, with Close
and Save actions. Range colors and shapes use the captured modal range set; docked
ranges compose its connecting indicator with docked date tokens.

Single-date modal calendars support horizontal pointer swipes and month arrows.
Modal range calendars scroll vertically, adding six months at either scroll edge
within the supported 1900–2100 interval. This month window is renderer behavior.
Year menus scroll vertically. Reduced-motion preferences disable slide transitions.

The calendar icon opens the picker; typing alone leaves it closed. This follows
the accessibility page's more specific entry-point guidance. Text entry accepts
ISO year-first or US month-first dates with dashes, slashes, dots or spaces and
normalizes to YYYY-MM-DD after entry. Ambiguous month-first dates use en-US order.
There is no input mask. Enter commits a complete selection, Escape cancels, and
Cancel restores the committed dates. Arrow keys move by day/week, Home/End by
month boundary, Page Up/Down by month, and Shift+Page Up/Down by year. Shift+M and
Shift+Y open the available month/year menus. The input's Alt+ArrowDown shortcut
also opens the picker. Full dates and format hints are exposed to assistive
technology, with one roving calendar tab stop and 48px pointer targets.

Modal headers include the documented divider. Input layout uses the supplied
measurement diagram: 24px sides, 16px top, 36px supporting-text gap, 10px
from the headline row to the divider and from the divider to the floating label,
16px below the field, and 8px/12px above/below the action row. Each input reserves
half the floating label line height above its outline. This also keeps the
24px gap between range fields clear of the second floating label.
These are renderer rules using system measurement tokens.
`forge/material/overrides.json` replaces the captured 512px input-container minimum
height with 273px, the sum of the measured header, label clearance, field and
action-row geometry at the captured typography sizes. Additional fields
and validation errors can increase the height. The original capture is preserved.

The full-screen calendar uses the full-screen dialog shape and elevation tokens
for square corners and zero elevation. Its comparison specimen fills a bounded
preview. Switching to text entry uses the shared centered modal-input dialog,
including Cancel/OK and its normal corners and elevation. The calendar button
returns to the original full-screen calendar with the current selection. When a
bounded comparison opens text input, confirming or cancelling also restores the
bounded calendar.

Locale-specific formatting, assistive-technology testing and mobile-device gesture checks remain separate
validation work.

Source guidance: [Material date picker guidelines](https://m3.material.io/components/date-pickers/guidelines)
and [specifications](https://m3.material.io/components/date-pickers/specs).

## Motion schemes and lifecycle

The Catalog selects `Expressive` in `forge/material/site-theme.json`. The build
exports that captured context separately and applies its system spring values
to the site theme. Public DTCG and token-CSS downloads retain their selected
context, which defaults to Standard. Component aliases continue to reference
canonical system spring names. No captured values are overwritten.

`forge/components/static/motion.cjs` supplies the shared unit-mass spring model.
The Node build samples it into private `--_motion-*-duration` and
`--_motion-*-easing` properties in generated component CSS. These are derived
renderer values, not additional Material source tokens. Browser CSS and Web
Animations use those same curves. `motion.js` must load before component scripts;
Catalog serves that URL from the shared `.cjs` source.

| Property or component | Assignment |
| --- | --- |
| Menu reveal, submenu corners, switch, radio dot, slider bubble, field label, tab indicator and split-button shape/icon | Fast spatial |
| Small opacity and color changes, basic dialogs and date/time disclosure | Fast effects |
| Sheet movement/settling, list expansion/reorder and expanded rail geometry | Default spatial |
| Rail content/scrim effects and sheet scrim | Default effects |
| Full-screen search expansion and full-screen dialog entry | Slow spatial |
| Full-screen search return and full-screen dialog exit | Default spatial |
| FAB-menu item width/opacity/stagger | Fast spatial / fast effects / slow effects |

The component spring aliases and explicit assignments follow the
[Material motion specification](https://m3.material.io/styles/motion/overview/specs)
and component guidance. The
[official menu implementation](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/Menu.kt)
assigns fast spatial to scale/shape and fast effects to opacity/color even though
those assignments are absent from the captured menu tables. Full-screen dialog
slide choreography remains an authored browser presentation.

Buttons and standard button groups preserve spring velocity when retargeted.
Sheet release carries drag velocity into its settling spring. Date/time picker
exit starts at the rendered opacity and makes the closing panel inert. Changes
to reduced motion finish pending picker dismissal without stranding the dialog.
Search leading icons and date-picker arrows also stop transitioning.

Native popover overlay transitions retain menus and FAB menus during exit.
Without overlay support, an inert, ARIA-hidden copy remains outside the popover
lifecycle. Ordinary menus retain separate spatial/effects animations; FAB menus
mirror their item springs and delays. Reopening, completion, cancellation,
reduced motion and a timeout remove copies. No fallback calls `showPopover()`. Exit snapshots read a fixed set of used
geometry, paint and state-layer properties. They never enumerate computed
styles, which include the publication tokens inherited by every descendant.

Ripple timings and loading/progress algorithms retain their component-specific
Material implementations. Static components do not gain an arbitrary entry
animation. Hero loop choreography remains custom and uses the selected springs.

For browser regression checks, start Catalog on port 8081, then run from the root:

```sh
go run ./forge/conformance
```

Open `http://127.0.0.1:8082/motion.html` and choose **Run motion checks**. The
fixture uses real pages and scripts in isolated frames and simulates reduced
motion and missing overlay support. It covers six appearances, menu/FAB cleanup,
interrupted picker dismissal, search icon motion and sheet entry/exit. These
simulations supplement actual Firefox/WebKit and operating-system preference
checks; they do not establish full browser conformance.
