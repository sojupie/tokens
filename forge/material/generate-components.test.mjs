import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { generateComponents } from "./generate-components.mjs";
import { compile } from "../build-tokens.mjs";

const contract = JSON.parse(
  await readFile(new URL("./component-contract.json", import.meta.url)),
);
function fixture() {
  const document = {},
    values = {};
  let i = 0;
  for (const [set, suffixes] of Object.entries(contract))
    for (const suffix of suffixes) {
      const name = set + "." + suffix;
      document["token" + i++] = {
        $type: "number",
        $value: 1,
        $extensions: {
          "dev.tokencatalog.material": {
            sourceName: name,
            presentation: { setName: set },
          },
        },
      };
      values[name.replaceAll(".", "-")] = "1";
      if (["label-text", "menu-item.label-text", "label-text.font", "title.font", "subtitle.font", "search.label-text.font"].includes(suffix)) values[name.replaceAll(".", "-")+"-letter-spacing"] = "0px";
    }
  return { document, values };
}
test("reviewed contracts map every token or report its explicit exclusion", () => {
  const { document, values } = fixture();
  const { coverage, css } = generateComponents(document, values);
  assert.equal(
    coverage.mapped.length + coverage.excluded.length,
    Object.keys(document).length,
  );
  assert.ok(coverage.mapped.length > 750);
  assert.match(
    css,
    /--_card-pressed-state-opacity: var\(--md-comp-elevated-card-pressed-state-layer-opacity\)/,
  );
  assert.match(
    css,
    /--_card-elevation: var\(--md-comp-elevated-card-focus-container-elevation\)/,
  );
  assert.match(
    css,
    /--_icon-pressed-shape: var\(--md-comp-icon-button-small-pressed-container-shape\)/,
  );
  assert.match(
    css,
    /--_field-focus-color: var\(--md-comp-outlined-text-field-focus-indicator-outline-color\)/,
  );
  assert.match(
    css,
    /--_field-trailing-icon-size: var\(--md-comp-outlined-select-text-field-trailing-icon-size\)/,
  );
  assert.match(
    css,
    /--_card-background-state-opacity: var\(--md-comp-elevated-card-hover-state-layer-opacity\)/,
  );
  assert.doesNotMatch(
    css,
    /--_card-background-state-opacity: var\(--md-comp-elevated-card-pressed/,
  );
  const focus = coverage.mapped.find(
    (r) => r.token === "md.comp.outlined-card.focus.indicator.color",
  );
  assert.doesNotMatch(focus.selector, /data-component-state/);
  for (const row of coverage.mapped.filter((r) =>
    r.token.includes(".narrow.leading-space"),
  )) {
    for (const selector of row.selector.split(", "))
      assert.ok(selector.includes('[data-width="narrow"]'), selector);
  }
  assert.equal(
    coverage.configurations.filter((c) => c.family === "icon").length,
    1200,
  );
  assert.equal(
    coverage.configurations.filter((c) => c.family === "card").length,
    18,
  );
  const shapePosition = (token) =>
    css.indexOf(`var(--md-comp-menus-menu-item-${token})`);
  assert.ok(shapePosition("shape") < shapePosition("first-child-shape"));
  assert.ok(shapePosition("shape") < shapePosition("last-child-shape"));
  assert.ok(
    shapePosition("last-child-shape") < shapePosition("selected-shape"),
  );
  for (const position of ["first", "last"]) {
    const token = `md.comp.menus.menu-item.${position}-child.inner-corner.corner-size`;
    const row = coverage.mapped.find((row) => row.token === token);
    assert.ok(row, token);
    assert.match(row.selector, /:not\(:only-child\)/);
    assert.equal(
      row.property,
      `--_menu-item-${position === "first" ? "bottom" : "top"}-shape`,
    );
    assert.ok(!coverage.excluded.some((row) => row.token === token));
  }
});
test("selected menu colors take precedence over generic interaction colors", () => {
  const { document, values } = fixture();
  const { css, coverage } = generateComponents(document, values);
  const hovered = coverage.mapped.find(row => row.token === "md.comp.menus.vibrant.menu-item.hovered.leading-icon.color");
  assert.match(hovered.selector, /\.md-menu-item:where\(/);
  for (const variant of ["standard", "vibrant"]) {
    for (const role of ["leading-icon", "trailing-icon", "supporting-text", "trailing-supporting-text", "label-text"]) {
      const position = (suffix) => css.indexOf(`var(--md-comp-menus-${variant}-menu-item-${suffix})`);
      for (const state of ["hovered", "focused", "pressed", "disabled"]) {
        assert.ok(position(`selected-${role}-color`) > position(`${state}-${role}-color`), `${variant} ${state} ${role}`);
      }
    }
    assert.ok(css.indexOf(`var(--md-comp-menus-${variant}-menu-item-selected-disabled-label-text-opacity)`) > css.indexOf(`var(--md-comp-menus-${variant}-menu-item-selected-label-text-color)`));
  }
});
test("missing, unknown and unconverted component tokens stop generation", () => {
  const { document, values } = fixture();
  const broken = structuredClone(document);
  delete broken.token0;
  assert.throws(() => generateComponents(broken, values), /Missing required/);
  const unknown = structuredClone(document);
  unknown.new = {
    $type: "number",
    $value: 1,
    $extensions: {
      "dev.tokencatalog.material": {
        sourceName: "md.comp.filled-card.unreviewed.color",
        presentation: { setName: "md.comp.filled-card" },
      },
    },
  };
  assert.throws(
    () => generateComponents(unknown, values),
    /Unknown component token/,
  );
  const invalidSpring = {
    ...values,
    "md-comp-icon-button-small-pressed-container-corner-size-motion-spring-stiffness":
      "0",
  };
  assert.throws(
    () => generateComponents(document, invalidSpring),
    /Invalid component spring/,
  );
  const missing = { ...values };
  delete missing["md-comp-filled-card-container-color"];
  assert.throws(
    () => generateComponents(document, missing),
    /Missing converted/,
  );
  unknown.new.$extensions["dev.tokencatalog.material"].presentation.setName =
    "md.comp.icon-button.new";
  assert.throws(
    () => generateComponents(unknown, values),
    /Unknown component set/,
  );
});
test(
  "captured components build the same bindings in all six appearances",
  { skip: !process.env.TOKEN_ASSETS_DIR },
  async () => {
    let css;
    let bindings;
    for (const appearance of [
      "light-default",
      "light-medium",
      "light-high",
      "dark-default",
      "dark-medium",
      "dark-high",
    ]) {
      const document = JSON.parse(
        await readFile(
          `${process.env.TOKEN_ASSETS_DIR}/dtcg/${appearance}/tokens.tokens.json`,
        ),
      );
      const compiled = await compile(document, ":root", { appTheme: true });
      const generated = generateComponents(document, compiled.values);
      assert.equal(generated.coverage.mapped.length, 3655);
      const names = generated.coverage.mapped.map(({ token }) => token).sort();
      if (bindings !== undefined) assert.deepEqual(names, bindings, appearance);
      bindings = names;
      if (css !== undefined) assert.equal(generated.css, css);
      css = generated.css;
      assert.doesNotMatch(css, /undefined|NaN|\[object Object\]/);
    }
  },
);


test("selection controls map source states, errors, geometry and disabled opacity", () => {
  const {document, values} = fixture();
  const {coverage, css} = generateComponents(document, values);
  for (const family of ["checkbox", "switch"]) {
    assert.equal(coverage.mapped.filter(row => row.token.startsWith(`md.comp.${family}.`)).length, contract[`md.comp.${family}`].length);
    assert.ok(coverage.configurations.some(c => c.family === family && c.state === "disabled" && c.selected));
  }
  assert.ok(css.indexOf('var(--md-comp-switch-with-icon-handle-width)') > css.indexOf('var(--md-comp-switch-unselected-handle-width)'));
  assert.ok(css.indexOf('var(--md-comp-switch-pressed-handle-width)') > css.indexOf('var(--md-comp-switch-with-icon-handle-width)'));
  assert.match(css, /--_switch-handle-width: var\(--md-comp-switch-pressed-handle-width\)/);
  assert.match(css, /--_checkbox-container-opacity: var\(--md-comp-checkbox-unselected-disabled-container-opacity\)/);
  const error = coverage.mapped.find(r => r.token === "md.comp.checkbox.selected.error.hover.container.color");
  assert.ok(error.selector.includes('[aria-invalid="true"]'));
  assert.ok(error.selector.includes('input:checked'));
  assert.ok(css.indexOf('var(--md-comp-checkbox-selected-error-hover-container-color)') > css.indexOf('var(--md-comp-checkbox-selected-hover-container-color)'));
});


test("FABs bind current geometry and colors, expose real states, and retain explicit exclusions", () => {
 const {document,values}=fixture(); const {coverage,css}=generateComponents(document,values);
 for(const family of ["fab","extended-fab"]) {
  const configs=coverage.configurations.filter(c=>c.family===family);
  assert.equal(configs.length,96);
  assert.equal(configs.some(c=>c.state==="disabled"),false);
  assert.ok(coverage.mapped.some(r=>r.token===`md.comp.${family}.medium.container.shape`));
  assert.ok(coverage.mapped.some(r=>r.token===`md.comp.${family}.primary-container.pressed.state-layer.opacity`));
  assert.ok(coverage.excluded.some(r=>r.token.startsWith(`md.comp.${family}.surface.`)));
 }
 assert.match(css,/--_fab-pressed-state-opacity: var\(--md-comp-fab-primary-container-pressed-state-layer-opacity\)/);
 assert.match(css,/--_fab-tracking: var\(--md-comp-extended-fab-medium-label-text-letter-spacing\)/);
});

test("feedback components keep badge variants separate and map action/ripple states", () => {
 const {document,values}=fixture(); const {coverage,css}=generateComponents(document,values);
 const dot=coverage.mapped.find(r=>r.token==="md.comp.badge.size");
 const labeled=coverage.mapped.find(r=>r.token==="md.comp.badge.large.size");
 assert.ok(dot.selector.includes(':not([data-variant="labeled"])'));
 assert.ok(labeled.selector.includes('[data-variant="labeled"]'));
 assert.match(css,/--_support-action-pressed-state-opacity: var\(--md-comp-snackbar-action-pressed-state-layer-opacity\)/);
 assert.match(css,/--_support-icon-icon: var\(--md-comp-snackbar-icon-hover-icon-color\)/);
 assert.ok(coverage.configurations.some(c=>c.family==="tooltip"&&c.variant==="rich"&&c.state==="pressed"));
 assert.ok(coverage.configurations.some(c=>c.family==="snackbar"&&c.variant==="two-lines"));
});

test('navigation uses current source sets, orientation geometry and selection states', () => {
 const {document,values}=fixture();
 const {coverage,css}=generateComponents(document,values);
 assert.ok(coverage.excluded.some(r=>r.token==='md.comp.navigation-bar.container.height'));
 for (const family of ['nav-bar','nav-rail','nav-drawer']) assert.ok(coverage.configurations.some(c=>c.family===family));
 const row=coverage.mapped.find(r=>r.token==='md.comp.nav-bar.container.height');
 assert.equal(row.selector,'.md-navigation-bar');
 assert.match(css,/\.md-navigation-bar \{ --_nav-height: var\(--md-comp-nav-bar-container-height\)/);
 const narrow=coverage.mapped.find(r=>r.token==='md.comp.nav-rail.collapsed.narrow.container.width');
 assert.match(narrow.selector,/data-variant="collapsed"/);assert.match(narrow.selector,/data-width="narrow"/);
 const selected=coverage.mapped.find(r=>r.token==='md.comp.nav-bar.item.active.icon.color');
 assert.match(selected.selector,/aria-current="page"/);
 assert.ok(coverage.excluded.some(r=>r.token==='md.comp.navigation-drawer.bottom.container.shape'));
 assert.ok(!coverage.configurations.some(c=>c.family==='nav-drawer'&&c.variant==='bottom'));
 assert.ok(!coverage.configurations.some(c=>c.family==='nav-rail'&&c.variant==='modal'));
 assert.match(css,/data-layout="modal"/);
 assert.equal(coverage.mapped.find(r=>r.token==='md.comp.nav-rail.item.container.vertical-space').property,'--_nav-item-vertical-padding');
 const horizontal=coverage.mapped.find(r=>r.token==='md.comp.nav-rail.item.horizontal.label-text.font');
 assert.match(horizontal.selector,/data-orientation="horizontal"/);
});

test('tabs and segments map selection, state layers and source indicator geometry',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 for(const family of ['tabs','segmented']) assert.ok(coverage.configurations.some(c=>c.family===family));
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.outlined-segmented-button.')).length,45);
 const height=coverage.mapped.find(r=>r.token==='md.comp.primary-navigation-tab.with-icon-and-label-text.container.height');
 assert.match(height.selector,/data-control="icon-label"/);
 assert.match(css,/--_tab-indicator-shape: var\(--md-comp-primary-navigation-tab-active-indicator-shape\)/);
 const state=coverage.mapped.find(r=>r.token==='md.comp.primary-navigation-tab.active.pressed.state-layer.color');assert.match(state.selector,/aria-selected="true"/);
 assert.match(css,/--_segment-pressed-state-opacity: var\(--md-comp-outlined-segmented-button-pressed-state-layer-opacity\)/);
});

test('sliders map every current size, label style and native-input state',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.slider.')).length,124);
 assert.ok(coverage.configurations.some(c=>c.family==='slider'&&c.control==='range'&&c.size==='xlarge'&&c.orientation==='vertical'));
 for(const size of ['xsmall','small','medium','large','xlarge']) {
  const height=coverage.mapped.find(r=>r.token===`md.comp.slider.${size}.active.track.height`);
  assert.match(height.selector,new RegExp(`data-size="${size}"`));
 }
 assert.match(css,/--_slider-control-width: var\(--md-comp-slider-pressed-handle-width\)/);
 assert.match(css,/--_slider-value-text-font-size: var\(--md-comp-slider-value-indicator-label-label-text-size\)/);
 assert.match(css,/--_slider-label-text-font-size: var\(--md-comp-slider-label-label-text-size\)/);
 assert.doesNotMatch(css,/--_slider[^:]*\./);
 assert.match(css,/--_slider-active-shape-tr: var\(--md-comp-slider-active-track-shape\)/);
});


test("progress contracts map every current and baseline geometry without interaction states",()=>{
 const {document,values}=fixture();
 const {coverage,css}=generateComponents(document,values);
 const rows=coverage.mapped.filter(row=>/^md\.comp\.(progress-indicator|linear-progress-indicator|circular-progress-indicator)\./.test(row.token));
 assert.equal(rows.length,56);
 assert.ok(rows.every(row=>!/:hover|:active|:focus/.test(row.selector)));
 assert.ok(coverage.configurations.some(c=>c.family==="progress"&&c.shape==="wave"&&c.size==="thick"));
 assert.ok(coverage.configurations.some(c=>c.family==="progress"&&c.shape==="four-color"&&c.control==="indeterminate"));
 assert.match(css,/--_progress-wavelength: var\(--md-comp-progress-indicator-linear-indeterminate-active-indicator-wave-wavelength\)/);
 const normal=css.indexOf('--_progress-size: var(--md-comp-progress-indicator-circular-with-wave-size)');
 const thick=css.indexOf('--_progress-size: var(--md-comp-progress-indicator-circular-thick-size)');
 assert.ok(thick>normal,"thick circular size must survive the wave size binding");
});

test("loading mappings preserve current colors, size and container shape with an explicit unused base color",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(row=>row.token.startsWith('md.comp.loading-indicator.')).length,7);
 assert.ok(coverage.configurations.some(c=>c.family==='loading'&&c.variant==='uncontained'&&c.control==='determinate'));
 assert.match(css,/--_loading-size: var\(--md-comp-loading-indicator-active-indicator-size\)/);
 assert.match(css,/--_loading-shape: var\(--md-comp-loading-indicator-container-shape\)/);
 assert.ok(coverage.excluded.some(row=>row.token==='md.comp.loading-indicator.container.color'));
 assert.ok(css.indexOf('--_loading-color: var(--md-comp-loading-indicator-contained-active-indicator-color)')>css.indexOf('--_loading-color: var(--md-comp-loading-indicator-active-indicator-color)'), 'contained foreground must override the base active color');
});


