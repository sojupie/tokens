(() => {
 function scrollEdges(top,height,total){return {start:top>0,end:Math.ceil(top+height)<total};}
 function isDirty(state){return state.input.value!==state.committed;}
 if(typeof module!=='undefined')module.exports={scrollEdges,isDirty};
 if(typeof document==='undefined')return;
 const states=new WeakMap(),reduced=matchMedia('(prefers-reduced-motion: reduce)');
 const stateFor=target=>states.get(target.closest?.('.md-dialog-demo')?.querySelector('[data-dialog-preview]'));
 function edges(state){const e=scrollEdges(state.body.scrollTop,state.body.clientHeight,state.body.scrollHeight);state.panel.dataset.topScrolled=String(e.start);state.panel.dataset.bottomScrolled=String(e.end);}
 function preset(state){const value=state.panel.dataset.scrollStart;state.body.scrollTop=value==='bottom'?state.body.scrollHeight:value==='middle'?Math.min(Number.parseFloat(getComputedStyle(state.panel.querySelector('.md-full-header')).height),(state.body.scrollHeight-state.body.clientHeight)/2):0;edges(state);}
 function timing(state,closing){const s=getComputedStyle(state.panel),phase=closing?'exit':'enter';return{duration:Number.parseFloat(s.getPropertyValue('--_dialog-'+phase+'-duration')),easing:s.getPropertyValue('--_dialog-'+phase+'-easing').trim(),fill:'both'};}
 function validity(state,show=false){const valid=state.input.validity.valid;state.input.setAttribute('aria-invalid',String(show&&!valid));state.error.textContent=show&&!valid?'Enter a name.':'Required';for(const button of state.panel.querySelectorAll('[data-dialog-save]'))button.disabled=state.panel.dataset.variant!=='full-screen'&&!valid;return valid;}
 function rest(state){state.revision++;state.animation?.cancel();state.animation=null;state.closing=false;state.panel.inert=false;state.input.value=state.committed;validity(state);state.opener?.setAttribute('aria-expanded','false');if(state.opener?.isConnected)state.opener.focus({preventScroll:true});}
 function open(state){
  if(state.panel.open)return;state.revision++;state.animation?.cancel();state.closing=false;state.panel.inert=false;state.input.value=state.committed;validity(state);state.panel.showModal();preset(state);state.opener.setAttribute('aria-expanded','true');
  state.panel.querySelector('button,input,select,textarea,a[href]').focus({preventScroll:true});
  if(!reduced.matches){const full=state.panel.dataset.variant==='full-screen';state.animation=state.panel.animate(full?[{transform:'translateY(100%)'},{transform:'none'}]:[{opacity:0},{opacity:1}],timing(state,false));}
 }
 async function close(state,save,discard=false){
  if(!state.panel.open)return;
  if(!save&&!discard&&state.discard&&isDirty(state)){if(!state.discard.open){state.discard.showModal();state.discard.querySelector('[data-dialog-keep]').focus();}return;}
  if(save){if(!validity(state,true)){state.input.focus();return;}state.committed=state.input.value;state.output.textContent='Saved '+state.committed;}else{state.input.value=state.committed;validity(state);}
  if(state.panel.hasAttribute('data-dialog-inline'))return;
  const current=getComputedStyle(state.panel),from={transform:current.transform,opacity:current.opacity},revision=++state.revision;state.animation?.cancel();state.animation=null;state.closing=true;state.panel.inert=true;
  if(!reduced.matches){const full=state.panel.dataset.variant==='full-screen',animation=state.panel.animate([from,full?{transform:'translateY(100%)',opacity:1}:{opacity:0}],timing(state,true));state.animation=animation;try{await animation.finished;}catch{return;}if(state.revision!==revision)return;}
  state.panel.close();rest(state);
 }
 function cycleFocus(event){
  const panel=event.currentTarget;if(event.key!=='Tab'||!panel.matches(':modal'))return;
  const controls=[...panel.querySelectorAll('button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),a[href],[tabindex="0"]')].filter(el=>el.getClientRects().length);
  const first=controls[0],last=controls.at(-1);if(!first)return;
  if(event.shiftKey&&document.activeElement===first){event.preventDefault();last.focus();}else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus();}
 }
 function init(){for(const panel of document.querySelectorAll('[data-dialog-preview]')){
  if(states.has(panel))continue;const demo=panel.closest('.md-dialog-demo');if(!demo)continue;const state={panel,body:panel.querySelector('[data-dialog-content]'),input:panel.querySelector('[data-dialog-value]'),error:panel.querySelector('[data-dialog-error]'),discard:demo.querySelector('[data-dialog-discard]'),opener:demo.querySelector('[data-dialog-open]'),output:demo.querySelector('[data-dialog-output]'),revision:0,animation:null};state.committed=state.input.value;states.set(panel,state);
  if(panel.open&&!panel.hasAttribute('data-dialog-inline'))panel.close();
  panel.addEventListener('keydown',cycleFocus);state.discard?.addEventListener('keydown',cycleFocus);
  state.discard?.addEventListener('cancel',event=>{event.preventDefault();state.discard.close();});
  panel.addEventListener('cancel',event=>{event.preventDefault();close(state,false);});panel.addEventListener('close',()=>{if(state.panel.open)return;rest(state);});
  state.body.addEventListener('scroll',()=>edges(state));state.observer=new ResizeObserver(()=>edges(state));state.observer.observe(state.body);if(panel.hasAttribute('data-dialog-inline'))preset(state);
 }}
 document.addEventListener('click',event=>{const state=stateFor(event.target);if(!state)return;if(event.target.closest('[data-dialog-keep]'))state.discard.close();else if(event.target.closest('[data-dialog-discard-confirm]')){state.discard.close();close(state,false,true);}else if(event.target.closest('[data-dialog-open]'))open(state);else if(event.target.closest('[data-dialog-save]'))close(state,true);else if(event.target.closest('[data-dialog-cancel]'))close(state,false);else if(event.target===state.panel&&!state.panel.hasAttribute('data-dialog-inline')&&state.panel.dataset.variant!=='full-screen'){const r=state.panel.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)close(state,false);}});
 document.addEventListener('input',event=>{if(event.target.matches?.('[data-dialog-value]')){const state=stateFor(event.target);if(state)validity(state,true);}});
 document.addEventListener('htmx:beforeCleanupElement',event=>{const node=event.detail?.elt;if(!node)return;for(const panel of [...(node.matches?.('[data-dialog-preview]')?[node]:[]),...node.querySelectorAll('[data-dialog-preview]')]){const state=states.get(panel);if(!state)continue;state.revision++;state.animation?.cancel();state.observer.disconnect();if(state.discard?.open)state.discard.close();if(panel.open)panel.close();states.delete(panel);}});
 reduced.addEventListener('change',()=>{if(!reduced.matches)return;for(const panel of document.querySelectorAll('[data-dialog-preview]')){const state=states.get(panel);if(!state?.animation)continue;state.revision++;state.animation.cancel();state.animation=null;if(state.closing){panel.close();rest(state);}}});
 document.addEventListener('htmx:afterSwap',init);document.addEventListener('htmx:historyRestore',init);if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',init);else init();
})();
