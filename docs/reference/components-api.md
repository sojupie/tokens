# Go/templ components

Import `github.com/sojupie/tokens/forge/components/ui` from a Go/templ caller.
Forge owns component anatomy, token bindings, accessibility attributes and browser
mechanics. Catalog owns page layouts, sample content, playground configurations
and application actions. Scenes use the same Forge package as other callers.

The package is part of this repository's Go module. It is not a separately
versioned library. Component CSS and scripts must come from the same build.

## Authoring

Use typed props for normal configuration. Supply content through templ children
or component slots. Supply unique IDs for fields and panels. Use native links for
navigation and native buttons for actions.

```templ
package example

import "github.com/sojupie/tokens/forge/components/ui"

templ Notifications() {
    @ui.TextField(ui.TextFieldProps{
        ID: "display-name", Name: "display_name", Label: "Display name",
        Value: "Maya", Required: true,
    })
    @ui.Switch(ui.ChoiceProps{
        ID: "notifications", Name: "notifications", Label: "Notifications",
        Value: "enabled", Checked: true,
    })
    @ui.Action(ui.ButtonProps{Type: "submit"}) {
        Save changes
    }
}

templ Inbox() {
    @ui.List(ui.ListProps{Label: "Messages", Leading: ui.LeadingAvatar}) {
        @ui.ListLink("/messages/42", ui.ListItemProps{
            Label: "Exhibition opening", Supporting: "The final plan is ready.",
            Avatar: "M", LeadingKind: ui.LeadingAvatar, Unread: true,
            Badge: ui.Badge("small", "", nil),
        })
    }
}
```

`TextFieldProps`, `ChoiceProps`, `SliderProps` and `TimePickerProps` accept form
names and values. A slider's `Value` is its upper value; a non-nil `Start` enables
a range. Progress values are percentages; nil selects indeterminate progress.
Time-picker hours use 0–23 even when the display uses a 12-hour clock. Its hidden
form field updates on OK and dispatches native `input` and `change` events.

`Card`, `CardLink` and `CardButton` distinguish content, navigation and actions.
A link card must not contain another interactive control. Put independent action
buttons in a content card's `Actions` slot. `CardButton` accepts phrasing content
and does not render an actions slot.

`Modal` accepts a title, optional description, body children and an `Actions`
component. `FullScreen` selects the full-screen anatomy; `BottomActions` places
its actions in the bottom action bar. `DialogTrigger` opens it; `DialogClose` supplies a result. `Sheet`
supplies normal header/body/actions slots. `SheetSurface` supplies panel and handle
anatomy when the application needs its own header. Place interactive sheets inside
`SheetHost`; render a trigger alongside the sheet. `Hidden: true` starts a standard
sheet closed. Modal sheets start closed through native dialog state.

`Attrs` and `InputAttrs` extend application hooks or adapt playground states.
Normal callers should use typed props rather than copy `md-*` classes or forced
`data-component-state` values. Existing attribute-based helpers remain compatible;
they are lower-level entry points, not a second component implementation.

## Assets

A render records modules through `ui.WithAssets`. Nested components register their
own modules. A badge inside a list therefore records both `badge` and `list`.
The common module is always included when a component registers dependencies.

```go
assets := ui.Assets{}
var body bytes.Buffer
ctx := ui.WithAssets(request.Context(), assets)
if err := page.Render(ctx, &body); err != nil {
    return err
}
// Resolve assets through Forge's modules.json, write the head, then the body.
```

Buffer before emitting the page head, or perform the host's existing collection
pass. Keep the map scoped to one render; it is not a concurrent global registry.
The host resolves module names using `forge/components/modules.json`, includes
common CSS/runtime first, and serves generated theme bindings with authored CSS.
`WithAssets` collects requirements; it does not generate URLs or insert script
and link elements. Catalog's normal render and nested asset scopes attach it
already. A new server must provide the corresponding asset delivery.

Run `npm run build:components` after browser-behavior changes. Run
`npm run build:tokens -- --published <publication>` after CSS or token-binding
changes. These builds serve different purposes.

## State and events

| Owner | Responsibilities |
| --- | --- |
| Forge controller | Open/close, focus return, Escape, gestures, animation interruption, cleanup and native control synchronization |
| Application | Data loading, unread state, saving, validation policy, deletion and routing |
| Catalog playground | Example values, forced visual states, comparison controls and sample save/discard behavior |
| Server pipeline | Capture completeness, publication integrity and persisted update states |