test("app bar mappings retain current and baseline typography, subtitle heights and scroll colors",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.ok(coverage.configurations.some(c=>c.family==='appbar'&&c.variant==='bottom'&&c.control==='fab'));
 assert.ok(coverage.configurations.some(c=>c.family==='appbar'&&c.variant==='baseline-centered'&&c.control==='avatar'));
 assert.match(css,/--_appbar-title-tracking: var\(--md-comp-app-bar-large-flexible-title-font-letter-spacing\)/);
 assert.match(css,/--_appbar-height: var\(--md-comp-app-bar-medium-flexible-with-subtitle-container-height\)/);
 assert.ok(css.indexOf('--_appbar-container: var(--md-comp-app-bar-on-scroll-container-color)')>css.indexOf('--_appbar-container: var(--md-comp-app-bar-container-color)'));
});
test("toolbar disabled states override floating color defaults and all geometries are available",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.ok(coverage.configurations.some(c=>c.family==='toolbar'&&c.variant==='floating'&&c.orientation==='vertical'&&c.size==='medium'));
 assert.ok(!coverage.configurations.some(c=>c.family==='toolbar'&&c.orientation==='vertical'&&c.shape!=='icon'));
 assert.ok(coverage.configurations.some(c=>c.family==='toolbar'&&c.variant==='docked'&&c.control==='vibrant'));
 assert.ok(!coverage.configurations.some(c=>c.family==='toolbar'&&c.variant==='floating-fab'));
 assert.ok(css.indexOf('--_toolbar-item-icon: var(--md-comp-toolbar-vibrant-disabled-icon-color)')>css.indexOf('--_toolbar-item-icon: var(--md-comp-toolbar-floating-vibrant-button-unselected-icon-color)'));
 assert.match(css,/--_toolbar-wrapper-padding: var\(--md-comp-toolbar-floating-container-external-padding\)/);
});


