const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/fields.js`,'utf8'),{module:out});const {matches,popupPosition,typeahead}=out.exports;
test('autocomplete matches case-insensitively without changing original indices',()=>{assert.deepEqual(Array.from(matches(['One','Another','ONE TWO'],' one ')),[0,2]);assert.equal(matches(['One'],'none').length,0);});
test('popup placement uses available viewport space and keeps the field width',()=>{let p=popupPosition({left:50,top:30,bottom:86,width:240},{width:800,height:600},200,8);assert.equal(p.top,86);assert.equal(p.width,240);p=popupPosition({left:750,top:530,bottom:586,width:240},{width:800,height:600},200,8);assert.equal(p.top,330);assert.equal(p.left,552);assert.equal(p.height,522);p=popupPosition({left:0,top:0,bottom:56,width:400},{width:200,height:100},200,8);assert.equal(p.width,184);assert.equal(p.left,8);assert.equal(p.height,36);});
test('select typeahead searches prefixes and cycles repeated letters',()=>{assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'aa'),1);assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'ap'),1);assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'z'),-1);});

test('readonly inputs and pill buttons toggle and survive native light-dismiss',()=>{
 for(const useButton of [false,true]) {
 const listeners={},attrs={},positions={},empty={hidden:true};let shown=false,focused=0,prevented=0;
 const options=['One','Two'].map((name,index)=>({id:`option-${index}`,hidden:false,dataset:{fieldOption:name},classList:{contains:c=>c==='md-menu-item'},setAttribute(){},scrollIntoView(){}}));
 const menu={dataset:{},scrollWidth:260,scrollHeight:2000,offsetHeight:120,style:{setProperty:(name,value)=>positions[name]=value},matches:()=>shown,showPopover(){shown=true;},hidePopover(){shown=false;},addEventListener(){},querySelectorAll:()=>options,querySelector:s=>s==='.md-field-empty'?empty:null};
 const root={dataset:{},getBoundingClientRect:()=>({left:20,top:300,bottom:356,width:200}),querySelector:s=>s==='.md-field-menu'?menu:null};
 const input={classList:{contains:name=>useButton&&name==='md-filter-chip'},readOnly:!useButton,hasAttribute:name=>useButton&&name==='data-field-select',value:'One',closest:s=>s==='[data-field-example]'?root:null,focus(){focused++;},setAttribute(k,v){attrs[k]=v;},removeAttribute(k){delete attrs[k];}};
 runInNewContext(readFileSync(`${__dirname}/static/fields.js`,'utf8'),{
  document:{readyState:'complete',documentElement:{clientWidth:800},querySelectorAll:s=>s==='[data-field-combo]'?[input]:[],addEventListener:(name,fn)=>listeners[name]=fn},
  window:{innerHeight:600,addEventListener(){}},getComputedStyle:()=>({getPropertyValue:()=> '8px'})
 });
 const click=()=>listeners.click({target:input,detail:1,preventDefault(){}});
 const down=()=>listeners.pointerdown({target:input,preventDefault(){prevented++;}});
 down();click();assert.equal(shown,true);assert.equal(attrs['aria-expanded'],'true');assert.equal(positions['--_field-popup-top'],'356px','capped menu fits below instead of flipping above for its full scroll height');assert.equal(positions['--_field-popup-width'],useButton?'260px':'200px','chip popups fit their options while text fields retain their width');
 down();shown=false;click();assert.equal(shown,false,'must not reopen after native pointer dismissal');
 down();click();assert.equal(shown,true);
 options[0].closest=selector=>selector==='[data-field-example]'?root:selector==='[data-field-option]'?options[0]:null;
 listeners.pointerdown({target:options[0],pointerType:'touch',pointerId:7,preventDefault(){}});
 listeners.pointercancel({pointerId:7});
 assert.equal(shown,true,'scrolling a menu must cancel the pending option tap without dismissing the menu');
 listeners.pointerup({target:options[0],pointerId:7,preventDefault(){throw Error('cancelled tap was committed');}});
 assert.equal(input.value,'One');
 listeners.keydown({target:input,key:'ArrowDown',preventDefault(){}});
 assert.equal(options[0].dataset.componentState,'focused');
 listeners.keydown({target:input,key:'Escape',preventDefault(){},stopPropagation(){}});
 assert.equal(shown,false);assert.equal(attrs['aria-expanded'],'false');assert.equal(options[0].dataset.componentState,undefined);
 assert.equal(input.value,'One');assert.equal(focused,3);assert.equal(prevented,3);
 }
});

test('chip option content can widen a popup without escaping the viewport',()=>{
 const field={left:680,top:40,bottom:72,width:100},viewport={width:800,height:600};
 let p=popupPosition(field,viewport,100,8,180);
 assert.equal(p.width,180);assert.equal(p.left,612);
 p=popupPosition(field,viewport,100,8,1200);
 assert.equal(p.width,784);assert.equal(p.left,8);
 assert.equal(popupPosition(field,viewport,100,8,60).width,100);
});
