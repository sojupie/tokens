(() => {
  // Apply the arrow-free style only where the browser reserves a scrollbar gutter.
  if (globalThis.CSS?.supports('selector(::-webkit-scrollbar-button)')) {
    const probe = document.createElement('div');
    probe.style.cssText = 'position:fixed;top:0;left:-100px;width:32px;height:32px;overflow:scroll;visibility:hidden;scrollbar-width:auto;scrollbar-color:auto';
    document.body.append(probe);
    document.documentElement.toggleAttribute('data-classic-scrollbars', probe.clientWidth < probe.offsetWidth);
    probe.remove();
  }

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
  // HTMX's listener cancels click defaults before checking link modifiers.
  // Leave those clicks to the browser before they reach the link handler.
  document.addEventListener('click', event => {
    if (!event.ctrlKey && !event.metaKey && !event.shiftKey && !event.altKey) return;
    const link = event.target.closest?.('a[href]');
    if (link?.closest('[hx-boost]')?.getAttribute('hx-boost') === 'true') event.stopPropagation();
  }, true);
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
  // Check deployed assets before hx-head reads and merges a response.
  const assets=doc=>Array.from(doc.head.querySelectorAll('script[src]:not([data-component-script]),link[rel="stylesheet"]:not([data-component-stylesheet])'),node=>[node.getAttribute('src')||node.getAttribute('href'),node.getAttribute('data-sources')||'',node.getAttribute('data-shapes-version')||''].join('\n')).join('\n');
  const shellVersion=doc=>doc.head?.querySelector?.('meta[name="tokens-shell"]')?.getAttribute('content')||'';
  const shellChanged=incoming=>{
    const next=shellVersion(incoming),current=shellVersion(document);
    if(next&&current)return next!==current;
    return assets(incoming)!==assets(document);
  };
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
      ctx.catalogAssetChange=(ctx.target===document.body&&shellChanged(incoming))||globalThis.TokenAssets.stylesChanged(incoming);
      if(ctx.catalogAssetChange)return '';
      globalThis.TokenBenchmark?.mark("styles-prepare-start");
      await globalThis.TokenAssets.prepareStyles(incoming,ctx.request.signal);
      globalThis.TokenBenchmark?.mark("styles-prepare-end");
      // Dependencies now belong to the head, never to replaceable regions.
      return html.replace(/<link\b[^>]*\bdata-component-stylesheet\b[^>]*>/gi,'');
    };
  });
  document.addEventListener('htmx:head:before:add',event=>{
    const node=event.detail.headElement;
    if(!node?.matches)return;
    if(node.matches('meta[name="tokens-shell"],link[data-component-stylesheet]')){event.preventDefault();return;}
    if(node.matches('meta[name="description"]')){
      const existing=document.head?.querySelector?.('meta[name="description"]');
      if(existing){existing.setAttribute('content',node.getAttribute('content')||'');event.preventDefault();}
    }
  });
  document.addEventListener('htmx:head:before:remove',event=>{
    if(event.detail.headElement.matches('[data-component-script],[data-component-stylesheet]'))event.preventDefault();
  });
  function syncExportDialog(main){
    const panel=document.getElementById('tokens-site-download');
    if(!panel||!main?.dataset)return;
    for(const [dst,src] of [['componentExport','exportComponent'],['defaultScope','exportDefaultScope'],['hasCss','exportHasCss'],['hasImplementation','exportHasImplementation']]){
      if(src in main.dataset)panel.dataset[dst]=main.dataset[src];
    }
    const form=panel.querySelector('form');
    if(!form)return;
    for(const [name,key] of [['kind','exportKind'],['slug','exportSlug'],['set','exportSet'],['q','exportQuery']]){
      const input=form.querySelector(`input[name="${name}"]`);
      if(input&&key in main.dataset)input.value=main.dataset[key];
    }
  }
  document.addEventListener('htmx:before:swap', event => {
    const {ctx, tasks} = event.detail;
    const task = tasks.find(task => task.type === 'main');
    // HTMX's attribute settling would restore the server's initial range
    // attributes after the sizing controller has measured the new preview.
    if (task?.fragment.querySelector('[data-preview-sizing-pending]')) task.swapSpec.settle = 0;
    if (ctx.target !== document.body) return;
    const main = document.getElementById('tokens-site-main');
    const incomingMain = task?.fragment.querySelector('#tokens-site-main');
    const page = document.getElementById('tokens-site-page');
    const incoming = task?.fragment.querySelector('#tokens-site-page');
    if (main && incomingMain && !incoming) {
      // Lean boosted response: keep both header and footer/export dialog mounted.
      syncExportDialog(incomingMain);
      task.target = main;
      task.fragment = document.createDocumentFragment();
      task.fragment.append(incomingMain);
      task.swapSpec.style = 'outerHTML';
      if (task.swapSpec.show === 'top') task.swapSpec.showTarget = 'window';
    } else if (page && incoming) {
      // Keep the shell mounted. The response's partials still update its current links.
      task.target = page;
      task.fragment = document.createDocumentFragment();
      task.fragment.append(incoming);
      task.swapSpec.style = 'outerHTML';
      if (task.swapSpec.show === 'top') task.swapSpec.showTarget = 'window';
    }
  });
  // HTMX 4 only fires before:cleanup on [data-htmx-powered] nodes. Emit it for
  // replaced non-HTMX containers so component/controller lifecycles always dispose.
  document.addEventListener('htmx:before:swap', event => {
    if (event.defaultPrevented) return;
    for (const task of event.detail.tasks || []) {
      const target = task.target;
      if (target && !target._htmx && typeof target.dispatchEvent === 'function' &&
          ['outerHTML', 'innerHTML', 'outerSync'].includes(task.swapSpec?.style)) {
        target.dispatchEvent(new CustomEvent('htmx:beforeCleanupElement', {bubbles: true, detail: {elt: target}}));
      }
    }
  });
  document.addEventListener('htmx:after:request',event=>{
    const ctx=event.detail.ctx;
    if(ctx.catalogAssetChange){globalThis.TokenPrefetch?.clear?.();event.preventDefault();location.assign(ctx.target===document.body?ctx.response.raw.url:location.href);return;}
    if(ctx.target===document.body&&ctx.response.status===200){
      const menu=document.getElementById('tokens-site-menu');if(menu?.matches(':modal'))menu.close();
    }
  });
  const pageRequests = new Set();
  let pageProgressTimer;
  document.addEventListener('htmx:before:request', event => {
    const {ctx} = event.detail;
    if (ctx.target !== document.body || !document.getElementById('tokens-page-progress')) return;
    if (pageRequests.size === 0)
      pageProgressTimer = setTimeout(() => document.documentElement.setAttribute('data-page-loading', ''), 200);
    pageRequests.add(ctx);
  });
  document.addEventListener('htmx:finally:request', event => {
    if (!pageRequests.delete(event.detail.ctx) || pageRequests.size) return;
    clearTimeout(pageProgressTimer);
    document.documentElement.removeAttribute('data-page-loading');
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
