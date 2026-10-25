const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
function harness({step='1',direction='ltr',orientation='horizontal',variant='continuous'}={}){
 const handlers={};const root={dataset:{orientation,variant}};
 const input={valueAsNumber:40,min:'0',max:'100',step,disabled:false,events:[],
  matches:()=>!input.disabled,closest:selector=>selector==='.md-slider'?root:null,
  dispatchEvent:event=>input.events.push(event.type)};
 const document={querySelectorAll:()=>[],addEventListener:(type,fn)=>(handlers[type]??=[]).push(fn)};
 runInNewContext(readFileSync(`${__dirname}/behaviors/sliders.js`,'utf8'),{
  document,window:{addEventListener:(type,fn)=>(handlers[type]??=[]).push(fn)},
  MaterialComponents:{},ResizeObserver:class{observe(){}},getComputedStyle:()=>({direction}),Event});
 return {input,key(key,type='keydown',options={}){const e={key,target:input,preventDefault(){this.defaultPrevented=true},...options};for(const fn of handlers[type]||[])fn(e);return !!e.defaultPrevented}};
}
test('Space plus arrows adjusts continuous sliders by ten percent',()=>{
 const h=harness();assert.equal(h.key(' '),true);assert.equal(h.key('ArrowRight'),true);
 assert.equal(h.input.valueAsNumber,50);assert.deepEqual(h.input.events,['input','change']);
 h.key('ArrowLeft');assert.equal(h.input.valueAsNumber,40);
});
test('Discrete sliders advance one stop and clamp at the limit',()=>{
 const h=harness({step:'2',variant:'discrete'});h.input.valueAsNumber=100;h.key(' ');h.key('ArrowRight');assert.equal(h.input.valueAsNumber,100);
 h.key('ArrowLeft');assert.equal(h.input.valueAsNumber,98);
});
test('RTL horizontal direction reverses while vertical up increases',()=>{
 const h=harness({direction:'rtl'});h.key(' ');h.key('ArrowRight');assert.equal(h.input.valueAsNumber,30);
 const v=harness({direction:'rtl',orientation:'vertical'});v.key(' ');v.key('ArrowUp');assert.equal(v.input.valueAsNumber,50);
});
test('Native keys remain native without Space and after release or blur',()=>{
 const h=harness();assert.equal(h.key('ArrowRight'),false);assert.equal(h.key('Home'),false);
 h.key(' ');h.key(' ','keyup');assert.equal(h.key('ArrowRight'),false);
 h.key(' ');h.key('','blur');assert.equal(h.key('ArrowRight'),false);
 h.key(' ');h.key('','focusout',{target:{closest:()=>null}});assert.equal(h.key('ArrowRight'),false);
 assert.equal(h.input.valueAsNumber,40);
});
