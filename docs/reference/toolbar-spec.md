# Toolbar specification

Measurements, color roles, accessibility, and renderer behavior for Material
Design 3 toolbars. Captured tokens are authoritative for their contexts;
supplemental measurements below fill fields absent from the token tables.

Sources: [Material specs](https://m3.material.io/components/toolbars/specs),
[guidelines](https://m3.material.io/components/toolbars/guidelines), and
[accessibility](https://m3.material.io/components/toolbars/accessibility).

## Variants and configurations

| Variant | Placement | Notes |
| --- | --- | --- |
| Docked toolbar | Full-width at the bottom edge | Expressive; replaces legacy bottom app bar |
| Floating toolbar | Above body content (horizontal or vertical) | Expressive; supports optional separate FAB |
| Bottom app bar | Legacy bottom edge | Baseline; supported for compatibility |

- **Color profiles:** Standard (`surface-container`) and Vibrant (`primary-container`).
- **FAB beside floating toolbar:** Composed as a separate sibling control beside
  the floating toolbar, not a third variant.

## Measurements (dp / CSS px)

| Attribute | Docked | Floating horizontal | Floating vertical |
| --- | --- | --- | --- |
| Container height | 64 | 64 | Intrinsic |
| Container width | Pane width | Intrinsic (fits pane) | 64 |
| Corner shape | `0` | Full (`9999px`) | Full (`9999px`) |
| Main-axis padding | 16 / 16 | 8 / 8 | 8 / 8 |
| Item spacing | 4–32 (default 32) | 4 | 4 |
| Cross-axis padding | Centers 48px targets in 64px | 8 | 8 |
| Outside clearance | 16 horizontal | 16 from pane edges | 24 from pane edges |
| Elevation | None | Level 3 (6dp) | Level 3 (6dp) |
| Hit target per control | ≥ 48 × 48 | ≥ 48 × 48 | ≥ 48 × 48 |

- Small icon buttons paint at 40dp inside a 48dp hit target.
- Bottom margin on web uses 16dp plus `env(safe-area-inset-bottom)` rather than
  a hardcoded 40dp mobile gesture inset.
- **Floating FAB:** Small/default is 56 × 56 (radius 16, icon 24, Level 1);
  Medium is 80 × 80 (radius 20, icon 28, Level 2). Both sit 8dp from the
  toolbar, centered on the cross axis, using `secondary-container` (Standard) or
  `tertiary-container` (Vibrant).
- **Legacy bottom app bar:** `md.comp.bottom-app-bar.container.height` is 80dp,
  while `md.comp.bottom-app-bar.with-fab.container.height` is 72dp (giving 8dp
  centered clearance around a 56dp FAB instead of the 80dp diagram's 12dp).
  Captured heights are preserved; supplemental 4dp leading, 16dp trailing, and
  48dp targets apply without overriding them.

## Colors and states

- **Standard:** `surface-container` background, `on-surface-variant` actions,
  `secondary-container` / `on-secondary-container` selected actions.
- **Vibrant:** `primary-container` background, `on-primary-container` actions,
  `surface-container` / `on-surface` selected actions.
- **States:** Disabled uses `on-surface` at `0.38` opacity; hover/focus/press
  state layers use `0.08` / `0.10` / `0.10`.
- `md.comp.toolbar.{standard,vibrant,floating}.container.shape` are
  extension-only circular tokens (`SHAPE_FAMILY_CIRCULAR`) rendered as fully
  rounded in CSS.

## Accessibility and renderer behavior

- Container uses `role="toolbar"`, an accessible name, and `aria-orientation`.
- Roving focus moves across controls with oriented Arrow keys, Home, and End;
  Space/Enter activate; floating FAB remains a separate Tab stop; RTL mirrors
  directional order.
- Trailing actions that don't fit the pane collapse into the shared overflow
  menu (delegating to the original action and restoring focus on close) and
  reappear when the pane widens.
- Legacy `floating-fab` preview URLs remain supported.
