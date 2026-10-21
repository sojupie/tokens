# Generated Material Design 3 buttons

`forge/material/generate-buttons.mjs` maps the current Material button color and size
sets to CSS. `npm run build:tokens -- --published published` runs it against
all six appearance contexts and writes `forge/components/static/buttons.css`.
Commit this generated file with changes to its mapping. Build the assets before
building the Go server so its embedded CSS matches the asset manifest.

The site uses these classes:

```html
<button class="md-button md-tonal-button" type="button">Save</button>
<button class="md-button md-outlined-button" type="button"
        data-size="medium" data-shape="square">Reply</button>
<button class="md-button md-filled-button" type="button"
        data-button-toggle aria-pressed="false">Selected</button>
```

Variants are `elevated`, `filled`, `outlined`, `text` and `tonal`. Sizes are
`xsmall`, `small`, `medium`, `large` and `xlarge`. Small and round are the
defaults. The color sets support selected and unselected toggle configurations
for every variant except text. The shared controller updates `aria-pressed`
when a button with `data-button-toggle` is activated.

## Source coverage

A color set supplies container, label, icon, outline and state-layer values.
A size set supplies height, typography, spacing, icon size, shapes and the
pressed-shape spring. The generator combines these sets and generates enabled,
hovered, focused, pressed and disabled rules. Shared state rules inherit the
base values where the source provides no override.

Every source token in those sets must have a known mapping. Required tokens,
state families and positive spring parameters are checked during the build.
Unknown properties, unknown button sets, omitted current button values and
missing required tokens fail the build. Superseded baseline and deprecated
button sets appear as explicit exclusions in the coverage report.

`published/assets/buttons-coverage.json` records source-token mappings,
exclusions and supported configurations. The current capture supplies 295
mapped tokens and 650 combinations of variant, size, shape, selection and state.
The button page at `/md3/components/buttons` uses this configuration list for
its live preview. Compare states, sizes, or variants while fixing the other
settings, or choose Interactive for one live button. Preview query parameters
are described in the [component reference](components.md).

## Interaction behavior

`forge/components/static/buttons.js` delegates events so buttons inserted by HTMX
receive the same behavior. It handles pointer and keyboard presses, release,
cancellation, disabled controls, selection and cleanup after removal or loss of
focus. Disabled anchors have their activation suppressed by the controller.

Pressed shapes use the size set's damping ratio and stiffness. The renderer
uses a unit-mass spring, samples its response and animates the four corner
radii through the Web Animations API. A new press or release starts from the
current rendered shape. These renderer rules are explicit platform mappings;
the source payload contains parameters rather than browser animation code.

Ripple geometry follows the press position and the button's bounds. Keyboard
activation starts at the center. Expansion uses `md.sys.motion.duration.medium1`;
fade and state transitions use `md.sys.motion.duration.short3`, with standard
motion easing. These duration assignments are the web renderer's choices.
State-layer colors and opacities come from the active button tokens.

Reduced-motion preferences disable the animations while preserving immediate
state feedback. JavaScript-free controls retain their CSS states and native
button behavior. Toggle state changes require the controller or application
code.

## Checks

```sh
npm run test:ui
TOKEN_ASSETS_DIR="$PWD/published/assets" npm run test:tokens
TOKEN_CATALOG_THEME_DB="$PWD/published/tokens.sqlite" go test ./...
```

Generator checks cover mapping completeness and failure on missing or unknown
source tokens. Interaction checks cover springs, pointer and keyboard input,
release, cancellation, disablement, reduced motion and animation cleanup.
The gallery allows rendered-property checks across the generated configurations
and all six appearance settings. `forge/conformance/check-button-rendering.mjs` exports
a read-only browser check for colors, opacity, typography, spacing, icons,
outlines, elevation bindings and shapes.

[Other generated components](components.md) cover cards, icon buttons, chips,
fields, menus, dialogs, radio buttons, and dividers. They share the press
controller and shadow renderer with common buttons.
