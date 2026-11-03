const test=require('node:test'),assert=require('node:assert/strict');
const {readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{module:out});
const {layout,itemFrame,scrollTargets,uncontainedFrames,captionOpacity}=out.exports;
test('every resting arrangement retains its size hierarchy and interpolates without overlap',()=>{
 for(const width of [177,328,736,740,1168])for(const variant of ['multi-browse','hero','center-aligned-hero']){
  const plan=layout(width,240,8,variant,6),max=variant==='center-aligned-hero'?5:6-plan.big;
  for(let stop=0;stop<=max;stop++){
   const visible=Array.from({length:6},(_,i)=>itemFrame(plan,i,stop,6)).filter(f=>f.left>=-.01&&f.left+f.size<=width+.01);
   assert.ok(visible.some(f=>Math.abs(f.size-plan.large)<.01),`${variant}/${width}/${stop}: large`);
   assert.ok(visible.some(f=>Math.abs(f.size-plan.small)<.01),`${variant}/${width}/${stop}: small`);
   if(variant==='multi-browse')assert.ok(visible.some(f=>f.size>plan.small&&f.size<plan.large));
   if(variant==='center-aligned-hero')assert.equal(visible.filter(f=>Math.abs(f.size-plan.large)<.01).length,1);
  }
  for(let n=0;n<=max*100;n++)for(let i=0;i<5;i++){
   const a=itemFrame(plan,i,n/100,6),b=itemFrame(plan,i+1,n/100,6);
   assert.ok(b.left-a.left-a.size>=8-.001,`${variant}/${width}/${n}: overlap`);
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
  const listeners={},scrollListeners={},calls=[],buttons=Array.from({length:6},()=>({style:{},disabled:false,querySelector:()=>({style:{}})}));
  const snaps=buttons.map((_,i)=>({getBoundingClientRect:()=>({left:i*157.71875})}));
  const output={},controls={},viewport={
   clientWidth:656,scrollWidth:1445,scrollLeft:0,
   style:{removeProperty(){delete this.scrollSnapType;}},
   addEventListener:(name,fn)=>scrollListeners[name]=fn,
   setPointerCapture(){},hasPointerCapture:()=>false,
   scrollTo(options){calls.push({...options,snap:this.style.scrollSnapType});this.scrollLeft=Math.round(options.left);},
  };
  const root={dataset:{variant:'multi-browse'},style:{setProperty(){}},
   querySelectorAll:selector=>selector==='.md-carousel-caption'?[]:['.md-carousel-snap','.md-carousel-slot'].includes(selector)?snaps:buttons,
   querySelector:selector=>selector==='.md-carousel-pagination'?null:selector==='.md-carousel-scroll-track'?{}:selector==='.md-carousel-viewport'?viewport:selector==='output'?output:controls[selector]??={},
  };
  const target={closest:selector=>selector==='.md-carousel'?root:selector==='.md-carousel-viewport'?viewport:null};
  const document={readyState:'complete',querySelectorAll:()=>[root],addEventListener:(name,fn)=>listeners[name]=fn};
  runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{
   document,matchMedia:()=>({matches:reduced}),window:{addEventListener(){}},
   getComputedStyle:()=>({direction:'ltr',paddingInlineStart:'8px',paddingInlineEnd:'8px',getPropertyValue:name=>name==='--_carousel-maximum'?'':name==='--_carousel-preferred'?'160px':name==='--md-sys-motion-duration-medium2'?'300ms':name==='--md-sys-motion-easing-standard'?'cubic-bezier(0.2,0,0,1)':'8px'}),
   Animation:class{constructor(effect){this.effect=effect;this.playState='running';clock=this;}play(){}cancel(){}},KeyframeEffect:class{getComputedTiming(){return{progress:clock.playState==='finished'?1:.5};}},requestAnimationFrame:fn=>{nextFrame=fn;return 1;},cancelAnimationFrame(){},
   ResizeObserver:class{constructor(fn){this.fn=fn;}observe(target){this.fn([{target}]);}},
  });
  const initialWidths=buttons.map(button=>button.style.width);
  viewport.scrollLeft=-20;scrollListeners.scroll();
  assert.deepEqual(buttons.map(button=>button.style.width),initialWidths,'start overscroll must not become forward motion');
  viewport.scrollLeft=0;
  calls.length=0;listeners.pointerdown({target,pointerType:'mouse',button:0,pointerId:1,clientX:500,timeStamp:0});
  listeners.pointermove({pointerId:1,clientX:440,timeStamp:60,preventDefault(){}});
  assert.equal(viewport.scrollLeft,60);
  scrollListeners.scrollend(); // An idle moment while held is not the end of the gesture.
  assert.equal(root.dataset.carouselDragging,'');
  assert.equal(root.dataset.carouselSettled,undefined);
  listeners.pointerup({pointerId:1,timeStamp:fast?70:160});
  if(!reduced){
   assert.equal(calls.length,2); // Resolve target and restore the current native position.
   nextFrame();assert.equal(calls.length,2,'visual frames must not scroll the sticky layer');
   assert.equal(viewport.scrollLeft,60);
   scrollListeners.scrollend();assert.equal(root.dataset.carouselSettled,undefined);
   clock.playState='finished';nextFrame();
  }
  assert.equal(calls.length,reduced?1:3);
  assert.equal(calls.at(-1).left,fast?(reduced?157.71875:158):0);
  assert.equal(calls[0].behavior,'instant');
  assert.equal(calls[0].snap,reduced?undefined:'none');
  assert.equal(root.dataset.carouselDragging,undefined);
  viewport.scrollLeft=fast?158:0;
  scrollListeners.scrollend();
  assert.equal(viewport.style.scrollSnapType,undefined);
  assert.equal(root.dataset.carouselSettled,fast?'1':'0');
 }
});

