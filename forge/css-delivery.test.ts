import test from 'node:test';
import assert from 'node:assert/strict';
import {deliveryStyles} from './css-delivery.ts';

test('delivery shortening preserves runtime prefixes, inline properties, literals and names that could collide',()=>{
 const source={
  '/theme.css':':root{--a:red;--brand-long:red;--motion-fast-duration:1s;--motion-slow-duration:2s;--inline-color:blue}',
  '/static/app.css':'.x{color:var(--brand-long);background:var(--inline-color);animation-duration:var(--motion-fast-duration);content:"--brand-long"}'
 };
 const result=deliveryStyles(source,['style.getPropertyValue(`--motion-${speed}-duration`);','style="--inline-color:red"']);
 assert.equal(result.shortened,2);
 assert.match(result.styles['/static/app.css'],/color:var\(--[b-zA-Z0-9_-]+\)/);
 assert.match(result.styles['/static/app.css'],/var\(--inline-color\)/);
 assert.match(result.styles['/static/app.css'],/var\(--motion-fast-duration\)/);
 assert.match(result.styles['/static/app.css'],/content:"--brand-long"/);
 assert.deepEqual(deliveryStyles(source,['style="--inline-color:red"','`--motion-${speed}-duration`']),result);
 assert.match(source['/theme.css'],/--brand-long/,'input/download CSS is not mutated');
});

test('literal runtime names protect only exact matches and empty rules are dropped from delivery',()=>{
 const css=':root{--surface:red;--surface-variant:blue;--concat-size:4px;--format-color:green}:root[data-theme=dark]{}';
 const result=deliveryStyles({'/theme.css':css},[
  'getPropertyValue("--surface");getPropertyValue("--concat-"+key);',
  'fmt.Sprintf("--format-%s", key);',
 ]);
 assert.equal(result.shortened,1);
 const delivered=result.styles['/theme.css'];
 assert.match(delivered,/--surface:red/);
 assert.match(delivered,/--concat-size:4px/);
 assert.match(delivered,/--format-color:green/);
 assert.doesNotMatch(delivered,/--surface-variant|data-theme/);
 assert.match(css,/--surface-variant|data-theme/);
});
