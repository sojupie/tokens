const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const exported={exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/loading.js`,'utf8'),{module:exported});
const {loadingFrame,loadingPath}=exported.exports;
const shapes=JSON.parse(readFileSync(`${__dirname}/static/loading-shapes.json`,'utf8'));

test('loading geometry contains all seven matched morphs and the determinate pair',()=>{
 assert.equal(shapes.indeterminate.morphs.length,7);
 assert.equal(shapes.determinate.morphs.length,1);
 for(const sequence of Object.values(shapes)) {
  assert.ok(sequence.scale>0&&sequence.scale<=1);
  for(const [a,b]of sequence.morphs) {
   assert.equal(a.length,b.length);assert.ok(a.length>4);
   for(const curves of [a,b])for(let i=0;i<curves.length;i++) {
    assert.equal(curves[i].length,8);assert.ok(curves[i].every(Number.isFinite));
    const next=curves[(i+1)%curves.length];
    assert.ok(Math.abs(curves[i][6]-next[0])<.00001&&Math.abs(curves[i][7]-next[1])<.00001,'morph path must close without a seam');
   }
  }
 }
});
test('loading morphs retain finite closed paths through spring overshoot',()=>{
 for(const sequence of Object.values(shapes))for(const pair of sequence.morphs)for(const progress of [0,.25,.4,.5,.75,1,1.095]) {
  const path=loadingPath(pair,progress,sequence.scale);
  assert.match(path,/^M[\d.-]+,[\d.-]+C/);assert.ok(path.endsWith('Z'));assert.doesNotMatch(path,/NaN|Infinity/);
 }
});
test('loading motion cycles all seven shapes without a rotation reset',()=>{
 const indices=new Set();
 for(let i=0;i<14;i++){
  const current=loadingFrame(i*650);indices.add(current.index);
  assert.equal(current.progress,0);assert.equal(current.rotation,(i+1)*90+i*650/4666*360);
  if(i){const before=loadingFrame(i*650-.01);assert.ok(Math.abs(current.rotation-before.rotation)<1,'rotation must remain continuous at morph boundaries');}
 }
 assert.equal(indices.size,7);
 assert.equal(loadingFrame(-1).index,0);
 assert.deepEqual(JSON.parse(JSON.stringify(loadingFrame(1000,.4))),{index:0,progress:.4,rotation:-72});
 assert.equal(loadingFrame(1000,2).progress,1);assert.equal(loadingFrame(1000,-1).progress,0);
});
test('server-rendered first frames match browser geometry within float precision',()=>{
 for(const control of ['determinate','indeterminate']) {
  const sequence=shapes[control],path=loadingPath(sequence.morphs[0],control==='determinate'?.4:0,sequence.scale);
  const numbers=s=>s.match(/-?\d+\.\d+/g).map(Number);
  const actual=numbers(path),expected=numbers(sequence.initial);assert.equal(actual.length,expected.length);
  actual.forEach((n,i)=>assert.ok(Math.abs(n-expected[i])<.000002));
 }
});

test('inert catalog illustrations neither fetch shape data nor schedule animation',()=>{
 let requests=0,frames=0;
 const media={matches:false,addEventListener(){}};
 runInNewContext(readFileSync(`${__dirname}/static/loading.js`,'utf8'),{
  URL,document:{currentScript:{src:'http://localhost/md3/static/loading.js',dataset:{shapesVersion:'test'}},querySelectorAll:()=>[{closest:()=>true}],addEventListener(){}},
  matchMedia:()=>media,IntersectionObserver:class{},fetch:()=>{requests++;},requestAnimationFrame:()=>{frames++;}
 });
 assert.equal(requests,0);assert.equal(frames,0);
});

test('reduced motion renders a stationary shape and resumes animation when disabled',async()=>{
 let frames=0,onPreferenceChange;const drawn={};
 const media={matches:true,addEventListener(type,fn){onPreferenceChange=fn;}};
 const element={dataset:{control:'indeterminate'},isConnected:true,closest:()=>null,
  querySelector:tag=>({setAttribute(name,value){drawn[tag+':'+name]=value;}})};
 runInNewContext(readFileSync(`${__dirname}/static/loading.js`,'utf8'),{
  URL,performance:{now:()=>0},console,
  document:{currentScript:{src:'http://localhost/md3/static/loading.js',dataset:{shapesVersion:'test'}},querySelectorAll:()=>[element],addEventListener(){}},
  matchMedia:()=>media,IntersectionObserver:class{constructor(fn){this.fn=fn;}observe(target){this.fn([{target,isIntersecting:true}]);}},
  fetch:async()=>({ok:true,json:async()=>shapes}),requestAnimationFrame:()=>++frames
 });
 await new Promise(setImmediate);
 assert.ok(drawn['path:d'].startsWith('M'));assert.equal(drawn['g:transform'],'rotate(90 .5 .5)');assert.equal(frames,0);
 media.matches=false;onPreferenceChange();assert.equal(frames,1);
});
