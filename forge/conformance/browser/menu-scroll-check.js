const run=document.querySelector('#run'),out=document.querySelector('#result'),samples=document.querySelector('#samples');
const settle=w=>new Promise(resolve=>w.requestAnimationFrame(()=>w.requestAnimationFrame(resolve)));
async function load(url){
 const frame=document.createElement('iframe');samples.append(frame);
 await new Promise(resolve=>{frame.onload=resolve;frame.src=url;});
 await frame.contentWindow.TokenAssets.ready;
 return frame;
}
run.onclick=async()=>{
 run.disabled=true;const checks=[];
 const assert=(ok,message)=>{if(!ok)throw Error(message);checks.push(message);};
 try{
  samples.replaceChildren();
  const frame=await load('/md3/components/menus?preview.compare=interactive');
  const w=frame.contentWindow,d=frame.contentDocument;
  const opener=d.querySelector('#component-preview [data-menu-open]'),menu=d.getElementById(opener.dataset.menuOpen),list=menu.querySelector('ul');
  for(let i=0;i<24;i++)list.append(list.querySelector('li').cloneNode(true));
  const style=d.createElement('style');style.textContent='.md-menu{--_menu-gap:8px}';d.head.append(style);
  opener.click();
  assert(menu.hasAttribute('data-menu-scrollable'),'Overflowing menu selects scrolling treatment');
  assert(w.getComputedStyle(menu).overflowY==='scroll','Menu requests a persistent scrollbar');
  for(const region of menu.querySelectorAll('section,ul'))assert(parseFloat(w.getComputedStyle(region).gap)===0,'Scrollable menu has no gaps');
  await settle(w);await Promise.all(menu.getAnimations().map(animation=>animation.finished));
  const before=menu.getBoundingClientRect();menu.scrollTop=120;await settle(w);
  assert(menu.scrollTop>0,'Menu can scroll');
  const after=menu.getBoundingClientRect();
  assert(after.top===before.top&&after.left===before.left&&after.height===before.height,'Scrolling menu contents preserves its placement and height');
  menu.hidePopover();list.replaceChildren(list.firstElementChild);opener.click();
  assert(!menu.hasAttribute('data-menu-scrollable'),'Short menu clears scrolling treatment');

  // Check the real table consumer, including both sides of its anchor.
  const table=await load('/components/buttons'),tw=table.contentWindow,td=table.contentDocument;
  const picker=td.querySelector('.tokens-token-set'),panel=td.querySelector('.tokens-token-set-menu');
  const gap=parseFloat(tw.getComputedStyle(panel).getPropertyValue('--md-sys-measurement-space100'));
  const bounds=()=>panel.getBoundingClientRect();
  for(const targetTop of [80,540]){
   tw.scrollTo(0,tw.scrollY+picker.getBoundingClientRect().top-targetTop);await settle(tw);
   picker.click();await settle(tw);await Promise.all(panel.getAnimations().map(animation=>animation.finished));
   const initial=bounds(),anchor=picker.getBoundingClientRect(),above=panel.dataset.menuAbove==='true';
   assert(Math.abs((above?anchor.top-initial.bottom:initial.top-anchor.bottom)-gap)<1,'Table menu uses the same gap on opening ('+(above?'above':'below')+')');
   assert(initial.top>=gap-1&&initial.bottom<=tw.innerHeight-gap+1,'Table menu stays within the viewport');
   panel.scrollTop=160;await settle(tw);
   assert(panel.scrollTop>0,'Table menu contents scroll');
   assert(bounds().top===initial.top&&bounds().height===initial.height,'Table menu scrolling does not move or resize it');
   const links=[...panel.querySelectorAll('[role^="menuitem"]')],active=td.activeElement;
   active.dispatchEvent(new tw.KeyboardEvent('keydown',{key:'ArrowDown',bubbles:true}));
   assert(td.activeElement===links[(links.indexOf(active)+1)%links.length],'One arrow key advances exactly one table item');
   const pageY=tw.scrollY;
   td.activeElement.dispatchEvent(new tw.KeyboardEvent('keydown',{key:'End',bubbles:true}));await settle(tw);
   assert(td.activeElement===links.at(-1)&&links.at(-1).getBoundingClientRect().bottom<=bounds().bottom+1,'Keyboard navigation scrolls its focused item into view');
   assert(tw.scrollY===pageY,'Keyboard menu navigation does not scroll the page');
   panel.scrollTop=panel.scrollHeight;await settle(tw);
   const scrollTop=panel.scrollTop;tw.scrollBy(0,40);await settle(tw);
   const moved=bounds(),movedAnchor=picker.getBoundingClientRect(),movedAbove=panel.dataset.menuAbove==='true';
   assert(Math.abs((movedAbove?movedAnchor.top-moved.bottom:moved.top-movedAnchor.bottom)-gap)<1,'Page scrolling keeps the table menu anchored with the same gap');
   assert(Math.abs(panel.scrollTop-Math.min(scrollTop,panel.scrollHeight-panel.clientHeight))<1,'Repositioning preserves the menu scroll position within its new bounds');
   panel.hidePopover();await settle(tw);
  }
  out.textContent=JSON.stringify({passed:checks.length,checks},null,2);
 }catch(error){out.textContent=JSON.stringify({passed:checks.length,checks,error:error.stack},null,2);}
 finally{run.disabled=false;}
};
