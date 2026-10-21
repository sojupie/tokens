(() => {
  // AndroidX distance/velocity defaults; the web renderer adds a minimized peek anchor.
  function sheetTarget(start, height, originHeight, peek, velocity, distance, speed) {
    if(velocity < -speed)return 'expanded';
    if(height < peek/2)return 'hidden';
    if(velocity > speed)return start==='expanded'?'minimized':'hidden';
    if(start==='expanded')return originHeight-height>=distance?'minimized':'expanded';
    return height-originHeight>=distance?'expanded':'minimized';
  }
  function sheetExit(bottom, start, rtl) { return bottom?'translateY(100%)':`translateX(${start!==rtl?-100:100}%)`; }
  if(typeof module!=='undefined')module.exports={sheetTarget,sheetExit};
  if(typeof document==='undefined')return;
  const states=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
  function demo(sheet){return sheet.closest('.md-sheet-demo');}
  function numeric(sheet,key){return Number.parseFloat(getComputedStyle(sheet).getPropertyValue('--_sheet-'+key));}
  function entry(sheet){return states.get(sheet);}
  function content(sheet,visible,interactive=visible){const body=sheet.querySelector('.md-sheet-content');body.hidden=!visible;body.inert=!interactive;}
  function updateHandle(sheet){
    const handle=sheet.querySelector('[data-sheet-handle]');if(!handle)return;
    const expanded=sheet.dataset.variant==='expanded';handle.setAttribute('aria-expanded',String(expanded));handle.setAttribute('aria-label',expanded?'Minimize sheet':'Expand sheet');
  }
  function exit(sheet){return sheetExit(sheet.classList.contains('md-bottom-sheet'),sheet.dataset.orientation==='start',getComputedStyle(sheet).direction==='rtl');}
  function animate(sheet,frames,closing=false){
    const state=entry(sheet);state.animation?.cancel();state.animation=null;
    if(reduced.matches)return Promise.resolve(true);
    const style=getComputedStyle(sheet),duration=numeric(sheet,'duration');
    const animation=sheet.animate(frames,{duration,easing:style.getPropertyValue(closing?'--_sheet-exit-easing':'--_sheet-enter-easing').trim(),fill:'both'});
    state.animation=animation;
    return animation.finished.then(()=>{if(state.animation!==animation)return false;state.animation=null;animation.cancel();return true;},()=>false);
  }
  function bounds(sheet,variant){
    const old=sheet.dataset.variant,height=sheet.style.height;
    sheet.style.height='';sheet.dataset.variant=variant;
    const value=sheet.getBoundingClientRect().height;
    sheet.dataset.variant=old;sheet.style.height=height;return value;
  }
  async function resize(sheet,variant){
    const state=entry(sheet);if(state.phase==='closing'||state.phase==='closed')return;
    const from=sheet.getBoundingClientRect().height;
    state.animation?.cancel();state.phase='resizing';
    content(sheet,true,variant==='expanded');sheet.dataset.variant=variant;updateHandle(sheet);
    const to=bounds(sheet,variant);sheet.style.height=to+'px';
    if(await animate(sheet,[{height:from+'px'},{height:to+'px'}])){sheet.style.height='';state.phase='open';content(sheet,variant==='expanded');}
  }
  async function open(sheet,opener){
    const state=entry(sheet);if(!state)return;
    if(['open','opening','resizing'].includes(state.phase)){sheet.querySelector(':is([data-sheet-initial-focus],[data-sheet-handle],.md-sheet-close):not(:disabled)')?.focus({preventScroll:true});return;}
    const interrupted=state.phase==='closing',style=getComputedStyle(sheet),from=interrupted?style.transform:exit(sheet),fromHeight=interrupted?sheet.getBoundingClientRect().height:0;
    state.animation?.cancel();state.phase='opening';state.opener=opener;state.saved=sheet.querySelector('input')?.value;
    const status=demo(sheet).querySelector('[data-sheet-status]');if(status)status.textContent='';
    sheet.inert=false;sheet.removeAttribute('data-sheet-closing');
    if(sheet instanceof HTMLDialogElement){if(!sheet.open){if(sheet.dataset.control==='standard')sheet.show();else sheet.showModal();}}else sheet.hidden=false;
    content(sheet,sheet.dataset.variant!=='minimized');updateHandle(sheet);
    sheet.style.height='';const to=sheet.getBoundingClientRect().height;sheet.style.height=to+'px';
    opener.setAttribute('aria-expanded','true');
    sheet.querySelector(':is([data-sheet-initial-focus],[data-sheet-handle],.md-sheet-close):not(:disabled)')?.focus({preventScroll:true});
    if(await animate(sheet,[{transform:from,height:(interrupted?fromHeight:to)+'px'},{transform:'none',height:to+'px'}])){sheet.style.height='';state.phase='open';}
  }
  async function close(sheet,save=false,restoreFocus=true){
    const state=entry(sheet);if(!state||['closed','closing'].includes(state.phase))return;
    const from=getComputedStyle(sheet).transform,height=sheet.getBoundingClientRect().height;
    const input=sheet.querySelector('input');if(input){if(save)state.saved=input.value;else input.value=state.saved;}
    sheet.style.height=height+'px';state.phase='closing';sheet.inert=true;sheet.setAttribute('data-sheet-closing','');
    if(!await animate(sheet,[{transform:from},{transform:exit(sheet)}],true))return;
    if(sheet instanceof HTMLDialogElement)sheet.close();else sheet.hidden=true;
    sheet.inert=false;sheet.style.height='';sheet.removeAttribute('data-sheet-closing');state.phase='closed';
    const opener=state.opener;opener?.setAttribute('aria-expanded','false');if(restoreFocus)opener?.focus({preventScroll:true});
    const status=demo(sheet).querySelector('[data-sheet-status]');if(status)status.textContent='Sheet closed.';
  }
  function initialize(){
    for(const sheet of document.querySelectorAll('.md-sheet-demo[data-interactive="true"] .md-sheet'))if(!states.has(sheet)){
      states.set(sheet,{phase:'closed',animation:null,drag:null,suppressClick:false});
      if(sheet instanceof HTMLDialogElement)sheet.addEventListener('cancel',event=>{event.preventDefault();close(sheet);});
    }
  }
  document.addEventListener('click',event=>{
    const button=event.target.closest?.('button'),container=button?.closest('.md-sheet-demo[data-interactive="true"]');if(!container)return;
    const sheet=container.querySelector('.md-sheet'),state=entry(sheet);if(!state)return;
    if(button.hasAttribute('data-sheet-open')){open(sheet,button);return;}
    if(button.hasAttribute('data-sheet-close')){close(sheet,button.hasAttribute('data-sheet-save'));return;}
    if(button.hasAttribute('data-sheet-handle')){
      if(state.suppressClick&&event.detail){state.suppressClick=false;event.preventDefault();return;}
      resize(sheet,sheet.dataset.variant==='minimized'?'expanded':'minimized');
    }
  });
  document.addEventListener('keydown',event=>{
    const sheet=event.target.closest?.('.md-sheet-demo[data-interactive="true"] .md-sheet');if(!sheet)return;
    if(event.key==='Escape'&&(!(sheet instanceof HTMLDialogElement)||sheet.dataset.control==='standard')){event.preventDefault();event.stopPropagation();close(sheet);return;}
    if(!event.target.matches('[data-sheet-handle]'))return;
    if(['ArrowUp','Home','ArrowDown','End'].includes(event.key)){event.preventDefault();resize(sheet,['ArrowUp','Home'].includes(event.key)?'expanded':'minimized');}
  });
  // Dismiss only when both ends of a pointer gesture are on the native backdrop.
  document.addEventListener('pointerdown',event=>{
    for(const openSheet of document.querySelectorAll('dialog.md-sheet[data-control="standard"][open]')){const state=entry(openSheet);if(state&&!openSheet.contains(event.target)&&!state.opener?.contains(event.target))close(openSheet,false,false);}
    const sheet=event.target.closest?.('.md-sheet-demo[data-interactive="true"] .md-sheet');if(!sheet)return;
    const state=entry(sheet);if(!state||event.button!==0||event.isPrimary===false||state.phase==='closing')return;
    if(sheet instanceof HTMLDialogElement&&event.target===sheet){const r=sheet.getBoundingClientRect();state.backdrop=event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom;}
    const handle=event.target.closest('[data-sheet-handle]');if(!handle)return;
    state.suppressClick=false;
    const rect=sheet.getBoundingClientRect(),bottom=sheet instanceof HTMLDialogElement?window.innerHeight:sheet.closest('.md-sheet-frame').getBoundingClientRect().bottom;
    const height=rect.height-Math.max(0,rect.bottom-bottom);
    state.animation?.cancel();state.animation=null;state.phase='open';sheet.style.height=height+'px';
    state.drag={id:event.pointerId,startY:event.clientY,height,origin:sheet.dataset.variant,moved:false,samples:[[event.timeStamp,event.clientY]],current:height,min:bounds(sheet,'minimized'),max:bounds(sheet,'expanded')};
    handle.setPointerCapture(event.pointerId);
  });
  document.addEventListener('pointermove',event=>{
    const sheet=event.target.closest?.('.md-sheet-demo[data-interactive="true"] .md-sheet'),state=sheet&&entry(sheet),drag=state?.drag;if(!drag||drag.id!==event.pointerId)return;
    const delta=event.clientY-drag.startY;
    if(!drag.moved&&Math.abs(delta)<Number.parseFloat(getComputedStyle(sheet).getPropertyValue('--md-sys-measurement-space100')))return;
    if(!drag.moved){drag.moved=true;sheet.dataset.variant='expanded';content(sheet,true,false);}
    state.suppressClick=true;event.preventDefault();
    drag.current=Math.max(0,Math.min(drag.max,drag.height-delta));sheet.style.height=drag.current+'px';
    drag.samples.push([event.timeStamp,event.clientY]);drag.samples=drag.samples.filter(([time])=>event.timeStamp-time<=100);
  });
  function release(event,cancelled){
    const sheet=event.target.closest?.('.md-sheet-demo[data-interactive="true"] .md-sheet'),state=sheet&&entry(sheet);if(!state)return;
    const drag=state.drag;
    if(drag&&drag.id===event.pointerId){
      state.drag=null;const handle=sheet.querySelector('[data-sheet-handle]');if(handle.hasPointerCapture(event.pointerId))handle.releasePointerCapture(event.pointerId);
      if(cancelled){resize(sheet,drag.origin);return;}
      if(!drag.moved)return;
      drag.samples.push([event.timeStamp,event.clientY]);const recent=drag.samples.filter(([time])=>event.timeStamp-time<=100),first=recent[0];
      const velocity=first&&event.timeStamp>first[0]?(event.clientY-first[1])/(event.timeStamp-first[0]):0;
      const target=sheetTarget(drag.origin,drag.current,drag.height,drag.min,velocity,numeric(sheet,'drag-distance'),numeric(sheet,'drag-velocity')/1000);
      if(target==='hidden')close(sheet);else resize(sheet,target);
    }else if(!cancelled&&state.backdrop&&event.target===sheet){
      const r=sheet.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)close(sheet);
    }
    state.backdrop=false;
  }
  document.addEventListener('pointerup',event=>release(event,false));
  document.addEventListener('pointercancel',event=>release(event,true));
  document.addEventListener('htmx:afterSwap',initialize);
  document.addEventListener('htmx:beforeCleanupElement',event=>{
    const root=event.detail.elt;
    for(const sheet of root.matches?.('.md-sheet')?[root]:root.querySelectorAll?.('.md-sheet')||[]){const state=entry(sheet);state?.animation?.cancel();if(sheet instanceof HTMLDialogElement&&sheet.open)sheet.close();states.delete(sheet);}
  });
  reduced.addEventListener('change',()=>{if(reduced.matches)for(const sheet of document.querySelectorAll('.md-sheet'))entry(sheet)?.animation?.finish();});
  initialize();
})();
