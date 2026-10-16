# Token Forge roadmap

Historical plan. See [current status](README.md) before following these instructions.


**Status:** Parked indefinitely
**Created:** 2026-08-11  
**Scope:** Greenfield Material 3 token compiler and portable component system

## 1. Goal

Token Forge compiler, canonical token promotion, component implementation, and
package work are paused indefinitely. Source Loom is the active project. This
roadmap records the intended future consumer, but it is not an active execution
plan.

Build a library-first Material 3 system from first principles. Previous
application code supplies measurements, screenshots, behavior expectations,
and failure reports only. It creates no compatibility obligation.

The intended product is:

> Canonical Material 3 foundation + selected Material 3 Expressive component
> revisions + documented platform adaptations.

Skip variable fonts, decorative Expressive shapes, and broad motion work until
a component requires them. Prefer the established Material 3 model where the
newer Expressive revision does not provide a clear, useful web contract.

## 2. Sources of truth

Keep three kinds of truth separate:

1. Source Loom snapshots record what Material published at a point in time.
2. Canonical DTCG files record what Token Forge intentionally supports.
3. Platform code translates canonical decisions into CSS, DOM behavior, and
   adapters.

Source snapshots are evidence, not production build inputs. Extraction can
generate candidates, but only review and promotion can change canonical tokens.

When evidence conflicts, use this order and record the conflict:

1. Current Material token-table data for token identities and values.
2. Current Material documentation for anatomy, states, measurements, and intent.
3. Official Material platform repositories as corroborating evidence.
4. A documented Forge decision where Material is incomplete or unsuitable.

## 3. Token foundation

Use platform-neutral JSON that conforms to the stable DTCG 2025.10 Format,
Color, and Resolver modules. Preserve Material names as nested token paths:

```text
md.ref.*   approved reference values
md.sys.*   semantic system roles and theme remapping
md.comp.*  component element and state roles
```

Reference tokens normally contain literals. System tokens normally alias
reference tokens. Component tokens normally alias system or reference tokens.
Literal component values require upstream evidence or a recorded local decision.

Material component states such as hovered, pressed, selected, and disabled are
component token identities. They are not global resolver contexts. Use resolver
sets for cross-cutting contexts such as light and dark theme.

Keep detailed source records in a provenance registry. Token `$extensions`
refer to compact source IDs instead of copying long URLs into each token.

### Units

- Represent color with DTCG color values, not CSS strings.
- Represent Material dp geometry as DTCG `dimension` values in idealized `px`.
- Preserve the original Material unit in provenance.
- Preserve typography sp during source import until Forge approves a web unit
  policy.
- Represent opacity as `number`, duration as `duration`, and easing as
  `cubicBezier`.
- Keep aliases in canonical source and resolve them for validation or output.

Do not encode HTML, keyboard behavior, selectors, or component anatomy as
tokens. Tokens hold reusable design values, not a UI description language.

## 4. Repository shape

Create directories when their first artifact exists:

```text
Token-Forge/
|-- cmd/tokenc/
|-- internal/token/{parse,ir,resolve,validate,diagnostics}/
|-- internal/emit/{json,css}/
|-- sources/
|   `-- m3-source.lock.json     # historical filename; Source Loom input
|-- tokens/{reference,system,component,themes,adaptations}/
|-- packages/{tokens,css,dom,react}/
|-- apps/lab/
|-- examples/
|-- conformance/
`-- docs/{architecture,decisions,plans}/
```

Do not pre-create empty packages. Keep Go packages internal until an external Go
consumer establishes a stable API.

## 5. Compiler

Implement the compiler in Go:

```text
DTCG documents
    -> schema parsing
    -> typed unresolved IR
    -> reference graph validation
    -> resolver composition
    -> resolved IR
    -> deterministic emitters
```

Initial outputs:

```text
dist/tokens.resolved.json
dist/tokens.css
dist/manifest.json
```

Add token-name TypeScript declarations only when a real package API needs them.
The compiler validates token completeness but does not compile component markup
or behavior.

## 6. Web architecture

Keep the canonical web implementation in three layers:

1. Generated token CSS.
2. Handwritten portable component CSS.
3. Framework-neutral TypeScript DOM behavior.

The public contract consists of semantic HTML, documented classes and data
attributes, CSS custom properties, DOM events, and accessibility semantics.
React is an adapter over this contract, not the owner of it.

Use plain CSS as the portable source. Build tools may resolve imports, validate,
prefix, minify, and create entry points. Consumers must not need Tailwind or
another utility framework.

Use explicit state-layer elements with stable stacking. State-layer color and
opacity are separate tokens. CSS owns color, geometry, typography, transitions,
and state layers. JavaScript owns behavior, accessibility, keyboard handling,
and semantic state attributes.

## 7. Support policy

Define browser support from required capabilities, not from an inherited app:

- CSS custom properties and cascade layers;
- logical properties;
- `:focus-visible`;
- flex and grid gap;
- ES modules and Pointer Events;
- ResizeObserver where components need it.

Choose the exact browser matrix before publishing component CSS. Compatibility
bundles can be added later without weakening the canonical source.

## 8. First vertical slice

