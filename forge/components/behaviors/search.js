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
    state.panel.dataset.variant=searchLayout(state.panel.dataset.layout,(state.root.querySelector('[data-search-viewport]')||state.scene).clientWidth);
    state.panel.style.setProperty('--_search-results-height',state.results.scrollHeight+'px');
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
    for(const region of state.root.querySelectorAll('[data-search-background]'))region.inert=open;
    state.back.hidden=false;state.back.inert=!open;
    const leading=state.panel.querySelector('.md-search-navigation,.md-search-leading');
    if(leading){leading.hidden=false;leading.inert=open;}
    if(open)filter(state);else state.panel.querySelector('output').textContent='';
    measure(state);
    state.opener?.setAttribute('aria-expanded',String(open));
    if(focus){state.restoring=!open;(open?state.input:state.opener||state.input).focus({preventScroll:true});state.restoring=false;}
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
    const opener=event.target.closest('button[data-search-open]');if(opener){state.opener=opener;setOpen(state,true);return;}
    if(event.target.closest('.md-search-bar')&&!event.target.closest('button,.md-search-options')){setOpen(state,true);return;}
    if(event.target.closest('[data-search-back],.md-search-scrim')){setOpen(state,false);return;}
    if(event.target.closest('[data-search-clear]')){state.input.value='';state.submitted=false;filter(state);if(state.open)state.input.focus({preventScroll:true});return;}
    const result=event.target.closest('[data-search-result]');
    if(result){
      const label=result.dataset.searchResult;
      setOpen(state,false);
      state.root.dispatchEvent(new CustomEvent('md-search-select',{bubbles:true,detail:{label}}));
      return;
    }
    if(event.target.closest('.md-search-avatar'))state.root.dispatchEvent(new CustomEvent('md-search-profile',{bubbles:true}));
  });
  document.addEventListener('change',event=>{
    if(!event.target.matches?.('[data-search-exact]'))return;
    const state=stateFor(event.target);if(state){state.exact=event.target.checked;filter(state);}
  });
  document.addEventListener('input',event=>{
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
    const node=(event.detail?.elt || event.target);if(typeof node?.querySelectorAll!=='function')return;
    for(const root of [...(node.matches?.('.md-search-demo')?[node]:[]),...node.querySelectorAll('.md-search-demo')]){const state=examples.get(root);if(state)scenes.unobserve(state.scene);examples.delete(root);}
  });
  document.addEventListener('htmx:afterSwap',init);document.addEventListener('htmx:historyRestore',init);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',init);else init();
  document.addEventListener('tokens:assets-ready',init);
})();
