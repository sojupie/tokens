(() => {
  const disclosures = new WeakMap();
  document.addEventListener('click', async event => {
    const summary = event.target.closest?.('details.tokens-export-disclosure > summary');
    if (!summary) return;
    event.preventDefault();
    const detail = summary.parentElement, previous = disclosures.get(detail);
    const opening = !(previous ? previous.opening : detail.open);
    const from = detail.getBoundingClientRect().height;
    previous?.animation.cancel();
    detail.open = true;
    detail.style.height = '';
    const style = getComputedStyle(detail);
    const closedHeight = summary.getBoundingClientRect().height + parseFloat(style.borderTopWidth) + parseFloat(style.borderBottomWidth);
    const to = opening ? detail.getBoundingClientRect().height : closedHeight;
    const children = [...detail.children].filter(child => child !== summary);
    for (const child of children) child.inert = !opening;
    if (matchMedia('(prefers-reduced-motion: reduce)').matches) {
      detail.open = opening;
      for (const child of children) child.inert = false;
      detail.removeAttribute('data-expanding');disclosures.delete(detail);return;
    }
    detail.setAttribute('data-expanding', String(opening));
    const animation = detail.animate([{height:from+'px'}, {height:to+'px'}], {...MaterialMotion.timing(detail,'default-spatial'),fill:'both'});
    disclosures.set(detail, {animation,opening});
    try {
      await animation.finished;
      if (disclosures.get(detail)?.animation !== animation) return;
      detail.open = opening;
      for (const child of children) child.inert = false;
      animation.cancel();detail.removeAttribute('data-expanding');disclosures.delete(detail);
    } catch { /* A second activation retargets from the current rendered height. */ }
  });
  document.addEventListener('close', event => {
    for (const detail of event.target.querySelectorAll?.('details.tokens-export-disclosure') || []) {
      const state = disclosures.get(detail);if (!state) continue;
      state.animation.cancel();detail.open = state.opening;detail.removeAttribute('data-expanding');
      for (const child of detail.children) child.inert = false;
      disclosures.delete(detail);
    }
  }, true);
  // Safari can retain :focus-visible after touch-driven programmatic focus.
  document.addEventListener('pointerdown',()=>document.documentElement.setAttribute('data-pointer-focus',''),true);
  document.addEventListener('keydown',event=>{if(!event.metaKey&&!event.altKey&&!event.ctrlKey)document.documentElement.removeAttribute('data-pointer-focus');},true);
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
  document.addEventListener('htmx:beforeProcessNode',event=>{
    const node=event.detail?.elt || event.target;
    if(node.matches?.('a[href]'))navigationLinks(node);
  });
  // Check deployed assets before hx-head reads and merges a full response.
  const assets=doc=>Array.from(doc.head.querySelectorAll('script[src]:not([data-component-script]),link[rel="stylesheet"]:not([data-component-stylesheet])'),node=>[node.getAttribute('src')||node.getAttribute('href'),node.getAttribute('data-sources')||'',node.getAttribute('data-shapes-version')||''].join('\n')).join('\n');
  document.addEventListener('htmx:head:before:merge',event=>{
    const ctx=event.detail.ctx;
    if(ctx.response.status!==200){event.preventDefault();return;}
    if(ctx.target!==document.body)event.preventDefault();
    const read=ctx.response.raw.text.bind(ctx.response.raw);
    ctx.response.raw.text=async()=>{
      globalThis.TokenBenchmark?.mark("response-read-start");
      const html=await read();
      globalThis.TokenBenchmark?.mark("response-read-end");
      globalThis.TokenBenchmark?.mark("html-parse-start");
      const incoming=new DOMParser().parseFromString(html,'text/html');
      globalThis.TokenBenchmark?.mark("html-parse-end");
      ctx.catalogAssetChange=(ctx.target===document.body&&assets(incoming)!==assets(document))||globalThis.TokenAssets.stylesChanged(incoming);
      if(ctx.catalogAssetChange)return '';
      globalThis.TokenBenchmark?.mark("styles-prepare-start");
      await globalThis.TokenAssets.prepareStyles(incoming,ctx.request.signal);
      globalThis.TokenBenchmark?.mark("styles-prepare-end");
      // Dependencies now belong to the head, never to replaceable regions.
      return html.replace(/<link\b[^>]*\bdata-component-stylesheet\b[^>]*>/gi,'');
    };
  });
  document.addEventListener('htmx:head:before:remove',event=>{
    if(event.detail.headElement.matches('[data-component-script],[data-component-stylesheet]'))event.preventDefault();
  });
  document.addEventListener('htmx:after:request',event=>{
    const ctx=event.detail.ctx;
    if(ctx.catalogAssetChange){event.preventDefault();location.assign(ctx.target===document.body?ctx.response.raw.url:location.href);return;}
    if(ctx.target===document.body&&ctx.response.status===200){
      const menu=document.getElementById('tokens-site-menu');if(menu?.matches(':modal'))menu.close();
    }
  });
  // Explicit gallery links and browser history both return to the saved offset.
  const galleryKey=()=>/\/(components|styles)\/?$/.test(location.pathname)?location.pathname+location.search:null;
  function saveGallery(){const key=galleryKey();if(key&&document.getElementById('subject-list'))try{sessionStorage.setItem('tokens-scroll:'+key,String(window.scrollY));}catch{}}
  function restoreGallery(){const key=galleryKey();if(!key)return;let saved;try{saved=sessionStorage.getItem('tokens-scroll:'+key);}catch{return;}if(saved===null)return;
    Promise.resolve(globalThis.TokenAssets?.ready).then(()=>requestAnimationFrame(()=>window.scrollTo({top:Number(saved),left:0,behavior:'instant'})));
  }
  document.addEventListener('htmx:before:request',event=>{if(event.detail.ctx.request.headers['HX-Boosted']==='true')saveGallery();});
  document.addEventListener('htmx:after:swap',event=>{const headers=event.detail.ctx.request.headers;if(headers['HX-Boosted']==='true'||headers['HX-History-Restore-Request']==='true')restoreGallery();});
  window.addEventListener('pagehide',saveGallery);
  function initialize(){
    navigationLinks();
  }
  document.addEventListener('htmx:afterSwap',initialize);document.addEventListener('htmx:historyRestore',initialize);
  if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',initialize);else initialize();
})();