test("sheets map handle opacity, mirrored shape corners, focus rings and action states",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(row=>row.token.startsWith('md.comp.sheet.')).length,41);
 assert.ok(coverage.configurations.some(c=>c.family==='bottom-sheet'&&c.variant==='minimized'&&c.control==='modal'));
 assert.ok(coverage.configurations.some(c=>c.family==='side-sheet'&&c.variant==='detached'&&c.orientation==='start'));
 assert.match(css,/--_sheet-handle-opacity: var\(--md-comp-sheet-bottom-docked-drag-handle-opacity\)/);
 assert.match(css,/--_sheet-action-pressed-state-color: var\(--md-comp-sheet-side-docked-action-pressed-state-layer-color\)/);
 assert.ok(css.indexOf('--_sheet-shape: var(--md-comp-sheet-side-detached-container-shape)')>css.indexOf('--_sheet-shape: var(--md-comp-sheet-side-docked-modal-container-shape)'));
 assert.match(css,/--_sheet-shape-tl: var\(--md-comp-sheet-side-docked-modal-container-shape\)/);
});


test("split buttons map every geometry token and preserve separate hover, pressed and selected targets",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(row=>row.token.startsWith('md.comp.split-button.')).length,61);
 assert.equal(coverage.configurations.filter(c=>c.family==='split').length,100);
 assert.match(css,/--_split-outer: var\(--md-comp-split-button-small-container-shape\)/);
 assert.match(css,/--_split-pressed-inner: var\(--md-comp-split-button-small-inner-corner-pressed-corner-size\)/);
 assert.match(css,/--_split-item-selected-inner: var\(--md-comp-split-button-small-trailing-button-inner-corner-selected-corner-size\)/);
 assert.match(css,/> \.md-button:where\(:where\(:hover/);
});


