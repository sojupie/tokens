(() => {
  // Snackbar timing is renderer behavior; Material's captured tokens contain no timeout.
  function snackbarTimer(close, clock) {
    let timer, remaining = 0, started = 0, active = false;
    const pauses = new Set();
    const clear = () => { clock.clearTimeout(timer); timer = undefined; };
    const schedule = () => {
      if (!active || pauses.size || !remaining) return;
      started = clock.now();
      timer = clock.setTimeout(() => { active = false; close(); }, remaining);
    };
    return {
      start(duration) { clear(); pauses.clear(); remaining = duration; active = duration > 0; schedule(); },
      stop() { clear(); active = false; pauses.clear(); },
      pause(reason) { if (!pauses.size && active) { remaining = Math.max(0, remaining - (clock.now() - started)); clear(); } pauses.add(reason); },
      resume(reason) { if (!pauses.delete(reason)) return; if (active && !pauses.size) { if (remaining <= 0) { active = false; close(); } else schedule(); } },
    };
  }
  function navigationIndex(index, length, key, horizontal, rtl=false) {
    if (key === 'Home') return 0;
    if (key === 'End') return length-1;
    const forward = horizontal ? (rtl ? 'ArrowLeft' : 'ArrowRight') : 'ArrowDown';
    const backward = horizontal ? (rtl ? 'ArrowRight' : 'ArrowLeft') : 'ArrowUp';
    if (key !== forward && key !== backward) return null;
    return (index + (key===forward ? 1 : -1) + length) % length;
  }
  function toolbarGap(available,widths,min,max) {
    if(widths.length<2)return 0;
    return Math.max(min,Math.min(max,(available-widths.reduce((sum,width)=>sum+width,0))/(widths.length-1)));
  }
  function sliderValues(values,index,min,max) {
    const result=values.map(value=>Math.max(min,Math.min(max,value)));
    if(result.length===2) { if(index===0) result[0]=Math.min(result[0],result[1]);else result[1]=Math.max(result[0],result[1]); }
    return result;
  }
  function splitMenuPosition(trigger,group,menu,viewport,rtl,gap,margin) {
    const fits=x=>x>=margin&&x+menu.width<=viewport.width-margin;
    // Prefer the trailing edge, then the other trailing edge and group edges.
    const edges=rtl?[trigger.left,trigger.right-menu.width,group.left,group.right-menu.width]:[trigger.right-menu.width,trigger.left,group.right-menu.width,group.left];
    const left=edges.find(fits)??Math.max(margin,Math.min(edges[0],viewport.width-menu.width-margin));
    const below=trigger.bottom+gap,above=trigger.top-menu.height-gap;
    const top=below+menu.height<=viewport.height-margin?below:above>=margin?above:Math.max(margin,Math.min(below,viewport.height-menu.height-margin));
    return {left,top};
  }
  if (typeof module !== "undefined") module.exports = { snackbarTimer, navigationIndex, sliderValues, toolbarGap, splitMenuPosition };
  if (typeof document === "undefined") return;
  document.addEventListener('click',event=>{
    const select=event.target.closest?.('[data-chip-select]'),remove=event.target.closest?.('[data-chip-remove]'),action=event.target.closest?.('[data-chip-action]');
    const chip=(select||remove||action)?.closest('[data-family="chip"]');if(!chip||chip.getAttribute('aria-disabled')==='true')return;
    const demo=chip.closest('.md-chip-demo'),output=demo.querySelector('[data-chip-output]');
    if(select){const selected=select.getAttribute('aria-pressed')!=='true';select.setAttribute('aria-pressed',String(selected));chip.dataset.selected=String(selected);output.textContent=selected?'Selected':'Unselected';}
    if(action)output.textContent=action.textContent.trim()+' activated';
    if(remove){const figure=demo.closest('figure'),fallback=figure?.closest('#component-preview')?.querySelector('select,button');chip.hidden=true;output.textContent='Removed '+chip.querySelector('.md-chip-label').textContent;fallback?.focus({preventScroll:true});}
  });
  document.addEventListener('keydown',event=>{
    const chip=event.target.closest?.('[data-family="chip"]');if(!chip||chip.getAttribute('aria-disabled')==='true')return;
    const primary=chip.querySelector('[data-chip-select]'),remove=chip.querySelector('[data-chip-remove]');if(!primary||!remove)return;
    if(event.key==='Delete'||event.key==='Backspace'){event.preventDefault();remove.click();return;}
    const rtl=getComputedStyle(chip).direction==='rtl',forward=event.key===(rtl?'ArrowLeft':'ArrowRight'),backward=event.key===(rtl?'ArrowRight':'ArrowLeft');
    if(forward&&event.target===primary||backward&&event.target===remove){event.preventDefault();(forward?remove:primary).focus({preventScroll:true});}
  });
  document.addEventListener('focusin',event=>{const remove=event.target.closest?.('[data-chip-remove]');if(remove)remove.parentElement.querySelector('[data-chip-select]').tabIndex=-1;});
  document.addEventListener('focusout',event=>{const remove=event.target.closest?.('[data-chip-remove]');if(remove)remove.parentElement.querySelector('[data-chip-select]').tabIndex=0;});
  const tabAnimations=new WeakMap();
  function selectTab(tab) {
    const list=tab.closest('[role="tablist"]'), example=list.closest('.md-tabs-example');
    const previous=list.querySelector('[aria-selected="true"]');
    if(previous===tab) return;
    const from=previous?.querySelector('.md-tab-active-indicator').getBoundingClientRect();
    tabAnimations.get(list)?.cancel();
    for(const item of list.querySelectorAll('[role="tab"]')) { const selected=item===tab;item.setAttribute('aria-selected',String(selected));item.tabIndex=selected?0:-1; }
    for(const panel of example.querySelectorAll('[role="tabpanel"]')) panel.hidden=panel.id!==tab.getAttribute('aria-controls');
    if(!from || matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const indicator=tab.querySelector('.md-tab-active-indicator'), to=indicator.getBoundingClientRect();
    if(!to.width) return;
    const animation=indicator.animate([{transform:`translateX(${from.left-to.left}px) scaleX(${from.width/to.width})`},{transform:'none'}],MaterialMotion.timing(tab));
    tabAnimations.set(list,animation);
  }
  document.addEventListener('click',event=>{
    const proxy=event.target.closest?.('[data-toolbar-proxy]');
    if(proxy&&!proxy.disabled)proxy.closest('.md-toolbar-demo').querySelector('[data-toolbar-item="'+proxy.dataset.toolbarProxy+'"]').click();
    const tool=event.target.closest?.('[data-toolbar-item="0"]');
    if(tool)queueMicrotask(()=>{const peer=tool.closest('.md-toolbar-demo').querySelector('[data-toolbar-proxy="0"]');if(!peer)return;const selected=tool.getAttribute('aria-pressed')==='true';peer.setAttribute('aria-checked',String(selected));peer.querySelector('[data-menu-checked-icon]').hidden=!selected;peer.querySelector('[data-menu-unchecked-icon]').hidden=selected;});
  });
  document.addEventListener('focusin',event=>{
    const tool=event.target.closest?.('[data-toolbar-item]');
    if(tool)for(const item of tool.closest('.md-toolbar').querySelectorAll('[data-toolbar-item]'))item.tabIndex=item===tool?0:-1;
    const tab=event.target.closest?.('.md-tab');if(tab) selectTab(tab);
  });
  document.addEventListener('keydown',event=>{
    const tool=event.target.closest?.('[data-toolbar-item]');
    if(tool) {
      const group=tool.closest('.md-toolbar'),items=[...group.querySelectorAll('[data-toolbar-item]:not(:disabled):not([hidden])')];
      const next=navigationIndex(items.indexOf(tool),items.length,event.key,group.dataset.orientation!=='vertical',getComputedStyle(group).direction==='rtl');
      if(next!==null){event.preventDefault();items[next]?.focus();}return;
    }
    const tab=event.target.closest?.('.md-tab');if(!tab)return;
    const list=tab.closest('[role="tablist"]'),items=[...list.querySelectorAll('[role="tab"]')];
    const next=navigationIndex(items.indexOf(tab),items.length,event.key,true,getComputedStyle(list).direction==='rtl');
    if(next===null)return;
    event.preventDefault();items[next].focus();selectTab(items[next]);
  });
  document.addEventListener('click',event=>{
    const button=event.target.closest?.('[data-group-button]');if(!button||button.disabled)return;
    const group=button.closest('.md-button-group');
    if(group.dataset.control==='connected'&&group.dataset.shape==='single')for(const item of group.querySelectorAll('button'))item.setAttribute('aria-pressed',String(item===button));
    const selected=[...group.querySelectorAll('button[aria-pressed="true"]')].map(item=>item.textContent.trim());
    group.closest('.md-button-group-demo').querySelector('output').textContent=group.dataset.control==='standard'?button.textContent.trim()+' activated.':selected.length?selected.join(', ')+' selected.':'No selection.';
  });
  document.addEventListener('keydown',event=>{
    const button=event.target.closest?.('[data-group-button]');if(!button)return;
    const group=button.closest('.md-button-group'),items=[...group.querySelectorAll('button:not(:disabled)')];
    const next=navigationIndex(items.indexOf(button),items.length,event.key,true,getComputedStyle(group).direction==='rtl');
    if(next!==null){event.preventDefault();items[next]?.focus();}
  });
  const timers = new WeakMap();
  function selectNavigation(button) {
    const group=button.closest('[data-navigation-group]');
    if (!group) return;
    for (const item of group.querySelectorAll('[data-navigation-item]')) item.removeAttribute('aria-current');
    button.setAttribute('aria-current','page');
    const output=group.querySelector('output');
    const label=button.querySelector('.md-nav-label').textContent;
    if(output) output.textContent=label+' selected';
    const frame=group.closest('.md-rail-frame');
    if(frame){const title=frame.querySelector('[data-navigation-title]');title.textContent=label;frame.querySelector('[data-navigation-description]').textContent='Your '+label.toLowerCase()+' in this workspace.';const screen=frame.querySelector('.md-rail-screen');screen.scrollTop=0;if(!matchMedia('(prefers-reduced-motion: reduce)').matches)screen.animate([{opacity:0},{opacity:1}],MaterialMotion.timing(group,'fast-effects'));if(frame.dataset.modalOpen==='true')setRailExpanded(frame,false);}
    group.closest('dialog[data-navigation-modal]')?.close();
  }
  document.addEventListener('keydown', event => {
    const button=event.target.closest?.('[data-navigation-item]');
    if (!button) return;
    const group=button.closest('[data-navigation-group]'), items=[...group.querySelectorAll('[data-navigation-item]')];
    const next=navigationIndex(items.indexOf(button),items.length,event.key,group.classList.contains('md-navigation-bar'),getComputedStyle(group).direction==='rtl');
    if (next===null) return;
    event.preventDefault(); items[next].focus();
  });
  const railFrames=new WeakMap(),railLabels=new WeakSet();
  function measureRailLabels(labels){
    const sizes=Array.from(labels,label=>({item:label.closest('.md-nav-item'),height:label.getBoundingClientRect().height}));
    for(const {item,height} of sizes)if(item&&height)item.style.setProperty('--_nav-label-height',height+'px');
  }
  const railLabelSizes=new ResizeObserver(entries=>measureRailLabels(entries.map(entry=>entry.target)));
  function syncRail(frame){
    const rail=frame.querySelector('.md-navigation-rail'),expanded=rail.dataset.variant==='expanded',modal=expanded&&rail.dataset.layout==='modal',hidden=!expanded&&frame.dataset.collapse==='hidden';
    frame.dataset.modalOpen=String(modal);frame.dataset.railHidden=String(hidden);frame.querySelector('.md-navigation-surface').hidden=hidden;
    const surface=frame.querySelector('.md-navigation-surface');if(modal){surface.setAttribute('role','dialog');surface.setAttribute('aria-label','Expanded navigation');}else{surface.removeAttribute('role');surface.removeAttribute('aria-label');}
    frame.querySelector('.md-navigation-scrim').hidden=!modal;frame.querySelector('.md-rail-body').inert=modal;
    for(const menu of frame.querySelectorAll('[data-navigation-expand]')){menu.setAttribute('aria-expanded',String(expanded));menu.setAttribute('aria-label',expanded?'Collapse navigation':'Expand navigation');}
    const external=frame.querySelector('.md-rail-external-menu');external.hidden=!hidden&&!['fab','none'].includes(frame.dataset.header);
    const fab=rail.querySelector('.md-nav-fab');if(fab){fab.classList.toggle('md-fab',!expanded);fab.classList.toggle('md-extended-fab',expanded);}
  }
  function setRailExpanded(frame,expanded,focus=false){
    const rail=frame.querySelector('.md-navigation-rail'),state=railFrames.get(frame),start=rail.getBoundingClientRect().width,wasModal=frame.dataset.modalOpen==='true';
    const slot=frame.querySelector('.md-navigation-slot'),body=frame.querySelector('.md-rail-body'),bodyBefore=body.getBoundingClientRect();
    const textTargets='.md-nav-label,.md-nav-section,.md-nav-fab > span:not(.md-button-ripple)';
    const outgoing=[...rail.querySelectorAll(textTargets)].filter(e=>e.getClientRects().length).map(element=>{const style=getComputedStyle(element);return {text:element.textContent,rect:element.getBoundingClientRect(),font:style.font,color:style.color,spacing:style.letterSpacing,align:style.textAlign,opacity:style.opacity,direction:style.direction,section:element.matches('.md-nav-section')};});
    const motionTargets='.md-nav-indicator,.md-nav-icon,.md-nav-label,.md-rail-menu,.md-nav-fab,.md-nav-fab > span:not(.md-button-ripple)';
    const before=new Map([...rail.querySelectorAll(motionTargets)].filter(e=>e.getClientRects().length).map(e=>[e,e.getBoundingClientRect()]));
    for(const a of state?.animations||[])a.cancel();for(const copy of state?.copies||[])copy.remove();delete frame.dataset.transitioning;
    const fab=rail.querySelector('.md-nav-fab'),fabStart=before.get(fab)?.width;
    rail.dataset.variant=expanded?'expanded':'collapsed';rail.dataset.orientation=expanded?'horizontal':'vertical';syncRail(frame);
    // Resolve the new label metrics before capturing the FLIP destination.
    // ResizeObserver otherwise updates the CSS base transform mid-animation.
    measureRailLabels(rail.querySelectorAll('.md-nav-label'));
    const slotEnd=slot.getBoundingClientRect().width;
    const targetHidden=frame.dataset.railHidden==='true',end=targetHidden?0:rail.getBoundingClientRect().width,timing=MaterialMotion.timing(rail,'default-spatial'),effects=MaterialMotion.timing(rail,'default-effects');
    const restoreFocus=()=>{if(focus){const target=targetHidden?frame.querySelector('.md-rail-external-menu'):rail.querySelector('[data-navigation-expand],.md-nav-fab,[data-navigation-item]');target?.focus({preventScroll:true});}};
    if(matchMedia('(prefers-reduced-motion: reduce)').matches){restoreFocus();return;}
    if(targetHidden){frame.querySelector('.md-navigation-surface').hidden=false;frame.dataset.railHidden='false';}
    if(wasModal&&!expanded){frame.dataset.modalOpen='true';frame.querySelector('.md-rail-body').inert=true;frame.querySelector('.md-navigation-scrim').hidden=false;}
    frame.style.setProperty('--_rail-content-width',(targetHidden?start:end)+'px');
    frame.style.setProperty('--_rail-content-align',getComputedStyle(rail).direction==='rtl'?'flex-end':'flex-start');
    frame.style.setProperty('--_rail-slot-width',slotEnd+'px');frame.dataset.transitioning='';
    const current={animations:[],copies:[]};railFrames.set(frame,current);
    // Labels crossfade between layouts. Retain only the outgoing text, at its
    // current position and typography, while the real label takes its new slot.
    // Label phase follows the spatial transition, adapted from AndroidX
    // NavigationItem.kt, Copyright 2024 The Android Open Source Project,
    // Apache-2.0. See THIRD_PARTY_NOTICES.md.
    const style=getComputedStyle(rail),spring=MaterialMotion.springFrames(Number(style.getPropertyValue('--md-sys-motion-spring-default-spatial-damping')),Number(style.getPropertyValue('--md-sys-motion-spring-default-spatial-stiffness')));
    const phases=spring.samples.map((progress,i)=>({progress,offset:i/(spring.samples.length-1)})),textTiming={duration:timing.duration,easing:'linear'},frameRect=frame.getBoundingClientRect(),sign=style.direction==='rtl'?-1:1;
    const middle=phases.findIndex(({progress})=>progress>=.5),a=phases[middle-1],b=phases[middle];
    phases.splice(middle,0,{progress:.5,offset:a.offset+(.5-a.progress)/(b.progress-a.progress)*(b.offset-a.offset)});
    for(const old of outgoing){
      const copy=document.createElement('span');copy.textContent=old.text;copy.dataset.railTextExit='';copy.setAttribute('aria-hidden','true');copy.inert=true;if(old.section)copy.dataset.railSectionExit='';
      Object.assign(copy.style,{position:'absolute',pointerEvents:'none',zIndex:'4',left:(old.rect.x-frameRect.x-frame.clientLeft)+'px',top:(old.rect.y-frameRect.y-frame.clientTop)+'px',width:old.rect.width+'px',height:old.rect.height+'px',font:old.font,color:old.color,letterSpacing:old.spacing,textAlign:old.align,direction:old.direction});
      frame.append(copy);current.copies.push(copy);current.animations.push(copy.animate(old.section?[{opacity:old.opacity},{opacity:0}]:phases.map(({progress,offset})=>({offset,opacity:Number(old.opacity)*Math.max(0,1-2*Math.max(0,progress))**2})),{...(old.section?effects:textTiming),fill:'forwards'}));
    }
    // Fix the final inner layout while the outer rail resizes. Move existing
    // contents from their current positions instead of reflowing every frame.
    // Capture every destination before starting any animation. An animated
    // parent must not contaminate the measured destination of its children.
    const destinations=new Map([...rail.querySelectorAll(motionTargets)].map(element=>{
      const style=getComputedStyle(element);
      return [element,{rect:element.getBoundingClientRect(),left:parseFloat(style.left)||0,top:parseFloat(style.top)||0}];
    }));
    const heading=rail.querySelector('.md-nav-section');
    if(heading?.getClientRects().length)current.animations.push(heading.animate([{opacity:0},{opacity:1}],effects));
    const shifts=new Map();
    for(const [element,{rect:next,left,top}] of destinations){
      const prior=before.get(element);if(!next.width||!next.height)continue;
      const parent=element.parentElement.closest(motionTargets),parentShift=shifts.get(parent)||{x:0,y:0};
      const shift=prior?{x:prior.x-next.x,y:prior.y-next.y}:{x:0,y:0};shifts.set(element,shift);
      if(element.matches(textTargets)){
        // Exit text stays fixed. On entry, the horizontal label approaches
        // from the icon while nearly transparent; its vertical slot is fixed.
        const parentRect=destinations.get(parent)?.rect,approach=expanded&&!element.matches('.md-nav-section')&&parentRect?(sign===1?next.left-parentRect.left:parentRect.right-next.right):0;
        current.animations.push(element.animate(phases.map(({progress:p,offset})=>({offset,left:(left-parentShift.x*(1-p)+sign*approach*(p-1))+'px',top:(top-parentShift.y*(1-p))+'px',opacity:Math.max(0,Math.min(1,2*p-1))**2})),textTiming));
        continue;
      }
      // Paint text at its moving position rather than promoting it into a
      // transformed texture that changes rasterization when the effect ends.
      current.animations.push(element.animate(prior?[{left:(left+shift.x-parentShift.x)+'px',top:(top+shift.y-parentShift.y)+'px'},{left:left+'px',top:top+'px'}]:[{opacity:0},{opacity:1}],prior?timing:effects));
    }
    const bodyEnd=body.getBoundingClientRect(),bodyShift=bodyBefore.x-bodyEnd.x;
    if(Math.abs(bodyShift)>.01)current.animations.push(body.animate([{left:bodyShift+'px'},{left:'0px'}],timing));
    const animation=rail.animate([{width:start+'px',minWidth:start+'px'},{width:end+'px',minWidth:end+'px'}],{...timing,fill:'both'});current.animations.push(animation);
    if(fab)current.animations.push(fab.animate([{width:fabStart+'px',minWidth:fabStart+'px'},{width:fab.getBoundingClientRect().width+'px',minWidth:fab.getBoundingClientRect().width+'px'}],timing));
    if(wasModal&&!expanded)current.animations.push(frame.querySelector('.md-navigation-scrim').animate([{opacity:1},{opacity:0}],{...effects,fill:'forwards'}));
    if(expanded)restoreFocus();
    animation.finished.then(()=>{if(railFrames.get(frame)!==current)return;delete frame.dataset.transitioning;syncRail(frame);for(const a of current.animations)a.cancel();for(const copy of current.copies)copy.remove();if(!expanded)restoreFocus();},()=>{});
  }
  function initRailFrames(){for(const frame of document.querySelectorAll('.md-rail-frame')){if(!railFrames.has(frame)){railFrames.set(frame,{});syncRail(frame);}for(const label of frame.querySelectorAll('.md-nav-label'))if(!railLabels.has(label)){railLabels.add(label);railLabelSizes.observe(label);}}}
  document.addEventListener('keydown',event=>{const frame=event.target.closest?.('.md-rail-frame');if(!frame||frame.dataset.modalOpen!=='true')return;if(event.key==='Escape'){event.preventDefault();setRailExpanded(frame,false,true);return;}if(event.key==='Tab'){const controls=[...frame.querySelector('.md-navigation-rail').querySelectorAll('button')].filter(e=>e.getClientRects().length),first=controls[0],last=controls.at(-1);if(event.shiftKey&&event.target===first){event.preventDefault();last.focus();}else if(!event.shiftKey&&event.target===last){event.preventDefault();first.focus();}}});
  document.addEventListener('htmx:beforeCleanupElement',event=>{const root=event.detail.elt;for(const label of root.querySelectorAll?.('.md-rail-frame .md-nav-label')||[])railLabelSizes.unobserve(label);});
  document.addEventListener('htmx:afterSwap',initRailFrames);document.addEventListener('htmx:historyRestore',initRailFrames);if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',initRailFrames);else initRailFrames();
  const clock = { now: () => performance.now(), setTimeout: (fn, delay) => setTimeout(fn, delay), clearTimeout: id => clearTimeout(id) };
  function syncSlider(slider,changed) {
    const inputs=[...slider.querySelectorAll('input[type="range"]')];
    const min=Number(inputs[0].min),max=Number(inputs[0].max),span=max-min||1;
    const values=sliderValues(inputs.map(input=>input.valueAsNumber),inputs.indexOf(changed),min,max);
    const start=inputs.length===2?(values[0]-min)/span:0,end=(values.at(-1)-min)/span;
    slider.style.setProperty('--_slider-start',String(start));slider.style.setProperty('--_slider-end',String(end));slider.style.setProperty('--_slider-mid',String((start+end)/2));
    const vertical=slider.dataset.orientation==='vertical',rect=slider.getBoundingClientRect(),css=getComputedStyle(slider),stateSize=Number.parseFloat(css.getPropertyValue('--_slider-state-size'));
    const axis=(vertical?rect.height:rect.width)-stateSize;
    const widths=[];
    for(const [index,input]of inputs.entries()) {
      input.valueAsNumber=values[index];
      input.setAttribute('aria-valuemin',String(index===1?values[0]:min));input.setAttribute('aria-valuemax',String(inputs.length===2&&index===0?values[1]:max));
      const control=input.parentElement;
      control.querySelector('.md-slider-value').textContent=String(values[index]);
      widths.push(Number.parseFloat(getComputedStyle(control).getPropertyValue('--_slider-control-width')));
    }
    const leading=Number.parseFloat(css.getPropertyValue('--_slider-leading')),trailing=Number.parseFloat(css.getPropertyValue('--_slider-trailing'));
    slider.style.setProperty('--_slider-before-gap',`${widths[0]/2+leading}px`);slider.style.setProperty('--_slider-after-gap',`${widths.at(-1)/2+trailing}px`);
    slider.style.setProperty('--_slider-start-gap',inputs.length===2?`${widths[0]/2+trailing}px`:'0px');slider.style.setProperty('--_slider-end-gap',`${widths.at(-1)/2+leading}px`);
    const fractions=values.map(value=>(value-min)/span);
    const decorate=(element,fraction,size=0)=>{
      element.dataset.active=String(fraction>=start&&fraction<=end);
      element.style.visibility=fractions.some((v,i)=>Math.abs(v-fraction)*axis<widths[i]/2+leading+size/2)?'hidden':'visible';
    };
    for(const tick of slider.querySelectorAll('.md-slider-tick')) decorate(tick,(Number(tick.dataset.value)-min)/span);
    decorate(slider.querySelector('.md-slider-stop-start'),0);decorate(slider.querySelector('.md-slider-stop-end'),1);
    const iconSize=Number.parseFloat(css.getPropertyValue('--_slider-icon-size')),iconPadding=Number.parseFloat(css.getPropertyValue('--_slider-icon-padding'));
    for(const icon of slider.querySelectorAll('.md-slider-icon')) { const distance=(iconPadding+iconSize/2)/(axis||1);decorate(icon,icon.classList.contains('md-slider-leading')?distance:1-distance,iconSize); }
    const nubs=[...slider.querySelectorAll('.md-slider-nub')].map(nub=>nub.getBoundingClientRect());
    slider.dataset.overlap=String(nubs.length===2&&nubs[0].left<=nubs[1].right&&nubs[1].left<=nubs[0].right&&nubs[0].top<=nubs[1].bottom&&nubs[1].top<=nubs[0].bottom);
  }
  const sliderResize=new ResizeObserver(entries=>{for(const {target}of entries)syncSlider(target);});
  for(const type of ['input','focusin','focusout','pointerdown','pointerup','pointercancel','transitionend','click']) document.addEventListener(type,event=>{
    const slider=event.target.closest?.('.md-slider');if(!slider)return;
    if(event.type==='pointerdown')event.target.closest('.md-slider-input')?.toggleAttribute('data-slider-hover',event.pointerType!=='touch');
    if(event.type==='pointercancel')for(const control of slider.querySelectorAll('[data-slider-hover]'))control.removeAttribute('data-slider-hover');
    syncSlider(slider,event.target);
  },type==='input');
  document.addEventListener('pointermove',event=>{
    const slider=event.target.closest?.('.md-slider');if(!slider)return;
    for(const control of slider.querySelectorAll('.md-slider-input')) {
      const thumb=control.querySelector('.md-slider-thumb').getBoundingClientRect(),nub=control.querySelector('.md-slider-nub').getBoundingClientRect();
      const hover=event.pointerType!=='touch'&&!control.querySelector('input').disabled&&event.target.closest('.md-slider-input')===control&&event.clientX>=Math.min(thumb.left,nub.left)&&event.clientX<=Math.max(thumb.right,nub.right)&&event.clientY>=Math.min(thumb.top,nub.top)&&event.clientY<=Math.max(thumb.bottom,nub.bottom);
      control.toggleAttribute('data-slider-hover',hover);
    }
  });
  document.addEventListener('pointerout',event=>{
    const slider=event.target.closest?.('.md-slider');if(slider&&!slider.contains(event.relatedTarget))for(const control of slider.querySelectorAll('[data-slider-hover]'))control.removeAttribute('data-slider-hover');
  });
  document.addEventListener('reset',event=>queueMicrotask(()=>{for(const slider of event.target.querySelectorAll('.md-slider'))syncSlider(slider);}));
  function syncToolbar(group) {
    const style=getComputedStyle(group),value=key=>Number.parseFloat(style.getPropertyValue('--_toolbar-'+key));
    const demo=group.parentElement,scene=demo.closest('.md-toolbar-scene'),items=[...group.querySelectorAll('[data-toolbar-item]')];
    items.forEach(item=>item.hidden=false);
    const width=item=>{const cs=getComputedStyle(item);return item.offsetWidth+parseFloat(cs.marginLeft)+parseFloat(cs.marginRight);};
    if(group.dataset.variant==='docked'){
      const available=group.clientWidth-Number.parseFloat(style.paddingLeft)-Number.parseFloat(style.paddingRight);
      const fits=()=>{const visible=items.filter(item=>!item.hidden);return visible.reduce((sum,item)=>sum+width(item),0)+Math.max(0,visible.length-1)*value('min-gap')<=available;};
      for(const item of items.slice(0,-1).reverse())if(!fits())item.hidden=true;
      group.style.setProperty('--_toolbar-gap',toolbarGap(available,items.filter(item=>!item.hidden).map(width),value('min-gap'),value('max-gap'))+'px');
    }else{
      if(!scene)return;
      const vertical=group.dataset.orientation==='vertical',clearance=parseFloat(getComputedStyle(demo).getPropertyValue('--_toolbar-clearance'))||16;
      const fab=demo.querySelector('.md-toolbar-fab'),fabSpace=fab?(vertical?fab.offsetHeight:fab.offsetWidth)+parseFloat(getComputedStyle(fab)[vertical?'marginTop':'marginInlineStart']):0;
      const available=(vertical?scene.clientHeight:scene.clientWidth)-2*clearance-fabSpace;
      const fits=()=>vertical?group.scrollHeight<=available:group.scrollWidth<=available;
      for(const item of items.slice(0,-1).reverse())if(!fits())item.hidden=true;
    }
    for(const item of items){
      const proxy=demo.querySelector('[data-toolbar-overflow="'+item.dataset.toolbarItem+'"]');
      if(proxy){proxy.hidden=!item.hidden;proxy.querySelector('button').disabled=item.disabled;}
    }
    if(!items.some(item=>!item.hidden&&item.tabIndex===0)){const first=items.find(item=>!item.hidden&&!item.disabled);items.forEach(item=>item.tabIndex=item===first?0:-1);}
  }
  const toolbarResize=new ResizeObserver(records=>{for(const {target}of records){const group=target.matches('.md-toolbar-scene')?target.querySelector('.md-toolbar'):target.closest('.md-toolbar');if(group)syncToolbar(group);}});
  document.addEventListener('scroll',event=>{
    const demo=event.target;
    if(demo.matches?.('[data-appbar-demo][data-interactive="true"]'))demo.querySelector('.md-app-bar').toggleAttribute('data-scrolled',demo.scrollTop>0);
  },true);
  function initialize(root = document) {
    for(const group of root.querySelectorAll('.md-toolbar:not([data-toolbar-ready])')) {
      group.setAttribute('data-toolbar-ready','');
      const items=[...group.querySelectorAll('[data-toolbar-item]')],first=items.find(item=>!item.disabled);
      items.forEach(item=>item.tabIndex=item===first?0:-1);
      syncToolbar(group);toolbarResize.observe(group);items.forEach(item=>toolbarResize.observe(item));const scene=group.closest('.md-toolbar-scene');if(scene)toolbarResize.observe(scene);
    }
    for(const slider of root.querySelectorAll('.md-slider:not([data-slider-ready])')) { slider.setAttribute('data-slider-ready','');syncSlider(slider);sliderResize.observe(slider); }

    for (const bar of root.querySelectorAll('.md-snackbar[popover]')) {
      if (timers.has(bar)) continue;
      const timer = snackbarTimer(() => { if (bar.isConnected && bar.matches(':popover-open')) bar.hidePopover(); }, clock);
      timers.set(bar, timer);
      bar.addEventListener('toggle', () => { if (!bar.matches(':popover-open')) timer.stop(); });
      bar.addEventListener('pointerenter', () => timer.pause('pointer'));
      bar.addEventListener('pointerleave', () => timer.resume('pointer'));
      bar.addEventListener('focusin', () => timer.pause('focus'));
      bar.addEventListener('focusout', event => { if (!bar.contains(event.relatedTarget)) timer.resume('focus'); });
    }
  }
  function positionFABMenu(panel) {
    if(!panel.matches(':popover-open'))return;
    const trigger=panel.closest('.md-fab-menu').querySelector('.md-fab-menu-opener');
    const anchor=trigger.getBoundingClientRect(),box=panel.getBoundingClientRect();
    const gap=Number.parseFloat(getComputedStyle(panel).getPropertyValue('--md-sys-measurement-space200')),closeGap=Number.parseFloat(getComputedStyle(panel).getPropertyValue('--_fab-menu-close-gap'));
    const left=getComputedStyle(panel).direction==='rtl'?anchor.left:anchor.right-box.width;
    panel.style.left=Math.max(gap,Math.min(left,document.documentElement.clientWidth-box.width-gap))+'px';
    panel.style.top=Math.max(gap,anchor.top-closeGap-box.height)+'px';
  }
  document.addEventListener('toggle',event=>{
    const panel=event.target;if(!panel.matches?.('.md-fab-menu-panel[popover]'))return;
    const trigger=panel.closest('.md-fab-menu').querySelector('.md-fab-menu-opener');
    const opening=panel.matches(':popover-open');
    trigger.classList.toggle('md-fab-menu-close',opening);trigger.setAttribute('aria-expanded',String(opening));trigger.setAttribute('aria-label',opening?'Close menu':'Create');
    if(panel.matches(':popover-open')){
      positionFABMenu(panel);const items=panel.querySelectorAll('[role="menuitem"]');
      if(panel.dataset.initialFocus)items[panel.dataset.initialFocus==='last'?items.length-1:0]?.focus();else trigger.focus({preventScroll:true});delete panel.dataset.initialFocus;
    }else if(panel.contains(document.activeElement)||document.activeElement===document.body)trigger.focus({preventScroll:true});
  },true);
  document.addEventListener('keydown',event=>{
    const trigger=event.target.closest?.('.md-fab-menu-opener');
    if(trigger&&event.key==='Tab'&&trigger.getAttribute('aria-expanded')==='true')document.getElementById(trigger.getAttribute('popovertarget')).hidePopover();
    if(trigger&&['ArrowDown','ArrowUp'].includes(event.key)){
      event.preventDefault();const panel=document.getElementById(trigger.getAttribute('popovertarget'));
      const edge=event.key==='ArrowUp'?'last':'first';
      if(panel.matches(':popover-open')){const items=panel.querySelectorAll('[role="menuitem"]');items[edge==='last'?items.length-1:0]?.focus();}
      else{panel.dataset.initialFocus=edge;panel.showPopover();}return;
    }
    const panel=event.target.closest?.('.md-fab-menu-panel[popover]');if(!panel)return;
    const items=[...panel.querySelectorAll('[role="menuitem"]')],index=items.indexOf(event.target.closest('[role="menuitem"]'));
    if(event.key==='Tab'){panel.hidePopover();panel.closest('.md-fab-menu').querySelector('.md-fab-menu-opener').focus();return;}
    const next=navigationIndex(index,items.length,event.key,false);if(next!==null){event.preventDefault();items[next]?.focus();}
  });
  document.addEventListener('click',event=>{
    const button=event.target.closest?.('[data-fab-menu-action]');if(!button)return;
    const root=button.closest('.md-fab-menu'),panel=button.closest('.md-fab-menu-panel');
    root.querySelector('output').textContent=button.textContent.trim()+' selected.';
    if(panel.matches('[popover]')){panel.hidePopover();root.querySelector('.md-fab-menu-opener').focus();}
  });
  for(const type of ['resize','scroll'])window.addEventListener(type,()=>{
    for(const panel of document.querySelectorAll('.md-fab-menu-panel:popover-open'))positionFABMenu(panel);
  },true);
  function positionSplitMenu(menu) {
    if(!menu.matches(':popover-open'))return;
    const trigger=menu.closest('.md-split-demo').querySelector('.md-split-trailing');
    const r=trigger.getBoundingClientRect(),box=menu.getBoundingClientRect();
    const style=getComputedStyle(trigger),gap=Number.parseFloat(style.getPropertyValue('--md-sys-measurement-space50')),margin=Number.parseFloat(style.getPropertyValue('--md-sys-measurement-space100'));
    const pos=splitMenuPosition(r,trigger.closest('.md-split-button').getBoundingClientRect(),box,{width:document.documentElement.clientWidth,height:window.innerHeight},style.direction==='rtl',gap,margin);
    menu.style.left=pos.left+'px';menu.style.top=pos.top+'px';
  }
  document.addEventListener('toggle',event=>{
    const menu=event.target;if(!menu.matches?.('.md-split-menu'))return;
    menu.closest('.md-split-demo').querySelector('.md-split-trailing').setAttribute('aria-expanded',String(menu.matches(':popover-open')));
    positionSplitMenu(menu);
    if(menu.matches(':popover-open')){const items=menu.querySelectorAll('button');items[menu.dataset.initialFocus==='last'?items.length-1:0]?.focus();delete menu.dataset.initialFocus;}
  },true);
  document.addEventListener('keydown',event=>{
    const trigger=event.target.closest?.('.md-split-trailing');
    if(trigger&&!trigger.disabled&&['ArrowDown','ArrowUp'].includes(event.key)) {
      event.preventDefault();const menu=document.getElementById(trigger.getAttribute('popovertarget'));
      menu.dataset.initialFocus=event.key==='ArrowUp'?'last':'first';menu.showPopover();return;
    }
    const item=event.target.closest?.('.md-split-menu button');if(!item)return;
    if(event.key==='Tab'){const demo=item.closest('.md-split-demo');item.closest('.md-split-menu').hidePopover();demo.querySelector('.md-split-trailing').focus();return;}
    const items=[...item.closest('.md-split-menu').querySelectorAll('button:not(:disabled)')];
    const next=navigationIndex(items.indexOf(item),items.length,event.key,false);
    if(next!==null){event.preventDefault();items[next]?.focus();}
  });
  document.addEventListener('click',event=>{
    const button=event.target.closest?.('[data-split-primary],[data-split-option]');if(!button)return;
    const demo=button.closest('.md-split-demo');demo.querySelector('output').textContent=button.hasAttribute('data-split-primary')?'Added to cart.':button.textContent+'.';
    const menu=button.closest('.md-split-menu');if(menu){menu.hidePopover();demo.querySelector('.md-split-trailing').focus();}
  });
  for(const type of ['resize','scroll'])window.addEventListener(type,()=>{
    for(const menu of document.querySelectorAll('.md-split-menu:popover-open'))positionSplitMenu(menu);
  },true);
  function positionTooltip(tip) {
    if (CSS.supports('top', 'anchor(bottom)') || !tip.matches(':popover-open')) return;
    const trigger = tip.closest('.md-tooltip-demo')?.querySelector('button');
    if (!trigger) return;
    const anchor = trigger.getBoundingClientRect(), box = tip.getBoundingClientRect();
    const gap = Number.parseFloat(getComputedStyle(tip).getPropertyValue('--md-sys-measurement-space100'));
    const width = document.documentElement.clientWidth, height = document.documentElement.clientHeight;
    tip.style.margin = '0';
    tip.style.left = `${Math.max(gap, Math.min(anchor.left, width - box.width - gap))}px`;
    tip.style.top = `${Math.max(gap, anchor.bottom + box.height + gap <= height ? anchor.bottom + gap : anchor.top - box.height - gap)}px`;
    tip.setAttribute('data-positioned', '');
  }
  document.addEventListener('toggle', event => {
    if (event.target.matches?.('.md-tooltip[popover]')) positionTooltip(event.target);
  }, true);
  for (const type of ['resize', 'scroll']) window.addEventListener(type, () => {
    for (const tip of document.querySelectorAll('.md-tooltip:popover-open')) positionTooltip(tip);
  }, true);
  initialize();
  document.addEventListener('htmx:afterSwap', () => initialize());
  document.addEventListener('htmx:beforeCleanupElement', event => {
    const root = event.detail.elt;
    if(root.matches?.('.md-toolbar,[data-toolbar-item]'))toolbarResize.unobserve(root);
    for(const item of root.querySelectorAll?.('.md-toolbar,[data-toolbar-item]')||[])toolbarResize.unobserve(item);
    if(root.matches?.(".md-slider"))sliderResize.unobserve(root);
    for(const slider of root.querySelectorAll?.(".md-slider")||[])sliderResize.unobserve(slider);
    if (root.matches?.('.md-snackbar[popover]')) timers.get(root)?.stop();
    for (const bar of root.querySelectorAll?.('.md-snackbar[popover]') || []) timers.get(bar)?.stop();
  });
  document.addEventListener('click', event => {
    if (!(event.target instanceof Element)) return;
    const modal=event.target.closest('[data-navigation-modal]');
    if (modal && event.target===modal) {
      const rect=modal.getBoundingClientRect();
      if (event.clientX<rect.left || event.clientX>rect.right || event.clientY<rect.top || event.clientY>rect.bottom) modal.close();
    }
    const button = event.target.closest('button');
    if (!button) return;
    if (button.matches('.md-tab')) selectTab(button);
    if (button.hasAttribute('data-navigation-item')) selectNavigation(button);
    if (button.hasAttribute('data-navigation-open')) button.closest('[data-navigation-demo]').querySelector('dialog').showModal();
    if (button.hasAttribute('data-navigation-expand')) {const frame=button.closest('.md-rail-frame');if(frame&&frame.dataset.interactive==='true')setRailExpanded(frame,frame.querySelector('.md-navigation-rail').dataset.variant!=='expanded',true);}
    if (button.hasAttribute('data-navigation-dismiss'))setRailExpanded(button.closest('.md-rail-frame'),false,true);
    if (button.hasAttribute('data-navigation-create')){const frame=button.closest('.md-rail-frame');if(frame)frame.querySelector('[data-navigation-title]').textContent='Create file';}
    const demo = button.closest('[data-snackbar-demo]');
    if (demo) {
      const bar = demo.querySelector('.md-snackbar');
      if (button.hasAttribute('data-snackbar-open')) {
        demo.querySelector('output').textContent = '';
        bar.showPopover();
        timers.get(bar).start(Number(bar.dataset.snackbarDuration));
      } else if (button.hasAttribute('data-snackbar-close') || button.hasAttribute('data-snackbar-action')) {
        timers.get(bar).stop();
        bar.hidePopover();
        demo.querySelector('output').textContent = button.hasAttribute('data-snackbar-action') ? 'Message undone.' : 'Notification dismissed.';
      }
    }
    if (button.hasAttribute('data-tooltip-close')) button.closest('.md-tooltip[popover]')?.hidePopover();
  });
  document.addEventListener('keydown', event => {
    if (event.key !== 'Escape') return;
    for (const bar of document.querySelectorAll('.md-snackbar:popover-open')) { timers.get(bar)?.stop(); bar.hidePopover(); }
    if (event.target instanceof Element) event.target.closest('.md-tooltip-demo')?.setAttribute('data-tooltip-dismissed', '');
  });
  for (const type of ['pointerover', 'focusin']) document.addEventListener(type, event => {
    if (!(event.target instanceof Element)) return;
    const demo = event.target.closest('.md-tooltip-demo');
    if (demo && !demo.contains(event.relatedTarget)) demo.removeAttribute('data-tooltip-dismissed');
  });
})();
