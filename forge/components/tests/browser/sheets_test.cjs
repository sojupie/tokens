const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const exported={exports:{}};
runInNewContext(readFileSync(`${__dirname}/../../generated/sheets.js`,'utf8'),{module:exported});
const {sheetTarget,sheetExit}=exported.exports;

const anchors=[{name:'minimized',height:56},{name:'default',height:360},{name:'full',height:644}];
test('sheet dragging can reach every anchor and dismiss below the peek',()=>{
 assert.equal(sheetTarget(anchors,'full',620,0,56,.125),'full');
 assert.equal(sheetTarget(anchors,'full',390,0,56,.125),'default');
 assert.equal(sheetTarget(anchors,'default',100,0,56,.125),'minimized');
 assert.equal(sheetTarget(anchors,'minimized',340,0,56,.125),'default');
 assert.equal(sheetTarget(anchors,'default',600,0,56,.125),'full');
 assert.equal(sheetTarget(anchors,'default',20,0,56,.125),'hidden');
});
test('sheet flings use the neighboring anchor in the final direction',()=>{
 assert.equal(sheetTarget(anchors,'default',350,-.2,56,.125),'full');
 assert.equal(sheetTarget(anchors,'full',630,.2,56,.125),'default');
 assert.equal(sheetTarget(anchors,'default',600,.2,56,.125),'default');
 assert.equal(sheetTarget(anchors,'minimized',55,.2,56,.125),'hidden');
 assert.equal(sheetTarget(anchors,'minimized',55,-.2,56,.125),'default');
 assert.equal(sheetTarget(anchors,'full',20,-.2,56,.125),'minimized');
});
test('side sheet enter and exit direction follows logical edge and text direction',()=>{
 assert.equal(sheetExit(true,false,false),'translateY(100%)');
 assert.equal(sheetExit(false,false,false),'translateX(100%)');
 assert.equal(sheetExit(false,true,false),'translateX(-100%)');
 assert.equal(sheetExit(false,false,true),'translateX(-100%)');
 assert.equal(sheetExit(false,true,true),'translateX(100%)');
});

function sheetFixture(allowFull=false,motion=false,sideVariant) {
 const animations=[];
 const listeners={},attrs=new Set(allowFull?['data-sheet-allow-full']:[]),body={hidden:false,inert:false};
 const handle={
  setAttribute(){},focus(){},setPointerCapture(){},releasePointerCapture(){},hasPointerCapture(){return true;},
  hasAttribute(name){return name==='data-sheet-handle';},
  matches(selector){return selector==='[data-sheet-handle]';},
  closest(selector){return selector==='button'||selector==='[data-sheet-handle]'?handle:selector.endsWith(' .md-sheet')?sheet:host;},
 };
 class Dialog {}
 const sheet=Object.assign(new Dialog(),{
  dataset:{variant:sideVariant||'expanded',control:'modal',sheetSize:'default'},style:{height:''},open:true,
  classList:{contains:name=>name===(sideVariant?'md-side-sheet':'md-bottom-sheet')},
  matches:selector=>!sideVariant&&selector==='dialog.md-bottom-sheet[data-control="modal"]',
  hasAttribute:name=>attrs.has(name),
  toggleAttribute(name,value){value?attrs.add(name):attrs.delete(name);},
  closest(){return host;},
  querySelector:selector=>selector==='.md-sheet-content'?body:sideVariant&&selector==='[data-sheet-handle]'?null:handle,
  getBoundingClientRect(){const height=parseFloat(this.style.height)||(this.dataset.variant==='minimized'?56:this.dataset.sheetSize==='full'?644:360);return {height,bottom:720,top:720-height};},
  addEventListener(){},
  setAttribute(name){attrs.add(name);},removeAttribute(name){attrs.delete(name);},dispatchEvent(){return true;},
  show(){this.open=true;},showModal(){this.open=true;},close(){this.open=false;},
  animate(frames,timing){const animation={frames,timing,finished:Promise.resolve(),cancel(){}};animations.push(animation);return animation;},
 });
 const host={querySelector:selector=>selector==='.md-sheet'?sheet:null};
 const context={
  document:{querySelectorAll:selector=>selector.startsWith(':is(')?[sheet]:[],addEventListener(name,fn){listeners[name]=fn;}},
  matchMedia:()=>({matches:!motion,addEventListener(){}}),
  CustomEvent:class {constructor(type,options){this.type=type;Object.assign(this,options);}},
  MaterialMotion:{attachedTiming:()=>({duration:400,easing:'cubic-bezier(0.2, 0, 0, 1)'}),timing:()=>({duration:400,easing:'linear(0, 1.1, 1)'})},
  HTMLDialogElement:Dialog,window:{innerHeight:720},
  getComputedStyle:()=>({transform:'none',getPropertyValue:key=>key==='--md-sys-measurement-space100'?'8':key==='--_sheet-drag-distance'?'56':'125'}),
 };
 runInNewContext(readFileSync(`${__dirname}/../../generated/sheets.js`,'utf8'),context);
 const click=(detail=1)=>listeners.click({target:handle,detail,preventDefault(){}});
 const pointer=(type,y,time=0)=>listeners[type]({target:handle,button:0,isPrimary:true,pointerType:'pen',pointerId:1,clientY:y,timeStamp:time,preventDefault(){}});
 return {sheet,body,listeners,handle,click,pointer,host,animations};
}

