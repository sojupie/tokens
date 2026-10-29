import {test} from 'node:test';
import assert from 'node:assert/strict';
import {circularShapeRadius} from './circular-shapes.mjs';

test('circular rendering uses the same captured box for selected and unselected shapes', () => {
  const values = {'md-comp-button-small-container-height': '40px'};
  const expected = '20px';
  assert.equal(circularShapeRadius('md-comp-button-small-container-shape-round', values), expected);
  assert.equal(circularShapeRadius('md-comp-button-small-selected-container-shape-square', values), expected);
  assert.equal(values['md-comp-button-small-container-height'], '40px');
});

test('circular rendering takes the smaller box dimension and rejects missing measurements', () => {
  const values = {'md-comp-slider-handle-height': '44px', 'md-comp-slider-handle-width': '4px'};
  assert.equal(circularShapeRadius('md-comp-slider-handle-shape', values),
    '2px');
  assert.throws(() => circularShapeRadius('md-comp-unknown-container-shape', values), /Missing finite circular geometry/);
});