test("button groups bind connected corners and standard width springs at all sizes",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(row=>row.token.startsWith('md.comp.button-group.')).length,55);
 assert.ok(coverage.configurations.some(c=>c.family==='button-group'&&c.control==='connected'&&c.shape==='multiple'&&c.selected&&c.size==='xlarge'));
 assert.match(css,/--_group-selected-inner: var\(--md-comp-button-group-connected-small-selected-inner-corner-corner-size\)/);
 assert.match(css,/--_group-width-damping: var\(--md-comp-button-group-standard-small-pressed-item-width-motion-spring-dampening\)/);
 assert.match(css,/--_group-width-growth: var\(--md-comp-button-group-standard-small-pressed-item-width-multiplier\)/);
 values['md-comp-button-group-standard-small-pressed-item-width-motion-spring-stiffness']='0';
 assert.throws(()=>generateComponents(document,values),/Invalid component spring/);
});


test("FAB menus map every item and close-button token with independent state profiles",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(row=>row.token.startsWith('md.comp.fab-menu.')).length,117);
 assert.equal(coverage.configurations.filter(c=>c.family==='fab-menu').length,24);
 assert.match(css,/--_fab-menu-close-gap: var\(--md-comp-fab-menu-close-button-between-space\)/);
 assert.match(css,/--_fab-menu-item-font: var\(--md-comp-fab-menu-menu-item-label-text\)/);
 assert.match(css,/--_fab-menu-item-tracking: var\(--md-comp-fab-menu-menu-item-label-text-letter-spacing\)/);
 assert.match(css,/--_fab-menu-item-pressed-state-opacity: var\(--md-comp-fab-menu-primary-container-list-item-pressed-state-layer-opacity\)/);
});


