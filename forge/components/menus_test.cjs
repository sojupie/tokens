function runMotionScript(code, globals) {
  runInNewContext(readFileSync(require.resolve('./static/motion.cjs'), 'utf8'), globals);
  return runInNewContext(code, globals);
}
const test=require('node:test'),assert=require('node:assert/strict'),{readFileSync}=require('node:fs'),{runInNewContext}=require('node:vm');const out={exports:{}};runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{module:out});const {menuIndex}=out.exports;
test('menu navigation wraps within its orientation and mirrors horizontal keys',()=>{assert.equal(menuIndex(2,3,'ArrowDown',false,false),0);assert.equal(menuIndex(0,3,'ArrowUp',false,false),2);assert.equal(menuIndex(0,3,'ArrowLeft',true,true),1);assert.equal(menuIndex(1,3,'ArrowRight',true,true),0);assert.equal(menuIndex(1,3,'Home',true,false),0);assert.equal(menuIndex(0,3,'End',false,false),2);assert.equal(menuIndex(0,3,'ArrowRight',false,false),null);});
test('pressing a submenu trigger cancels its delayed hover opener',()=>{
 const events={},timers=new Set();let sequence=0;
 const root={dataset:{},closest:()=>demo,matches:()=>false,hasAttribute:key=>key==='popover'},demo={querySelectorAll:()=>[root]};
 const submenu={hasAttribute:()=>true,matches:()=>false},trigger={dataset:{menuSubmenu:'child'},contains:()=>false,closest:()=>root,hasAttribute:()=>true};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[],getElementById:()=>submenu};
 runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){},setTimeout(fn,ms){assert.equal(ms,200);timers.add(++sequence);return sequence;},clearTimeout(id){timers.delete(id);}});
 const event={target:{closest:()=>trigger},pointerType:'mouse',buttons:0,relatedTarget:null};
 events.pointerover(event);assert.equal(timers.size,1);events.pointerdown(event);assert.equal(timers.size,0);
 events.pointerover({...event,buttons:1});assert.equal(timers.size,0);
});
test('each selectable menu item updates its state and check icon',()=>{
 const events={},output={textContent:''};
 const demo={querySelector:selector=>selector==='[data-menu-output]'?output:null};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[]};
 runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){}});
 for(const label of ['Favorite','Download','Edit']){
  const attributes={'role':'menuitemcheckbox','aria-label':label,'aria-checked':'false'},checked={hidden:true},unchecked={hidden:false};
  const item={dataset:{},getAttribute:name=>attributes[name],setAttribute:(name,value)=>attributes[name]=value,
   closest:selector=>selector==='[data-menu-example]'?demo:null,
   querySelector:selector=>selector==='[data-menu-checked-icon]'?checked:unchecked};
  const event={target:{closest:selector=>selector==='[data-menu-back]'||selector==='[data-menu-submenu],[data-menu-open]'?null:item}};
  events.click(event);
  assert.equal(attributes['aria-checked'],'true');assert.equal(item.dataset.selected,'true');assert.equal(checked.hidden,false);assert.equal(unchecked.hidden,true);assert.equal(output.textContent,label+' selected');
  events.click(event);assert.equal(attributes['aria-checked'],'false');assert.equal(checked.hidden,true);assert.equal(unchecked.hidden,false);
 }
});
test('single selection replaces the previous item and multi selection stays open',()=>{
 const events={},output={textContent:''};
 const demo={querySelector:selector=>selector==='[data-menu-output]'?output:null};
 const menu={dataset:{menuSelection:'single'},querySelectorAll:()=>nodes};
 const nodes=['Favorite','Download','Edit'].map((label,index)=>{
  const attrs={role:'menuitemradio','aria-label':label,'aria-checked':String(index===0)},check={hidden:index!==0};
  return {dataset:{},getAttribute:name=>attrs[name],setAttribute:(name,value)=>attrs[name]=value,
   closest:selector=>selector==='[data-menu-panel]'?menu:selector==='[data-menu-example]'?demo:null,
   querySelector:selector=>selector==='[data-menu-checked-icon]'?check:null,check};
 });
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[]};
 runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){}});
 const click=index=>events.click({target:{closest:selector=>selector==='[data-menu-back]'||selector==='[data-menu-submenu],[data-menu-open]'?null:nodes[index]}});
 click(2);assert.deepEqual(nodes.map(n=>n.getAttribute('aria-checked')),['false','false','true']);click(1);assert.deepEqual(nodes.map(n=>n.getAttribute('aria-checked')),['false','true','false']);
 menu.dataset.menuSelection='multiple';for(const node of nodes)node.setAttribute('role','menuitemcheckbox');
 // Output remains available, but querying a popover for dismissal is forbidden.
 demo.querySelector=selector=>{if(selector==='[data-menu-output]')return output;throw Error('multi selection closed its menu');};
 click(2);assert.deepEqual(nodes.map(n=>n.getAttribute('aria-checked')),['false','true','true']);
});
test('Firefox exit is painted synchronously without re-entering the popover lifecycle',async()=>{
 for(const modal of [false,true]) {
 const events={},values=new Map(),frames=[],timers=new Map();let sequence=0,removed=false,appended=false,unpopovered=false,copies=0;
 const copy={style:{setProperty:(key,value)=>values.set(key,value)},removeAttribute(name){if(name==='popover')unpopovered=true;},querySelectorAll:()=>[],setAttribute(){},
  showPopover(){throw Error('exit must not re-enter native popover dismissal');},
  remove(){removed=true;},animate(keyframes,options){assert.equal(appended,true);frames.push({keyframes,options});return {finished:Promise.resolve(),cancel(){},effect:{getComputedTiming:()=>({endTime:options.duration})}};}};
 const style=Object.assign(['display','opacity','--_menu-container','--md-sys-color-primary'],{opacity:'1',getPropertyValue:name=>name==='--_motion-fast-effects-duration'?'150ms':name==='--_motion-fast-spatial-duration'?'350ms':name.endsWith('-easing')?'linear(0, 1)':name==='--_menu-container'?'rgb(1, 2, 3)':name==='display'?'block':'1'});
 const parent={append(){appended=true;}};
 const menu={scrollTop:80,scrollLeft:0,matches:selector=>selector!=='.md-fab-menu-panel',hasAttribute:()=>false,closest:()=>modal?parent:null,cloneNode:()=>{copies++;return copy;}};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,querySelectorAll:()=>[],body:modal?{append(){throw Error('modal copy must remain inside the dialog');}}:parent};
 runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){},CSS:{supports:()=>false},matchMedia:()=>({matches:false}),getComputedStyle:()=>style,setTimeout(fn,ms){timers.set(++sequence,{fn,ms});return sequence;},clearTimeout:id=>timers.delete(id)});
 events.beforetoggle({newState:'closed',target:menu});
 assert.equal(appended,true,'copy must be present before the original disappears');assert.equal(unpopovered,true);assert.equal(copies,1);
 events.beforetoggle({newState:'closed',target:menu});assert.equal(copies,1);
 assert.equal(copy.inert,true);assert.equal(copy.scrollTop,80);assert.match(copy.style.cssText,/display:block;/);assert.equal(values.get('pointer-events'),'none');assert.match(copy.style.cssText,/--_menu-container:rgb\(1, 2, 3\);/);assert.equal(values.has('--md-sys-color-primary'),false);
 assert.equal(frames[0].options.duration,150);assert.equal(frames[0].keyframes[0].opacity,'1');assert.equal(frames[0].keyframes[1].opacity,0);
 assert.equal(Array.from(timers.values()).some(t=>t.ms===0),false,'no delayed copy handoff');
 const watchdog=Array.from(timers.values()).find(t=>t.ms===400);assert.ok(watchdog);watchdog.fn();assert.equal(removed,true,'cleanup must not depend on animation completion');
 await Promise.resolve();assert.equal(removed,true);
 }
});
test('clicking an open menu trigger closes after native light-dismiss',()=>{
 const events={};let shown=true,focused=false;
 const trigger={dataset:{menuOpen:'main'},setAttribute(){},closest:()=>null,focus(){focused=true;}};
 const demo={querySelector:()=>trigger};
 const menu={matches:()=>shown,hasAttribute:()=>true,closest:()=>demo,hidePopover(){shown=false;},showPopover(){throw Error('native dismissal must not reopen the menu');}};
 const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,getElementById:()=>menu,querySelectorAll:()=>[]};
 runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){},clearTimeout(){}});
 const event={target:{closest:selector=>selector==='[data-menu-back]'?null:trigger},detail:1};events.pointerdown(event);shown=false;events.click(event);assert.equal(shown,false);assert.equal(focused,true);
});

