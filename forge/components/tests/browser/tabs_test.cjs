const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');
function harness(rtl=false) {
 const handlers={},events=[]; let focused=0,removed=false;
 const panels=[0,1,2].map(i=>({id:`panel-${i}`,hidden:i!==0,closest:()=>example}));
 const example={querySelectorAll:()=>panels};
 const list={dispatchEvent(event){events.push(event);return true},dataset:{},closest:()=>example,querySelectorAll:()=>items,querySelector:()=>items.find(i=>i.selected==='true')};
 const items=[0,1,2].map(i=>({id:`tab-${i}`,selected:String(i===0),tabIndex:i===0?0:-1,disabled:false,
  closest:s=>s==='[inert]'?null:s==='.md-tab'?items[i]:list,
  getClientRects:()=>[{}], scrollIntoView(){},
  querySelector:s=>s==='.md-badge'?{remove(){removed=true}}:{getBoundingClientRect:()=>({left:i*100,width:30})},
  getAttribute:s=>s==='aria-selected'?items[i].selected:`panel-${i}`,
  setAttribute(s,v){if(s==='aria-selected') this.selected=v},
  focus(){focused=i;handlers.focusin({target:this})}
 }));
 runInNewContext(readFileSync(`${__dirname}/../../behaviors/tabs.js`,'utf8'),{
  CustomEvent:class {constructor(type,options){this.type=type;Object.assign(this,options)}},
  document:{addEventListener:(type,fn)=>{handlers[type]=fn}},
  getComputedStyle:()=>({direction:rtl?'rtl':'ltr',visibility:'visible'}),matchMedia:()=>({matches:true})
 });
 return {items,panels,events,get removed(){return removed},swipe(dx,dy=0){
  const panel={parentElement:{querySelector:()=>list}};
  const target={closest:s=>s.includes('tabpanel')?panel:null};
  handlers.pointerdown({pointerType:'touch',target,pointerId:1,clientX:100,clientY:100});
  handlers.pointerup({pointerId:1,clientX:100+dx,clientY:100+dy});
 },get focused(){return focused},key(index,key){handlers.keydown({target:items[index],key,preventDefault(){}})},click(index){handlers.click({target:items[index]})}};
}
test('tab arrows move focus without activating, and do not loop',()=>{
 const h=harness();h.key(0,'ArrowRight');assert.equal(h.focused,1);assert.equal(h.items[0].selected,'true');
 h.key(2,'ArrowRight');assert.equal(h.focused,2);h.key(0,'ArrowLeft');assert.equal(h.focused,0);
 h.key(0,'End');assert.equal(h.focused,2);h.key(2,'Home');assert.equal(h.focused,0);
});
test('tab activation changes exactly its associated panel',()=>{
 const h=harness();h.click(2);assert.deepEqual(h.panels.map(p=>p.hidden),[true,true,false]);
 assert.deepEqual(h.items.map(t=>t.selected),['false','false','true']);
});
test('RTL arrows reverse and unavailable tabs are skipped',()=>{
 const h=harness(true);h.items[1].disabled=true;h.key(0,'ArrowLeft');assert.equal(h.focused,2);
 h.key(2,'ArrowRight');assert.equal(h.focused,0);
});

test('fixed content swipe changes adjacent tab, ignores vertical motion and stops at ends',()=>{
 const h=harness();h.swipe(-80);assert.equal(h.items[1].selected,'true');
 h.swipe(-10,100);assert.equal(h.items[1].selected,'true');
 h.swipe(-80);assert.equal(h.items[2].selected,'true');
 h.swipe(-80);assert.equal(h.items[2].selected,'true');
 h.swipe(80);assert.equal(h.items[1].selected,'true');
});

test('tab activation leaves notification state to the host and emits the selected IDs',()=>{
 const h=harness();h.click(1);assert.equal(h.removed,false);assert.equal(h.events.length,1);
 assert.equal(h.events[0].type,'md-tab-change');assert.equal(h.events[0].detail.id,'tab-1');assert.equal(h.events[0].detail.panel,'panel-1');
});
