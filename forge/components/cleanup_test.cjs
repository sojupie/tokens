const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');

test('HTMX cleanup ignores text nodes while component controllers stay registered',()=>{
 for(const file of ['fields','menus','date-picker','time-picker','dialogs','search','../../../catalog/web/static/site-ui']){
  const listeners=new Map();
  const document={readyState:'loading',querySelectorAll:()=>[],querySelector:()=>null,addEventListener:(name,fn)=>{const handlers=listeners.get(name)||[];handlers.push(fn);listeners.set(name,handlers);}};
  runInNewContext(readFileSync(`${__dirname}/static/${file}.js`,'utf8'),{
   document,window:{addEventListener(){}},addEventListener(){},matchMedia:()=>({matches:false,addEventListener(){}}),ResizeObserver:class {observe(){} unobserve(){}}
  });
  const handlers=listeners.get('htmx:beforeCleanupElement');assert.ok(handlers?.length,`${file} has a cleanup controller`);
  for(const handler of handlers)assert.doesNotThrow(()=>handler({detail:{elt:{nodeType:3,textContent:' '}}}),file);
 }
});