The first slice covers the complete button family:

- labelled action button;
- icon-only action button;
- labelled toggle button;
- icon-only toggle button;
- standard groups of related independent controls;
- connected single-select and multi-select toggle groups.

Share private rendering, token resolution, geometry, loading, and state-layer
machinery. Keep semantic public APIs distinct so invalid action, toggle, and
group combinations are hard to express.

Verify controlled state, disabled selected state, roving radio focus, independent
multi-selection, icon-only accessible names, loading suppression, connected
geometry, shape morphs, and selected/unselected color roles.

## 9. Conformance lab

Build a new lab as a test instrument. It consumes only public artifacts and has
no privileged access to component internals.

Provide live controls for variants, size, shape, theme, selection model,
disabled and loading state, surface context, and viewport. Show event history,
focus and pressed diagnostics, ARIA and data attributes, and token provenance.

Prefer bare surfaces. Add a contrasting nested surface only when a color role,
scrim, elevation, or boundary cannot be verified without it. Avoid decorative
cards that hide the component's real surface relationship.

## 10. Testing

Automate stable contracts:

- DTCG parsing, references, resolver behavior, and deterministic output;
- source lock integrity and candidate import diagnostics;
- action activation and disabled or loading suppression;
- controlled toggle transitions and disabled selection preservation;
- single-select exclusivity and multi-select independence;
- accessible names, radio semantics, `aria-pressed`, and roving focus;
- dynamic group ordering, removal, and disabled-item navigation;
- documented semantic data attributes;
- raw DOM and adapter behavior parity.

Use browser conformance tests and the lab for colors, compositing, geometry,
focus rings, spacing, typography, elevation, themes, and responsive behavior.

Do not freeze broad DOM snapshots, internal wrappers, framework utility order,
every pixel, or every possible visual combination.

## 11. Release model

Version these independently:

- pinned Source Loom snapshot;
- canonical Forge token revision;
- compiler version;
- package versions.

Candidate import is review tooling. Published packages contain canonical tokens
and generated artifacts, not raw upstream captures. Application consumers never
run Source Loom or contact the upstream site.

Initial package candidates are:

- `@token-forge/tokens`;
- `@token-forge/css`;
- `@token-forge/dom`;
- `@token-forge/react`.

Confirm package names and registry ownership before release work.

## 12. Execution phases

These phases remain future work. Source Loom does not wait for them.

### Phase 0: charter and repository bootstrap

- Accept the two-repository boundary.
- Add architecture decisions and naming rules.
- Define the support and release policies.
- Pin DTCG 2025.10 as the initial token standard.

**Exit:** project identity and boundaries are documented.

### Phase 1: first approved source snapshot

- Review the Source Loom snapshot schema and coverage report.
- Pin one immutable snapshot by manifest digest.
- Import a loss-preserving candidate graph.
- Establish provenance and review records.

**Exit:** Forge can reproduce the candidate import without the live site.

### Phase 2: canonical foundation

- Promote reviewed reference, system, shape, type, motion, and theme tokens.
- Promote button, icon-button, and button-group component tokens.
- Record Expressive overrides and web adaptations separately.

**Exit:** every first-slice token is traceable and understandable.

### Phase 3: generic compiler

- Parse DTCG groups, types, aliases, `$root`, and `$extends`.
- Validate missing targets, cycles, incompatible types, and resolver inputs.
- Emit byte-stable resolved JSON, CSS custom properties, and a manifest.

**Exit:** identical inputs produce byte-identical outputs and useful errors.

### Phase 4: button implementation

- Define semantic HTML and CSS contracts.
- Implement plain CSS, state layers, DOM behavior, and conformance fixtures.
- Add the first framework adapter after the raw DOM contract is stable.

**Exit:** the full button family works without a framework.

### Phase 5: lab and consumer verification

- Build the interactive conformance lab.
- Integrate the public packages into one realistic application flow.
- Fix density, responsive, form, bundling, and surrounding-CSS failures.

**Exit:** the example consumes only public package surfaces.

### Phase 6: packaging

- Finalize exports, CSS entry points, generated-file checks, and release notes.
- Publish provenance and source-lock information with releases.

**Exit:** external consumers can install and verify the first release.

## 13. Decision gates

Resolve these before their dependent phase:

- package scope and public repository URLs;
- Forge extension namespace;
- Source Loom redistribution and hosting policy;
- first-release DTCG conformance subset;
- typography sp and web-relative-unit policy;
- browser capability matrix;
- first adapter and package registry.

## 14. References

- Material Design 3, [Design tokens overview](https://m3.material.io/foundations/design-tokens/overview), reviewed 2026-08-10.
- Material Design 3, [Building with M3 Expressive](https://m3.material.io/blog/building-with-m3-expressive).
- DTCG, [Format Module 2025.10](https://www.w3.org/community/reports/design-tokens/CG-FINAL-format-20251028/).
- DTCG, [Color Module 2025.10](https://www.w3.org/community/reports/design-tokens/CG-FINAL-color-20251028/).
- DTCG, [Resolver Module 2025.10](https://www.w3.org/community/reports/design-tokens/CG-FINAL-resolver-20251028/).
