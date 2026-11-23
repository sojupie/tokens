(() => {
  const MAX_ENTRIES = 50;
  const MAX_BYTES = 4 * 1024 * 1024;
  const TTL_MS = 10 * 60 * 1000;
  const MAX_INFLIGHT = 2;
  const DWELL_MS = 65;

  const base = document.currentScript ? new URL('../', document.currentScript.src).pathname.replace(/\/$/, '') : '';
  const cache = new Map();
  const inflight = new Map();
  let totalBytes = 0;
  let activeNavigations = 0;
  let pending;
  let timer;

  function canPrefetch() {
    const connection = navigator.connection;
    return document.visibilityState !== 'hidden' &&
      !connection?.saveData &&
      !['slow-2g', '2g'].includes(connection?.effectiveType);
  }

  function eligibleURL(rawHref, allowCurrent = false) {
    if (!rawHref) return null;
    let url;
    try {
      url = new URL(rawHref, location.href);
    } catch {
      return null;
    }
    if (url.origin !== location.origin || url.hash || (!allowCurrent && url.href === location.href)) return null;
    const path = url.pathname;
    if (path !== base + '/' && path !== base + '/for-agents' && !['components', 'styles', 'changes'].some(section =>
      path === `${base}/${section}` || path.startsWith(`${base}/${section}/`))) return null;
    return url;
  }

  function destination(link, allowCurrent = false) {
    if (!link || link.hasAttribute('download') || link.closest('[hx-get], [data-hx-get], [hx-boost="false"], [data-hx-boost="false"]') ||
        (link.target && link.target !== '_self')) return null;
    return eligibleURL(link.href, allowCurrent);
  }

  function cacheKey(url) {
    return url.pathname + url.search;
  }

  function getCached(key) {
    const entry = cache.get(key);
    if (!entry) return null;
    if (Date.now() >= entry.expiresAt) {
      cache.delete(key);
      totalBytes -= entry.bytes;
      return null;
    }
    cache.delete(key);
    cache.set(key, entry);
    return entry;
  }

  function putCached(key, entry) {
    if (!entry || entry.bytes > MAX_BYTES / 2) return;
    const existing = cache.get(key);
    if (existing) {
      cache.delete(key);
      totalBytes -= existing.bytes;
    }
    cache.set(key, entry);
    totalBytes += entry.bytes;
    while (cache.size > MAX_ENTRIES || totalBytes > MAX_BYTES) {
      const oldestKey = cache.keys().next().value;
      if (oldestKey === undefined) break;
      const oldest = cache.get(oldestKey);
      cache.delete(oldestKey);
      totalBytes -= oldest.bytes;
    }
  }

  function clearCache() {
    cache.clear();
    inflight.clear();
    totalBytes = 0;
  }

  function parseDocument(html) {
    if (typeof DOMParser === 'undefined') return null;
    try {
      return new DOMParser().parseFromString(html, 'text/html');
    } catch {
      return null;
    }
  }

  function shellMismatched(doc) {
    if (!doc) return false;
    const current = document.head?.querySelector?.('meta[name="tokens-shell"]')?.getAttribute('content') || '';
    const next = doc.head?.querySelector?.('meta[name="tokens-shell"]')?.getAttribute('content') || '';
    return Boolean(current && next && current !== next);
  }

  function makeResponse(entry) {
    if (typeof Response === 'function') {
      const response = new Response(entry.html, {
        status: entry.status,
        headers: { 'Content-Type': entry.contentType },
      });
      try {
        Object.defineProperty(response, 'url', { value: entry.url });
      } catch {}
      return response;
    }
    return {
      ok: entry.status >= 200 && entry.status < 300,
      status: entry.status,
      url: entry.url,
      headers: { get: name => name.toLowerCase() === 'content-type' ? entry.contentType : null },
      text: async () => entry.html,
    };
  }

  function buildEntry(finalURL, response, html) {
    const contentType = response.headers?.get?.('content-type') || 'text/html; charset=utf-8';
    return {
      html,
      status: response.status || 200,
      contentType,
      url: finalURL,
      bytes: html.length * 2,
      expiresAt: Date.now() + TTL_MS,
    };
  }

  function prefetchURL(url, urgent = false) {
    const key = cacheKey(url);
    if (getCached(key) || inflight.has(key)) return;
    if (!canPrefetch()) return;
    if (!urgent && (activeNavigations > 0 || inflight.size >= MAX_INFLIGHT)) return;

    const request = {
      method: 'GET',
      credentials: 'same-origin',
      headers: {
        'HX-Request': 'true',
        'HX-Boosted': 'true',
        'HX-Request-Type': 'full',
        'Purpose': 'prefetch',
      },
    };
    const task = fetch(url.href, request).then(async response => {
      const contentType = response.headers?.get?.('content-type') || 'text/html';
      if (response.status !== 200 || !contentType.includes('text/html')) return null;
      const html = await response.text();
      const doc = parseDocument(html);
      if (shellMismatched(doc)) return null;
      const finalURL = response.url || url.href;
      const entry = buildEntry(finalURL, response, html);
      const finalParsed = eligibleURL(finalURL, true);
      if (finalParsed) putCached(cacheKey(finalParsed), entry);
      if (doc) globalThis.TokenAssets?.warm?.(doc);
      return entry;
    }).catch(() => null).finally(() => {
      if (inflight.get(key) === task) inflight.delete(key);
    });
    inflight.set(key, task);
  }

  function cancel() {
    clearTimeout(timer);
    pending = undefined;
  }

  function schedule(link) {
    if (pending === link) return;
    cancel();
    const url = destination(link);
    if (!url || !canPrefetch() || getCached(cacheKey(url)) || inflight.has(cacheKey(url))) return;
    pending = link;
    timer = setTimeout(() => {
      pending = undefined;
      const current = link.isConnected ? destination(link) : null;
      if (!current || current.href !== url.href) return;
      prefetchURL(current, false);
    }, DWELL_MS);
  }

  function prefetchImmediate(link) {
    cancel();
    const url = destination(link);
    if (!url) return;
    prefetchURL(url, true);
  }

  function register() {
    globalThis.htmx?.registerExtension('tokens-prefetch', {
      htmx_before_request(elt, { ctx }) {
        if (ctx.target !== document.body) return;
        const method = (ctx.request.method || 'GET').toUpperCase();
        if (method !== 'GET') return;
        const headers = ctx.request.headers || {};
        if (headers['HX-Boosted'] !== 'true' && headers['HX-History-Restore-Request'] !== 'true') return;
        const url = eligibleURL(ctx.request.action, true);
        if (!url) return;
        const key = cacheKey(url);
        const cached = getCached(key);
        if (cached) {
          ctx.fetch = () => Promise.resolve(makeResponse(cached));
          return;
        }
        const inProgress = inflight.get(key);
        if (inProgress) {
          ctx.fetch = () => inProgress.then(entry => entry ? makeResponse(entry) : fetch(ctx.request.action, ctx.request));
          return;
        }
        const origFetch = ctx.fetch || fetch;
        ctx.fetch = async (action, init) => {
          const response = await origFetch(action, init);
          const contentType = response.headers?.get?.('content-type') || 'text/html';
          if (response.status !== 200 || !contentType.includes('text/html')) return response;
          const html = await response.text();
          const doc = parseDocument(html);
          const finalURL = response.url || new URL(action, location.href).href;
          const entry = buildEntry(finalURL, response, html);
          if (!shellMismatched(doc)) {
            const finalParsed = eligibleURL(finalURL, true);
            if (finalParsed) putCached(cacheKey(finalParsed), entry);
          }
          return makeResponse(entry);
        };
      },
    });
  }

  register();

  document.addEventListener('htmx:before:request', event => {
    if (event.detail?.ctx?.target === document.body) activeNavigations++;
  });
  document.addEventListener('htmx:finally:request', event => {
    if (event.detail?.ctx?.target === document.body && activeNavigations > 0) activeNavigations--;
  });

  document.addEventListener('pointerover', event => {
    if (event.pointerType === 'mouse' || event.pointerType === 'pen') schedule(event.target?.closest?.('a[href]'));
  }, { capture: true, passive: true });
  document.addEventListener('pointerout', event => {
    if (pending && !pending.contains(event.relatedTarget)) cancel();
  }, { capture: true, passive: true });
  document.addEventListener('pointerdown', event => {
    if (event.button > 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
    const link = event.target?.closest?.('a[href]');
    if (link) prefetchImmediate(link);
  }, { capture: true, passive: true });
  document.addEventListener('focusin', event => schedule(event.target?.closest?.('a[href]')), { capture: true, passive: true });
  document.addEventListener('focusout', event => {
    if (pending && !pending.contains(event.relatedTarget)) cancel();
  }, { capture: true, passive: true });
  document.addEventListener('visibilitychange', cancel);
  window.addEventListener('pagehide', cancel);

  globalThis.TokenPrefetch = {
    clear: clearCache,
    has: key => !!getCached(key),
    size: () => cache.size,
  };
})();
