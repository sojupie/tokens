// The shared shell uses one native modal on compact windows; only its current
// navigation level scrolls. Wide windows dock the primary rail beside that list.
{
 const header=document.getElementById('tokens-site-header');
 if(header){
  const drawer=header.querySelector('#tokens-site-menu');
  const opener=header.querySelector('[data-navigation-open]');
  const appearance=header.querySelector('.tokens-drawer-appearance');
  const wide=matchMedia(`(min-width:${header.dataset.drawerBreakpoint}px)`);
  const rail=matchMedia('(min-width:600px)');
  const viewport=window.visualViewport;
  const scrollPositions=new Map();
  let viewportFrame=0,navigation,navigationState,pointerType,latched=false,pendingSection;
  function rememberScroll(){
   for(const list of header.querySelectorAll('[data-site-scroll]')) if(list.getClientRects().length)scrollPositions.set(list.dataset.siteScroll,list.scrollTop);
  }
  function showSection(section,focus=false){
   const previousSection=header.dataset.section;
   header.dataset.section=section;
   header.dataset.mobileLevel=section?'section':'main';
   for(const panel of header.querySelectorAll('[data-site-panel]')){
    const hidden=panel.dataset.sitePanel!==section;
    panel.toggleAttribute('data-site-enter',panel.hidden&&!hidden);
    panel.hidden=hidden;panel.inert=hidden;
   }
   for(const button of header.querySelectorAll('[data-site-section]'))button.setAttribute('aria-expanded',String(button.dataset.siteSection===section));
   if(focus){
    const target=section?header.querySelector('[data-site-panel]:not([hidden]) [data-site-back]'):header.querySelector(`[data-site-section="${previousSection}"]`);
    (target??header.querySelector('.tokens-site-main a'))?.focus({preventScroll:true});
   }
  }
  function refreshNavigation(){
   const next=header.querySelector('[data-site-navigation]');
   if(!next)return;
   const state=next.dataset.activeSection+'\n'+next.dataset.activePage;
   if(navigation===next&&navigationState===state)return;
   navigation=next;navigationState=state;
   const section=next.dataset.activeSection;
   document.documentElement.dataset.siteSection=section;
   header.dataset.drawerPersistent=String(wide.matches&&!!section);
   const shown=header.querySelector(`[data-site-panel="${section}"]`)?section:rail.matches?header.dataset.section:'';
   showSection(header.querySelector(`[data-site-panel="${shown}"]`)?shown:'');
   latched=rail.matches&&!wide.matches&&pendingSection===section;
   pendingSection=undefined;
   setPanelOpen(latched);
   for(const list of header.querySelectorAll('[data-site-scroll]')){
    if(scrollPositions.has(list.dataset.siteScroll))list.scrollTop=scrollPositions.get(list.dataset.siteScroll);
    else if(list.getClientRects().length)list.querySelector('[aria-current="page"]')?.scrollIntoView({block:'nearest'});
   }
  }
  header.addEventListener('click',event=>{
   const link=event.target.closest('[data-site-rail-section]');
   if(link&&rail.matches&&!event.ctrlKey&&!event.metaKey&&!event.shiftKey&&!event.altKey&&(!event.button||event.button===0)){
    const section=link.dataset.siteRailSection;
    const hasPanel=!!header.querySelector(`[data-site-panel="${section}"]`);
    const current=new URL(link.href,location.href).pathname===location.pathname;
    if(current){
     event.preventDefault();event.stopPropagation();
     if(!wide.matches&&hasPanel){
      latched=!latched;rememberScroll();showSection(section);setPanelOpen(latched);
     }
    }else if(!wide.matches&&hasPanel&&(pointerType==='touch'||pointerType==='pen'))pendingSection=section;
    return;
   }
   const section=event.target.closest('button[data-site-section]'),back=event.target.closest('[data-site-back]');
   if(!section&&!back)return;
   event.preventDefault();event.stopPropagation();rememberScroll();
   showSection(section?.dataset.siteSection??'',true);
  },true);
  function setPanelOpen(open){
   header.dataset.panelOpen=String(open);
   drawer.inert=rail.matches&&!open&&header.dataset.drawerPersistent!=='true';
  }
  function closePanel(){
   latched=false;
   setPanelOpen(false);
  }
  function revealSection(event){
   if(!rail.matches||event.pointerType==='touch'||event.pointerType==='pen'||(event.type==='focusin'&&(pointerType==='touch'||pointerType==='pen')))return;
   const link=event.target.closest('[data-site-rail-section]');
   if(!link)return;
   if(!header.querySelector(`[data-site-panel="${link.dataset.siteRailSection}"]`))return;
   rememberScroll();showSection(link.dataset.siteRailSection);setPanelOpen(true);
  }
  document.addEventListener('pointerdown',event=>{
   pointerType=event.pointerType;
   if(latched&&!header.contains(event.target))closePanel();
  },true);
  header.addEventListener('pointerover',revealSection);
  header.addEventListener('focusin',revealSection);
  header.addEventListener('pointerleave',()=>{if(!latched)closePanel();});
  header.addEventListener('focusout',event=>{if(!latched&&!header.contains(event.relatedTarget))closePanel();});
  header.addEventListener('keydown',event=>{pointerType=undefined;if(event.key==='Escape'&&rail.matches){closePanel();event.preventDefault();}});
  header.addEventListener('scroll',event=>{
   if(event.target.matches('[data-site-scroll]'))scrollPositions.set(event.target.dataset.siteScroll,event.target.scrollTop);
  },true);
  document.addEventListener('htmx:beforeSwap',()=>{
   rememberScroll();
   // Preserved headers temporarily enter htmx's hidden pantry. Reinsertion is
   // not a section entrance and must not restart its opacity transition.
   for(const panel of header.querySelectorAll('[data-site-panel]'))panel.removeAttribute('data-site-enter');
  });
  document.addEventListener('htmx:afterSwap',refreshNavigation);
  function syncViewport(){
   viewportFrame=0;
   // Follow browser-chrome height, not elastic-scroll offsets or pinch zoom.
   if(rail.matches&&viewport?.scale===1)header.style.setProperty('--site-drawer-height',`${viewport.height}px`);
   else header.style.removeProperty('--site-drawer-height');
  }
  function scheduleViewport(){if(!viewportFrame)viewportFrame=requestAnimationFrame(syncViewport);}
  viewport?.addEventListener('resize',scheduleViewport);
  viewport?.addEventListener('scroll',scheduleViewport);
  function sync(){
   header.dataset.docked=String(rail.matches);
   header.dataset.siteLayout=wide.matches?'wide':rail.matches?'medium':'mobile';
   document.documentElement.dataset.siteLayout=header.dataset.siteLayout;
   document.documentElement.dataset.siteDrawer=rail.matches?'expanded':'modal';
   opener.setAttribute('aria-expanded',String(drawer.open));
   opener.setAttribute('aria-haspopup',rail.matches?'false':'dialog');
   document.documentElement.style.setProperty('--site-header-height',rail.matches?'0px':`${header.getBoundingClientRect().height}px`);
  }
  function arrange(){
   rememberScroll();syncViewport();
   header.dataset.panelOpen='false';
   if(appearance)(rail.matches?header:drawer).append(appearance);
   const variant=rail.matches?'standard':'modal';
   // The inline layout already opens the docked drawer. Preserve that surface
   // and its focus; close only when changing native modal/docked modes.
   if(drawer.open&&(drawer.dataset.variant!==variant||!rail.matches))drawer.close();
   drawer.dataset.variant=variant;
   if(rail.matches){
    drawer.setAttribute('role','navigation');
    if(!drawer.open){
     const previous=document.activeElement;drawer.show();
     if(previous===document.body)document.activeElement?.blur();
     else if(previous?.getClientRects().length)previous.focus({preventScroll:true});
    }
   } else drawer.removeAttribute('role');
   sync();navigation=undefined;refreshNavigation();
  }
  drawer.addEventListener('close',sync);
  wide.addEventListener('change',arrange);
  rail.addEventListener('change',arrange);
  new ResizeObserver(sync).observe(header);
  arrange();
 }
}
