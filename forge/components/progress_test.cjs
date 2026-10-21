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
