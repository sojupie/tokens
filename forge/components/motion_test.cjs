const test = require('node:test');
const assert = require('node:assert/strict');
const { springState, springFrames, springTrack, cssTiming } = require('./static/motion.cjs');
test('spring retargeting retains position and velocity for under, critical and overdamping', () => {
  for (const damping of [.6, 1, 1.4]) {
    const state = springState(damping, 800, .08, 24, 8, 0);
    const restart = springState(damping, 800, 0, state.value, 24, state.velocity);
    assert.ok(Math.abs(restart.value - state.value) < 1e-9);
    assert.ok(Math.abs(restart.velocity - state.velocity) < 1e-9);
    const next = springState(damping, 800, .00001, state.value, 24, state.velocity);
    assert.ok(Math.abs((next.value - restart.value) / .00001 - state.velocity) < .15);
  }
});
test('spring width tracks conserve total width through an interrupted press and release', () => {
  const from = [100, 100, 100], to = [90, 120, 90];
  const state = from.map((n, i) => springState(.6, 800, .08, n, to[i]));
  const result = springTrack(.6, 800, state.map(s => s.value), from, state.map(s => s.velocity));
  for (const sample of result.samples) assert.ok(Math.abs(sample.reduce((a,b) => a+b,0)-300) < 1e-7);
  assert.deepEqual(result.samples.at(-1), from);
});
test('generated CSS curves carry the captured spring response and effects never overshoot', () => {
  const spatial = springFrames(.6, 800), effects = springFrames(1, 3800);
  assert.ok(Math.max(...spatial.samples) > 1);
  assert.ok(effects.samples.every(n => n >= 0 && n <= 1));
  const css = cssTiming(.6, 800);
  assert.match(css.easing, /^linear\(/);
  assert.equal(parseFloat(css.duration).toFixed(3), spatial.duration.toFixed(3));
});

test('popup fallback mirrors FAB item springs and removes copies on reopen or reduced motion', () => {
  const {runInNewContext} = require('node:vm'), {readFileSync} = require('node:fs');
  const listeners = {}, timers = new Map(), copies = [];
  let timer = 0;
  const reads=[];
  const style = Object.assign(['display','opacity','grid-template-areas','--_motion-fast-spatial-easing','--_button-container'], {getPropertyValue(name){reads.push(name);if(name==='grid-template-areas'||name==='--_motion-fast-spatial-easing')throw Error('Unneeded computed style read');return name==='display'?'block':'1';}});
  style[Symbol.iterator]=function(){throw Error('Must not enumerate inherited publication tokens');};
  const document={body:{append(copy){copies.push(copy);}},addEventListener(){}};
  const globals={document,CSS:{supports:()=>false},matchMedia:()=>({matches:false,addEventListener:(name,fn)=>listeners[name]=fn}),getComputedStyle:()=>style,setTimeout(fn,ms){timers.set(++timer,{fn,ms});return timer;},clearTimeout:id=>timers.delete(id)};
  runInNewContext(readFileSync(require.resolve('./static/motion.cjs'),'utf8'),globals);
  const original = {effect:{target:null,getKeyframes:()=>[{width:'80px',opacity:1},{width:'0px',opacity:0}],getTiming:()=>({duration:420,delay:75,fill:'both'})},currentTime:12};
  const item={getAnimations:()=>[original]};original.effect.target=item;
  const mirrored=[];
  const node=()=>({style:{setProperty(){}},animate(frames,options){const a={cancel(){this.cancelled=true;},effect:{getComputedTiming:()=>({endTime:options.duration+options.delay})},finished:new Promise(()=>{})};mirrored.push({frames,options,a});return a;}});
  const element={scrollTop:4,scrollLeft:0,closest:()=>null,querySelectorAll:()=>[item],getAnimations:()=>[],cloneNode(){const child=node();return{...node(),querySelectorAll:selector=>selector==='*'?[child]:[],removeAttribute(){},setAttribute(){},remove(){this.removed=true;}};}};
  globals.MaterialMotion.retainExit(element,true);
  assert.equal(copies.length,1);assert.equal(copies[0].inert,true);
  assert.match(copies[0].style.cssText,/--_button-container:1;/,'keep the resolved component paint variables');
  assert.ok(reads.length<300,'snapshot only geometry and paint, not every computed property');
  assert.doesNotMatch(copies[0].style.cssText,/--_motion-/,'copied WAAPI tracks supply their own timing');
  assert.equal(mirrored[0].options.delay,75);assert.equal(mirrored[0].options.duration,420);assert.equal(mirrored[0].a.currentTime,12);
  assert.ok([...timers.values()].some(t=>t.ms===545));
  globals.MaterialMotion.clearExit(element);
  assert.equal(copies[0].removed,true);assert.equal(mirrored[0].a.cancelled,true);
  globals.MaterialMotion.retainExit(element,true);listeners.change({matches:true});
  assert.equal(copies[1].removed,true);assert.equal(timers.size,0);
});