test('hero layouts are distinct and dynamic small items stay in the captured 40–56px range',()=>{
 for(const width of [176,280,384,640,900]) {
  for(const variant of ['multi-browse','hero','center-aligned-hero']) {
   const plan=layout(width,240,8,variant,6);
   if(plan.variant==='uncontained')continue;
   assert.ok(plan.small>=40&&plan.small<=56);
   const first=itemFrame(plan,variant==='center-aligned-hero'?1:0,variant==='center-aligned-hero'?1:0,6);
   if(variant==='center-aligned-hero')assert.ok(Math.abs(first.left+first.size/2-width/2)<.001);
   else assert.ok(Math.abs(first.left)<.001);
   if(variant==='hero')assert.ok(plan.keylines.filter(k=>k.size===plan.small).length===2);
  }
 }
});

function nativeCarousel({variant='uncontained',reduced=false,scrolling='free',showControls=true,viewportWidth=336,paddingStart=16,paddingEnd=0}={}) {
 const listeners={},scrollListeners={},output={},moves=[];
 let nextFrame,clock;
 const widths=variant==='uncontained-multi-aspect-ratio'?[160,90,120,135,160,90]:[160,160,160,160,160,160];
 let start=0;
 const slots=widths.map(width=>{const left=start;start+=width+8;return{getBoundingClientRect:()=>({left,width})};});
 const buttons=widths.map(()=>({style:{},disabled:false,querySelector:()=>({style:{}}),focus(){}}));
 const viewport={clientWidth:viewportWidth,clientHeight:600,scrollHeight:3640,scrollTop:0,scrollWidth:start-8+16,scrollLeft:0,
  style:{removeProperty(){delete this.scrollSnapType;}},addEventListener:(name,fn)=>scrollListeners[name]=fn,
  scrollTo(options){moves.push(options);if(options.top!==undefined)this.scrollTop=options.top;else this.scrollLeft=Math.max(0,Math.min(options.left,this.scrollWidth-this.clientWidth));},
  setPointerCapture(){},hasPointerCapture:()=>false,
 };
 const pagination={children:[],attributes:{},replaceChildren(...children){this.children=children;},setAttribute(name,value){this.attributes[name]=value;}};
 const controls={},root={dataset:{variant,scrolling},style:{setProperty(){}},
  querySelectorAll:selector=>selector==='.md-carousel-caption'?[]:['.md-carousel-slot','.md-carousel-snap'].includes(selector)?slots:buttons,
  querySelector:selector=>selector==='.md-carousel-pagination'?pagination:selector==='.md-carousel-scroll-track'?{}:selector==='.md-carousel-viewport'?viewport:selector==='output'?(showControls?output:null):!showControls&&selector.startsWith('[data-carousel-step=')?null:controls[selector]??={},
 };
 const target={closest:selector=>selector==='.md-carousel'?root:selector==='.md-carousel-viewport'?viewport:null};
 const document={readyState:'complete',querySelectorAll:()=>[root],addEventListener:(name,fn)=>listeners[name]=fn,createElement:()=>({active:false,setAttribute(){},toggleAttribute(name,on){this.active=on;}})};
 runInNewContext(readFileSync(`${__dirname}/static/carousel.js`,'utf8'),{
  document,matchMedia:()=>({matches:reduced}),window:{addEventListener(){}},
  getComputedStyle:()=>({direction:'ltr',paddingInlineStart:paddingStart+'px',paddingInlineEnd:paddingEnd+'px',getPropertyValue:name=>name==='--_carousel-maximum'?'':name==='--_carousel-preferred'?'160px':'8px'}),
  Animation:class{constructor(effect){this.effect=effect;this.playState='running';clock=this;}play(){}cancel(){}},KeyframeEffect:class{getComputedTiming(){return{progress:clock.playState==='finished'?1:.5};}},requestAnimationFrame:fn=>{nextFrame=fn;return 1;},cancelAnimationFrame(){},
  ResizeObserver:class{observe(){}},
 });
 moves.length=0;return{root,viewport,buttons,moves,pagination,scroll(position){viewport.scrollLeft=position;scrollListeners.scroll();},step(direction){listeners.click({target:{closest:selector=>selector==='.md-carousel'?root:selector==='[data-carousel-step]'?{dataset:{carouselStep:String(direction)}}:null},detail:0});},finishAnimation(){clock.playState='finished';nextFrame();},drag(distance=60,fast=false){
  listeners.pointerdown({target,pointerType:'mouse',button:0,pointerId:1,clientX:200,timeStamp:0});
  listeners.pointermove({pointerId:1,clientX:200-distance,timeStamp:60,preventDefault(){}});
  listeners.pointerup({pointerId:1,timeStamp:fast?70:200});
 }};
}
test('uncontained variants retain measured media widths while their visible masks adapt',()=>{
 for(const variant of ['uncontained','uncontained-multi-aspect-ratio']) {
  const c=nativeCarousel({variant});c.drag();
  assert.equal(c.root.dataset.carouselLayout,'flow');
  assert.equal(c.viewport.scrollLeft,60);
  assert.equal(c.moves.length,0,'free release does not seek a snap target');
  assert.ok(c.buttons.every(b=>Number.isFinite(parseFloat(b.style.width))));
  assert.ok(c.buttons.some((b,i)=>parseFloat(b.style.width)<c.root.querySelectorAll('.md-carousel-slot')[i].getBoundingClientRect().width));
 }
});
test('optional uncontained snapping uses native flow, including unequal item widths',()=>{
 const c=nativeCarousel({variant:'uncontained-multi-aspect-ratio',scrolling:'snap'});c.drag();
 assert.equal(c.viewport.style.scrollSnapType,'none');
 c.finishAnimation();
 assert.equal(c.viewport.scrollLeft,0);
 assert.equal(c.viewport.style.scrollSnapType,undefined);
 assert.ok(c.moves.every(move=>move.behavior==='instant'),'one shared token clock owns animated frames');
});
test('reduced motion uses complete stable items without width masking or parallax',()=>{
 for(const variant of ['multi-browse','uncontained','uncontained-multi-aspect-ratio','hero','center-aligned-hero','full-screen']) {
  const c=nativeCarousel({variant,reduced:true});c.drag();
  assert.equal(c.root.dataset.carouselLayout,'flow');
  if(variant!=='full-screen')assert.ok(c.buttons.every(b=>b.style.width===''&&b.style.transform===''));
  if(!variant.startsWith('uncontained'))assert.equal(c.moves[0].behavior,'instant');
 }
});

