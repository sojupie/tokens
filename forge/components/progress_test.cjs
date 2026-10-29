const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const moduleObject={exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/progress.js`,'utf8'),{module:moduleObject});
const {progressSegments,progressPath,complement}=moduleObject.exports;
test('progress tracks exclude all active spans, including the circular wrap',()=>{
 assert.deepEqual(JSON.parse(JSON.stringify(complement([[.9,1.2]],.05,true))),[[.24999999999999994,.85]]);
 assert.deepEqual(JSON.parse(JSON.stringify(complement([[.1,.2],[.4,.7]],.05))),[[0,.05],[.25,.35000000000000003],[.75,1]]);
 assert.equal(complement([[0,1]],.1).length,0);
 assert.equal(complement([[0,0]],.1).length,1);
});
test('progress animations retain finite ordered segments over repeated cycles',()=>{
 for(const circular of [false,true])for(let t=0;t<15000;t+=17) {
  for(const [a,b]of progressSegments(t,circular)) {
   assert.ok(Number.isFinite(a)&&Number.isFinite(b));assert.ok(a<=b);
   if(!circular)assert.ok(a>=0&&b<=1);
  }
 }
});
test('flat and wavy paths stay within source geometry and complete circles join',()=>{
 for(const circular of [false,true]) {
  const g={width:circular?48:240,height:circular?48:14,thickness:8,circular,amplitude:circular?1.6:3,wavelength:circular?15:40};
  for(const phase of [0,.23,.9]) {
   const path=progressPath(0,1,g,phase);assert.doesNotMatch(path,/NaN|Infinity/);
   const points=[...path.matchAll(/[ML]([\d.-]+),([\d.-]+)/g)].map(m=>[+m[1],+m[2]]);
   for(const [x,y]of points){assert.ok(x>=g.thickness/2-.001&&x<=g.width-g.thickness/2+.001);assert.ok(y>=g.thickness/2-.001&&y<=g.height-g.thickness/2+.001);}
   if(circular)assert.deepEqual(points[0],points.at(-1));
  }
 }
 assert.equal(progressPath(.5,.5,{width:20,height:4,thickness:4}), '');
});

test('source circular shapes resolve to infinity while explicit zero remains zero',()=>{
 const {progressNumber}=moduleObject.exports;
 assert.equal(progressNumber('calc(infinity * 1px)'),Infinity);
 assert.equal(progressNumber('0px',4),0);
 assert.equal(progressNumber('8px'),8);
 assert.equal(progressNumber('',4),4);
});

test('programmatic progress and mode changes repaint without slider input',()=>{
 let changed;
 const node=()=>({attrs:{},style:{},setAttribute(k,v){this.attrs[k]=v}});
 const track=node(),active=node(),second=node(),stop=node(),svg=node();
 const attributes={'aria-valuenow':'10'};
 const element={dataset:{variant:'linear',shape:'flat',control:'determinate'},clientWidth:240,clientHeight:4,isConnected:true,
  getAttribute:k=>attributes[k],closest:()=>null,
  querySelector:s=>({svg,'.md-progress-track':track,'.md-progress-stop':stop}[s]),querySelectorAll:()=>[active,second]};
 const values={'--_progress-thickness':'4px','--_progress-shape':'2px','--_progress-track-shape':'2px','--_progress-gap':'4px'};
 runInNewContext(readFileSync(`${__dirname}/static/progress.js`,'utf8'),{
  document:{hidden:false,querySelectorAll:()=>[element],addEventListener(){}},
  getComputedStyle:()=>({getPropertyValue:k=>values[k]||'0'}),performance:{now:()=>0},
  matchMedia:()=>({matches:true,addEventListener(){}}),
  ResizeObserver:class{observe(){}},IntersectionObserver:class{observe(){}},
  MutationObserver:class{constructor(fn){changed=fn}observe(){}disconnect(){}},
 });
 const before=active.attrs.d;
 attributes['aria-valuenow']='75';changed();assert.notEqual(active.attrs.d,before);
 const determinate=active.attrs.d;
 element.dataset.control='indeterminate';changed();assert.notEqual(active.attrs.d,determinate);
});

test('circular waveform rotates with its active arc instead of sliding through it',()=>{
 const g={width:52,height:52,thickness:8,circular:true,amplitude:1.6,wavelength:15};
 const points=p=>[...p.matchAll(/[ML]([\d.-]+),([\d.-]+)/g)].map(m=>[+m[1],+m[2]]);
 const a=points(progressPath(0,.4,g,.3)),b=points(progressPath(.25,.65,g,.3));
 assert.equal(a.length,b.length);
 for(let i=0;i<a.length;i++) {
  assert.ok(Math.abs(b[i][0]-(52-a[i][1]))<.002);
  assert.ok(Math.abs(b[i][1]-a[i][0])<.002);
 }
 // The track reaches the outer circle rather than shrinking to the wave's mean.
 const track=points(progressPath(0,1,g,0,false));
 assert.equal(track[0][1],4);
});

// AndroidX keyframe easing belongs to the interval after the named frame.
test('circular sweep and turns follow reference keyframe intervals',()=>{
 const [start,end]=progressSegments(150,true)[0];
 assert.ok(Math.abs(start-.2)<.000001);
 assert.ok(Math.abs((end-start)-(.1+.77*.05))<.000001);
 const [a,b]=progressSegments(1500,true)[0];
 assert.ok(Math.abs((b-a)-.485)<.000001);
 assert.ok(Math.abs((progressSegments(3000,true)[0][1]-progressSegments(3000,true)[0][0])-.87)<.000001);
});
