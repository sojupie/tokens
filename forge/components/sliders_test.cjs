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

test('A single inset icon moves to the inactive end when the active track is too short',()=>{
 const classes=new Set(['md-slider-leading']);
 const icon={dataset:{},style:{},hasAttribute:name=>name==='data-slider-inset',classList:{
  contains:name=>classes.has(name),toggle(name,on){if(on)classes.add(name);else classes.delete(name);}
 }};
 const bubble={textContent:''};
 const control={querySelector:()=>bubble};
 const input={min:'0',max:'100',valueAsNumber:40,parentElement:control,setAttribute(){}};
 const stop=()=>({dataset:{},style:{}});
 const startStop=stop(),endStop=stop();
 const slider={dataset:{},style:{setProperty(){}},setAttribute(){},getBoundingClientRect:()=>({width:400,height:52}),
  querySelectorAll(selector){return selector==='input[type="range"]'?[input]:selector==='.md-slider-icon'?[icon]:[];},
  querySelector:selector=>selector==='.md-slider-stop-start'?startStop:endStop};
 let resize;
 runInNewContext(readFileSync(`${__dirname}/behaviors/sliders.js`,'utf8'),{
  document:{addEventListener(){},querySelectorAll:()=>[slider]},window:{addEventListener(){}},
  MaterialComponents:{sliderValues:values=>values},
  ResizeObserver:class{constructor(fn){resize=()=>fn([{target:slider}]);}observe(){}},
  getComputedStyle:()=>({getPropertyValue:name=>({'--_slider-state-size':'44','--_slider-control-width':'4','--_slider-leading':'6','--_slider-trailing':'6','--_slider-icon-size':'24','--_slider-icon-padding':'8'}[name])})
 });
 assert.equal(slider.dataset.insetPosition,'start');assert.equal(icon.dataset.active,'true');
 input.valueAsNumber=0;resize();
 assert.equal(slider.dataset.insetPosition,'end');assert.equal(icon.dataset.active,'false');assert.equal(icon.style.visibility,'visible');
 input.valueAsNumber=100;resize();
 assert.equal(slider.dataset.insetPosition,'start');assert.equal(icon.style.visibility,'visible');
});