test('mixed-width snap targets use logical leading edges in both directions',()=>{
 const ltr=[{left:16,right:336},{left:344,right:434},{left:442,right:622}];
 const rtl=[{left:664,right:984},{left:566,right:656},{left:378,right:558}];
 assert.deepEqual(Array.from(scrollTargets(ltr,false,800)),[0,328,426]);
 assert.deepEqual(Array.from(scrollTargets(rtl,true,800)),[0,328,426]);
});

test('narrow compact containers retain large, medium and bounded small keylines',()=>{
 const plan=layout(177,240,8,'multi-browse',6);
 assert.equal(plan.variant,'multi-browse');
 const visible=[0,1,2].map(i=>itemFrame(plan,i,0,6));
 assert.ok(visible[0].size>visible[1].size&&visible[1].size>visible[2].size);
 assert.equal(visible[2].size,40);
});

test('carousel scenes can omit previous and next controls',()=>{
 for(const variant of ['multi-browse','uncontained']) {
  const c=nativeCarousel({variant,showControls:false});
  assert.equal(c.root.dataset.carouselReady,'');
  assert.ok(c.buttons.every(button=>Number.isFinite(parseFloat(button.style.width))||c.root.dataset.carouselLayout==='flow'));
 }
});

test('an explicit maximum bounds all large slots',()=>{
 for(const variant of ['multi-browse','hero','center-aligned-hero']){
  const plan=layout(320,240,8,variant,6,200);
  assert.ok(plan.large<=200);
 }
});

