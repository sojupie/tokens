# Forge component library proposal

Status: deferred draft for discussion; the library migration has not started.
The current release keeps the native implementation and separates its templates,
styles, and behaviors. See [Architecture](architecture.md#component-authoring) for
the maintained source paths. Paths, package names,
and APIs below are proposed. Examples are illustrative rather than runnable
implementations.

Forge should produce a usable Material Design 3 web library, independent of
Catalog. Its build machinery should also support a smaller design system with
different tokens, anatomy, and interaction choices. Razor MVC, React islands,
and SCSS are the first integration environments to prove alongside Catalog.

## Proposed decisions

1. Keep Source, Forge, and Catalog in this repository. Give Forge an independent
   package build that consumes saved token inputs without running Catalog.
2. Use authored component contracts, anatomy, CSS, and TypeScript behavior.
   Generate bindings, integration helpers, metadata, and distributable assets.
3. Ship Web Components as the first browser integration. Keep token CSS usable
   independently and component behavior separate from custom-element registration.
4. Start Button and Menu with native HTML in light DOM, enhanced by custom
   elements. Prove this choice before applying it to the rest of Material.
5. Keep Material implementations explicit. Share browser mechanics when a second
   implementation demonstrates the same need.
6. Migrate one component at a time. Catalog must use the released component API;
   it must not retain a separate implementation of a migrated component.

These decisions preserve an extension path for other outputs. They do not promise
that browser templates, CSS, or TypeScript behavior can generate native platform
components unchanged.

## Current problem

The current ownership map is in [Architecture](architecture.md#component-authoring).
Forge generates token bindings, while Catalog supplies component specimens through
family templates and Go attribute helpers. Component styles and behaviors now have
separate source files. They still use Catalog's preview case model and existing
DOM contracts. Some browser behavior still knows about demo containers and HTMX
lifecycle events. This proposal addresses the remaining library boundary, beyond
the source cleanup for the gallery release.

`forge/build-tokens.mjs` both compiles design assets and assembles Catalog styles.
The root Node package is private build tooling, not an installable component
package. Splitting those files alone would not establish a public library.

The replacement must remove those dependencies for migrated components. A consumer
must not need a Go server, a publication database, gallery attributes, or knowledge
of Catalog's helpers to use a button or menu.

## Ownership and build boundary

| Owner | Maintained inputs | Responsibilities |
| --- | --- | --- |
| Source | Captured payloads, extraction rules, verified snapshots | Preserve values, references, contexts, provenance, and completeness; export DTCG and conversion information. |
| Forge build | Component contract format and output generators | Validate inputs, compile CSS, generate helpers and declarations, assemble packages. |
| Design-system implementation | Component anatomy, bindings, CSS, behavior and evidence references | Define the actual Material or other system's components. |
| Browser utilities | Small TypeScript modules | Focus operations, animation sampling, positioning, cleanup; contain no Material token names or Catalog actions. |
| Catalog | Examples, controls, documentation and illustrations | Consume the package and show supported configurations. |
| Application | Page layout, application state and business actions | Integrate components with routing, forms, Redux and backend APIs. |

```text
saved DTCG + source metadata       authored component definitions
             |                                |
             +--------------+-----------------+
                            v
                   validation and binding
                            |
             +--------------+------------------+
             v              v                  v
        token CSS     component CSS     generated HTML helpers
                            |                  |
                            +---------+--------+
                                      |
                      authored TS behavior + registration
                                      |
                                      v
                         versioned browser package
                                      |
                        Catalog / Razor / React
```

Go continues to own CLI validation and reports. Node continues to own Style
Dictionary and browser builds. TypeScript type-checking supplies author feedback;
the Go validator checks a serialized, versioned contract manifest. It does not
execute TypeScript. Node and Go must share format fixtures to prevent different
interpretations of the manifest.

Split the reusable Forge build from Catalog theme assembly. The existing root
build can orchestrate both. A library build accepts explicit saved token contexts
and their identities; it must not import Catalog CSS or perform a live crawl.
Material value overrides stay in `forge/material/overrides.json`.

## Authored source and generated output

Use one directory per component within each design system. The following is a
target layout, not a request to move the existing tree immediately:

```text
forge/
  build/                         # validators and output generators
  browser/                       # shared browser utilities
  material/
    overrides.json
    components/
      button/
        definition.ts            # contract and finite anatomy tree
        bindings.ts              # explicit source token bindings
        button.css               # layout and state rules
        button.ts                # authored behavior
        button.test.ts
      menu/
        definition.ts
        bindings.ts
        menu.css
        menu.ts
        menu.test.ts
  fixtures/simple-system/        # synthetic second system, not company assets
  packages/material/             # package manifest and export configuration
  integrations/                  # small Catalog, Razor and React consumers
```

Co-locate supplementary source-guidance references with definitions or tests.
Do not duplicate captured evidence in a manually maintained approval registry.
Generated CSS usage and binding reports remain derived from source.

The definition describes public options, content positions, events, supported
states, and structural anatomy. Keep the anatomy format deliberately finite:
native elements, literal attributes, typed attribute inputs, escaped text, and
named content positions. Reject unsupported constructs. Do not add expressions,
arbitrary code strings, async operations, or a new programming language.

For example, a proposed Button definition could contain:

```ts
export const button = defineComponent({
  name: 'button',
  element: 'md3-button',
  properties: {
    appearance: enumProperty(['filled', 'tonal', 'outlined', 'text', 'elevated']),
  },
  anatomy: element('button', {
    class: 'md3-button__control',
    type: input('type', ['button', 'submit', 'reset'], 'button'),
    disabled: booleanInput('disabled'),
  }, [
    content('icon', { optional: true }),
    element('span', { class: 'md3-button__label' }, [text('label')]),
  ]),
});
```

This is a small declarative anatomy format, and therefore a limited DSL. Its
purpose is to emit the same native markup in different host languages. Complex
interaction stays in ordinary TypeScript. The complete definition must also
describe content requirements, native attributes, and applicable states; this
excerpt is not a complete Button contract.

The generator uses that definition for typed helpers, documentation tables, and
registration metadata. It does not infer keyboard behavior from an event list or
invent accessibility from token names. Generated helpers must use each target's
normal escaping for text and attribute values. Rich content is supplied through
typed host-language composition, never interpolated as untrusted raw HTML.

Keep token bindings separate from this anatomy. A Material binding entry records
the source set, property suffix, component role, and applicability. For example,
the existing button generator maps `container.color` to a container role and
`pressed.container.shape` from the size set to a pressed shape role. The new
binding format must preserve that distinction rather than flattening both into
one unqualified property name.

Generated output would include rules such as this illustrative excerpt:

```css
md3-button[appearance="filled"] > .md3-button__control {
  background: var(--md3-button-container, var(--md-EXACT-SOURCE-TOKEN));
}
```

`--md-EXACT-SOURCE-TOKEN` is a placeholder for the real exported variable, not a
new token name. Public hooks such as `--md3-button-container` are implementation
API; captured source variables retain their original identity. Private computed
properties are not public hooks. Required missing bindings fail the build;
deliberate omissions have explicit, source-derived explanations.

## Native HTML and Web Components

For the first implementation, recommend autonomous custom elements around useful
native markup. The generated Button helpers would produce:

```html
<md3-button appearance="filled">
  <button class="md3-button__control" type="submit" name="action" value="save">
    <span class="md3-button__label">Save</span>
  </button>
</md3-button>
```

The native button owns activation, focus, disabled state, form association and
submission. The wrapper owns enhancement and configuration. There is one tab stop
and no duplicate button role on the wrapper. Native attributes belong on the
button; the custom element must not require redundant copies of `disabled`.

Generate the structural markup from the anatomy definition for templ, Razor and
React. The custom-element class binds behavior to that structure without replacing
consumer-owned content. Helpers are recommended, while the documented HTML remains
usable directly. Package CSS styles it before JavaScript registers the element.

Use a plain TypeScript controller with explicit connection and cleanup, then a
small custom-element adapter. Avoid a general controller framework. Connecting
twice must not duplicate listeners; disconnecting cancels animation and observers.
Child changes must have a documented update path. Constructors must not assume
parsed children already exist; initialization must account for parser timing.
See [custom-element lifecycle](https://developer.mozilla.org/en-US/docs/Web/API/Web_components/Using_custom_elements).

This changes the earlier preference for Shadow DOM. Light DOM better fits the
first Razor/SCSS experiment and avoids requiring JavaScript SSR at request time.
It also exposes anatomy as a supported integration surface and permits global CSS
interference. Namespace selectors, keep them scoped to component roots, and test
against hostile global resets. Document which content positions are public and
which details consumers must not override.

Lit remains an alternative when components need substantial owned rendering.
Its default Shadow DOM provides isolation; its SSR package currently uses a
JavaScript server environment and is experimental. Adopting it would need an
explicit Razor rendering strategy. Do not support light and shadow variants of
every component merely to postpone the choice. [Lit Shadow DOM](https://lit.dev/docs/components/shadow-dom/),
[Lit SSR](https://lit.dev/docs/ssr/overview/).

## Behavior and motion contracts

Menu is the stateful proof. Start with an action menu: native trigger, labelled
menu surface, item composition, disabled items, keyboard navigation, typeahead,
outside dismissal and focus restoration. Keep filter listboxes and site navigation
distinct; they must not acquire action-menu semantics just because they look similar.
Use the [WAI menu-button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/)
alongside captured Material guidance.

Define `open`, `show()`, `hide()`, and selection events in the public contract.
An illustrative `md3-select` event carries a stable item value, bubbles, and is
composed. Specify whether events report committed state or request a change;
do not mix those meanings. The initial Menu owns its open state. A React wrapper
may request changes through its methods; a separate controlled-state mode needs
an explicit contract before implementation.

Browser controllers own listeners, focus movement, measured geometry, transient
ARIA state and temporary effects. React owns its authored children and application
data. Controllers must not reorder or replace React-owned content. Generated
React bindings install and remove native event listeners for compatibility with
the supported React version.

Material maps particular transitions to captured motion roles and supporting
guidance. Spring sampling can be shared; the choice of spring, affected property,
direction, delay and interruption policy belongs to the component implementation.
Do not apply one expressive spring to every visual property. Reduced motion,
rapid reversal, interruption and settled geometry are required behavior.

Hidden menus must stay hidden before upgrade. The pilot must document its
no-JavaScript fallback rather than exposing an inert menu trigger as functional.
Use ordinary visible links or actions as the fallback where those actions are
essential. An HTML helper alone does not supply a working non-JavaScript menu.

## Packages and application integration

Use a provisional package name such as `@token-catalog/material`; registry scope
and availability remain to be decided. Expose explicit imports:

```ts
import '@token-catalog/material/theme/light-default.css';
import '@token-catalog/material/button.css';
import '@token-catalog/material/register/button';
```

The package contains compiled ESM, declarations, component CSS, saved theme CSS,
the public manifest, license notices and the needed assets. Per-component imports
must not register every element. Registration entry points have declared package
side effects; metadata and helper imports must be safe in server environments.

Token CSS is available without loading component JavaScript. Runtime theme
switching changes CSS variables, including the six existing Material appearance
contexts. The Catalog palette remains distinct from captured download values.
Fonts and icons need explicit delivery and license decisions; do not assume a
Catalog font stylesheet or Go icon helper exists in the consuming application.

Razor and templ helpers are thin generated markup adapters with native escaping,
not independent behavior implementations. Generate them under explicit output
paths; their exact packaging is part of the pilot. Do not create a NuGet package
or publish an npm package just to demonstrate consumption.

React bindings render the same anatomy and expose typed events and refs. SCSS
controls page composition and documented CSS hooks. Company or agency business
logic remains outside the library. In Litium, retain the existing HTTP service,
request context and Redux actions; the component reports an interaction rather
than calling commerce endpoints itself. [Litium MVC development](https://docs.litium.dev/platform/get-started/developer-introduction/mvc-accelerator-development).

Build a package archive and install it into the integration fixtures. A test that
imports Forge source files directly does not prove that the package is usable.
Record the component version, token input identity, context and generator version
with each build. Token refreshes and upstream changes must not silently change an
already released package. Public anatomy, properties, hooks and events participate
in versioning. Only one compatible definition of each custom-element tag may be
registered per document; handle version conflicts explicitly rather than silently
accepting whichever version loaded first.

## Second design system and future outputs

Use synthetic, repository-owned tokens and Button/Menu definitions for a small
second system. Give it two button appearances, different markup or icon placement,
and non-spring motion. It must have its own token names and component element
prefix. This fixture is not a model of confidential company branding.

Success means the same compiler accepts it without editing the Material
generator, importing Material token names, or pretending unsupported Material
variants exist. Different behavior is allowed. Shared utilities should be extracted
from actual duplication, not required by inheritance from a universal Button.

The serialized contract and anatomy make another HTML helper backend feasible.
They do not represent a complete implementation-neutral UI language. A future
framework-native or native-platform target may reuse tokens, contracts and tests,
but still needs authored rendering and platform effects. Reject unsupported
features instead of silently generating partial components. Mitosis or a richer
compiler can be reconsidered if actual target needs exceed this boundary.

## Implementation sequence and acceptance

Keep each stage reviewable and commit behavior changes separately from generated
assets and documentation where practical.

1. **Button package:** implement the small contract/anatomy format, binding
   validation, native markup helpers, scoped CSS and enhancement. Pack the package
   and consume it outside Catalog. Include native submit/reset, disabled controls,
   pointer cancellation, focus indication and reduced motion checks.
2. **Menu package:** prove the lifecycle, focus, composition, event, positioning,
   viewport constraints and interrupted-animation contracts. Test multiple instances,
   nested content, scroll, keyboard, touch blur ordering and cleanup.
3. **Consumers:** use the packed artifacts in a minimal Razor MVC application and
   a React island with SCSS. Test rendering before registration, delayed scripts,
   native forms, application actions, remounts and event teardown. Verify the
   company's actual Litium/React/toolchain versions before claiming compatibility
   with its storefront; the fixture alone proves the integration pattern.
4. **Second system:** build both fixture components with different tokens and
   anatomy. Change a semantic token and verify consistent output in both consumers.
5. **Catalog migration:** replace Button and Menu specimens with the package and
   generated templ helpers. Verify actual Catalog uses as well as the component lab.
   Remove migrated markup, bindings and initialization from the old path. Avoid
   loading old document-wide handlers against new elements during coexistence.

Completion requires reproducible builds, failure diagnostics with component and
token identities, keyboard and screen-reader review, and browser checks on Chromium,
Firefox and WebKit. Real iOS Safari remains a separate check from desktop WebKit.
Test light/dark and all contrast modes, RTL, zoom, long translations, native forms,
animation interruption and detach/reconnect. Missing browsers or untested device
behavior must remain visible in the implementation report.

The pilot is complete when another developer can install the package, implement
the examples from documentation, trace a visual value to its binding, and release
a change without relying on chat history. It does not constitute full MD3 coverage.
After that proof, migrate the remaining families in separate changes and state
their supported variants and limits.

## Decisions to confirm after the pilot

- Does light DOM remain manageable under realistic storefront CSS, or does a
  specific family justify Shadow DOM and a separate rendering strategy?
- Is the limited anatomy format enough for these helpers without accumulating
  conditionals and custom expression syntax?
- Which browser and React versions will the first release support, and which
  theme, font and icon entry points should be public?
- Who owns package publishing, version approval and release credentials in a
  future company installation? The company should own those resources; agency
  access should not make the agency the sole release owner.

This draft supersedes the proposed extraction sequence in the earlier
[component handover](handover-components-and-exports.md#how-to-offer-a-library-and-adapters).
That handover's description of the existing implementation remains useful.
