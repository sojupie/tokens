const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const exported={exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/sheets.js`,'utf8'),{module:exported});
const {sheetTarget,sheetExit}=exported.exports;

test('sheet dragging respects positional thresholds and available anchors',()=>{
 assert.equal(sheetTarget('expanded',360,384,56,0,56,.125),'expanded');
 assert.equal(sheetTarget('expanded',300,384,56,0,56,.125),'minimized');
 assert.equal(sheetTarget('expanded',20,384,56,0,56,.125),'hidden');
 assert.equal(sheetTarget('minimized',100,56,56,0,56,.125),'minimized');
 assert.equal(sheetTarget('minimized',120,56,56,0,56,.125),'expanded');
});
test('sheet flings follow the last direction even after a reverse drag',()=>{
 assert.equal(sheetTarget('expanded',370,384,56,.2,56,.125),'minimized');
 assert.equal(sheetTarget('minimized',55,56,56,.2,56,.125),'hidden');
 assert.equal(sheetTarget('minimized',55,56,56,-.2,56,.125),'expanded');
 assert.equal(sheetTarget('expanded',20,384,56,-.2,56,.125),'expanded');
});
test('side sheet enter and exit direction follows logical edge and text direction',()=>{
 assert.equal(sheetExit(true,false,false),'translateY(100%)');
 assert.equal(sheetExit(false,false,false),'translateX(100%)');
 assert.equal(sheetExit(false,true,false),'translateX(-100%)');
 assert.equal(sheetExit(false,false,true),'translateX(-100%)');
 assert.equal(sheetExit(false,true,true),'translateX(100%)');
});

function sheetFixture() {
 const listeners={},attrs=new Set(),body={hidden:false,inert:false};
 const handle={
  setAttribute(){},focus(){},setPointerCapture(){},releasePointerCapture(){},hasPointerCapture(){return true;},
  hasAttribute(name){return name==='data-sheet-handle';},
  matches(selector){return selector==='[data-sheet-handle]';},
  closest(selector){return selector==='button'||selector==='[data-sheet-handle]'?handle:selector.endsWith(' .md-sheet')?sheet:host;},
 };
 class Dialog {}
 const sheet=Object.assign(new Dialog(),{
  dataset:{variant:'expanded',control:'modal'},style:{height:''},open:true,
  classList:{contains:name=>name==='md-bottom-sheet'},
  matches:selector=>selector==='dialog.md-bottom-sheet[data-control="modal"]',
  hasAttribute:name=>attrs.has(name),
  toggleAttribute(name,value){value?attrs.add(name):attrs.delete(name);},
  closest(){return host;},
  querySelector:selector=>selector==='.md-sheet-content'?body:handle,
  getBoundingClientRect(){const height=parseFloat(this.style.height)||(this.dataset.variant==='minimized'?56:attrs.has('data-sheet-full-height')?644:360);return {height,bottom:720,top:720-height};},
  addEventListener(){},
 });
 const host={querySelector:selector=>selector==='.md-sheet'?sheet:null};
 const context={
  document:{querySelectorAll:selector=>selector.startsWith(':is(')?[sheet]:[],addEventListener(name,fn){listeners[name]=fn;}},
  matchMedia:()=>({matches:true,addEventListener(){}}),
  HTMLDialogElement:Dialog,window:{innerHeight:720},
  getComputedStyle:()=>({getPropertyValue:key=>key==='--md-sys-measurement-space100'?'8':key==='--_sheet-drag-distance'?'56':'125'}),
 };
 runInNewContext(readFileSync(`${__dirname}/static/sheets.js`,'utf8'),context);
 const click=(detail=1)=>listeners.click({target:handle,detail,preventDefault(){}});
 const pointer=(type,y,time=0)=>listeners[type]({target:handle,button:0,isPrimary:true,pointerType:'pen',pointerId:1,clientY:y,timeStamp:time,preventDefault(){}});
 return {sheet,body,listeners,handle,click,pointer};
}

test('sheet handle taps keep variant, height and content synchronized across cycles',async()=>{
 const f=sheetFixture();
 for(const variant of ['minimized','expanded','minimized']) {
  f.pointer('pointerdown',100);f.pointer('pointerup',100,10);f.click(0);
  await Promise.resolve();
  assert.equal(f.sheet.dataset.variant,variant);
  assert.equal(f.sheet.style.height,'','settled state must release its measured height');
  assert.equal(f.body.hidden,variant==='minimized');
  assert.equal(f.body.inert,variant==='minimized');
  assert.equal(f.sheet.getBoundingClientRect().height,variant==='minimized'?56:644);
 }
});

test('a zero-detail compatibility click cannot toggle a sheet again after dragging',async()=>{
 const f=sheetFixture();
 f.pointer('pointerdown',100);f.pointer('pointermove',200,100);f.pointer('pointerup',200,300);
 await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'minimized');
 f.click(0);await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'minimized','drag must not be followed by a second toggle');
 f.listeners.keydown({target:f.handle,key:'Enter'});f.click(0);await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'expanded','keyboard activation remains available');
 const keyboard=sheetFixture();
 keyboard.pointer('pointerdown',100);keyboard.pointer('pointermove',200,100);keyboard.pointer('pointerup',200,300);
 await Promise.resolve();
 keyboard.listeners.keydown({target:keyboard.handle,key:'Enter'});keyboard.click(0);await Promise.resolve();
 assert.equal(keyboard.sheet.dataset.variant,'expanded','a drag without a compatibility click must not swallow keyboard activation');
});

test('a cancelled sheet drag restores its original state',async()=>{
 const f=sheetFixture();
 f.pointer('pointerdown',100);f.pointer('pointermove',200,100);f.pointer('pointercancel',200,150);
 await Promise.resolve();
 assert.equal(f.sheet.dataset.variant,'expanded');assert.equal(f.body.hidden,false);assert.equal(f.sheet.style.height,'');
});
