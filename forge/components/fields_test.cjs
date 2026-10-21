const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');const out={exports:{}};runInNewContext(readFileSync(`${__dirname}/static/fields.js`,'utf8'),{module:out});const {matches,popupPosition,typeahead}=out.exports;
test('autocomplete matches case-insensitively without changing original indices',()=>{assert.deepEqual(Array.from(matches(['One','Another','ONE TWO'],' one ')),[0,2]);assert.equal(matches(['One'],'none').length,0);});
test('popup placement uses available viewport space and keeps the field width',()=>{let p=popupPosition({left:50,top:30,bottom:86,width:240},{width:800,height:600},200,8);assert.equal(p.top,86);assert.equal(p.width,240);p=popupPosition({left:750,top:530,bottom:586,width:240},{width:800,height:600},200,8);assert.equal(p.top,330);assert.equal(p.left,552);assert.equal(p.height,522);p=popupPosition({left:0,top:0,bottom:56,width:400},{width:200,height:100},200,8);assert.equal(p.width,184);assert.equal(p.left,8);assert.equal(p.height,36);});
test('select typeahead searches prefixes and cycles repeated letters',()=>{assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'aa'),1);assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'ap'),1);assert.equal(typeahead(['Alpha','Apricot','Beta'],0,'z'),-1);});

test('readonly inputs and pill buttons toggle and survive native light-dismiss',()=>{
 for(const useButton of [false,true]) {
 const listeners={},attrs={},positions={},empty={hidden:true};let shown=false,focused=0,prevented=0;
 const options=['One','Two'].map((name,index)=>({id:`option-${index}`,hidden:false,dataset:{fieldOption:name},classList:{contains:c=>c==='md-menu-item'},setAttribute(){},scrollIntoView(){}}));
 const menu={dataset:{},scrollHeight:2000,offsetHeight:120,style:{setProperty:(name,value)=>positions[name]=value},matches:()=>shown,showPopover(){shown=true;},hidePopover(){shown=false;},addEventListener(){},querySelectorAll:()=>options,querySelector:s=>s==='.md-field-empty'?empty:null};
 const root={dataset:{},getBoundingClientRect:()=>({left:20,top:300,bottom:356,width:200}),querySelector:s=>s==='.md-field-menu'?menu:null};
 const input={readOnly:!useButton,hasAttribute:name=>useButton&&name==='data-field-select',value:'One',closest:s=>s==='[data-field-example]'?root:null,focus(){focused++;},setAttribute(k,v){attrs[k]=v;},removeAttribute(k){delete attrs[k];}};
 runInNewContext(readFileSync(`${__dirname}/static/fields.js`,'utf8'),{
  document:{readyState:'complete',documentElement:{clientWidth:800},querySelectorAll:s=>s==='[data-field-combo]'?[input]:[],addEventListener:(name,fn)=>listeners[name]=fn},
  window:{innerHeight:600,addEventListener(){}},getComputedStyle:()=>({getPropertyValue:()=> '8px'})
 });
 const click=()=>listeners.click({target:input,detail:1,preventDefault(){}});
 const down=()=>listeners.pointerdown({target:input,preventDefault(){prevented++;}});
 down();click();assert.equal(shown,true);assert.equal(attrs['aria-expanded'],'true');assert.equal(positions['--_field-popup-top'],'356px','capped menu fits below instead of flipping above for its full scroll height');
 down();shown=false;click();assert.equal(shown,false,'must not reopen after native pointer dismissal');
 down();click();assert.equal(shown,true);
 listeners.keydown({target:input,key:'ArrowDown',preventDefault(){}});
 assert.equal(options[0].dataset.componentState,'focused');
 listeners.keydown({target:input,key:'Escape',preventDefault(){},stopPropagation(){}});
 assert.equal(shown,false);assert.equal(attrs['aria-expanded'],'false');assert.equal(options[0].dataset.componentState,undefined);
 assert.equal(input.value,'One');assert.equal(focused,3);assert.equal(prevented,3);
 }
});
