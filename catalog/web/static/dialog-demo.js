// Catalog's example form rules. Forge owns dialog presentation and dismissal.
(() => {
 function isDirty(state){return state.input.value!==state.committed;}
 if(typeof module!=='undefined')module.exports={isDirty};
 if(typeof document==='undefined')return;
 const states=new WeakMap();
 const stateFor=target=>states.get(target.closest?.('.md-dialog-demo')?.querySelector('[data-dialog-preview]'));
 function validity(state,show=false){
  const valid=state.input.validity.valid;state.input.setAttribute('aria-invalid',String(show&&!valid));state.error.textContent=show&&!valid?'Enter a name.':'Required';
  for(const button of state.panel.querySelectorAll('[data-dialog-save]'))button.disabled=state.panel.dataset.variant!=='full-screen'&&!valid;
  return valid;
 }
 function preset(state){
  const value=state.panel.dataset.scrollStart;
  state.body.scrollTop=value==='bottom'?state.body.scrollHeight:value==='middle'?Math.min(Number.parseFloat(getComputedStyle(state.panel.querySelector('.md-full-header')).height),(state.body.scrollHeight-state.body.clientHeight)/2):0;
 }
 function init(){for(const panel of document.querySelectorAll('[data-dialog-preview]')){
  if(states.has(panel))continue;const demo=panel.closest('.md-dialog-demo');if(!demo)continue;
  const state={panel,body:panel.querySelector('[data-dialog-content]'),input:panel.querySelector('[data-dialog-value]'),error:panel.querySelector('[data-dialog-error]'),discard:demo.querySelector('[data-dialog-discard]'),output:demo.querySelector('[data-dialog-output]')};
  state.committed=state.input.value;states.set(panel,state);
  panel.addEventListener('md-dialog-beforeopen',()=>{state.input.value=state.committed;validity(state)});
  panel.addEventListener('md-dialog-opened',()=>preset(state));
  panel.addEventListener('md-dialog-beforeclose',event=>{
   if(event.detail.result==='save'){
    if(!validity(state,true)){event.preventDefault();state.input.focus();return}
    state.committed=state.input.value;state.output.textContent='Saved '+state.committed;
   }else if(event.detail.result!=='discard'&&state.discard&&isDirty(state)){
    event.preventDefault();if(!state.discard.open)state.discard.showModal();
   }
  });
  panel.addEventListener('md-dialog-closed',()=>{state.input.value=state.committed;validity(state)});
  if(panel.hasAttribute('data-dialog-inline'))preset(state);
 }}
 document.addEventListener('click',event=>{
  const state=stateFor(event.target);if(!state)return;
  if(event.target.closest('[data-dialog-keep]'))state.discard.close();
  else if(event.target.closest('[data-dialog-discard-confirm]')){state.discard.close();MaterialDialogs.close(state.panel,'discard','action')}
  else if(event.target.closest('[data-dialog-save]'))MaterialDialogs.close(state.panel,'save','action');
  else if(event.target.closest('[data-dialog-cancel]'))MaterialDialogs.close(state.panel,'cancel','action');
 });
 document.addEventListener('input',event=>{if(event.target.matches?.('[data-dialog-value]')){const state=stateFor(event.target);if(state)validity(state,true)}});
 document.addEventListener('htmx:beforeCleanupElement',event=>{const node=event.detail?.elt;if(typeof node?.querySelectorAll!=='function')return;for(const panel of [...(node.matches?.('[data-dialog-preview]')?[node]:[]),...node.querySelectorAll('[data-dialog-preview]')]){const state=states.get(panel);if(state?.discard?.open)state.discard.close();states.delete(panel)}});
 for(const name of ['DOMContentLoaded','htmx:afterSwap','htmx:historyRestore','tokens:assets-ready'])document.addEventListener(name,init);init();
})();
