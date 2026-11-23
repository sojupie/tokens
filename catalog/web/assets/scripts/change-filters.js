(() => {
  const PAGE_SIZE = 50;
  const jsonCache = new Map();
  const parsedCountsCache = new WeakMap();
  const requestSeq = new WeakMap();
  const hydratedContainers = new WeakSet();

  function parsedCounts(card) {
    let counts = parsedCountsCache.get(card);
    if (!counts) {
      try {
        counts = JSON.parse(card.dataset.changeCounts || '{}');
      } catch {
        counts = {};
      }
      parsedCountsCache.set(card, counts);
    }
    return counts;
  }

  function readFilters(form) {
    return {
      subject: form.querySelector('select[name="subject"]')?.value ?? '',
      snapshot: form.querySelector('select[name="snapshot"]')?.value ?? '',
      type: form.querySelector('select[name="type"]')?.value ?? '',
      q: (form.querySelector('input[name="q"]')?.value ?? '').trim(),
    };
  }

  function syncSelect(form, name, value) {
    const select = form.querySelector(`select[name="${name}"]`);
    if (!select) return;
    select.value = value;
    if (select.value !== value) select.value = '';
    const actual = select.value;
    const root = select.closest?.('[data-field-example]');
    if (!root) return;
    let label = select.selectedOptions?.[0]?.textContent?.trim() ?? '';
    for (const option of root.querySelectorAll('[data-field-option]')) {
      const selected = (option.dataset.fieldValue ?? option.dataset.fieldOption) === actual;
      option.setAttribute('aria-selected', String(selected));
      if (selected && option.dataset.fieldOption) label = option.dataset.fieldOption;
    }
    const display = root.querySelector('[data-field-display]');
    if (display && label) display.textContent = label;
    const combo = root.querySelector('[data-field-combo]');
    if (combo && combo.tagName === 'INPUT') combo.value = label;
  }

  function syncForm(form, filters) {
    syncSelect(form, 'subject', filters.subject);
    syncSelect(form, 'snapshot', filters.snapshot);
    syncSelect(form, 'type', filters.type);
    const input = form.querySelector('input[name="q"]');
    if (input) {
      input.value = filters.q;
      const clear = form.querySelector('[data-search-clear-target]');
      if (clear) clear.hidden = !input.value;
    }
  }

  function withinMountedDOM(container, form, filters, page) {
    if (hydratedContainers.has(container)) return false;
    const mode = form.dataset.changeFilters;
    if (mode === 'global' || mode === 'all') return true;
    if (page !== 1) return false;
    if (mode === 'snapshot') {
      return filters.snapshot === (form.dataset.changeMountedSnapshot ?? '');
    }
    if (mode === 'query') {
      return filters.snapshot === (form.dataset.changeMountedSnapshot ?? '') &&
        filters.type === (form.dataset.changeMountedType ?? '');
    }
    return false;
  }

  function buildFilterURL(form, filters, page = 1) {
    const url = new URL(form.getAttribute('action') || form.action, location.href);
    url.search = '';
    for (const [key, value] of [
      ['subject', filters.subject],
      ['snapshot', filters.snapshot],
      ['type', filters.type],
      ['q', filters.q],
    ]) {
      if (value) url.searchParams.set(key, value);
    }
    if (page > 1) url.searchParams.set('page', String(page));
    return url;
  }

  function updateHistory(form, filters, kind, page = 1) {
    const url = buildFilterURL(form, filters, page);
    if (url.href === location.href) return;
    const state = { ...history.state, htmx: true };
    if (kind === 'search' && state.changeFilter === 'search') {
      history.replaceState(state, '', url);
    } else {
      history.replaceState(state, '', location.href);
      history.pushState({ ...state, changeFilter: kind }, '', url);
    }
  }

  function loadJSON(url) {
    if (!url) return Promise.resolve(null);
    let pending = jsonCache.get(url);
    if (!pending) {
      pending = fetch(url, { headers: { Accept: 'application/json' } }).then(response => {
        if (!response.ok) throw new Error(`Change data request failed (${response.status})`);
        return response.json();
      }).catch(error => {
        jsonCache.delete(url);
        throw error;
      });
      jsonCache.set(url, pending);
    }
    return pending;
  }

  function endpointURL(form, file, version) {
    if (!version) return '';
    const base = new URL(form.getAttribute('action') || form.action, location.href).pathname.replace(/\/$/, '');
    return `${base}/${file}?v=${encodeURIComponent(version)}`;
  }

  function applyGlobalCounts(container, filters, countsMap) {
    let anyVisible = false;
    for (const card of container.querySelectorAll('[data-change-summary]')) {
      const snap = card.dataset.changeSnapshot ?? '';
      const subj = card.dataset.changeSubject ?? '';
      let count = 0;
      if ((!filters.subject || subj === filters.subject) && (!filters.snapshot || snap === filters.snapshot)) {
        count = countsMap
          ? (countsMap.get(`${snap}\0${subj}`) || 0)
          : (parsedCounts(card)[filters.type] || 0);
      }
      card.hidden = count === 0;
      if (count > 0) {
        anyVisible = true;
        const label = card.querySelector('[data-change-count]');
        if (label) label.textContent = count === 1 ? '1 item changed' : `${count} items changed`;
      }
      const link = card.querySelector('[data-change-group-link]');
      if (link) {
        const href = new URL(link.getAttribute('href') || link.href, location.href);
        href.search = '';
        if (filters.q) href.searchParams.set('q', filters.q);
        href.searchParams.set('snapshot', link.dataset.changeSnapshot ?? snap);
        if (filters.type) href.searchParams.set('type', filters.type);
        link.setAttribute('href', href.pathname + href.search);
      }
    }
    for (const batch of container.querySelectorAll('.tokens-timeline-batch')) {
      batch.hidden = !batch.querySelector('[data-change-summary]:not([hidden])');
    }
    const empty = container.querySelector('[data-change-empty]');
    if (empty) empty.hidden = anyVisible;
  }

  function filterGlobal(container, form, filters) {
    const seq = (requestSeq.get(container) || 0) + 1;
    requestSeq.set(container, seq);
    if (!filters.q) {
      applyGlobalCounts(container, filters, null);
      return;
    }
    loadJSON(endpointURL(form, 'search-index.json', form.dataset.changeSearchVersion)).then(index => {
      if (requestSeq.get(container) !== seq || container.isConnected === false) return;
      const countsMap = new Map();
      if (index) {
        const term = filters.q.toLowerCase();
        const typeIndex = filters.type ? index.types.indexOf(filters.type) : -1;
        const typeBit = filters.type ? (typeIndex >= 0 ? (1 << typeIndex) : -1) : 0;
        if (typeBit !== -1) {
          for (const item of index.items) {
            const [sIdx, uIdx, mask, search] = item;
            if (typeBit && (mask & typeBit) === 0) continue;
            if (!search.includes(term)) continue;
            const snap = index.snapshots[sIdx];
            const subj = index.subjects[uIdx];
            if (filters.snapshot && snap !== filters.snapshot) continue;
            if (filters.subject && subj !== filters.subject) continue;
            const key = `${snap}\0${subj}`;
            countsMap.set(key, (countsMap.get(key) || 0) + 1);
          }
        }
      }
      applyGlobalCounts(container, filters, countsMap);
    }, () => {});
  }

  function updateSubjectSummary(container, page, pageCount, total, added, changed, removed) {
    const banner = container.querySelector('[data-change-counts-banner]');
    if (banner) banner.hidden = total === 0;
    const pageSummary = container.querySelector('[data-change-page-summary]');
    if (pageSummary) pageSummary.textContent = `Showing page ${page} of ${pageCount} · ${total} total changes`;
    const typeSummary = container.querySelector('[data-change-type-summary]');
    if (typeSummary) {
      typeSummary.textContent = `${added} added · ${changed} updated · ${removed} removed. One token can have changes in several settings.`;
    }
    const empty = container.querySelector('[data-change-empty]');
    if (empty) empty.hidden = total > 0;
  }

  function filterMountedSubject(container, filters) {
    const term = filters.q.toLowerCase();
    let total = 0;
    let added = 0;
    let changed = 0;
    let removed = 0;
    for (const card of container.querySelectorAll('[data-change-card]')) {
      const snap = card.dataset.changeSnapshot ?? '';
      const type = card.dataset.changeType ?? '';
      const search = card.dataset.changeSearch ?? '';
      const matches = (!filters.snapshot || snap === filters.snapshot) &&
        (!filters.type || type === filters.type) &&
        (!term || search.includes(term));
      card.hidden = !matches;
      if (matches) {
        total++;
        if (type === 'added') added++;
        else if (type === 'changed') changed++;
        else if (type === 'removed') removed++;
      }
    }
    for (const batch of container.querySelectorAll('.tokens-timeline-batch')) {
      batch.hidden = !batch.querySelector('[data-change-card]:not([hidden])');
    }
    updateSubjectSummary(container, 1, 1, total, added, changed, removed);
  }

  function cloneIcon(container, selector) {
    const tpl = container.querySelector(selector);
    return tpl?.content ? tpl.content.cloneNode(true) : null;
  }

  function el(tag, attrs, ...children) {
    const node = document.createElement(tag);
    if (attrs) {
      for (const [key, value] of Object.entries(attrs)) {
        if (value !== undefined && value !== null && value !== false) node.setAttribute(key, value === true ? '' : String(value));
      }
    }
    for (const child of children) {
      if (child !== undefined && child !== null && child !== false) node.append(child);
    }
    return node;
  }

  function renderSubjectFromData(container, form, data, filters, requestedPage) {
    hydratedContainers.add(container);
    const term = filters.q.toLowerCase();
    const snapIndex = filters.snapshot ? data.snapshots.findIndex(s => s.id === filters.snapshot) : -1;
    const typeIndex = filters.type ? data.types.findIndex(t => t.key === filters.type) : -1;
    const matches = [];
    let added = 0;
    let changed = 0;
    let removed = 0;
    if ((!filters.snapshot || snapIndex >= 0) && (!filters.type || typeIndex >= 0)) {
      for (const item of data.items) {
        if (filters.snapshot && item.s !== snapIndex) continue;
        if (filters.type && item.t !== typeIndex) continue;
        if (term && !item.q.includes(term)) continue;
        matches.push(item);
        const key = data.types[item.t]?.key;
        if (key === 'added') added++;
        else if (key === 'changed') changed++;
        else if (key === 'removed') removed++;
      }
    }

    const total = matches.length;
    const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));
    const page = Math.min(Math.max(1, requestedPage), pageCount);
    updateSubjectSummary(container, page, pageCount, total, added, changed, removed);

    const results = container.querySelector('[data-change-results]');
    if (!results) return page;
    const fragment = document.createDocumentFragment();
    const slice = matches.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

    let currentSnap = -1;
    let currentRecords = null;
    for (const item of slice) {
      if (item.s !== currentSnap) {
        currentSnap = item.s;
        const meta = data.snapshots[item.s];
        const heading = el('header', { class: 'tokens-observation-heading' },
          el('h2', { class: 'md-typescale-title-large' },
            'Observed ',
            el('time', { datetime: meta.detectedAt }, meta.observedLabel),
          ),
          meta.previousLabel ? el('p', { class: 'tokens-metadata' }, `Compared with ${meta.previousLabel}`) : null,
        );
        currentRecords = el('div', { class: 'tokens-observation-records' });
        if (data.browseUrl) {
          currentRecords.append(
            el('p', { class: 'tokens-change-component-link' },
              el('a', { class: 'md-button md-outlined-button', 'data-shape': 'round', 'data-size': 'small', href: data.browseUrl },
                'Browse tokens',
                cloneIcon(container, 'template[data-change-arrow]'),
              ),
            ),
          );
        }
        fragment.append(
          el('section', { class: 'tokens-section tokens-timeline-batch', 'data-change-snapshot': meta.id }, heading, currentRecords),
        );
      }

      const typeMeta = data.types[item.t];
      const summary = el('summary', { class: 'tokens-change-heading' },
        cloneIcon(container, 'template[data-change-chevron]'),
        el('h4', null, item.n),
        el('span', { class: 'tokens-change-status md-typescale-label-large', 'data-change-type': typeMeta.key }, typeMeta.label),
        item.c ? el('span', { class: 'tokens-change-context tokens-change-context-summary' }, item.c) : null,
      );
      const details = el('details', {
        class: 'md-card md-outlined-card tokens-change-card tokens-change-disclosure',
        'data-change-card': '',
        'data-change-snapshot': data.snapshots[item.s].id,
        'data-change-type': typeMeta.key,
        'data-change-search': item.q,
      }, summary);

      if (item.h?.length) {
        const list = el('ol', { class: 'tokens-change-hierarchy', 'aria-label': 'Token hierarchy' });
        for (let i = 0; i < item.h.length; i++) {
          list.append(el('li', null, i > 0 ? cloneIcon(container, 'template[data-change-chevron]') : null, item.h[i]));
        }
        details.append(list);
      }

      const tbody = el('tbody');
      for (const row of item.v || []) {
        tbody.append(el('tr', null, el('th', { scope: 'row' }, row[0]), el('td', null, row[1]), el('td', null, row[2])));
      }
      details.append(
        el('table', { class: 'tokens-change-table' },
          el('thead', null, el('tr', null, el('th', { scope: 'col' }, 'Property'), el('th', { scope: 'col' }, 'Previous value'), el('th', { scope: 'col' }, 'New value'))),
          tbody,
        ),
      );
      currentRecords.append(details);
    }

    if (page > 1 || page < pageCount) {
      const nav = el('nav', { class: 'tokens-pagination', 'aria-label': 'Change log pages' });
      if (page > 1) {
        const prevURL = buildFilterURL(form, filters, page - 1);
        nav.append(el('a', {
          class: 'md-button md-outlined-button', 'data-shape': 'round', 'data-size': 'small',
          href: prevURL.pathname + prevURL.search, 'data-change-page': String(page - 1), 'hx-boost': 'false',
        }, 'Previous'));
      }
      nav.append(el('span', { class: 'md-typescale-body-small' }, `Page ${page}`));
      if (page < pageCount) {
        const nextURL = buildFilterURL(form, filters, page + 1);
        nav.append(el('a', {
          class: 'md-button md-outlined-button', 'data-shape': 'round', 'data-size': 'small',
          href: nextURL.pathname + nextURL.search, 'data-change-page': String(page + 1), 'hx-boost': 'false',
        }, 'Next'));
      }
      fragment.append(nav);
    }

    results.replaceChildren(fragment);
    return page;
  }

  function applyFilters(form, filters, page = 1, historyKind = null) {
    const container = form.closest('#tokens-change-history');
    if (!container) return;
    if (form.dataset.changeFilters === 'global') {
      filterGlobal(container, form, filters);
      if (historyKind) updateHistory(form, filters, historyKind, 1);
      return;
    }
    if (withinMountedDOM(container, form, filters, page)) {
      const seq = (requestSeq.get(container) || 0) + 1;
      requestSeq.set(container, seq);
      filterMountedSubject(container, filters);
      if (historyKind) updateHistory(form, filters, historyKind, 1);
      return;
    }
    const version = form.dataset.changeDataVersion;
    if (!version) return;
    const seq = (requestSeq.get(container) || 0) + 1;
    requestSeq.set(container, seq);
    if (historyKind) updateHistory(form, filters, historyKind, page);
    loadJSON(endpointURL(form, 'data.json', version)).then(data => {
      if (requestSeq.get(container) !== seq || container.isConnected === false || !data) return;
      const actualPage = renderSubjectFromData(container, form, data, filters, page);
      if (historyKind && actualPage !== page) updateHistory(form, filters, 'search', actualPage);
    }, () => {});
  }

  function restore(path) {
    const form = document.querySelector('form[data-change-filters]');
    if (!form || form.dataset.changeFilters === 'server') return false;
    const url = new URL(path, location.href);
    const action = new URL(form.getAttribute('action') || form.action, location.href);
    if (url.pathname !== action.pathname) return false;
    const rawPage = Number.parseInt(url.searchParams.get('page') || '1', 10);
    const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;
    const filters = {
      subject: url.searchParams.get('subject') ?? '',
      snapshot: url.searchParams.get('snapshot') ?? '',
      type: url.searchParams.get('type') ?? '',
      q: (url.searchParams.get('q') ?? '').trim(),
    };
    const container = form.closest('#tokens-change-history');
    if (!container) return false;
    if (!withinMountedDOM(container, form, filters, page) && !form.dataset.changeDataVersion) return false;
    syncForm(form, filters);
    applyFilters(form, filters, page, null);
    return true;
  }

  document.addEventListener('input', event => {
    const target = event.target;
    if (!target?.matches?.('form[data-change-filters] input[name="q"]')) return;
    const form = target.closest('form[data-change-filters]');
    if (!form || form.dataset.changeFilters === 'server') return;
    applyFilters(form, readFilters(form), 1, 'search');
  });

  document.addEventListener('change', event => {
    const target = event.target;
    const form = target?.closest?.('form[data-change-filters]');
    if (!form || form.dataset.changeFilters === 'server') return;
    if (target.matches?.('input[name="q"]')) return;
    applyFilters(form, readFilters(form), 1, 'select');
  });

  document.addEventListener('submit', event => {
    const form = event.target;
    if (!form?.matches?.('form[data-change-filters]') || form.dataset.changeFilters === 'server') return;
    event.preventDefault();
    event.stopImmediatePropagation();
    applyFilters(form, readFilters(form), 1, 'search');
  }, true);

  document.addEventListener('click', event => {
    if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey || event.button) return;
    const reset = event.target?.closest?.('form[data-change-filters] [data-change-reset]');
    if (reset) {
      const form = reset.closest('form[data-change-filters]');
      if (!form || form.dataset.changeFilters === 'server') return;
      event.preventDefault();
      event.stopImmediatePropagation();
      const cleared = { subject: '', snapshot: '', type: '', q: '' };
      syncForm(form, cleared);
      applyFilters(form, cleared, 1, 'select');
      return;
    }
    const pageLink = event.target?.closest?.('#tokens-change-history [data-change-page]');
    if (pageLink) {
      const container = pageLink.closest('#tokens-change-history');
      const form = container?.querySelector('form[data-change-filters]');
      if (!form || !form.dataset.changeDataVersion) return;
      const targetPage = Number.parseInt(pageLink.dataset.changePage || '1', 10);
      if (!Number.isFinite(targetPage) || targetPage < 1) return;
      event.preventDefault();
      event.stopImmediatePropagation();
      applyFilters(form, readFilters(form), targetPage, 'select');
      container.scrollIntoView?.({ block: 'start' });
    }
  }, true);

  function canPrewarm() {
    const connection = typeof navigator !== 'undefined' ? navigator.connection : null;
    return !connection?.saveData && !['slow-2g', '2g'].includes(connection?.effectiveType);
  }

  function warmChangeData(form, includeNonPaginated = false) {
    if (!form || !canPrewarm()) return;
    if (form.dataset.changeFilters === 'global' && form.dataset.changeSearchVersion) {
      loadJSON(endpointURL(form, 'search-index.json', form.dataset.changeSearchVersion)).catch(() => {});
    } else if (form.dataset.changeDataVersion && (includeNonPaginated || form.dataset.changeFilters === 'paginated')) {
      loadJSON(endpointURL(form, 'data.json', form.dataset.changeDataVersion)).catch(() => {});
    }
  }

  document.addEventListener('focusin', event => {
    const target = event.target;
    if (target?.matches?.('form[data-change-filters] input[name="q"]')) {
      warmChangeData(target.closest('form[data-change-filters]'));
      return;
    }
    const pageLink = target?.closest?.('#tokens-change-history [data-change-page]');
    if (pageLink) {
      warmChangeData(pageLink.closest('#tokens-change-history')?.querySelector('form[data-change-filters]'), true);
    }
  });

  document.addEventListener('pointerover', event => {
    if (event.pointerType !== 'mouse' && event.pointerType !== 'pen') return;
    const pageLink = event.target?.closest?.('#tokens-change-history [data-change-page]');
    if (pageLink) {
      warmChangeData(pageLink.closest('#tokens-change-history')?.querySelector('form[data-change-filters]'), true);
    }
  }, { passive: true });

  window.addEventListener('popstate', () => restore(location.href));
  document.addEventListener('htmx:before:history:restore', event => {
    if (restore(event.detail.path)) event.preventDefault();
  });
})();
