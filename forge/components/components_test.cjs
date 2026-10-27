const {test} = require('node:test');
const assert = require('node:assert/strict');
const {readFileSync} = require('node:fs');
const {runInNewContext} = require('node:vm');
const exported = {exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/components.js`,'utf8'),{module:exported});
const {snackbarTimer} = exported.exports;
function clock() {
 let time=0, next=0; const pending=new Map();
 return {now:()=>time,setTimeout(fn,delay){pending.set(++next,{fn,at:time+delay});return next;},clearTimeout(id){pending.delete(id);},advance(ms){time+=ms;for(const[id,t]of pending)if(t.at<=time){pending.delete(id);t.fn();}},size:()=>pending.size};
}
test('snackbar timeout pauses for both focus and pointer and resumes remaining time',()=>{
 const c=clock();let closed=0;const timer=snackbarTimer(()=>closed++,c);
 timer.start(6000);c.advance(1000);timer.pause('pointer');timer.pause('focus');c.advance(10000);
 assert.equal(closed,0);timer.resume('pointer');c.advance(10000);assert.equal(closed,0);
 timer.resume('focus');timer.resume('focus');assert.equal(c.size(),1);c.advance(4999);assert.equal(closed,0);c.advance(1);assert.equal(closed,1);
});
test('snackbar restart and dismissal clear prior timeouts',()=>{
 const c=clock();let closed=0;const timer=snackbarTimer(()=>closed++,c);
 timer.start(6000);c.advance(5000);timer.start(6000);assert.equal(c.size(),1);c.advance(1000);assert.equal(closed,0);
 timer.stop();c.advance(10000);assert.equal(closed,0);assert.equal(c.size(),0);
 timer.start(0);assert.equal(c.size(),0);
});

test('navigation keyboard follows orientation, wraps and supports RTL', () => {
  const { navigationIndex } = exported.exports;
  assert.equal(navigationIndex(2,3,'ArrowRight',true),0);
  assert.equal(navigationIndex(0,3,'ArrowLeft',true),2);
  assert.equal(navigationIndex(0,3,'ArrowRight',true,true),2);
  assert.equal(navigationIndex(0,3,'ArrowDown',false),1);
  assert.equal(navigationIndex(1,3,'Home',false),0);
  assert.equal(navigationIndex(0,3,'End',false),2);
  assert.equal(navigationIndex(1,3,'ArrowRight',false),null);
});

test('slider values clamp the changed range handle without moving its peer',()=>{
 const {sliderValues}=exported.exports;
 assert.deepEqual(Array.from(sliderValues([20,70],0,0,100)),[20,70]);
 assert.deepEqual(Array.from(sliderValues([90,70],0,0,100)),[70,70]);
 assert.deepEqual(Array.from(sliderValues([20,10],1,0,100)),[20,20]);
 assert.deepEqual(Array.from(sliderValues([-10],0,0,100)),[0]);
 assert.deepEqual(Array.from(sliderValues([120],0,0,100)),[100]);
 assert.deepEqual(Array.from(sliderValues([45,45],0,0,100)),[45,45]);
});

test('docked toolbar spacing respects both source limits and available button widths',()=>{
 const {toolbarGap}=exported.exports;
 assert.equal(toolbarGap(400,[40,40,40],4,32),32);
 assert.equal(toolbarGap(160,[40,40,40],4,32),20);
 assert.equal(toolbarGap(100,[40,40,40],4,32),4);
 assert.equal(toolbarGap(100,[40],4,32),0);
});

test('toolbar touch targets do not inherit the icon target centering translation', () => {
 const css = readFileSync(`${__dirname}/static/components-base.css`, 'utf8');
 const target = css.match(/\.md-toolbar \.md-toolbar-item::after\s*\{([^}]+)\}/)?.[1];
 assert.ok(target, 'shared toolbar target rule is present');
 assert.match(target, /translate:\s*none\s*;/, 'inset expansion must cancel the inherited half-target translation');
 assert.match(target, /inset:\s*calc\(-1 \* var\(--md-sys-measurement-space50\)\)/);
});

test('cards define their own surface foreground instead of inheriting a parent color', () => {
  const css = readFileSync(`${__dirname}/static/components-base.css`, 'utf8');
  const card = css.match(/^\.md-card\s*\{([^}]+)\}/m)?.[1];
  assert.match(card, /\bcolor:\s*var\(--md-sys-color-on-surface\)\s*;/);
});

const {splitMenuPosition} = exported.exports;
test('split menu keeps 4dp separation and aligns an available button edge',()=>{
 const trigger={left:200,right:248,top:100,bottom:140},group={left:100,right:248};
 const menu={width:180,height:100},viewport={width:800,height:600};
 assert.deepEqual({...splitMenuPosition(trigger,group,menu,viewport,false,4,8)},{left:68,top:144});
 assert.deepEqual({...splitMenuPosition(trigger,group,menu,viewport,true,4,8)},{left:200,top:144});
 const nearStart={left:20,right:68,top:500,bottom:540};
 assert.deepEqual({...splitMenuPosition(nearStart,{left:8,right:68},menu,viewport,false,4,8)},{left:20,top:396});
 const narrow={left:250,right:298,top:100,bottom:140};
 assert.deepEqual({...splitMenuPosition(narrow,{left:100,right:298},menu,{width:300,height:600},false,4,8)},{left:100,top:144});
});
