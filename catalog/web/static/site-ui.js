(() => {
  const initialized=new WeakSet(),openers=new WeakMap(),compact=matchMedia('(max-width: 960px)');
  function position(menu,toggle){const r=toggle.getBoundingClientRect();menu.style.top=(r.bottom+8)+'px';menu.style.maxHeight=`calc(100dvh - ${r.bottom+24}px)`;}
  function layout(){
    const menu=document.getElementById('tokens-site-menu'),toggle=document.getElementById('tokens-nav-toggle');if(!menu||!toggle)return;
    if(menu.matches(':popover-open'))menu.hidePopover();toggle.hidden=!compact.matches;menu.classList.toggle('md-menu',compact.matches);
    if(compact.matches){menu.setAttribute('popover','auto');position(menu,toggle);}else{menu.removeAttribute('popover');menu.style.top='';menu.style.maxHeight='';}
  }
  // HTML specimens avoid Safari foreignObject compositing and clipping bugs.
  const thumbnailSizes=new ResizeObserver(entries=>{for(const {target,contentRect} of entries)target.style.setProperty("--_thumbnail-scale",String(contentRect.width/360));});
  const thumbnails=new WeakSet();
  document.addEventListener("htmx:beforeCleanupElement",event=>{const node=event.detail.elt;if(node.matches?.(".tokens-component-thumbnail"))thumbnailSizes.unobserve(node);for(const thumbnail of node.querySelectorAll(".tokens-component-thumbnail"))thumbnailSizes.unobserve(thumbnail);});
  const descriptions={
    'tokens.tokens.json':['Download DTCG JSON','DTCG 2025.10 values and aliases, with source metadata.'],
    'tokens.css':['Download CSS','CSS custom properties only. Component styles, markup and behavior are not included.'],
    'report.json':['Download report','Selected context, omitted values and source extensions needed for conversion.']
  };
  function update(panel){
    const link=panel.querySelector('[data-export-link]');if(!link)return;
    const format=panel.querySelector('input[name="format"]').value,context=panel.querySelector('input[name="context"]').value;
    // Retain the server's collection mount, including after history restoration.
    const url=new URL(link.href);url.pathname=url.pathname.replace(/\/dtcg\/[^/]+\/[^/]+$/,`/dtcg/${context}/${format}`);link.href=url.href;
    link.textContent=descriptions[format][0];panel.querySelector('[data-export-description]').textContent=descriptions[format][1];
    const report=new URL(url);report.pathname=report.pathname.replace(/[^/]+$/,'report.json');panel.querySelector('[data-export-report]').href=report.href;
  }
  // Boost document links only. Downloads, API documents and external links stay native.
  function navigationLinks(root=document){
    const base=new URL('../',document.querySelector('script[src*="site-ui.js"]').src).pathname.replace(/\/$/,'');
    const links=Array.from(root.querySelectorAll('a[href]'));if(root.matches?.('a[href]'))links.push(root);
    links.forEach(link=>{
      const url=new URL(link.href,location.href),path=url.pathname.slice(base.length);
      if(url.origin!==location.origin||!url.pathname.startsWith(base+'/')||
        !(path==='/'||path==='/for-agents'||['components','styles','changes'].some(section=>path==='/'+section||path.startsWith('/'+section+'/'))))link.setAttribute('hx-boost','false');
    });
  }
  document.addEventListener('htmx:beforeProcessNode',event=>navigationLinks(event.detail.elt));
  function updateDocument(html,url){
    const incoming=new DOMParser().parseFromString(html,'text/html');
    const assets=doc=>Array.from(doc.head.querySelectorAll('script[src],link[rel="stylesheet"]'),node=>node.getAttribute('src')||node.getAttribute('href')).join('\n');
    // A deployed asset change needs the new head and behavior as well as new content.
    if(assets(incoming)!==assets(document)){location.assign(url);return false;}
    const menu=document.getElementById('tokens-site-menu');if(menu?.matches(':popover-open'))menu.hidePopover();
    document.querySelectorAll('.tokens-nav-link').forEach(link=>{
      const current=Array.from(incoming.querySelectorAll('.tokens-nav-link')).find(next=>next.getAttribute('href')===link.getAttribute('href'))?.getAttribute('aria-current');
      if(current)link.setAttribute('aria-current',current);else link.removeAttribute('aria-current');
    });
    for(const selector of ['meta[name="description"]','meta[name="robots"]','link[rel="canonical"]','link[rel="alternate"]']){
      document.head.querySelectorAll(selector).forEach(node=>node.remove());
      incoming.head.querySelectorAll(selector).forEach(node=>document.head.append(node.cloneNode(true)));
    }
    return true;
  }
  document.addEventListener('htmx:beforeSwap',event=>{
    if(event.detail.requestConfig?.boosted&&event.detail.xhr.status===200&&!updateDocument(event.detail.xhr.responseText,event.detail.xhr.responseURL))event.detail.shouldSwap=false;
  });
  document.addEventListener('htmx:historyCacheMissLoad',event=>updateDocument(event.detail.response,event.detail.path));
  function initialize(){
    navigationLinks();
    for(const thumbnail of document.querySelectorAll(".tokens-component-thumbnail"))if(!thumbnails.has(thumbnail)){thumbnails.add(thumbnail);thumbnailSizes.observe(thumbnail);}
    const menu=document.getElementById('tokens-site-menu'),toggle=document.getElementById('tokens-nav-toggle');
    if(menu&&toggle&&!initialized.has(menu)){
      initialized.add(menu);menu.addEventListener('toggle',()=>{const open=menu.matches(':popover-open');toggle.setAttribute('aria-expanded',String(open));if(open){position(menu,toggle);menu.querySelector('a')?.focus({preventScroll:true});}});layout();
    }
    const panel=document.getElementById('tokens-site-download');if(!panel||initialized.has(panel))return;
    initialized.add(panel);panel.addEventListener('change',()=>update(panel));panel.addEventListener('close',()=>{const opener=openers.get(panel);if(opener?.isConnected)opener.focus({preventScroll:true});});update(panel);
  }
  document.addEventListener('click',event=>{
    const panel=document.getElementById('tokens-site-download');if(!panel)return;
    const trigger=event.target.closest?.('[data-export-open]');if(trigger){openers.set(panel,trigger);panel.showModal();panel.querySelector('[role="combobox"],button')?.focus({preventScroll:true});return;}
    if(event.target.closest?.('[data-export-cancel]'))panel.close();
  });
  compact.addEventListener('change',layout);
  window.addEventListener('resize',()=>{const menu=document.getElementById('tokens-site-menu'),toggle=document.getElementById('tokens-nav-toggle');if(compact.matches&&menu&&toggle)position(menu,toggle);});
  document.addEventListener('htmx:afterSwap',initialize);document.addEventListener('htmx:historyRestore',initialize);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',initialize);else initialize();
})();
