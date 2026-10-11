(() => {
  const connection = navigator.connection;
  if (connection?.saveData || ['slow-2g', '2g'].includes(connection?.effectiveType)) return;

  const speculation = typeof HTMLScriptElement !== 'undefined' &&
    HTMLScriptElement.supports?.('speculationrules');
  const hint = document.createElement('link');
  if (!speculation && !hint.relList?.supports('prefetch')) return;

  const base = document.currentScript ? new URL('../', document.currentScript.src).pathname.replace(/\/$/, '') : '';
  const seen = new Set();
  let pending;
  let timer;

  function destination(link) {
    if (!link || link.hasAttribute('download') || link.closest('[hx-get], [data-hx-get]') ||
        (link.target && link.target !== '_self')) return;
    const url = new URL(link.href, location.href);
    if (url.origin !== location.origin || url.hash || url.href === location.href) return;
    const path = url.pathname;
    if (path !== base + '/' && !['components', 'styles', 'changes'].some(section =>
      path === `${base}/${section}` || path.startsWith(`${base}/${section}/`))) return;
    return url.href;
  }

  function cancel() {
    clearTimeout(timer);
    pending = undefined;
  }

  function schedule(link) {
    if (pending === link) return;
    cancel();
    const url = destination(link);
    if (!url || seen.has(url) || seen.size >= 8 || document.visibilityState === 'hidden') return;
    pending = link;
    timer = setTimeout(() => {
      pending = undefined;
      if (!link.isConnected || destination(link) !== url || document.visibilityState === 'hidden' ||
          navigator.connection?.saveData || ['slow-2g', '2g'].includes(navigator.connection?.effectiveType)) return;
      seen.add(url);
      if (speculation) {
        const rules = document.createElement('script');
        rules.type = 'speculationrules';
        rules.textContent = JSON.stringify({prefetch: [{source: 'list', urls: [url], eagerness: 'immediate'}]});
        document.head.append(rules);
      } else {
        const preload = document.createElement('link');
        preload.rel = 'prefetch';
        preload.as = 'document';
        preload.href = url;
        document.head.append(preload);
      }
    }, 65);
  }

  document.addEventListener('pointerover', event => {
    if (event.pointerType === 'mouse' || event.pointerType === 'pen') schedule(event.target?.closest?.('a[href]'));
  }, {capture: true, passive: true});
  document.addEventListener('pointerout', event => {
    if (pending && !pending.contains(event.relatedTarget)) cancel();
  }, {capture: true, passive: true});
  document.addEventListener('focusin', event => schedule(event.target?.closest?.('a[href]')), {capture: true, passive: true});
  document.addEventListener('focusout', event => {
    if (pending && !pending.contains(event.relatedTarget)) cancel();
  }, {capture: true, passive: true});
  document.addEventListener('visibilitychange', cancel);
  window.addEventListener('pagehide', cancel);
})();