`MaterialDialogs.open(dialog, opener)` and `MaterialDialogs.close(dialog, result,
reason)` provide programmatic dialog control once its script has loaded. The
controller emits `md-dialog-beforeopen`, `md-dialog-opened`,
`md-dialog-beforeclose` and `md-dialog-closed`. The two `before` events are
cancelable. Close events include the result; `beforeclose` also includes the
reason. Application listeners can prevent dismissal while validating or saving.
Native `dialog.close()` and `form method="dialog"` remain native close paths;
use the controller API for animated, cancelable dismissal.

Sheets emit cancelable `md-sheet-beforeclose` and completed `md-sheet-closed`
events. Their reason is `confirm` or `dismiss`. Closing a sheet does not save or
restore inputs. Tabs emit `md-tab-change` with the selected tab ID and panel ID;
selecting a tab does not remove its badges.

No shared state-machine framework is required. Each controller owns its local
interaction lifecycle. Persisted capture/publication workflows, export request
cancellation and page caches have separate owners and invariants.

## Component families

Every component family has an entry point in `forge/components/ui`. Templates
and props are grouped by component. The six application scenes use these APIs
without `componentCase` or specimen renderers.

| Family | Entry points |
| --- | --- |
| Buttons and icon buttons | `Action`, `ButtonLink`, `IconButton` |
| Button groups and split buttons | `ButtonGroupSurface`, `GroupItem`, `SplitHost`, `SplitButton`, `SplitMenu` |
| FAB and FAB menu | `FAB` with `Extended` when needed, `FABMenu`, `FABMenuItem` |
| Cards, badges and dividers | `Card`, `CardLink`, `CardButton`, `Badge`, `Divider` |
| Chips and choices | `Chip`, `CheckboxField`, `CheckboxControl`, `Switch`, `SwitchControl`, `Radio`, `RadioControl` |
| Text, select and autocomplete fields | `TextField`, `Field` with `Kind: "select"` or `Kind: "autocomplete"` |
| Date and time pickers | `DatePicker`, `DatePickerPanel`, `DateEntryField`, `TimePicker`, `TimePickerPanel` |
| Search and carousel | `Search`, `SearchView`, `SearchBarSurface`, `SearchResult`, `Carousel`, `CarouselItem` |
| Lists | `List`, `ListRow`, `ListButton`, `ListLink`, `ListContent` |
| Navigation | `NavigationRegion`, `NavigationItem`, `NavigationItemContent`, `Drawer`, `DrawerToggle` |
| Tabs and segmented controls | `Tabs`, `Tab`, `TabPanel`, `SegmentedButtons`, `Segment` |
| App bars and toolbars | `AppBar`, `ToolbarSurface` |
| Dialogs and sheets | `Modal`, `DialogTrigger`, `DialogClose`, `SheetHost`, `SheetTrigger`, `Sheet`, `SheetSurface` |
| Progress and loading | `Progress`, `Loading` |
| Feedback | `PlainTooltip`, `Tooltip`, `TooltipContent`, `SnackbarMessage` |
| Menus | `Menu`, `MenuSurface`, `MenuItem`, `MenuEntry` |

The date picker accepts civil dates through `Value`, `End`, `Today` and `Month`.
Use `Name` and `EndName` for its form fields. An empty `Month` follows the selected
date, or today when no date is selected. Its supported range is 1900–2100.
Supply `Today` when the rendered date must be deterministic. Sample dates and
forced states belong to the Catalog adapter.

Composite components use content slots. Put a split button and its `SplitMenu`
inside `SplitHost`; the menu ID uses the same base ID as the button. Carousel
items take application images and labels; the host supplies any header or
navigation controls. Search results, optional search options and page content
belong to the caller. A navigation-button search bar requires the host to provide
its matching navigation dialog. `MenuEntry` adds supporting text and shortcut slots to `MenuItem`. Their attributes carry command IDs, submenu
relationships and checked states; command execution belongs to the application.

Catalog still has adapters for its playground controls and older illustration
fixtures. They choose configurations and sample content. They are not the API for
building application scenes. Keep new screen composition in the application
scene templates, with pane/grid rules from `docs/reference/layout.md`.

The migration preserves the existing CSS implementation. API coverage does not
by itself certify every component against Material guidelines. Source fidelity,
layout and interaction checks still apply.
