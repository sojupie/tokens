# Third-party notices

Token Catalog's original code and documentation are licensed under the
Business Source License 1.1 (BSL 1.1, transitioning to Apache License 2.0
three years after publication). The following material retains its upstream copyright
and license terms.

## Material Design 3

Material Design 3 is created by Google. The current collection captures token
payloads published at [m3.material.io](https://m3.material.io/) and transforms
them into searchable tables, context-specific values and JSON downloads.
Token names, source values and upstream design guidance originate with Google.

Google's [Material Design licensing statement](https://m3.material.io/get-started)
identifies Apache 2.0 or CC BY 4.0 for Material content, guidelines, imagery and
resources. Item-specific terms and notices govern the material concerned.
The project's license does not relicense captured payloads, token values,
or upstream material preserved in snapshots and exports.

Source-page and payload URLs are preserved in the captured evidence. Derived
publications regroup and filter values and references. These transformations
are Token Catalog's processing of the source data, not Google's original files.

Material Design and Google names identify the source. Token Catalog is an
independent project and is not affiliated with or endorsed by Google.

## Bundled assets

| Material | Source | License notice |
| --- | --- | --- |
| HTMX 2.0.8 | [bigskysoftware/htmx](https://github.com/bigskysoftware/htmx/tree/v2.0.8) | [Zero-Clause BSD](licenses/HTMX-LICENSE.txt) |
| Roboto font files | [Google Fonts / Roboto](https://github.com/google/fonts/tree/main/ofl/roboto) | [SIL Open Font License 1.1](catalog/web/static/fonts/OFL.txt), Copyright 2011 The Roboto Project Authors |
| Material Web ripple geometry and interaction lifecycle (adapted in `forge/components/static/buttons.js` and generated ripple CSS) | [material-components/material-web](https://github.com/material-components/material-web/tree/main/ripple/internal), Copyright 2022 Google LLC | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Material Web native range clipping geometry (adapted in `forge/components/static/components-base.css`) | [material-components/material-web](https://github.com/material-components/material-web/blob/main/slider/internal/_slider.scss), Copyright 2023 Google LLC | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Material Web field layout (adapted in `forge/components/static/components-base.css`) | [material-components/material-web](https://github.com/material-components/material-web/tree/main/field/internal), Copyright 2021 Google LLC | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Material Web chip action navigation and layout (adapted in `forge/components/static/components.js` and `components-base.css`) | [material-components/material-web](https://github.com/material-components/material-web/tree/main/chips/internal), Copyright 2023 Google LLC | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Material Web elevation shadow formulas (adapted in `forge/components/static/elevation.css`) | [material-components/material-web](https://github.com/material-components/material-web/blob/main/elevation/internal/_elevation.scss), Copyright 2022 Google LLC | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| AndroidX Material 3 navigation label transition (adapted in `forge/components/static/components.js`) | [NavigationItem.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/NavigationItem.kt), Copyright 2024 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| AndroidX Material 3 progress animation timing (adapted in `forge/components/static/progress.js`) | [androidx/androidx](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/ProgressIndicator.kt), Copyright 2020 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| AndroidX Material 3 loading shapes and motion (adapted in `forge/material/reference/LoadingShapes.java`, `forge/components/static/loading.js`, and `loading-shapes.json`) | [MaterialShapes.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/MaterialShapes.kt), [LoadingIndicator.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/LoadingIndicator.kt), Copyright 2024 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| App-bar and floating-toolbar layout guidance | [AppBar.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/AppBar.kt), [FloatingToolbar.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/FloatingToolbar.kt), The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| FAB menu width, opacity and reveal order | [FloatingActionButtonMenu.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/FloatingActionButtonMenu.kt), Copyright 2024 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Standard button-group neighbor compression | [ButtonGroup.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/ButtonGroup.kt), Copyright 2024 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Sheet sizing, scrim and gesture defaults | [SheetDefaults.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/SheetDefaults.kt), [ScrimTokens.kt](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/tokens/ScrimTokens.kt), Copyright 2023 The Android Open Source Project | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |
| Material Symbols SVG paths | [Google Material Design icons](https://github.com/google/material-design-icons) | [Apache License 2.0](catalog/web/MATERIAL-SYMBOLS-LICENSE.txt) |

The existing [Moby license](deploy/MOBY-LICENSE) is preserved with the deployment
files. Docker images and their installed packages retain their respective
licenses.

## Dependencies and further collections

Go dependencies are pinned in [go.mod](go.mod) and retain their upstream licenses.
CSS build dependencies are pinned in [package-lock.json](package-lock.json).
[Style Dictionary](https://github.com/style-dictionary/style-dictionary) uses the
Apache License 2.0; its dependency package includes the upstream license.
When distributing binaries, preserve applicable dependency notices alongside
those for the bundled assets. Container distributions also include system
packages such as Chromium under their own terms.

Further design-system collections need their own source attribution and license
record. Material Design's terms do not apply to another publisher's data.
See [licensing and attribution](docs/licensing.md) for repository and data scope.

The site palette is generated with [Material Color Utilities](https://github.com/material-foundation/material-color-utilities), licensed under Apache License 2.0. The npm dependency includes its upstream license.