test("lists map captured state, content, expansion and gesture fields with explicit duplicate exclusions",()=>{
 const {document,values}=fixture();const {coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.list.')).length,198);
 assert.equal(coverage.excluded.filter(r=>r.token.startsWith('md.comp.list.')).length,8);
 assert.ok(coverage.configurations.some(c=>c.family==='list'&&c.control==='expand'&&c.expanded&&c.size==='three-line'));
 assert.ok(coverage.configurations.some(c=>c.family==='list'&&c.control==='select'&&c.selection==='single'&&c.width==='video'));
 assert.match(css,/--_list-item-pressed-shape: var\(--md-comp-list-list-item-pressed-container-expressive-shape\)/);
 assert.match(css,/--_list-item-avatar-size: var\(--md-comp-list-list-item-leading-avatar-size\)/);
 assert.match(css,/--_list-reveal-container: var\(--md-comp-list-reveal-list-item-action-icon-button-container-color\)/);
 assert.match(css,/--_list-drop-color: var\(--md-comp-list-reorder-list-item-drop-zone-color\)/);
 values['md-comp-list-md-comp-list-list-item-one-line-container-height']='2px';
 assert.throws(()=>generateComponents(document,values),/Conflicting duplicate list height/);
});

test("carousel maps item states and outline without reviving deprecated tint",()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.carousel-item.')).length,26);
 assert.equal(coverage.excluded.filter(r=>r.token.startsWith('md.comp.carousel-item.')).length,1);
 assert.equal(coverage.configurations.filter(c=>c.family==='carousel').length,120);
 assert.match(css,/--_carousel-item-pressed-state-opacity: var\(--md-comp-carousel-item-pressed-state-layer-opacity\)/);
 assert.match(css,/data-control="outlined".*--_carousel-item-outline-width: var\(--md-comp-carousel-item-with-outline-outline-width\)/);
});