test('submenus use the parent surface edge, flip, or replace when neither side fits',()=>{
 const {submenuPosition}=out.exports;
 assert.deepEqual({...submenuPosition({left:100,right:420},{width:200},1000,8)}, {left:420,replace:false});
 assert.deepEqual({...submenuPosition({left:650,right:970},{width:200},1000,8)}, {left:450,replace:false});
 assert.deepEqual({...submenuPosition({left:24,right:369},{width:200},393,8)}, {left:24,replace:true});
 assert.deepEqual({...submenuPosition({left:400,right:720},{width:200},1000,8,true)}, {left:200,replace:false});
});

test('replacement submenus require activation instead of opening beneath a stationary pointer',()=>{
 for(const replaced of [false,true]) {
  const events={};let scheduled=0;
  const demo={querySelectorAll:()=>[]};
  const menu={closest:()=>demo,querySelectorAll:()=>[],hasAttribute:key=>replaced&&key==='data-menu-replace'};
  const item={dataset:{menuSubmenu:'child'},closest:()=>menu,contains:()=>false,hasAttribute:key=>key==='data-menu-submenu'};
  const document={readyState:'loading',addEventListener:(name,fn)=>events[name]=fn,getElementById:()=>({hasAttribute:()=>true})};
  runMotionScript(readFileSync(`${__dirname}/static/menus.js`,'utf8'),{document,addEventListener(){},setTimeout(){scheduled++;return 1;}});
  events.pointerover({target:{closest:()=>item},pointerType:'mouse',buttons:0});
  assert.equal(scheduled,replaced?0:1);
 }
});
