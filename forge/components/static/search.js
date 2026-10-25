(() => {
  const matches=(label,query,exact=false)=>{
    const value=query.trim().toLocaleLowerCase(),name=label.toLocaleLowerCase();
    return !value || (exact?name===value:name.includes(value));
  };
  const searchLayout=(layout,width)=>layout==='adaptive'?(width<600?'full-screen':'docked'):layout;
  if(typeof module!=='undefined')module.exports={matches,searchLayout};
  if(typeof document==='undefined')return;
  const examples=new WeakMap();
  const stateFor=target=>examples.get(target.closest?.('.md-search-demo'));
  function measure(state){
    state.panel.dataset.variant=searchLayout(state.panel.dataset.layout,(state.root.querySelector('.md-search-workspace')||state.scene).clientWidth);
    state.panel.style.setProperty('--_search-results-height',state.results.scrollHeight+'px');
    const workspace=state.root.querySelector('.md-search-workspace');if(workspace)state.root.querySelector('[data-search-composition-width-label]').textContent=Math.round(workspace.getBoundingClientRect().width)+'px';
  }
  function filter(state){
    const query=state.input.value.trim();let count=0;
    for(const item of state.results.querySelectorAll('[data-search-result]')){
      item.hidden=!matches(item.dataset.searchResult,query,state.exact);if(!item.hidden)count++;
    }
    state.results.querySelector('[data-search-empty]').hidden=count!==0;
    state.results.querySelector('[data-search-heading]').textContent=state.submitted?(query?`Results for “${query}”`:'Results'):query?'Quick results':'Suggestions';
    const noun=query||state.submitted?'result':'suggestion';
    state.panel.querySelector('output').textContent=!state.open?'':`${count} ${noun}${count===1?'':'s'} available${query?` for ${query}`:''}`;
    for(const target of state.panel.querySelectorAll('[data-search-clear-target]'))target.hidden=!query;
    measure(state);
  }
  function setOpen(state,open,focus=true){
    state.panel.querySelector('.md-search-bar').classList.toggle('md-search-trigger',!open);
    state.open=open;state.panel.dataset.open=String(open);state.root.dataset.searchOpen=String(open);
    state.panel.setAttribute('role',open?'dialog':'search');
    state.input.setAttribute('aria-expanded',String(open));state.input.readOnly=!open;
    state.results.inert=!open;state.results.setAttribute('aria-hidden',String(!open));
    state.scrim.hidden=!open;
    for(const region of [state.page,...state.root.querySelectorAll('.md-search-mail-navigation,.md-search-mail-detail')])region.inert=open;
    state.back.hidden=false;state.back.inert=!open;
    const leading=state.panel.querySelector('.md-search-navigation,.md-search-leading');
    if(leading){leading.hidden=false;leading.inert=open;}
    if(open)filter(state);else state.panel.querySelector('output').textContent='';
    measure(state);
    if(focus){state.restoring=!open;state.input.focus({preventScroll:true});state.restoring=false;}
  }
  const scenes=new ResizeObserver(entries=>{for(const {target}of entries){const state=stateFor(target);if(state)measure(state);}});
  function init(){
    for(const root of document.querySelectorAll('.md-search-demo')){
      if(examples.has(root))continue;
      const state={root,panel:root.querySelector('.md-search-view'),input:root.querySelector('.md-search-input'),scene:root.querySelector('.md-search-scene'),results:root.querySelector('.md-search-results'),page:root.querySelector('.md-search-page'),scrim:root.querySelector('.md-search-scrim'),back:root.querySelector('[data-search-back]')};
      delete root.dataset.searchReady;examples.set(root,state);
      // Only the requested disclosure survives HTMX history; transient open state does not.
      state.results.hidden=false;filter(state);setOpen(state,root.dataset.disclosure==='true',false);scenes.observe(state.scene);state.panel.getBoundingClientRect();root.dataset.searchReady='';
    }
  }
  document.addEventListener('focusin',event=>{
    const state=stateFor(event.target);if(state&&event.target===state.input&&!state.open&&!state.restoring)setOpen(state,true,false);
  });
  document.addEventListener('click',event=>{
    const state=stateFor(event.target);if(!state)return;
    if(event.target.closest('.md-search-bar')&&!event.target.closest('button,.md-search-options')){setOpen(state,true);return;}
    if(event.target.closest('[data-search-back],.md-search-scrim')){setOpen(state,false);return;}
    if(event.target.closest('[data-search-clear]')){state.input.value='';state.submitted=false;filter(state);if(state.open)state.input.focus({preventScroll:true});return;}
    const result=event.target.closest('[data-search-result]');
    if(result){const label=result.dataset.searchResult;for(const item of state.root.querySelectorAll('[data-search-message]'))item.setAttribute('aria-pressed',String(item.dataset.searchMessage===label));state.root.querySelector('[data-search-page-title]').textContent=label;state.root.querySelector('[data-search-page-description]').textContent='Selected from search results.';if(state.panel.dataset.layout==='adaptive')state.root.dataset.messageOpen='true';state.root.querySelector('[data-search-feedback]').textContent=`${label} opened`;setOpen(state,false,state.panel.dataset.layout!=='adaptive');if(state.panel.dataset.layout==='adaptive')state.root.querySelector('[data-search-page-title]').focus({preventScroll:true});return;}
    const destination=event.target.closest('[data-search-mail-destination]');
    if(destination){for(const item of state.root.querySelectorAll('[data-search-mail-destination]')){if(item.dataset.searchMailDestination===destination.dataset.searchMailDestination)item.setAttribute('aria-current','page');else item.removeAttribute('aria-current');}state.page.querySelector('[data-search-folder-title]').textContent=destination.querySelector('.md-nav-label').textContent;return;}
    const message=event.target.closest('[data-search-message]');
    if(message){for(const item of state.root.querySelectorAll('[data-search-message]'))item.setAttribute('aria-pressed',String(item===message));state.root.querySelector('[data-search-page-title]').textContent=message.dataset.searchMessage;state.root.querySelector('[data-search-page-description]').textContent='Review the latest component updates and examples.';state.root.dataset.messageOpen='true';state.root.querySelector('[data-search-page-title]').focus({preventScroll:true});return;}
    if(event.target.closest('[data-search-message-back]')){delete state.root.dataset.messageOpen;state.root.querySelector('[data-search-message][aria-pressed="true"]')?.focus({preventScroll:true});return;}
    if(event.target.closest('[data-search-compose],[data-search-reply]')){state.root.querySelector('[data-search-feedback]').textContent='Example mail composition: use search to explore the component.';return;}
    if(event.target.closest('[data-search-navigation]'))state.root.querySelector('[data-search-navigation-panel]').showModal();
    if(event.target.closest('.md-search-avatar'))state.root.querySelector('[data-search-feedback]').textContent='Profile selected';
  });
  document.addEventListener('change',event=>{
    if(!event.target.matches?.('[data-search-exact]'))return;
    const state=stateFor(event.target);if(state){state.exact=event.target.checked;filter(state);}
  });
  document.addEventListener('input',event=>{
    if(event.target.matches?.('[data-search-composition-width]')){const root=event.target.closest('.md-search-demo');root.querySelector('.md-search-workspace').style.setProperty('--_search-composition-width',event.target.value+'px');measure(stateFor(event.target));return;}
    if(!event.target.matches?.('.md-search-input'))return;
    const state=stateFor(event.target);if(state){state.submitted=false;filter(state);}
  });
  document.addEventListener('keydown',event=>{
    const state=stateFor(event.target);if(!state||!state.open||state.root.querySelector('.md-search-options:popover-open'))return;
    const buttons=[...state.results.querySelectorAll('[data-search-result]:not([hidden]) > button')],index=buttons.indexOf(event.target);
    if(event.target===state.input){
      if(event.key==='ArrowDown'||event.key==='ArrowUp'){event.preventDefault();(event.key==='ArrowDown'?buttons[0]:buttons.at(-1))?.focus();}
      if(event.key==='Enter'){event.preventDefault();state.submitted=true;filter(state);state.results.querySelector('[data-search-heading]').focus({preventScroll:true});}
    }else if(index>=0){
      if(event.key==='ArrowDown'){event.preventDefault();buttons[Math.min(index+1,buttons.length-1)]?.focus();}
      if(event.key==='ArrowUp'){event.preventDefault();if(index===0)state.input.focus();else buttons[index-1].focus();}
    }
    if(event.key==='Escape'){event.preventDefault();setOpen(state,false);}
    if(event.key==='Tab'){
      const focusable=[...state.panel.querySelectorAll('button,input')].filter(el=>!el.closest('[hidden],[inert]')&&el.getClientRects().length),i=focusable.indexOf(document.activeElement);
      if(event.shiftKey&&i===0){event.preventDefault();focusable.at(-1)?.focus();}
      else if(!event.shiftKey&&i===focusable.length-1){event.preventDefault();focusable[0]?.focus();}
    }
  });
  document.addEventListener('htmx:beforeCleanupElement',event=>{
    const node=event.detail?.elt;if(!node)return;
    for(const root of [...(node.matches?.('.md-search-demo')?[node]:[]),...node.querySelectorAll('.md-search-demo')]){const state=examples.get(root);if(state)scenes.unobserve(state.scene);examples.delete(root);}
  });
  document.addEventListener('htmx:afterSwap',init);document.addEventListener('htmx:historyRestore',init);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',init);else init();
  document.addEventListener('tokens:assets-ready',init);
})();