test("search maps baseline and contained geometry, states and individual type metrics",()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.search-')).length,78);
 assert.equal(coverage.excluded.filter(r=>r.token.startsWith('md.comp.search-')).length,6);
 assert.match(css,/--_search-view-height: var\(--md-comp-search-view-contained-full-screen-bar-container-height\)/);
 assert.match(css,/--_search-view-results-shape: var\(--md-comp-search-view-contained-docked-results-shape\)/);
 assert.match(css,/--_search-pressed-state-opacity: var\(--md-comp-search-bar-pressed-state-layer-opacity\)/);
});

test("date pickers map days, years, docked month menus, ranges and input geometry",()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>/^md.comp.date-(picker|input)\./.test(r.token)).length,210);
 assert.equal(coverage.excluded.filter(r=>/^md.comp.date-(picker|input)\./.test(r.token)).length,19);
 assert.ok(coverage.configurations.filter(c=>c.family==='date'&&c.control==='menu-button').every(c=>c.selection===''));
 assert.ok(!coverage.mapped.find(r=>r.token==='md.comp.date-picker.docked.date.container.width').selector.includes(' .md-date-day'));
 assert.ok(!coverage.mapped.find(r=>r.token==='md.comp.date-picker.modal.date.state-layer.width').selector.includes(' .md-date-day'));
 assert.match(css,/--_date-day-state-width: var\(--md-comp-date-picker-docked-date-state-layer-width\)/);
 assert.match(css,/--_date-year-width: var\(--md-comp-date-picker-modal-year-selection-year-container-width\)/);
 assert.match(css,/--_date-range-color: var\(--md-comp-date-picker-modal-range-selection-active-indicator-container-color\)/);
 assert.match(coverage.mapped.find(r=>r.token==='md.comp.date-picker.modal.range-selection.date.in-range.label-text.color').selector,/:not\(:where\(\[data-date-selected="true"\]\)\)/);
 const typo='md.comp.date-picker.modal.range-selection.date.in-range.hover.state-layer.opcaity';values[typo.replaceAll('.','-')]='0.3';
 assert.throws(()=>generateComponents(document,values),/Conflicting duplicate date opacity/);
});

