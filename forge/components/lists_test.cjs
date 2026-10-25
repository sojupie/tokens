const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
const out={exports:{}};
runInNewContext(readFileSync(`${__dirname}/static/lists.js`,'utf8'),{module:out});
const {revealSnap,moveIndex}=out.exports;
test('list reveal uses half the action width and follows writing direction',()=>{
 assert.equal(revealSnap(-35,72),false);assert.equal(revealSnap(-36,72),true);assert.equal(revealSnap(36,72,true),true);assert.equal(revealSnap(-36,72,true),false);
});
test('keyboard reordering stays within the list and preserves the requested position',()=>{
 assert.equal(moveIndex(0,-1,3),0);assert.equal(moveIndex(2,1,3),2);assert.equal(moveIndex(1,-1,3),0);assert.equal(moveIndex(1,1,3),2);
});
test('tall-row alignment follows rendered height and observes swapped rows',()=>{
 const events={},changes=[];let resize;
 const row={getBoundingClientRect:()=>({height:88}),toggleAttribute:(name,on)=>changes.push([name,on])};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[row]};
 const sizes={observe:item=>assert.equal(item,row),unobserve:item=>assert.equal(item,row)};
 runInNewContext(readFileSync(`${__dirname}/static/lists.js`,'utf8'),{document,window:{addEventListener(){}},matchMedia:()=>({matches:false,addEventListener(){}}),ResizeObserver:class{constructor(fn){resize=fn;return sizes;}}});
 events.DOMContentLoaded();resize([{target:row,borderBoxSize:[{blockSize:88}]}]);
 row.getBoundingClientRect=()=>({height:72});resize([{target:row,borderBoxSize:[{blockSize:72}]}]);
 assert.deepEqual(changes,[['data-list-tall',true],['data-list-tall',false]]);
 events['htmx:afterSwap']();events['htmx:beforeCleanupElement']({detail:{elt:{querySelectorAll:()=>[row]}}});
});

test('reorder handle targets its row and updates bottom-sheet boundary actions',()=>{
 const events={},attrs={},up={dataset:{listMove:'up'}},down={dataset:{listMove:'down'}};
 let selected=false,focused=false;
 const sheet={setAttribute:(name,value)=>attrs[name]=value,querySelector:()=>({focus:()=>focused=true})};
 const demo={querySelector:()=>sheet,querySelectorAll:()=>[up,down]};
 const list={dataset:{control:'reorder'},closest:()=>demo,querySelectorAll:()=>[row]};
 const row={dataset:{listLabel:'First item'},querySelector:s=>s==='.md-list-row > .md-list-item'?{disabled:false}:null,closest:()=>list,hasAttribute:name=>name==='data-list-action-target'&&selected,toggleAttribute:(name,on)=>selected=on};
 const button={disabled:false,closest:s=>s==='[data-list-row]'?row:button,hasAttribute:s=>s==='data-list-grip'};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn};
 runInNewContext(readFileSync(`${__dirname}/static/lists.js`,'utf8'),{document,window:{addEventListener(){}},matchMedia:()=>({matches:false,addEventListener(){}}),ResizeObserver:class{}});
 events.click({target:button,detail:1});
 assert.equal(selected,true);assert.equal(attrs['aria-label'],'Reorder First item');assert.equal(up.disabled,true);assert.equal(down.disabled,true);assert.equal(focused,true);
});

test('ordinary list arrows wrap, include horizontal keys and skip unavailable items',()=>{
 const events={},list={dataset:{control:'select'},querySelectorAll:()=>items};let focused;
 const items=Array.from({length:3},(_,index)=>({disabled:false,inert:false,hidden:false,
  closest(selector){if(selector==='.md-list')return list;if(selector==='[data-list-row]')return {};if(selector==='[inert]')return this.inert?{}:null;return null},
  matches:selector=>selector==='.md-list-item',getClientRects(){return this.hidden?[]:[{}]},focus(){focused=this}}));
 const globals={document:{readyState:'loading',addEventListener:(name,fn)=>events[name]=fn},window:{addEventListener(){}},matchMedia:()=>({matches:false,addEventListener(){}}),ResizeObserver:class{},getComputedStyle:()=>({direction:'ltr',visibility:'visible'})};
 runInNewContext(readFileSync(`${__dirname}/behaviors/utils.js`,'utf8'),globals);runInNewContext(readFileSync(`${__dirname}/static/lists.js`,'utf8'),globals);
 const key=(target,key)=>events.keydown({target,key,preventDefault(){}});
 key(items[0],'ArrowUp');assert.equal(focused,items[2]);key(items[2],'ArrowDown');assert.equal(focused,items[0]);
 key(items[0],'ArrowRight');assert.equal(focused,items[1]);key(items[0],'ArrowLeft');assert.equal(focused,items[2]);
 items[1].hidden=true;key(items[0],'ArrowDown');assert.equal(focused,items[2]);
 key(items[0],'End');assert.equal(focused,items[2]);key(items[2],'Home');assert.equal(focused,items[0]);
});
