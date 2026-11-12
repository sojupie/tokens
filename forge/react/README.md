# Forge for React

`@token-catalog/forge-react` renders Forge’s Button, ButtonLink, IconButton,
Checkbox, TextField, and basic Dialog in React 19. The package carries Forge’s generated token bindings,
component styles, fonts, and button and dialog controllers. An installed
package works without a running Token Catalog or a checkout of this repository.

From the Token Catalog root, build against a local publication:

```sh
npm ci
npm ci --prefix forge/react
npm run build:tokens -- --published published --assets /tmp/forge-react-assets --css-names original
npm run build:react -- /tmp/forge-react-assets
npm pack ./forge/react
```

Install the resulting tarball and import the stylesheet once. Use
`data-component-assets` on the app root or on the container of any controls
rendered into a portal outside that root. For a different theme, set both
`data-theme` and `data-contrast` on the document element, for example
`data-theme="dark" data-contrast="default"`.

```tsx
import { useState } from 'react';
import { Button, Dialog, TextField } from '@token-catalog/forge-react';
import '@token-catalog/forge-react/styles.css';

export function EditName() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');

  return <main data-component-assets>
    <Button onClick={() => setOpen(true)}>Edit name</Button>
    <Dialog title="Edit name" open={open} onOpenChange={setOpen}
      actions={<Button variant="text" data-dialog-close="">Done</Button>}>
      <TextField label="Name" value={name} onChange={event => setName(event.target.value)} />
    </Dialog>
  </main>;
}
```

Button and IconButton accept native button props and Forge appearance options.
Give IconButton an accessible `label` and an icon as its child. To build a toggle,
pass `pressed` and update it in `onClick`. ButtonLink applies the same appearance
options to an anchor. Checkbox accepts native checkbox props, including a ref for
setting the input’s `indeterminate` property. Give it a visible or accessible label.
TextField accepts native input props
with a visible `label`, and can display supporting text, invalid styling, and icons.

Dialog uses the native top layer. Forge handles motion, Escape, backdrop dismissal,
and returning focus to the opener. Dismissal calls `onOpenChange(false)`; the
parent can keep it open while saving. Keep the dialog mounted while its `open`
prop changes so it can finish closing. `onAfterClose` is available for clearing
content after the exit animation. An action marked `data-dialog-close=""` requests
dismissal through the same controller.

The package ships the existing Forge common CSS bundle. It also contains styles
for other primitives. React and the catalog share the CSS and controllers, while
the JSX follows the Go templates’ markup contract.

## Add or change a component

1. Find the matching Go template in [`components/ui/templates`](../components/ui/templates/)
   and its props in [`components/ui/internal/model`](../components/ui/internal/model/).
   Use the [component API reference](../../docs/reference/components-api.md) to find
   the markup, slots, state attributes, and controller for the component.
2. Put the React renderer and its public prop types in [`index.tsx`](index.tsx).
   Accept native props where they fit. Match the Forge classes, slots, and state
   attributes so the same CSS and browser behavior work in both renderers.
3. Make shared style changes in [`components/styles`](../components/styles/) and
   shared behavior changes in [`components/behaviors`](../components/behaviors/).
   Use [`components/modules.json`](../components/modules.json) to find the assets
   a component needs. Add any new generated CSS or script inputs to
   [`build.ts`](build.ts). Build again when a shared input changes.
4. Give each interaction one owner. If React owns selection or a form value,
   write the corresponding attribute from React. If Forge owns the interaction,
   use its existing controller and connect it to React props and events. For
   example, React sets `Dialog.open`, while Forge handles Escape, backdrop
   dismissal, motion, and focus return.

Edit the source files above. The build writes JavaScript, declarations, CSS,
and fonts to `forge/react/dist/`. The tarball contains these files so consumers
can install it on its own. The build requires a verified publication at
`published/` or a path supplied with `--published`. See
[development setup](../../docs/development.md) for creating a publication.

Test the packed tarball in a separate React app. Use the same variant in the
Catalog playground and compare the layout and computed styles. Exercise the
mouse and keyboard, disabled states, and any focus changes introduced by the
component. For a popup, check Escape and the element that receives focus after
closing. Run `npm run check:types` in the Catalog and a production build in the
consumer. If a shared controller changes, run its existing browser tests too.