test("time pickers map separate clock, keyboard and period profiles",()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>/^md.comp.time-(picker|input)\./.test(r.token)).length,160);
 assert.equal(coverage.excluded.filter(r=>/^md.comp.time-(picker|input)\./.test(r.token)).length,14);
 const selected=coverage.mapped.find(r=>r.token==='md.comp.time-picker.clock-dial.selected.label-text.color');assert.ok(!selected.selector.includes('aria-pressed'));assert.equal(selected.property,'--_time-dial-selected-label');
 assert.match(css,/--_time-support-color: var\(--md-comp-time-input-time-input-field-supporting-text-color\)/);
 assert.match(css,/--_time-period-height: var\(--md-comp-time-picker-period-selector-horizontal-container-height\)/);
 assert.ok(css.indexOf('var(--md-comp-time-picker-time-selector-24h-vertical-container-width)')>css.indexOf('var(--md-comp-time-picker-time-selector-container-width)'));
 assert.match(css,/--_time-field-outline-width: var\(--md-comp-time-input-time-input-field-focus-outline-width\)/);
});

test('chip profiles keep separate icon slots, selection states and avatar geometry',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 for(const variant of ['assist','suggestion','input'])assert.ok(coverage.configurations.some(c=>c.family==='chip'&&c.variant===variant));
 assert.ok(!coverage.configurations.some(c=>c.family==='chip'&&['assist','suggestion'].includes(c.variant)&&c.selected));
 assert.ok(!coverage.configurations.some(c=>c.family==='chip'&&c.variant==='input'&&c.elevated));
 for(const slot of ['leading','trailing']) {
  const row=coverage.mapped.find(r=>r.token===`md.comp.input-chip.with-${slot}-icon.selected.hover.${slot}-icon.color`);
  assert.equal(row.property,`--_chip-${slot}-icon`);assert.match(row.selector,/data-selected/);assert.match(row.selector,/hover/);
 }
 assert.match(css,/--_chip-avatar-shape: var\(--md-comp-input-chip-with-avatar-avatar-shape\)/);
 assert.match(css,/:has\(> button:focus-visible\)/);
});

test('field profiles map filled indicators and custom popup geometry',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 for(const variant of ['outlined','filled'])for(const family of ['field','select','autocomplete'])assert.ok(coverage.configurations.some(c=>c.family===family&&c.variant===variant));
 for(const name of ['filled-select','outlined-select','filled-autocomplete','outlined-autocomplete']) {
  const row=coverage.mapped.find(r=>r.token===`md.comp.${name}.menu.list-item.selected.container.color`);assert.equal(row.property,'--_field-menu-selected-container');assert.doesNotMatch(row.selector,/data-selected|aria-selected/);
 }
 assert.match(css,/--_field-indicator-height: var\(--md-comp-filled-text-field-focus-active-indicator-height\)/);
 assert.match(css,/--_field-focus-width: var\(--md-comp-filled-text-field-focus-active-indicator-thickness\)/);
 assert.ok(!coverage.excluded.some(r=>r.reason.includes('Native select popup')));
});

