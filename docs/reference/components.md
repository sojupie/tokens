# Generated Material Design 3 components

Forge generates the remaining component families used by the catalog from the
captured Material token sets. Common buttons have their own
[reference](buttons.md).

| Family | Implemented configurations |
| --- | --- |
| Cards | Elevated, filled, and outlined; enabled, hover, focus, press, drag, and disabled states |
| Icon buttons | Standard, filled, tonal, and outlined; five sizes, two shapes, three widths, selection, and five interaction states |
| Filter chips | Flat and elevated; selected and unselected; enabled, hover, focus, press, drag, and disabled states |
| Radio buttons | Selected and unselected; enabled, hover, focus, press, and disabled states |
| Outlined text fields | Normal and error; enabled, hover, focus, and disabled states |
| Outlined selects | The native closed control, with normal, error, hover, focus, and disabled styles |
| Menus | Standard and vibrant vertical menus; selected, hover, focus, press, and disabled items |
| Basic dialogs | Container, headline, subhead, supporting text, icons, dividers, and text actions |
| Dividers | Color and thickness |

The current capture maps 777 source tokens. The gallery contains 1,290
configurations. Run the local server and open `/md3/lab/components`, or use
`?family=all` to see every configuration. The gallery is excluded from indexing.

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

`forge/components/static/components-base.css` defines browser anatomy.
`components.css` combines that anatomy, the shared shadow renderer, and generated
bindings. The server embeds it. Its checksum must match the publication asset
manifest, so regenerate assets and rebuild the server together.

The coverage report at `published/assets/components-coverage.json` records each
mapped token, selector, property, excluded token, and supported configuration.
The same selector structure is checked across all six appearances.

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

Menus retain navigation links and native popover behavior. They use normal Tab
navigation. Native dialogs retain browser focus management and Escape handling.
Reduced-motion preferences suppress animation while preserving state feedback.

## Renderer decisions and exclusions

Component tokens supply values. The web renderer supplies markup, graphic
geometry, event handling, and layout where the source has no matching token.
The renderer uses system measurement tokens for content spacing. Radio rings
use a 2 px stroke and a half-size center dot. Dialog backdrops use the captured
neutral-black reference at 32% opacity. Shared shadow geometry follows Material
Web's key and ambient formulas, with the original attribution retained.

The coverage report identifies these exclusions:

- Superseded baseline and deprecated icon-button and menu sets.
- Horizontal menus, submenu trigger states, embedded menu icon-button anatomy,
  and active/inactive menu-container shape morphing.
- Native select-popup tokens, whose rendering is controlled by the browser.
- Composite typography tokens where the renderer uses the corresponding
  individual font, size, weight, line-height, and tracking tokens.
- Surface-tint layers, since the renderer uses the supplied surface colors
  directly, and redundant menu inner-corner values already represented by the
  complete item shapes.

Other Material families, including full-screen dialogs, filled fields, switches,
sliders, navigation components, and the other chip types, still need their own
reviewed renderer before Forge can generate them. The catalog's token-table
layout, folders, preview tiles, and compact search remain application UI.

## Rendered checks

`forge/conformance/check-component-rendering.mjs` compares computed colors,
opacity, dimensions, typography, shapes, outlines, elevation bindings, and
state layers against source variables in the gallery. Run it in the browser
through the available browser-testing tool. It reads the DOM without changing
page state. Also check real catalog pages, including download and context
dialogs, menus, theme controls, and card navigation.

`forge/conformance/feedback-transitions.html` exercises press, release, pointer
exit, re-entry, and rapid repeated presses across all six appearances. It loads
the running catalog's component CSS and shared controller, samples rendered
layers on animation frames, and checks continuity, ripple overlap, and cleanup.
With the catalog running on port 8081, serve the fixture:

```sh
python3 -m http.server 8082 --bind 127.0.0.1 --directory forge/conformance
```

Open `http://127.0.0.1:8082/feedback-transitions.html` and click **Run transitions**.
The fixture contains representative markup copied from the component gallery.
