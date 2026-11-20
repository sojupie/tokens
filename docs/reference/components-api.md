# Go/templ components

Import `github.com/sojupie/tokens/forge/components/ui` from Go/templ callers.
Forge owns component anatomy, token bindings, accessibility attributes, and
browser controllers; callers own layouts, content, and application actions.
The package ships inside this Go module and must match the CSS and scripts from
the same build.

## Authoring basics

Use typed props for configuration, templ children/slots for content, unique IDs
for fields and panels, native links for navigation, and native buttons for
actions. `Attrs` and `InputAttrs` pass consumer hooks and merge `Attrs["class"]`
with component classes.

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

### Buttons and groups
- `ButtonProps.Color` accepts `primary`, `secondary`, `tertiary`, or `error`
  independently of `Variant` (`data-color="..."` in HTML). Tonal buttons use the
  family's container/on-container pair; other styles remap accent roles.
- Connected button groups fill inline width by default; set `GroupProps.Width`
  to `GroupHug` or `GroupFill` (standard groups default to `GroupHug`). Cap
  width via `max-inline-size` or `--forge-button-group-max-inline-size`.

### Forms, fields, and pickers
- `TextFieldProps`, `ChoiceProps`, `SliderProps`, and `TimePickerProps` accept
  form names and values. Slider `Value` is the upper value (non-nil `Start`
  enables a range). Progress values are percentages (`nil` = indeterminate).
- `SelectField` renders outlined or chip-style selects with separate display
  labels and submitted values; `Field` uses the same option list for select and
  autocomplete. `Select` keeps its hidden native `<select>` for compatibility.
- **Date pickers:** Accept civil dates via `Value`, `End`, `Today`, and `Month`
  (range 1900–2100; form fields `Name` / `EndName`).
- **Time pickers:** Use 0–23 hours internally and submit `HH:MM`, dispatching
  native `input` and `change` events on commit. Use `DialogTrigger` with
  `<picker ID>-time` for a standalone modal panel. `InlineContainer` adds a
  filled surface to inline pickers; `Width` accepts `TimePickerHug` or
  `TimePickerFill` (default).
- **ScrollField:** Accepts `Options` or `Items` (label + value), emits
  `md-scroll-field-change`, and supports `DirectEntry` paired with
  `ScrollFieldModeButton(ids, textEntry, props)`.
- Hosts replacing fragments outside HTMX call `initialize(scope)` after insert
  and `dispose(scope)` before removal on `MaterialTimePickers`,
  `MaterialScrollFields`, and `MaterialDialogs`.

### Cards, dialogs, and sheets
- Use `Card` (content + `Actions` slot), `CardLink` (navigation; never nest
  interactive controls), or `CardButton` (phrasing content, no actions slot).
- `Modal` takes a title, description, body children, and `Actions`
  (`FullScreen` / `BottomActions` for full-screen dialogs; `DialogProps.FormAttrs`
  makes the content region a native `<form>`). Open with `DialogTrigger`, close
  with `DialogClose`.
- Place interactive `Sheet` or `SheetSurface` inside `SheetHost` with a
  trigger (`Hidden: true` starts a standard sheet closed). Set
  `SheetProps.AllowFullHeight` on modal bottom sheets to add a `full` anchor
  alongside `default` and `minimized` (tracked in `data-sheet-size`).

### Compound components and host hooks
- **Navigation:** `NavigationRail` owns its header, FAB, and destination region
  (`PersistentMenu` reserves the slot for immersive `RailFrame`). Selection does
  not clear badges. `NavigationItemProps.IconContent` accepts a custom icon;
  child content renders a trailing indicator.
- **Lists:** `ListHost` provides live announcements and emits `md-list-change`.
  `ListRowFrameProps.Actions` keeps trailing controls (`IconButton`, `Menu`)
  outside the primary row button; `data-list-primary-action` activates a
  trailing button from the row. Use `ListItemProps.Trailing` for decorative
  content.
