const {test}=require('node:test');
const assert=require('node:assert/strict');
const {readFileSync}=require('node:fs');
const {runInNewContext}=require('node:vm');

test('a floating toolbar uses its allocated parent width without charging scene clearance again',()=>{
 const items=[0,1,2].map(index=>({dataset:{toolbarItem:String(index)}}));
 const parent={clientWidth:480,closest:()=>null,querySelector:()=>null};
 const group={parentElement:parent,closest:()=>null,dataset:{variant:'floating',orientation:'horizontal'},scrollWidth:480,querySelectorAll:()=>items,setAttribute(){},style:{setProperty(){}}};
 runInNewContext(readFileSync(`${__dirname}/behaviors/toolbars.js`,'utf8'),{
  MaterialComponents:{},ResizeObserver:class{observe(){}},
  getComputedStyle:()=>({getPropertyValue:()=>''}),
  document:{querySelectorAll:()=>[group],addEventListener(){}},
 });
 assert.ok(items.every(item=>!item.hidden),'all three actions fit the parent');
});
