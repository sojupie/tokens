(() => {
  // Heights are ordered snap anchors in CSS pixels; velocity is px/ms.
  function sheetTarget(anchors, start, height, velocity, distance, speed) {
    const origin=anchors.find(anchor=>anchor.name===start)||anchors[0];
    if(Math.abs(velocity)>speed){
      const from=Math.abs(height-origin.height)<distance?origin.height:height;
      if(velocity<0)return anchors.find(anchor=>anchor.height>from+1)?.name||anchors.at(-1).name;
      return anchors.findLast(anchor=>anchor.height<from-1)?.name||'hidden';
    }
    if(height<anchors[0].height/2)return 'hidden';
    if(Math.abs(height-origin.height)<distance)return origin.name;
    return anchors.reduce((nearest,anchor)=>Math.abs(anchor.height-height)<Math.abs(nearest.height-height)?anchor:nearest).name;
  }
  function sheetExit(bottom, start, rtl) { return bottom?'translateY(100%)':`translateX(${start!==rtl?-100:100}%)`; }
  if(typeof module!=='undefined')module.exports={sheetTarget,sheetExit};
  if(typeof document==='undefined')return;
  const states=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
  function demo(sheet){return sheet.closest('.md-sheet-demo,.md-sheet-host');}
  function numeric(sheet,key){return Number.parseFloat(getComputedStyle(sheet).getPropertyValue('--_sheet-'+key));}
  function entry(sheet){return states.get(sheet);}
  function content(sheet,visible,interactive=visible){const body=sheet.querySelector('.md-sheet-content');body.hidden=!visible;body.inert=!interactive;}
  function size(sheet){return sheet.dataset.sheetSize||(sheet.dataset.variant==='minimized'?'minimized':'default');}
  function setSize(sheet,value){sheet.dataset.sheetSize=value;sheet.dataset.variant=value==='minimized'?'minimized':'expanded';}
  function sizes(sheet){
    const names=['minimized','default'];
    if(sheet.matches('dialog.md-bottom-sheet[data-control="modal"]')&&sheet.hasAttribute('data-sheet-allow-full'))names.push('full');
    return names.map(name=>({name,height:bounds(sheet,name)})).filter((anchor,index,all)=>!index||anchor.height>all[index-1].height+1);
  }
  function nextSize(sheet,anchors=sizes(sheet)){
    return anchors[(anchors.findIndex(anchor=>anchor.name===size(sheet))+1)%anchors.length].name;
  }
  function updateHandle(sheet){
    const handle=sheet.querySelector('[data-sheet-handle]');if(!handle)return;
    const next=nextSize(sheet),labels={minimized:'Minimize sheet',default:'Restore sheet',full:'Expand sheet'};
    handle.setAttribute('aria-expanded',String(size(sheet)!=='minimized'));handle.setAttribute('aria-label',labels[next]);
  }
  function freeze(sheet){
    const height=sheet.getBoundingClientRect().height,state=entry(sheet);
    state.animation?.cancel();state.animation=null;sheet.style.height=height+'px';return height;
  }
  function exit(sheet){return sheetExit(sheet.classList.contains('md-bottom-sheet'),sheet.dataset.orientation==='start',getComputedStyle(sheet).direction==='rtl');}
  function animate(sheet,frames,velocity=0){
    const state=entry(sheet);state.animation?.cancel();state.animation=null;
    if(reduced.matches)return Promise.resolve(true);
    const style=getComputedStyle(sheet);let timing=frames[0].transform!==undefined?MaterialMotion.attachedTiming(sheet):MaterialMotion.timing(sheet,'default-spatial');
    if(velocity && frames[0].height && frames[1].height) {
      const from=parseFloat(frames[0].height),to=parseFloat(frames[1].height);
      if(from!==to){
        const spring=MaterialMotion.springFrames(parseFloat(style.getPropertyValue('--md-sys-motion-spring-default-spatial-damping')),parseFloat(style.getPropertyValue('--md-sys-motion-spring-default-spatial-stiffness')),velocity/(to-from));
        frames=spring.samples.map((p,i)=>({height:Math.max(0,from+(to-from)*p)+'px',offset:i/(spring.samples.length-1)}));timing={duration:spring.duration,easing:'linear'};
      }
    }
    const animation=sheet.animate(frames,{...timing,fill:'both'});
    state.animation=animation;
    return animation.finished.then(()=>{if(state.animation!==animation)return false;state.animation=null;animation.cancel();return true;},()=>false);
  }
  function bounds(sheet,value){
    const variant=sheet.dataset.variant,previous=sheet.dataset.sheetSize,height=sheet.style.height;
    sheet.style.height='';setSize(sheet,value);
    const result=sheet.getBoundingClientRect().height;
    sheet.dataset.variant=variant;if(previous===undefined)delete sheet.dataset.sheetSize;else sheet.dataset.sheetSize=previous;
    sheet.style.height=height;return result;
  }
  async function resize(sheet,value,velocity=0){
    const state=entry(sheet);if(state.phase==='closing'||state.phase==='closed')return;
    const from=freeze(sheet);state.phase='resizing';
    content(sheet,true,value!=='minimized');setSize(sheet,value);updateHandle(sheet);
    const to=bounds(sheet,value);sheet.style.height=to+'px';
    if(await animate(sheet,[{height:from+'px'},{height:to+'px'}],velocity)){sheet.style.height='';state.phase='open';content(sheet,value!=='minimized');}
  }
  async function open(sheet,opener){
    const state=entry(sheet);if(!state)return;
    if(['open','opening','resizing'].includes(state.phase)){sheet.querySelector(':is([data-sheet-initial-focus],[data-sheet-handle],.md-sheet-close):not(:disabled)')?.focus({preventScroll:true});return;}
    const interrupted=state.phase==='closing',style=getComputedStyle(sheet),from=interrupted?style.transform:exit(sheet),fromHeight=interrupted?sheet.getBoundingClientRect().height:0;
    state.animation?.cancel();state.phase='opening';state.opener=opener;
    const status=demo(sheet).querySelector('[data-sheet-status]');if(status)status.textContent='';
    if(!interrupted)setSize(sheet,state.initialSize);
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
    if(!sheet.dispatchEvent(new CustomEvent('md-sheet-beforeclose',{bubbles:true,cancelable:true,detail:{reason:save?'confirm':'dismiss'}})))return;
    const from=getComputedStyle(sheet).transform,height=sheet.getBoundingClientRect().height;
    sheet.style.height=height+'px';state.phase='closing';sheet.inert=true;sheet.setAttribute('data-sheet-closing','');
    if(!await animate(sheet,[{transform:from},{transform:exit(sheet)}]))return;
    if(sheet instanceof HTMLDialogElement)sheet.close();else sheet.hidden=true;
    sheet.inert=false;sheet.style.height='';sheet.removeAttribute('data-sheet-closing');state.phase='closed';sheet.dispatchEvent(new CustomEvent('md-sheet-closed',{bubbles:true,detail:{reason:save?'confirm':'dismiss'}}));
    const opener=state.opener;opener?.setAttribute('aria-expanded','false');if(restoreFocus)opener?.focus({preventScroll:true});
    const status=demo(sheet).querySelector('[data-sheet-status]');if(status)status.textContent='Sheet closed.';
  }
  function initialize(){
    for(const sheet of document.querySelectorAll(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"] .md-sheet'))if(!states.has(sheet)){
      states.set(sheet,{phase:sheet instanceof HTMLDialogElement?(sheet.open?'open':'closed'):(sheet.hidden?'closed':'open'),animation:null,drag:null,suppressClick:false,initialSize:size(sheet)});
      if(sheet instanceof HTMLDialogElement)sheet.addEventListener('cancel',event=>{event.preventDefault();close(sheet);});
    }
  }
  document.addEventListener('click',event=>{
    const button=event.target.closest?.('button'),container=button?.closest(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"]');if(!container)return;
    const sheet=container.querySelector('.md-sheet'),state=entry(sheet);if(!state)return;
    if(button.hasAttribute('data-sheet-open')){open(sheet,button);return;}
    if(button.hasAttribute('data-sheet-close')){close(sheet,button.hasAttribute('data-sheet-save'));return;}
    if(button.hasAttribute('data-sheet-handle')){
      if(state.suppressClick){state.suppressClick=false;event.preventDefault();return;}
      freeze(sheet);resize(sheet,nextSize(sheet));
    }
  });
  document.addEventListener('keydown',event=>{
    const sheet=event.target.closest?.(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"] .md-sheet');if(!sheet)return;
    if(event.key==='Escape'&&(!(sheet instanceof HTMLDialogElement)||sheet.dataset.control==='standard')){event.preventDefault();event.stopPropagation();close(sheet);return;}
    if(!event.target.matches('[data-sheet-handle]'))return;
    // Keyboard activation is independent of any preceding captured drag.
    if(event.key==='Enter'||event.key===' ')entry(sheet).suppressClick=false;
    if(['ArrowUp','Home','ArrowDown','End'].includes(event.key)){
      event.preventDefault();freeze(sheet);const anchors=sizes(sheet),index=anchors.findIndex(anchor=>anchor.name===size(sheet));
      const target=event.key==='Home'?anchors.length-1:event.key==='End'?0:Math.max(0,Math.min(anchors.length-1,index+(event.key==='ArrowUp'?1:-1)));
      resize(sheet,anchors[target].name);
    }
  });
  // Dismiss only when both ends of a pointer gesture are on the native backdrop.
  document.addEventListener('pointerdown',event=>{
    for(const openSheet of document.querySelectorAll('dialog.md-sheet[data-control="standard"][open]')){const state=entry(openSheet);if(state&&!openSheet.contains(event.target)&&!state.opener?.contains(event.target))close(openSheet,false,false);}
    const sheet=event.target.closest?.(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"] .md-sheet');if(!sheet)return;
    const state=entry(sheet);if(!state||event.button!==0||event.isPrimary===false||state.phase==='closing')return;
    if(sheet instanceof HTMLDialogElement&&event.target===sheet){const r=sheet.getBoundingClientRect();state.backdrop=event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom;}
    const handle=event.target.closest('[data-sheet-handle]');if(!handle)return;
    state.suppressClick=false;
    const rect=sheet.getBoundingClientRect(),bottom=sheet instanceof HTMLDialogElement?window.innerHeight:(sheet.closest('.md-sheet-frame')||demo(sheet)).getBoundingClientRect().bottom;
    const height=rect.height-Math.max(0,rect.bottom-bottom);
    freeze(sheet);state.phase='open';sheet.style.height=height+'px';
    state.drag={id:event.pointerId,startY:event.clientY,height,origin:size(sheet),moved:false,samples:[[event.timeStamp,event.clientY]],current:height,anchors:sizes(sheet)};
    handle.setPointerCapture(event.pointerId);
  });
  document.addEventListener('pointermove',event=>{
    const sheet=event.target.closest?.(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"] .md-sheet'),state=sheet&&entry(sheet),drag=state?.drag;if(!drag||drag.id!==event.pointerId)return;
    const delta=event.clientY-drag.startY;
    if(!drag.moved&&Math.abs(delta)<Number.parseFloat(getComputedStyle(sheet).getPropertyValue('--md-sys-measurement-space100')))return;
    if(!drag.moved){drag.moved=true;sheet.dataset.variant='expanded';content(sheet,true,false);}
    state.suppressClick=true;event.preventDefault();
    drag.current=Math.max(0,Math.min(drag.anchors.at(-1).height,drag.height-delta));sheet.style.height=drag.current+'px';
    drag.samples.push([event.timeStamp,event.clientY]);drag.samples=drag.samples.filter(([time])=>event.timeStamp-time<=100);
  });
  function release(event,cancelled){
    const sheet=event.target.closest?.(':is(.md-sheet-demo,.md-sheet-host)[data-interactive="true"] .md-sheet'),state=sheet&&entry(sheet);if(!state)return;
    const drag=state.drag;
    if(drag&&drag.id===event.pointerId){
      state.drag=null;const handle=sheet.querySelector('[data-sheet-handle]');if(handle.hasPointerCapture(event.pointerId))handle.releasePointerCapture(event.pointerId);
      if(cancelled){resize(sheet,drag.origin);return;}
      if(!drag.moved)return;
      drag.samples.push([event.timeStamp,event.clientY]);const recent=drag.samples.filter(([time])=>event.timeStamp-time<=100),first=recent[0];
      const velocity=first&&event.timeStamp>first[0]?(event.clientY-first[1])/(event.timeStamp-first[0]):0;
      const target=sheetTarget(drag.anchors,drag.origin,drag.current,velocity,numeric(sheet,'drag-distance'),numeric(sheet,'drag-velocity')/1000);
      if(target==='hidden')close(sheet);else resize(sheet,target,-velocity*1000);
    }else if(!cancelled&&state.backdrop&&event.target===sheet){
      const r=sheet.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)close(sheet);
    }
    state.backdrop=false;
  }
  document.addEventListener('pointerup',event=>release(event,false));
  document.addEventListener('pointercancel',event=>release(event,true));
  document.addEventListener('htmx:afterSwap',initialize);
  document.addEventListener('htmx:beforeCleanupElement',event=>{
    const root=(event.detail?.elt || event.target);
    for(const sheet of root.matches?.('.md-sheet')?[root]:root.querySelectorAll?.('.md-sheet')||[]){const state=entry(sheet);state?.animation?.cancel();if(sheet instanceof HTMLDialogElement&&sheet.open)sheet.close();states.delete(sheet);}
  });
  reduced.addEventListener('change',()=>{if(reduced.matches)for(const sheet of document.querySelectorAll('.md-sheet'))entry(sheet)?.animation?.finish();});
  initialize();
  document.addEventListener('tokens:assets-ready',initialize);
})();
