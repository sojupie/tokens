# Toolbar specification

This reference describes the toolbar renderer and its source measurements. The token
inventory and browser results below were recorded on 2026-10-20 against the named
publication. They do not establish verification of later changes.

## Sources and precedence

- [Material specifications](https://m3.material.io/components/toolbars/specs).
- [Material guidelines](https://m3.material.io/components/toolbars/guidelines).
- [Material accessibility](https://m3.material.io/components/toolbars/accessibility).
- Publication `material-2026-10-15-review`; `published-review/assets/dtcg/light-default/tokens.tokens.json`. Theme values change; retain the aliases below.
- Supplied figures: configuration table (1), variant table (2; figure 3 duplicates it), floating measurements (4), floating margins (5), baseline measurements (6).
- Additional official figures inspected: docked internal padding and docked alternative alignment.

Captured values and references remain authoritative for their contexts. Supplemental measurements fill missing fields. A figure does not replace a conflicting token. Conditional guidance and platform exceptions must remain explicit. No token override is required by this review.

## Variants, configurations, and availability

| Variant | M3 baseline | M3 Expressive | Placement |
| --- | --- | --- | --- |
| Docked toolbar | Not a named baseline component | Available | Full-width at the bottom by default |
| Floating toolbar | Unavailable | Available | Above body content; horizontal or vertical |
| Bottom app bar | Available | Supported but discouraged; prefer docked toolbar | Legacy bottom edge layout |

| Configuration | Availability and meaning |
| --- | --- |
| Standard color | Default; baseline equivalent is bottom app bar |
| Vibrant color | Expressive; higher emphasis |
| Horizontal floating | Default; Expressive |
| Vertical floating | Expressive; not a docked orientation |
| FAB beside floating toolbar | Configuration, not a third variant. Compose supports the combination directly; web composes separate toolbar and FAB primitives |
| FAB in baseline bottom app bar | Supported legacy configuration; not the floating arrangement |
| Slots | Buttons, icon buttons, text fields, images, or other appropriate components. Icon/label/icon-and-label examples are demonstrations, not an exhaustive slot model |

Platform APIs can differ: Compose implements floating toolbars separately from docked toolbars and baseline bottom app bars.

## Measurements

Dimensions are dp, represented as CSS px at the normal web scale. Distinguish visible control dimensions from hit targets and system safe areas.

| Attribute | Docked | Floating horizontal | Floating vertical |
| --- | --- | --- | --- |
| Container height | 64 | 64 | Intrinsic to contents |
| Container width | Available window/pane width | Intrinsic; must fit available pane | 64 |
| Corner shape | 0 in captured default | Fully rounded | Fully rounded |
| Leading/trailing padding | 16 / 16 | 8 / 8 | 8 / 8 along the main axis |
| Item spacing | Captured range 4–32; default example 32 | 4 | 4 |
| Cross-axis padding | Center 48-high targets within 64 | 8 for 48-high targets | 8 for 48-wide targets |
| Minimum outside clearance | 16 leading/trailing content clearance | 16 from window/pane edges | 24 from window/pane edges |
| Elevation | No dedicated docked elevation field | Level 3 (6) | Level 3 (6) |
| Target per placed control | At least 48×48 | At least 48×48 | At least 48×48 |

Floating figure 4 specifies 8dp default padding around the container, 4dp between items, 64dp toolbar height, a 56dp FAB, and an 8dp toolbar-to-FAB gap. A small icon button can paint at 40dp while its hit target occupies 48dp; do not make every visual control 48dp tall to meet accessibility.

Figure 5 specifies at least 16dp on either horizontal edge. Its bottom annotations show 16dp to the safe content boundary and 40dp to the physical device edge, including the pictured system gesture area. Do not impose a universal 40dp bottom margin on web: use 16dp plus the actual safe-area inset. The margin encloses the whole toolbar/FAB composition.

Official docked figures show 16dp outer padding and 32dp inter-item gaps, and an alternative with leading/middle/trailing groups separated by flexible space. Alternative examples include edge-aligned Back/Next and centered controls with 8dp gaps. Flexible spacers are a composition choice; they are not a replacement for the captured maximum-gap value.

### Floating FAB

| Attribute | Small/default | Medium |
| --- | --- | --- |
| Container width × height | 56×56 | 80×80 |
| Shape radius | 16 | 20 |
| Icon size | 24 | 28 |
| Rest elevation | Level 1 (1) | Level 2 (3) |
| Toolbar separation | 8 | 8 |
| Standard colors | Secondary container / on secondary container | Same |
| Vibrant colors | Tertiary container / on tertiary container | Same |

The FAB is outside the toolbar container and stays separately focusable. Center the two containers on the cross axis. Keep the complete pair within the pane margins.

### Legacy bottom app bar

Figure 6 specifies an 80dp bar, 4dp leading space before the first 48dp target, and an internal FAB with 12dp top/bottom clearance and 16dp trailing clearance. The 56dp FAB plus two 12dp spaces accounts for 80dp. It supports icon actions with or without that FAB. Keep these measurements in the baseline bottom-app-bar context; do not apply them to the 64dp Expressive toolbar. The capture has `md.comp.bottom-app-bar.container.height = 80dp` and a separate `md.comp.bottom-app-bar.with-fab.container.height = 72dp`. The latter conflicts with the pictured 80dp FAB arrangement. Preserve 72dp in that captured context; with a 56dp FAB, its current centered clearance is 8dp rather than the figure’s 12dp. The 4dp leading and 16dp trailing insets and 48dp targets are supplemental measurements and can be applied without replacing that height.

## Colors and interaction states

Use system roles, not the literal light-theme colors. Standard uses surface-container; vibrant uses primary-container. Captured standard action foregrounds use on-surface-variant and selected actions use secondary-container/on-secondary-container. Vibrant action foregrounds use on-primary-container; selected actions use surface-container/on-surface. Disabled labels/icons use on-surface at 0.38 opacity. Hover/focus/press layers retain their captured roles and 0.08/0.10/0.10 opacity.

The official color illustration also shows optional filled actions in primary/on-primary and standard actions in primary. These are child button style examples, not replacements for the toolbar's captured default on-surface-variant foreground. Keep child-style choices separate from the toolbar color configuration.

## Usage, placement, adaptation, and behavior

Use docked controls for recurring page-spanning actions and floating controls for contextual work. Avoid competing bottom navigation; local toolbar navigation must respect the page hierarchy. Primary destinations belong in navigation; toolbar actions usually appear on subsequent pages. Prioritize essential controls and avoid overwhelming action sets. Emphasize one action, not several; mixed slots should remain coherent. Avoid square filled controls in rounded floating containers, and wide controls in vertical layouts.

Floating containers must remain inside their pane: move trailing actions into a menu, revealing them as space returns; optional width caps are valid. Docked controls distribute in compact layouts; larger layouts can center or separate groups. Web/large-screen exceptions permit other docked placement, rounding, dividers, and smaller height/XS controls when space requires it. Keep these exceptions opt-in.

Prefer vertical floating controls at larger breakpoints, opposite a centered navigation rail. Compact layouts usually need one horizontal toolbar; vertical use needs few controls or simple content. Larger layouts may split related actions across opposite edges. Mirror directional content and meaningful order in RTL.

Scroll behavior is configurable: docked controls can persist or leave; floating controls can persist, leave, or condense. Compose supports condensation into a FAB/key action. Never condense and slide away simultaneously. Floating elevation can be omitted over a clearly distinct background.

These rules describe layout behavior; they do not add variants.

## Accessibility and input

The web container uses `role="toolbar"`, a meaningful accessible name, and orientation metadata. The container itself has no action. Each placed component retains its own accessibility contract and label. Targets are at least 48×48dp. Assistive technology must reach every action, menus and back controls, including after scrolling or collapsing.

Initial focus goes to the first enabled interactive control. Tab or the appropriate arrows navigate controls; Space/Enter activate them. Supply touch/click ripples and hover feedback. Disabled controls must not activate. A single Tab stop with arrows within the group is the current renderer convention; Tab traversal is also described by the official page, so verify the chosen behavior rather than treating either as a token property. Menus need keyboard access and focus return. RTL reverses directional arrow traversal.

## Renderer behavior

The renderer exposes Docked and Floating, with a separate None/Small/Medium FAB choice. Old `floating-fab` URLs remain supported. Variant comparisons use horizontal toolbars without a FAB so both variants appear; FAB-size comparisons use Floating. Hidden incompatible controls cannot imply unsupported combinations.

Previews place the actual primitives within a bounded page. Docked bars fill its bottom edge. Floating bars use their intrinsic content size, 16dp horizontal or 24dp vertical clearance, and an additional real bottom safe-area inset. Child visuals keep their captured size while hit regions occupy at least 48dp. Trailing actions move to the shared menu when the pane cannot fit them, for both variants. The menu delegates to the original action and returns focus; widening restores controls. Vibrant docked color composes the generic vibrant color token with docked geometry; the dedicated docked default color applies to Standard. No token value changes.

Roving toolbar focus remains the web convention, with oriented arrows/Home/End, Space/Enter, and a separate FAB Tab stop. Optional application compositions remain documented rather than exposed as extra variants: custom fields/images, grouped/flexible layouts, alternative child styles, reduced elevation, scroll policies, multiple toolbars and web-only shape exceptions. They require application context. The preview does not claim to demonstrate all of them.

## Captured token inventory

The following inventory preserves all 112 toolbar entries with a DTCG value in the reviewed light-default publication. Three additional shape entries are extension-only: `md.comp.toolbar.standard.container.shape`, `md.comp.toolbar.vibrant.container.shape`, and `md.comp.toolbar.floating.container.shape`. Each references `md.sys.shape.corner.full` and preserves `SHAPE_FAMILY_CIRCULAR` in its source extension; generated CSS uses a fully rounded shape. Do not treat these extension-only entries as absent capture data. The renderer maps 106 semantic fields because separate corner values are combined into shapes. Other themes resolve aliases differently. Object values are dimensions; shape corner fields are separate source values. Generated CSS also exposes composite shape variables. No completeness claim extends beyond this publication.

| Token | Captured value or reference |
| --- | --- |
| `md.comp.toolbar.docked.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.docked.container.height` | `{"value":64,"unit":"px"}` |
| `md.comp.toolbar.docked.container.leading-space` | `{md.sys.measurement.space200}` |
| `md.comp.toolbar.docked.container.max-spacing` | `{"value":32,"unit":"px"}` |
| `md.comp.toolbar.docked.container.min-spacing` | `{"value":4,"unit":"px"}` |
| `md.comp.toolbar.docked.container.shape.bottomLeft` | `{"value":0,"unit":"px"}` |
| `md.comp.toolbar.docked.container.shape.bottomRight` | `{"value":0,"unit":"px"}` |
| `md.comp.toolbar.docked.container.shape.topLeft` | `{"value":0,"unit":"px"}` |
| `md.comp.toolbar.docked.container.shape.topRight` | `{"value":0,"unit":"px"}` |
| `md.comp.toolbar.docked.container.trailing-space` | `{md.sys.measurement.space200}` |
| `md.comp.toolbar.floating.container.between-space` | `{md.sys.measurement.space50}` |
| `md.comp.toolbar.floating.container.elevation` | `{md.sys.elevation.level3}` |
| `md.comp.toolbar.floating.container.external-padding` | `{"value":16,"unit":"px"}` |
| `md.comp.toolbar.floating.container.height` | `{"value":64,"unit":"px"}` |
| `md.comp.toolbar.floating.container.leading-space` | `{md.sys.measurement.space100}` |
| `md.comp.toolbar.floating.container.trailing-space` | `{md.sys.measurement.space100}` |
| `md.comp.toolbar.floating.fab.between-space` | `{md.sys.measurement.space100}` |
| `md.comp.toolbar.floating.fab.container.elevation` | `{md.sys.elevation.level1}` |
| `md.comp.toolbar.floating.fab.container.height` | `{"value":56,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.container.shape.bottomLeft` | `{"value":16,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.container.shape.bottomRight` | `{"value":16,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.container.shape.topLeft` | `{"value":16,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.container.shape.topRight` | `{"value":16,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.container.width` | `{"value":56,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.icon.size` | `{"value":24,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.elevation` | `{md.sys.elevation.level2}` |
| `md.comp.toolbar.floating.fab.medium.container.height` | `{"value":80,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.shape.bottomLeft` | `{"value":20,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.shape.bottomRight` | `{"value":20,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.shape.topLeft` | `{"value":20,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.shape.topRight` | `{"value":20,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.container.width` | `{"value":80,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.medium.icon.size` | `{"value":28,"unit":"px"}` |
| `md.comp.toolbar.floating.fab.standard.container.color` | `{md.sys.color.secondary-container}` |
| `md.comp.toolbar.floating.fab.standard.icon.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.floating.fab.vibrant.container.color` | `{md.sys.color.tertiary-container}` |
| `md.comp.toolbar.floating.fab.vibrant.icon.color` | `{md.sys.color.on-tertiary-container}` |
| `md.comp.toolbar.floating.horizontal.container.external-space` | `{md.sys.measurement.space200}` |
| `md.comp.toolbar.floating.horizontal.container.height` | `{"value":64,"unit":"px"}` |
| `md.comp.toolbar.floating.standard.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.floating.vertical.container.external-space` | `{md.sys.measurement.space300}` |
| `md.comp.toolbar.floating.vertical.container.width` | `{"value":64,"unit":"px"}` |
| `md.comp.toolbar.floating.vibrant.button.selected.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.floating.vibrant.button.selected.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.floating.vibrant.button.selected.text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.floating.vibrant.button.unselected.icon.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.floating.vibrant.button.unselected.text.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.floating.vibrant.container.color` | `{md.sys.color.primary-container}` |
| `md.comp.toolbar.standard.button.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.standard.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.standard.disabled.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.standard.disabled.icon.opacity` | `0.38` |
| `md.comp.toolbar.standard.disabled.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.standard.disabled.label-text.opacity` | `0.38` |
| `md.comp.toolbar.standard.focused.icon.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.focused.label-text.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.focused.state-layer.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.focused.state-layer.opacity` | `{md.sys.state.focus.state-layer-opacity}` |
| `md.comp.toolbar.standard.hovered.icon.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.hovered.label-text.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.hovered.state-layer.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.hovered.state-layer.opacity` | `{md.sys.state.hover.state-layer-opacity}` |
| `md.comp.toolbar.standard.icon.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.label-text.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.pressed.icon.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.pressed.label-text.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.pressed.state-layer.color` | `{md.sys.color.on-surface-variant}` |
| `md.comp.toolbar.standard.pressed.state-layer.opacity` | `{md.sys.state.pressed.state-layer-opacity}` |
| `md.comp.toolbar.standard.selected.button.container.color` | `{md.sys.color.secondary-container}` |
| `md.comp.toolbar.standard.selected.focused.icon.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.focused.label-text.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.focused.state-layer.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.hovered.icon.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.hovered.label-text.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.hovered.state-layer.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.icon.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.label-text.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.pressed.icon.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.pressed.label-text.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.standard.selected.pressed.state-layer.color` | `{md.sys.color.on-secondary-container}` |
| `md.comp.toolbar.vibrant.button.container.color` | `{md.sys.color.primary-container}` |
| `md.comp.toolbar.vibrant.container.color` | `{md.sys.color.primary-container}` |
| `md.comp.toolbar.vibrant.disabled.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.disabled.icon.opacity` | `0.38` |
| `md.comp.toolbar.vibrant.disabled.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.disabled.label-text.opacity` | `0.38` |
| `md.comp.toolbar.vibrant.focused.icon.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.focused.label-text.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.focused.state-layer.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.focused.state-layer.opacity` | `{md.sys.state.focus.state-layer-opacity}` |
| `md.comp.toolbar.vibrant.hovered.icon.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.hovered.label-text.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.hovered.state-layer.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.hovered.state-layer.opacity` | `{md.sys.state.hover.state-layer-opacity}` |
| `md.comp.toolbar.vibrant.icon.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.label-text.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.pressed.icon.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.pressed.label-text.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.pressed.state-layer.color` | `{md.sys.color.on-primary-container}` |
| `md.comp.toolbar.vibrant.pressed.state-layer.opacity` | `{md.sys.state.pressed.state-layer-opacity}` |
| `md.comp.toolbar.vibrant.selected.button.container.color` | `{md.sys.color.surface-container}` |
| `md.comp.toolbar.vibrant.selected.focused.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.focused.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.focused.state-layer.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.hovered.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.hovered.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.hovered.state-layer.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.pressed.icon.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.pressed.label-text.color` | `{md.sys.color.on-surface}` |
| `md.comp.toolbar.vibrant.selected.pressed.state-layer.color` | `{md.sys.color.on-surface}` |

## Verification requirements

Exercise both variants, each supported color, both floating orientations, FAB absent/small/medium, icon/label/icon-and-label content, selection, all five interaction states, RTL, and narrow/wide panes. Invalid combinations must disappear or normalize coherently. Check computed source bindings and actual geometry, not just generated declarations. Test overflow activation and focus return, target hit-testing outside painted controls, disabled boundaries, theme changes, safe-area clearance, and no horizontal toolbar scrolling.

Keep the baseline bottom-app-bar check separate. Preserve captured token values, aliases, and contexts. Record any remaining implementation limit here rather than implying it was verified.

### Recorded verification boundary (2026-10-20)

The rendered matrix covers all 300 generated Web combinations in the default theme. Additional live checks cover compact panes, overflow selection and focus return, restoration after resizing, vertical geometry, a dark-theme sample, baseline 80dp/72dp contexts, and start-page reuse. RTL arrow logic has a unit check; an RTL page and a device with a nonzero safe-area inset were not rendered in this review. Their logical CSS and safe-area formula remain explicit above. The legacy baseline height conflict is intentionally unresolved.
