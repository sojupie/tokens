# Forge for React

`@token-catalog/forge-react` renders Forge components in React 19 with bundled
token CSS, fonts, and browser controllers. Installed tarballs work standalone
without a running Token Catalog or repo checkout.

## Build and install

From the repository root, build against a local publication:

```sh
npm ci
npm ci --prefix forge/react
npm run build:tokens -- --published published --assets /tmp/forge-react-assets --css-names original
npm run build:react -- /tmp/forge-react-assets
npm pack ./forge/react
```

Install the tarball and import `@token-catalog/forge-react/styles.css` once.
Add `data-component-assets` to the app root (and any portal container outside
it). To switch appearances, set both `data-theme` and `data-contrast` on
`<html>` (e.g. `data-theme="dark" data-contrast="default"`).

```tsx
import { useState } from 'react';
import { Button, Dialog, TextField } from '@token-catalog/forge-react';
import '@token-catalog/forge-react/styles.css';

export function EditName() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');

  return (
    <main data-component-assets>
      <Button onClick={() => setOpen(true)}>Edit name</Button>
      <Dialog
        title="Edit name"
        open={open}
        onOpenChange={setOpen}
        actions={<Button variant="text" data-dialog-close="">Done</Button>}
      >
        <TextField label="Name" value={name} onChange={e => setName(e.target.value)} />
      </Dialog>
    </main>
  );
}
```

## Components

| Component | Notes |
| --- | --- |
| `Button`, `ButtonLink`, `IconButton` | Native button/anchor props + Forge variants. `IconButton` requires `label` and an icon child; pass `pressed` for toggles. |
| `Checkbox` | Native checkbox props (including ref for `indeterminate`) with a visible or accessible label. |
| `TextField`, `TextArea` | Native input/textarea props with `label`, supporting text, invalid state, and icons (`rows` on `TextArea`). Set `--forge-field-padding-block` to adjust vertical padding. |
| `Select`, `ChipSelect` | Take `SelectOption` children; `Select` renders a field, `ChipSelect` renders a filter pill. Share Forge popup positioning and keyboard behavior. |
| `SearchBar` | Native search input with `leading` icon and optional `action` icon button. |
| `SplitButton` | Primary button (`md-split-leading`) + menu trigger (`md-split-trailing`, `aria-expanded`). |
| `ConnectedButtonGroup` | Filled single-selection group (`options`, `value`, `onChange`); Forge handles arrow-key focus. |
| `Fab` | Floating action button; can be passed as `NavigationRail` `action`. |
| `NavigationRail`, `NavigationRailItem` | Active item uses `aria-current="page"`; persistent controls go in `footer`. |
| `List`, `ListItemButton` | Wrap each `ListItemButton` in `<li>`; supports `leading`, `trailing`, `actions` (`data-actions-layout="inline"` on list), and `md-list-expand-content` (`data-expand-inset="inset"`). |
| `SideSheet` | Docked or detached inside `md-sheet-host`; React manages visibility and focus. |
| `Dialog` | Native top layer; Forge handles motion, Escape, backdrop dismissal, and focus return. Keep mounted while `open` changes so exit animations finish (`onAfterClose` fires after exit; `data-dialog-close=""` triggers dismissal). |

## Adding or changing a component

1. Check the matching Go template in [`components/ui/templates/`](../components/ui/templates/),
   props in [`components/ui/internal/model/`](../components/ui/internal/model/),
   and contract in [components API](../../docs/reference/components-api.md).
2. Add the React component and prop types in [`index.tsx`](index.tsx), matching
   Forge's classes, slots, and state attributes.
3. Put shared style or behavior changes in [`components/styles/`](../components/styles/)
   and [`components/behaviors/`](../components/behaviors/), and register any new
   built inputs in [`build.ts`](build.ts).
4. Give each interaction one owner—either React state or a Forge controller
   bridged via props/events.
5. Pack the tarball and verify in a separate React app against the Catalog
   playground (mouse, keyboard, disabled states, popups/focus return,
   `npm run check:types`, and browser tests if a controller changed).