test('sheet handle cycles through all enabled heights and returns to default',async()=>{
 for(const allowFull of [false,true]) {
  const f=sheetFixture(allowFull),cycle=allowFull?['full','minimized','default','full','minimized','default']:['minimized','default','minimized','default'];
  for(const size of cycle) {
   f.pointer('pointerdown',100);f.pointer('pointerup',100,10);f.click(0);
   await Promise.resolve();
   assert.equal(f.sheet.dataset.sheetSize,size);
   assert.equal(f.sheet.dataset.variant,size==='minimized'?'minimized':'expanded');
   assert.equal(f.sheet.style.height,'','settled state must release its measured height');
   assert.equal(f.body.hidden,size==='minimized');assert.equal(f.body.inert,size==='minimized');
   assert.equal(f.sheet.getBoundingClientRect().height,anchors.find(anchor=>anchor.name===size).height);
  }
 }
});

test('a zero-detail compatibility click cannot toggle a sheet again after dragging',async()=>{
 const f=sheetFixture();
 f.pointer('pointerdown',100);f.pointer('pointermove',360,100);f.pointer('pointerup',360,300);
 await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'minimized');
 f.click(0);await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'minimized','drag must not be followed by a second toggle');
 f.listeners.keydown({target:f.handle,key:'Enter'});f.click(0);await Promise.resolve();
 assert.equal(f.sheet.dataset.sheetSize,'default','keyboard activation restores the default height');
 const keyboard=sheetFixture();
 keyboard.pointer('pointerdown',100);keyboard.pointer('pointermove',360,100);keyboard.pointer('pointerup',360,300);
 await Promise.resolve();
 keyboard.listeners.keydown({target:keyboard.handle,key:'Enter'});keyboard.click(0);await Promise.resolve();
 assert.equal(keyboard.sheet.dataset.sheetSize,'default','a drag without a compatibility click must not swallow keyboard activation');
});

test('a cancelled sheet drag restores its original state',async()=>{
 const f=sheetFixture();
 f.pointer('pointerdown',100);f.pointer('pointermove',360,100);f.pointer('pointercancel',200,150);
 await Promise.resolve();
 assert.equal(f.sheet.dataset.sheetSize,'default');assert.equal(f.body.hidden,false);assert.equal(f.sheet.style.height,'');
});

test('sheet arrow keys visit adjacent anchors and Home/End select the extremes',async()=>{
 const f=sheetFixture(true);
 for(const [key,size] of [['ArrowUp','full'],['ArrowDown','default'],['ArrowDown','minimized'],['ArrowUp','default'],['Home','full'],['End','minimized']]) {
  f.listeners.keydown({target:f.handle,key,preventDefault(){}});await Promise.resolve();
  assert.equal(f.sheet.dataset.sheetSize,size);
 }
});


test('sheet translations use attached easing while anchored height changes keep spatial motion',async()=>{
 const f=sheetFixture(false,true);
 function action(attribute){const button={hasAttribute:name=>name===attribute,setAttribute(){},closest:selector=>selector==='button'?button:f.host};f.listeners.click({target:button,preventDefault(){}});}
 action('data-sheet-close');await new Promise(resolve=>setImmediate(resolve));
 assert.equal(f.sheet.open,false);assert.equal(f.animations[0].timing.easing,'cubic-bezier(0.2, 0, 0, 1)');
 action('data-sheet-open');await new Promise(resolve=>setImmediate(resolve));
 assert.equal(f.sheet.open,true);assert.equal(f.animations[1].timing.easing,'cubic-bezier(0.2, 0, 0, 1)');
 f.click();await new Promise(resolve=>setImmediate(resolve));
 assert.equal(f.animations[2].timing.easing,'linear(0, 1.1, 1)');assert.ok(f.animations[2].frames.every(frame=>frame.height));
});

test('side sheet variants survive closing and reopening for either behavior',async()=>{
 for(const variant of ['docked','detached'])for(const control of ['standard','modal']){
  const f=sheetFixture(false,true,variant);f.sheet.dataset.control=control;
  if(control==='standard'){Object.setPrototypeOf(f.sheet,Object.prototype);f.sheet.hidden=false;}
  const action=attribute=>{
   const button={hasAttribute:name=>name===attribute,setAttribute(){},focus(){},closest:selector=>selector==='button'?button:f.host};
   f.listeners.click({target:button,preventDefault(){}});
  };
  for(let cycle=0;cycle<2;cycle++){
   action('data-sheet-close');await new Promise(resolve=>setImmediate(resolve));
   assert.equal(control==='modal'?f.sheet.open:!f.sheet.hidden,false);
   action('data-sheet-open');await new Promise(resolve=>setImmediate(resolve));
   assert.equal(control==='modal'?f.sheet.open:!f.sheet.hidden,true);
   assert.equal(f.sheet.dataset.variant,variant,`${control} ${variant}`);
   assert.equal(f.sheet.dataset.sheetSize,'default');
   assert.equal(f.body.hidden,false);
   assert.equal(f.sheet.style.height,'');
  }
 }
});
