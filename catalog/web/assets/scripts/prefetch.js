(() => {
  const connection = navigator.connection;
  if (connection?.saveData || ['slow-2g', '2g'].includes(connection?.effectiveType)) return;

  const base = document.currentScript ? new URL('../', document.currentScript.src).pathname.replace(/\/$/, '') : '';
  const seen = new Set();
  let pending;
  let timer;

  function destination(link, allowCurrent = false) {
    if (!link || link.hasAttribute('download') || link.closest('[hx-get], [data-hx-get], [hx-boost="false"], [data-hx-boost="false"]') ||
        (link.target && link.target !== '_self')) return;
    const url = new URL(link.href, location.href);
    if (url.origin !== location.origin || url.hash || (!allowCurrent && url.href === location.href)) return;
    const path = url.pathname;
    if (path !== base + '/' && path !== base + '/for-agents' && !['components', 'styles', 'changes'].some(section =>
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
      link.dispatchEvent(new CustomEvent('tokens:prefetch'));
    }, 65);
  }

  // The upstream extension owns request construction, response reuse and expiry.
  // Arm persistent navigation links too; the current destination is skipped at use.
  document.addEventListener('htmx:beforeProcessNode', event => {
    const link = event.detail?.elt || event.target;
    if (link.matches?.('a[href]') && destination(link, true))
      link.setAttribute('hx-preload', 'tokens:prefetch');
  });

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