test('fullscreen dialogs bind parent geometry and separate scroll-edge elevations',()=>{
 const {document,values}=fixture(),{coverage}=generateComponents(document,values);
 const rows=coverage.mapped.filter(r=>r.token.startsWith('md.comp.full-screen-dialog.'));assert.equal(rows.length,52);
 for(const token of ['header.container.height','action-bar.container.height'])assert.equal(rows.find(r=>r.token.endsWith('.'+token)).selector,'.md-full-screen-dialog');
 assert.match(rows.find(r=>r.token.endsWith('header.on-scroll.container.elevation')).selector,/data-top-scrolled/);
 assert.match(rows.find(r=>r.token.endsWith('action-bar.on-scroll.container.elevation')).selector,/data-bottom-scrolled/);
 assert.ok(coverage.configurations.some(c=>c.family==='dialog'&&c.variant==='full-screen'&&c.orientation==='bottom'&&c.width==='bottom'));
 assert.ok(coverage.configurations.some(c=>c.family==='dialog'&&c.variant==='basic'&&c.expanded&&c.shape==='icon'));
 const dialogs=coverage.configurations.filter(c=>c.family==='dialog');assert.equal(dialogs.length,256);
 for(const variant of ['basic','full-screen'])for(const headline of ['shown','hidden'])assert.ok(dialogs.some(c=>c.variant===variant&&c.dialogHeadline===headline));
 assert.ok(dialogs.some(c=>c.variant==='basic'&&c.shape==='icon-divided'));
});

test('menus bind horizontal geometry, submenu focus and menu icon colors',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 const mapped=coverage.mapped.filter(r=>r.token.startsWith('md.comp.menus.'));
 assert.equal(mapped.length,213);assert.equal(coverage.excluded.filter(r=>r.token.startsWith('md.comp.menus.')).length,0);
 assert.match(css,/data-orientation="horizontal"/);assert.match(css,/--_menu-item-bottom-shape: var\(--md-comp-menus-horizontal-menu-item-pressed-shape\)/);
 assert.match(mapped.find(r=>r.token==='md.comp.menus.standard.icon-button.disabled.icon.color').selector,/aria-pressed|data-selected/);
 assert.equal(mapped.find(r=>r.token==='md.comp.menus.standard.disabled.icon-button.icon.color').property,'--_menu-icon-icon');
 assert.match(css,/data-menu-activity="inactive"/);assert.match(css,/--_menu-item-active-state-opacity: var\(--md-comp-menus-standard-menu-item-active-state-layer-opacity\)/);
 assert.ok(coverage.configurations.some(c=>c.family==='menu'&&c.orientation==='horizontal'&&c.shape==='icon-only'));
});

test('baseline menus use their own source sets and disabled colors override selection',()=>{
 const {document,values}=fixture(),{coverage,css}=generateComponents(document,values);
 assert.equal(coverage.mapped.filter(r=>r.token.startsWith('md.comp.menu.')).length,47);
 assert.equal(coverage.excluded.filter(r=>r.token.startsWith('md.comp.menu.')).length,2);
 for(const row of coverage.mapped.filter(r=>r.token.startsWith('md.comp.menus.')))assert.match(row.selector,/not\(\[data-variant="baseline"\]\)/);
 assert.ok(css.indexOf('var(--md-comp-menu-list-item-disabled-label-text-color)')>css.indexOf('var(--md-comp-menu-list-item-selected-label-text-color)'));
 assert.ok(coverage.configurations.some(c=>c.family==='menu'&&c.variant==='baseline'&&c.shape==='divided'));
});

test('date presentations and range selection remain independent',()=>{
 const {document,values}=fixture(),{coverage}=generateComponents(document,values);
 const dates=coverage.configurations.filter(c=>c.family==='date');
 assert.deepEqual([...new Set(dates.map(c=>c.variant))],['docked','modal','input','full-screen']);
 for(const variant of ['docked','modal','input','full-screen'])for(const dateRange of [false,true])assert.ok(dates.some(c=>c.variant===variant&&c.dateRange===dateRange&&c.control==='day'));
 assert.ok(!dates.some(c=>c.control==='text'));
});
