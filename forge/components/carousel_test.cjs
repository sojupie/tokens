const test=require('node:test'),assert=require('node:assert/strict');
const {readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{module:out});
const {layout,itemFrame}=out.exports;
test('multi-browse end transition has monotonic edges and widths',()=>{
 for(const width of [280,384,640,900])for(const preferred of [160,240,320]) {
  const plan=layout(width,preferred,8,'multi-browse',6);
  if(plan.variant==='uncontained')continue;
  const start=5-Math.min(plan.keylines.at(-1).i,5);
  for(let index=0;index<6;index++) {
   const first=itemFrame(plan,index,start,6),last=itemFrame(plan,index,5,6);
   let prior=first;
   for(let n=1;n<=500;n++) {
    const frame=itemFrame(plan,index,start+(5-start)*n/500,6);
    for(const edge of ['left','size']) {
     const direction=Math.sign(last[edge]-first[edge]);
     assert.ok(direction*(frame[edge]-prior[edge])>=-1e-8,`${width}/${preferred}: item ${index} ${edge} reversed`);
     if(direction===0)assert.ok(Math.abs(frame[edge]-first[edge])<1e-8);
    }
    prior=frame;
   }
  }
 }
});
test('carousel keylines preserve item size while masking, align endpoints, and remain ordered',()=>{
 for(const width of [160,280,384,640,900])for(const variant of ['multi-browse','uncontained','hero','full-screen']) {
  const plan=layout(width,240,8,variant,6),step=plan.large+plan.gap;
  assert.ok(Number.isFinite(step)&&step>0);
  assert.ok(Math.abs(itemFrame(plan,0,0,6).left)<.001);
  const last=itemFrame(plan,5,5,6);
  if(plan.variant!=='uncontained'||variant==='full-screen')assert.ok(Math.abs(last.left+last.size-width)<.001,`${variant} ${width}`);
  for(let p=0;p<=5;p+=.1){let previous=-Infinity;for(let i=0;i<6;i++){const f=itemFrame(plan,i,p,6);assert.ok(f.size>0&&Number.isFinite(f.left));assert.ok(f.left>=previous);previous=f.left; if(i<5&&f.left+f.size>0){const next=itemFrame(plan,i+1,p,6);if(next.left<width)assert.ok(next.left-f.left-f.size>=8-.001);}}}
 }
});
test('cleanup ignores carousel descendants that were never initialized',()=>{
 const listeners={},document={readyState:'complete',querySelectorAll:()=>[],addEventListener:(name,fn)=>listeners[name]=fn};
 runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{document,matchMedia:()=>({matches:false}),ResizeObserver:class{observe(){}unobserve(){}},window:{addEventListener(){}}});
 const child={matches:()=>true,querySelectorAll:()=>[]};
 assert.doesNotThrow(()=>listeners['htmx:beforeCleanupElement']({detail:{elt:child}}));
});
test('mouse release settles smoothly before restoring snapping',()=>{
 for(const reduced of [false,true])for(const fast of [false,true]) {
  let nextFrame,clock;
  const listeners={},scrollListeners={},calls=[],buttons=Array.from({length:6},()=>({style:{},disabled:false}));
  const output={},controls={},viewport={
   clientWidth:656,scrollWidth:1445,scrollLeft:0,
   style:{removeProperty(){delete this.scrollSnapType;}},
   addEventListener:(name,fn)=>scrollListeners[name]=fn,
   setPointerCapture(){},hasPointerCapture:()=>false,
   scrollTo(options){calls.push({...options,snap:this.style.scrollSnapType});},
  };
  const root={dataset:{variant:'multi-browse'},style:{setProperty(){}},
   querySelectorAll:()=>buttons,
   querySelector:selector=>selector==='.md-carousel-scroll-track'?{}:selector==='.md-carousel-viewport'?viewport:selector==='output'?output:controls[selector]??={},
  };
  const target={closest:selector=>selector==='.md-carousel'?root:selector==='.md-carousel-viewport'?viewport:null};
  const document={readyState:'complete',querySelectorAll:()=>[root],addEventListener:(name,fn)=>listeners[name]=fn};
  runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{
   document,matchMedia:()=>({matches:reduced}),window:{addEventListener(){}},
   getComputedStyle:()=>({direction:'ltr',paddingInlineStart:'8px',paddingInlineEnd:'8px',getPropertyValue:name=>name==='--_carousel-preferred'?'160px':name==='--md-sys-motion-duration-medium2'?'300ms':name==='--md-sys-motion-easing-standard'?'cubic-bezier(0.2,0,0,1)':'8px'}),
   Animation:class{constructor(effect){this.effect=effect;this.playState='running';clock=this;}play(){}cancel(){}},KeyframeEffect:class{getComputedTiming(){return{progress:clock.playState==='finished'?1:.5};}},requestAnimationFrame:fn=>{nextFrame=fn;return 1;},cancelAnimationFrame(){},
   ResizeObserver:class{constructor(fn){this.fn=fn;}observe(target){this.fn([{target}]);}},
  });
  listeners.pointerdown({target,pointerType:'mouse',button:0,pointerId:1,clientX:500,timeStamp:0});
  listeners.pointermove({pointerId:1,clientX:440,timeStamp:60,preventDefault(){}});
  assert.equal(viewport.scrollLeft,60);
  scrollListeners.scrollend(); // An idle moment while held is not the end of the gesture.
  assert.equal(root.dataset.carouselDragging,'');
  assert.equal(root.dataset.carouselSettled,undefined);
  listeners.pointerup({pointerId:1,timeStamp:fast?70:160});
  if(!reduced){nextFrame();assert.equal(calls[0].left,fast?109:30);clock.playState='finished';nextFrame();}
  assert.equal(calls.length,reduced?1:2);
  assert.equal(calls.at(-1).left,fast?158:0);
  assert.equal(calls[0].behavior,'instant');
  assert.equal(calls[0].snap,'none');
  assert.equal(root.dataset.carouselDragging,undefined);
  viewport.scrollLeft=fast?158:0;
  scrollListeners.scrollend();
  assert.equal(viewport.style.scrollSnapType,undefined);
  assert.equal(root.dataset.carouselSettled,fast?'1':'0');
 }
});
