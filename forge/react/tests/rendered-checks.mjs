import { componentRadius } from './check-button-rendering.mjs';

// These inspect output in the packed consumer fixture, not CSS declarations.
export function checkLayout() {
  const failures = [];
  let checked = 0;
  const check = (condition, message) => { checked++; if (!condition) failures.push(message); };
  const element = selector => {
    const found = document.querySelector(selector);
    if (!found) throw new Error('Missing fixture: ' + selector);
    return found;
  };
  const rect = selector => element(selector).getBoundingClientRect();
  const list = getComputedStyle(element('#single-list'));
  const item = getComputedStyle(element('#single-item'));
  for (const corner of ['TopLeft', 'TopRight', 'BottomLeft', 'BottomRight']) {
    check(item['border' + corner + 'Radius'] === list['border' + corner + 'Radius'], 'Single-row focus surface follows the ' + corner + ' container corner.');
  }
  const leading = rect('#action-item .md-list-leading');
  const body = rect('#action-item .md-list-body');
  const action = rect('#row-action');
  const row = rect('#action-list > li');
  check(leading.right <= body.left + 0.1 && body.right <= action.left + 0.1, 'Leading content, wrapping text and trailing action do not overlap.');
  check(action.right <= row.right + 0.1 && action.left >= row.left - 0.1, 'Trailing action stays inside the row.');
  check(Math.abs(leading.top + leading.height / 2 - body.top - body.height / 2) < 0.1, 'Wrapped text stays beside the leading icon.');
  const sample = document.createElement('span');
  sample.style.color = 'var(--md-sys-color-on-surface)';
  document.body.append(sample);
  check(getComputedStyle(element('#card-parent .md-card')).color === getComputedStyle(sample).color, 'Card uses its surface foreground inside a differently colored parent.');
  sample.remove();
  const input = element('#field');
  const field = input.closest('.md-outlined-text-field');
  check(!!field, 'Field fixture is mounted.');
  if (field) {
    const style = getComputedStyle(field);
    const ring = field.querySelector('.md-field-focus-outline');
    const expected = Number.parseFloat(style.getPropertyValue('--md-comp-outlined-text-field-focus-indicator-outline-thickness'));
    check(Number.isFinite(expected), 'Captured field focus width is available.');
    check(!!ring && Math.abs(Number.parseFloat(getComputedStyle(ring).borderTopWidth) - expected) < 0.1, 'Focused field outline uses its captured width.');
    check(getComputedStyle(input).outlineStyle === 'none' && style.outlineStyle === 'none', 'The focused field has one visible outline, without a native duplicate.');
  }
  return { checked, failureCount: failures.length, failures };
}

let resting;
const glyph = () => document.querySelector('#glyph-item .md-nav-icon > svg');
export function startGlyph() {
  const icon = glyph();
  if (!icon) throw new Error('Missing Material Symbol fixture.');
  resting = icon.getBoundingClientRect().width;
  if (!(resting > 0)) throw new Error('Glyph fixture is hidden.');
}

export async function checkGlyph(phase) {
  const failures = [];
  let checked = 0;
  const check = (ok, reason) => { checked++; if (!ok && !failures.includes(reason)) failures.push(reason); };
  const icon = glyph(), slot = icon.parentElement;
  const start = performance.now();
  let rect;
  // Sample the transition, including spring undershoot and off-center frames.
  do {
    await new Promise(requestAnimationFrame);
    rect = icon.getBoundingClientRect();
    const bounds = slot.getBoundingClientRect();
    check(rect.width >= resting - 0.1 && rect.height >= resting - 0.1, 'Glyph never shrinks below its resting size.');
    check(Math.abs(rect.x + rect.width / 2 - bounds.x - bounds.width / 2) < 0.1 && Math.abs(rect.y + rect.height / 2 - bounds.y - bounds.height / 2) < 0.1, 'Glyph stays centered in both axes throughout the transition.');
    check(getComputedStyle(icon).transform === 'none', 'Glyph is drawn at its current size without a scaling transform.');
  } while (performance.now() - start < 400);
  const hovered = phase === 'hover' || phase === 'release';
  check(Math.abs(rect.width - resting - (hovered ? 2 : 0)) < 0.1, 'Hover grows the glyph; press and leave restore its resting size.');
  const style = getComputedStyle(icon);
  check(hovered ? style.stroke !== 'none' && Number.parseFloat(style.strokeWidth) === 16 : style.stroke === 'none' || Number.parseFloat(style.strokeWidth) === 0, 'Hover adds weight; press and leave restore resting weight.');
  const selected = document.querySelector('#glyph-item').getAttribute('aria-current') === 'page';
  check(getComputedStyle(icon.querySelector('[data-symbol-variant="filled"]')).display === (selected ? 'inline' : 'none'), 'Selected state uses the filled Material Symbol.');
  return { checked, failureCount: failures.length, failures };
}

export async function checkSplit() {
  const failures = [];
  let checked = 0;
  // Wait for the real press/release animations, whichever mechanism owns them.
  for (let frame = 0; frame < 3; frame++) {
    for (const button of document.querySelectorAll('.md-split-button button')) getComputedStyle(button).borderRadius;
    await new Promise(requestAnimationFrame);
    await Promise.all([...document.querySelectorAll('.md-split-button button')].flatMap(button => button.getAnimations()).map(animation => animation.finished.catch(() => {})));
  }
  for (const [id, trailing] of [['split-primary', false], ['split-menu', true]]) {
    const button = document.getElementById(id), rect = button.getBoundingClientRect(), style = getComputedStyle(button);
    const source = name => style.getPropertyValue('--md-comp-split-button-small-' + name).trim();
    const pressed = button.matches(':active'), selected = trailing && button.getAttribute('aria-expanded') === 'true';
    const state = selected ? 'trailing-button-inner-corner-selected-corner-size' : pressed ? 'inner-corner-pressed-corner-size' : button.matches(':hover') ? 'inner-corner-hovered-corner-size' : 'inner-corner-corner-size';
    const inner = componentRadius(source(state), rect.width, rect.height), outer = componentRadius(source('container-shape'), rect.width, rect.height);
    for (const [corner, expected] of [['TopLeft', trailing ? inner : outer], ['TopRight', trailing ? outer : inner], ['BottomLeft', trailing ? inner : outer], ['BottomRight', trailing ? outer : inner]]) {
      const actual = componentRadius(style['border' + corner + 'Radius'], rect.width, rect.height);
      checked++;
      if (!Number.isFinite(actual) || !Number.isFinite(expected) || Math.abs(actual - expected) > 0.1) failures.push({ id, corner, raw: style['border' + corner + 'Radius'], actual, expected });
    }
  }
  return { checked, failureCount: failures.length, failures };
}