test('hero mouse fling advances at most one item, including reduced motion',()=>{
 for(const variant of ['hero','center-aligned-hero']) {
  const c=nativeCarousel({variant,reduced:true});c.drag(500,true);
  assert.equal(c.moves.at(-1).left,168);
 }
});

test('uncontained masks interpolate continuously with intact gaps and stable media',()=>{
 for(const widths of [[240,240,240,240,240,240],[320,108,192,144,240,180]]){
  const max=widths.reduce((a,b)=>a+b,0)+8*(widths.length-1)-328;
  for(let scroll=0;scroll<=max;scroll+=.5){
   const frames=uncontainedFrames(328,widths,8,scroll,max);
   for(let i=0;i<frames.length;i++){
    const f=frames[i];assert.equal(f.media,widths[i]);assert.ok(f.size>0&&f.size<=f.media+.001);
    if(i)assert.ok(Math.abs(f.left-frames[i-1].left-frames[i-1].size-8)<.001);
   }
  }
  const end=uncontainedFrames(328,widths,8,max,max).at(-1);assert.ok(Math.abs(end.left+end.size-328)<.001);
 }
});

test('arrows skip duplicate end targets and reverse without a stuck item index',()=>{
 const c=nativeCarousel({viewportWidth:640});
 for(let i=0;i<4;i++){c.step(1);c.finishAnimation();}
 assert.equal(c.viewport.scrollLeft,c.viewport.scrollWidth-c.viewport.clientWidth);
 const end=c.viewport.scrollLeft;
 c.step(-1);c.finishAnimation();
 assert.ok(c.viewport.scrollLeft<end);
 c.step(-1);c.step(-1);c.finishAnimation();
 assert.equal(c.viewport.scrollLeft,0,'rapid arrows advance from the requested destination');
});

test('pagination counts unique reachable positions and moves its active pill',()=>{
 const c=nativeCarousel({viewportWidth:640});
 assert.equal(c.pagination.children.length,4,'clamped duplicate end positions are one page');
 assert.equal(c.pagination.children.filter(dot=>dot.active).length,1);
 assert.equal(c.pagination.children[0].active,true);
 c.step(1);c.finishAnimation();
 assert.equal(c.pagination.children[1].active,true);
 assert.equal(c.pagination.attributes['aria-label'],'Carousel position 2 of 4');
});

test('items remain visible through viewport padding until their whole mask exits',()=>{
 for(const variant of ['multi-browse','hero','center-aligned-hero','uncontained']) for(const pressed of [false,true]) {
  const paddingStart=16,paddingEnd=16,c=nativeCarousel({variant,paddingStart,paddingEnd});
  if(pressed)c.buttons.forEach(b=>b.dataset={componentState:'pressed'});
  let partial=0;
  for(let scroll=0;scroll<=c.viewport.scrollWidth-c.viewport.clientWidth;scroll+=2){
   c.scroll(scroll);
   for(let i=0;i<c.buttons.length;i++){
    const b=c.buttons[i],translation=parseFloat(b.style.transform.slice(11));
    // Flow items also retain their native position in the scroll track.
    const left=translation+(variant==='uncontained'?i*168-scroll:0),right=left+parseFloat(b.style.width);
    const intersects=right>-paddingStart&&left<c.viewport.clientWidth-paddingStart;
    if(intersects&&(right<=0||left>=c.viewport.clientWidth-paddingStart-paddingEnd))partial++;
    assert.equal(b.style.visibility,intersects?'visible':'hidden',`${variant}/${scroll}/${i}: clipped a visible mask`);
   }
  }
  assert.ok(partial>0,`${variant}: test must cover masks inside side padding`);
 }
});

test('captions fade before the visible image clips any glyphs',()=>{
 const text={left:16,right:59,top:160,bottom:180};
 const clip={left:0,right:100,top:0,bottom:196};
 assert.equal(captionOpacity(text,clip,10),1,'fully fitting short labels remain visible');
 assert.equal(captionOpacity(text,{...clip,right:64},10),.5,'fade as the mask approaches the text');
 assert.equal(captionOpacity(text,{...clip,right:59},10),0,'hide at the clipping edge');
 assert.equal(captionOpacity(text,{...clip,right:56},10),0,'a narrow item cannot expose clipped text');
 assert.equal(captionOpacity(text,{...clip,left:20},10),0,'viewport clipping uses the same rule');
});
