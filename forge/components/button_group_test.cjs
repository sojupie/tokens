const {test}=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');
test('button group navigation moves its sole Tab stop and excludes unavailable controls',()=>{
 const listeners={};let focused;
 const group={querySelectorAll:selector=>selector==='button'?items:items.filter(i=>!i.disabled)};
 const items=Array.from({length:3},()=>({tabIndex:-1,getClientRects:()=>[{}],closest:s=>s==='[data-group-button]'?null:s==='[inert]'?null:group,focus(){focused=this},getAttribute:()=>null}));
 items[0].tabIndex=0;for(const item of items){const closest=item.closest;item.closest=s=>s==='[data-group-button]'?item:closest(s)}items[1].disabled=true;
 const context={document:{addEventListener:(name,fn)=>listeners[name]=fn},getComputedStyle:()=>({visibility:'visible',direction:'ltr'})};runInNewContext(readFileSync(`${__dirname}/behaviors/utils.js`,'utf8'),context);runInNewContext(readFileSync(`${__dirname}/behaviors/button-groups.js`,'utf8'),context);
 const event={target:items[0],key:'ArrowRight',preventDefault(){this.prevented=true}};listeners.keydown(event);assert.equal(focused,items[2]);assert.deepEqual(items.map(i=>i.tabIndex),[-1,-1,0]);assert.equal(event.prevented,true);
});