- **Search:** Emits `md-search-select` (`detail.label`) and `md-search-profile`.
  Mark inert background regions with `data-search-background` and adaptive
  viewports with `data-search-viewport`.
- **App bars, menus, toolbars, loading:** `AppBarProps.ScrollTarget` names a
  scroll container (or `document`). `MenuHost` coordinates commands/submenus.
  `data-toolbar-viewport` identifies available space for floating toolbars. Use
  `Loading` for shared loading geometry.

## Component families

| Family | Entry points |
| --- | --- |
| Buttons and icon buttons | `Action`, `ButtonLink`, `IconButton` |
| Button groups and split buttons | `ButtonGroupSurface`, `GroupItem`, `SplitHost`, `SplitButton`, `SplitMenu` |
| FAB and FAB menu | `FAB`, `FABMenu`, `FABMenuItem` |
| Cards, badges, dividers | `Card`, `CardLink`, `CardButton`, `Badge`, `Divider` |
| Chips and choices | `Chip`, `CheckboxField`, `CheckboxControl`, `Switch`, `SwitchControl`, `Radio`, `RadioControl` |
| Text, select, autocomplete | `TextField`, `SelectField`, `Field` (`Kind: "select"` / `"autocomplete"`) |
| Date and time pickers | `DatePicker`, `DatePickerPanel`, `DateEntryField`, `TimePicker`, `TimePickerPanel`, `ScrollField`, `ScrollFieldModeButton` |
| Search and carousel | `Search`, `SearchView`, `SearchBarSurface`, `SearchResult`, `Carousel`, `CarouselItem` |
| Lists | `ListHost`, `List`, `ListRow`, `ListButton`, `ListLink`, `ListContent` |
| Navigation | `NavigationRail`, `RailToggle`, `NavigationRegion`, `NavigationItem`, `NavigationItemContent`, `Drawer`, `DrawerToggle` |
| Tabs and segmented controls | `Tabs`, `Tab`, `TabPanel`, `SegmentedButtons`, `Segment` |
| App bars and toolbars | `AppBar`, `ToolbarSurface` |
| Dialogs and sheets | `Modal`, `DialogTrigger`, `DialogClose`, `SheetHost`, `SheetTrigger`, `Sheet`, `SheetSurface` |
| Progress and loading | `Progress`, `Loading` |
| Feedback | `PlainTooltip`, `Tooltip`, `TooltipContent`, `SnackbarMessage` |
| Menus | `Menu`, `MenuSurface`, `MenuItem`, `MenuEntry`, `MenuHost` |

## Assets

Components register their module dependencies via `ui.WithAssets` (including
nested components; `common` is always included):

```go
assets := ui.Assets{}
var body bytes.Buffer
ctx := ui.WithAssets(request.Context(), assets)
if err := page.Render(ctx, &body); err != nil {
    return err
}
// Resolve assets through forge/components/modules.json, write <head>, then body.
```

Keep `ui.Assets` scoped to one render and buffer the body before writing `<head>`.
Run `npm run build:components` after behavior changes and
`npm run build:tokens -- --published <dir>` after CSS/token-binding changes.

## State and events

| Owner | Responsibilities |
| --- | --- |
| Forge controller | Open/close, focus return, Escape, gestures, animation interruption, cleanup, native sync |
| Application | Data loading, unread state, saving, validation policy, deletion, routing |
| Catalog playground | Sample values, forced visual states, comparison controls |

- **Dialogs:** `MaterialDialogs.open(dialog, opener)` and
  `MaterialDialogs.close(dialog, result, reason)` emit `md-dialog-beforeopen`
  (cancelable), `md-dialog-opened`, `md-dialog-beforeclose` (cancelable,
  includes `result` and `reason`), and `md-dialog-closed`.
- **Sheets:** Emit `md-sheet-beforeclose` (cancelable) and `md-sheet-closed`
  with reason `confirm` or `dismiss`.
- **Tabs:** Emit `md-tab-change` with the selected tab and panel IDs.
