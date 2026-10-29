(() => {
  // Safari can retain :focus-visible after touch-driven programmatic focus.
  document.addEventListener('pointerdown',()=>document.documentElement.setAttribute('data-pointer-focus',''),true);
  document.addEventListener('keydown',event=>{if(!event.metaKey&&!event.altKey&&!event.ctrlKey)document.documentElement.removeAttribute('data-pointer-focus');},true);
  const initialized=new WeakSet(),openers=new WeakMap();
  // HTML specimens avoid Safari foreignObject compositing and clipping bugs.
  const thumbnailSizes=new ResizeObserver(entries=>{for(const {target,contentRect} of entries)target.style.setProperty("--_thumbnail-scale",String(contentRect.width/360));});
  const thumbnails=new WeakSet();
  // Styles and measured scale must both be settled before removing the blur.
  function settleThumbnails(failed=false){
    const pending=[...document.querySelectorAll(".tokens-component-thumbnail:not([data-preview-ready])")];
    for(const thumbnail of pending){
      thumbnail.style.setProperty("--_thumbnail-scale",String(thumbnail.getBoundingClientRect().width/360));
      if(failed)thumbnail.setAttribute("data-preview-error","");
    }
    if(!failed)requestAnimationFrame(()=>{
      for(const thumbnail of pending){
        thumbnail.removeAttribute("data-preview-error");
        thumbnail.setAttribute("data-preview-ready","");
      }
    });
  }
  document.addEventListener("tokens:styles-ready",()=>settleThumbnails());
  document.addEventListener("tokens:styles-error",()=>settleThumbnails(true));
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
    const assets=doc=>Array.from(doc.head.querySelectorAll('script[src]:not([data-component-script]),link[rel="stylesheet"]'),node=>[node.getAttribute('src')||node.getAttribute('href'),node.getAttribute('data-sources')||'',node.getAttribute('data-shapes-version')||''].join('\n')).join('\n');
    // A deployed asset change needs the new head and behavior as well as new content.
    if(assets(incoming)!==assets(document)){location.assign(url);return false;}
    const menu=document.getElementById('tokens-site-menu');if(menu?.matches(":modal"))menu.close();
    document.querySelectorAll('#tokens-site-menu [data-navigation-item]').forEach(link=>{
      const current=Array.from(incoming.querySelectorAll('#tokens-site-menu [data-navigation-item]')).find(next=>next.getAttribute('href')===link.getAttribute('href'))?.getAttribute('aria-current');
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
  // Explicit gallery links and browser history both return to the saved offset.
  const galleryKey=()=>/\/(components|styles)\/?$/.test(location.pathname)?location.pathname+location.search:null;
  function saveGallery(){const key=galleryKey();if(key&&document.getElementById('subject-list'))try{sessionStorage.setItem('tokens-scroll:'+key,String(window.scrollY));}catch{}}
  function restoreGallery(){const key=galleryKey();if(!key)return;let saved;try{saved=sessionStorage.getItem('tokens-scroll:'+key);}catch{return;}if(saved===null)return;
    Promise.resolve(globalThis.TokenAssets?.ready).then(()=>requestAnimationFrame(()=>window.scrollTo({top:Number(saved),left:0,behavior:'instant'})));
  }
  document.addEventListener('htmx:beforeRequest',event=>{if(event.detail.requestConfig?.boosted)saveGallery();});
  document.addEventListener('htmx:afterSettle',event=>{if(event.detail.requestConfig?.boosted)restoreGallery();});
  document.addEventListener('htmx:historyRestore',restoreGallery);
  window.addEventListener('pagehide',saveGallery);
  window.addEventListener('pageshow',restoreGallery);
  function initialize(){
    navigationLinks();
    for(const thumbnail of document.querySelectorAll(".tokens-component-thumbnail"))if(!thumbnails.has(thumbnail)){thumbnails.add(thumbnail);thumbnailSizes.observe(thumbnail);}
    const panel=document.getElementById('tokens-site-download');if(!panel||initialized.has(panel))return;
    initialized.add(panel);panel.addEventListener('change',()=>update(panel));panel.addEventListener('close',()=>{const opener=openers.get(panel);if(opener?.isConnected)opener.focus({preventScroll:true});});update(panel);
  }
  document.addEventListener('click',event=>{
    const panel=document.getElementById('tokens-site-download');if(!panel)return;
    const trigger=event.target.closest?.('[data-export-open]');if(trigger){openers.set(panel,trigger);panel.showModal();panel.querySelector('[role="combobox"],button')?.focus({preventScroll:true});return;}
    if(event.target.closest?.('[data-export-cancel]'))panel.close();
  });
  document.addEventListener('htmx:afterSwap',initialize);document.addEventListener('htmx:historyRestore',initialize);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',initialize);else initialize();
})();
