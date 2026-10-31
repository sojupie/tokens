(() => {
  const initialized = new WeakSet(), openers = new WeakMap(), pending = new WeakMap();
  const inputs = (form, name) => Array.from(form.querySelectorAll('input[name]')).filter(input => input.name === name);
  const value = (form, name) => new FormData(form).get(name) || '';
  function set(form, name, next) {
    for (const input of inputs(form, name)) {
      if (input.type === 'radio' || input.type === 'checkbox') input.checked = input.value === next;
      else input.value = next;
    }
  }
  function disableOption(form, name, option, disabled) {
    for (const input of inputs(form, name)) if (input.value === option) input.disabled = disabled;
  }
  function contextSummary(panel, pending) {
    const scope = inputs(panel.querySelector('form'), 'scope').find(input => input.checked);
    const scopeSummary = panel.querySelector('[data-export-scope-summary]');
    if (scope && scopeSummary) scopeSummary.textContent = scope.closest('label').textContent.trim();
    if (pending) return;
    const summary = panel.querySelector('[data-export-context-summary]');
    const groups = Array.from(panel.querySelectorAll('[data-export-context]:not([hidden])'));
    summary.textContent = groups.map(group => {
      const selected = Array.from(group.querySelectorAll('input')).filter(input => input.checked && !input.disabled).map(input => input.value);
      return selected.length ? group.dataset.exportContext + ': ' + selected.join(', ') : '';
    }).filter(Boolean).join(' · ');
  }
  function busy(panel, active) {
    contextSummary(panel, active);
    panel.dataset.exportBusy = String(active);
    panel.querySelector('form').setAttribute('aria-busy', String(active));
    panel.querySelector('[data-export-submit]').setAttribute('aria-busy', String(active));
  }
  async function update(panel) {
    const form = panel.querySelector('form'), implementation = value(form, 'content') === 'implementation';
    if (implementation && (!['subject', 'collection'].includes(value(form, 'scope')) || panel.dataset.componentExport !== 'true')) {
      set(form, 'scope', panel.dataset.componentExport === 'true' ? 'subject' : 'collection');
    }
    for (const scope of ['view', 'set']) disableOption(form, 'scope', scope, implementation);
    disableOption(form, 'scope', 'subject', implementation && panel.dataset.componentExport !== 'true');
    disableOption(form, 'content', 'implementation', panel.dataset.hasImplementation !== 'true');
    panel.querySelector('[data-export-token-formats]').hidden = implementation;
    for (const input of inputs(form, 'format')) input.disabled = implementation || input.value === 'css' && panel.dataset.hasCss !== 'true';
    panel.querySelector('[data-export-description]').textContent = implementation ? 'ZIP archive with component CSS, example HTML, and required JavaScript, including nested components.' : 'Source JSON retains captured values and reference metadata.';
    for (const node of panel.querySelectorAll('[data-export-context]')) {
      const unused = implementation && !['Theme', 'Contrast'].includes(node.dataset.exportContext);
      if (implementation) node.hidden = unused;
      for (const input of node.querySelectorAll('input')) input.disabled = unused;
    }
    const submit = panel.querySelector('[data-export-submit]'), error = panel.querySelector('[data-export-error]'), counts = panel.querySelector('[data-export-counts]');
    pending.get(panel)?.abort();
    const controller = new AbortController(); pending.set(panel, controller);
    submit.disabled = true; error.hidden = true; counts.textContent = 'Checking selection…';
    busy(panel, true);
    try {
      if (!implementation) {
        // Resolve choices against captured combinations, not platform-name guesses.
        const optionsURL = new URL(form.action);
        optionsURL.search = new URLSearchParams(new FormData(form)).toString();
        optionsURL.searchParams.set('options', '1');
        const response = await fetch(optionsURL, {signal: controller.signal});
        if (!response.ok) throw Error(await response.text());
        const choices = await response.json();
        if (controller.signal.aborted) return;
        for (const node of panel.querySelectorAll('[data-export-context]')) {
          const allowed = choices[node.dataset.exportContext] || [];
          node.hidden = allowed.length === 0;
          const options = Array.from(node.querySelectorAll('input'));
          for (const input of options) input.disabled = !allowed.includes(input.value);
          if (allowed.length && options.some(input => input.checked) && !options.some(input => input.checked && !input.disabled)) {
            const first = options.find(input => !input.disabled);
            set(form, first.name, first.value);
          }
        }
      }
      const params = new URLSearchParams(new FormData(form));
      if (implementation) params.set('format', 'zip');
      const missing = !params.has('format') || Array.from(panel.querySelectorAll('[data-export-context]:not([hidden])')).some(node => !params.has('context.' + node.dataset.exportContext));
      if (missing) { counts.textContent = 'Select at least one file format and one value for each context.'; return; }
      const endpoint = new URL(form.action); endpoint.search = params.toString();
      const report = panel.querySelector('[data-export-report]'); report.hidden = implementation;
      const reportURL = new URL(endpoint); reportURL.searchParams.set('format', 'report'); report.href = reportURL.href;
      endpoint.searchParams.set('summary', '1');
      const response = await fetch(endpoint, {signal: controller.signal});
      if (!response.ok) throw Error(await response.text());
      const summary = await response.json();
      if (controller.signal.aborted) return;
      counts.textContent = summary.combinations ? `${summary.combinations} format/context combinations · ${summary.files} files in one ZIP.` : implementation ? `${summary.files} files in the implementation archive.` : `${summary.selected} selected tokens · ${summary.exported} exported · ${summary.dependencies} required references · ${summary.omitted} omitted`;
      submit.disabled = false;
      submit.querySelector('[data-button-label]').textContent = implementation || summary.combinations ? 'Download ZIP' : 'Download';
    } catch (reason) {
      if (reason.name === 'AbortError') return;
      counts.textContent = ''; error.textContent = reason.message; error.hidden = false;
    } finally {
      if (pending.get(panel) === controller) busy(panel, false);
    }
  }

  function sync(panel, trigger) {
    const form = panel.querySelector('form'), table = document.querySelector('.tokens-token-filters');
    if (['tokens', 'implementation'].includes(trigger.dataset.exportContent)) set(form, 'content', trigger.dataset.exportContent);
    for (const detail of panel.querySelectorAll('.tokens-export-disclosure')) detail.open = trigger.dataset.exportContent === 'custom';
    const params = new URLSearchParams(location.search);
    if (table) for (const [key, val] of new FormData(table)) params.set(key, val);
    for (const [key, val] of params) if (key === 'set' || key === 'q' || key.startsWith('context.')) set(form, key, val);
    set(form, 'scope', trigger.dataset.exportScope || panel.dataset.defaultScope || value(form, 'scope'));
  }
  document.addEventListener('click', event => {
    const panel = document.getElementById('tokens-site-download'); if (!panel) return;
    const trigger = event.target.closest?.('[data-export-open]');
    if (trigger) { const menu = trigger.closest?.('.md-split-menu'); const opener = menu ? menu.closest('.md-split-demo').querySelector('.md-split-trailing') : trigger; menu?.hidePopover(); sync(panel, trigger); openers.set(panel, opener); panel.showModal(); update(panel); panel.querySelector('input:checked:not(:disabled)')?.focus({preventScroll: true}); }
    if (event.target.closest?.('[data-export-cancel]')) panel.close();
  });
  function initialize() {
    const panel = document.getElementById('tokens-site-download'); if (!panel || initialized.has(panel)) return;
    initialized.add(panel);
    panel.querySelector('form').addEventListener('submit', event => {
      // Implementation archives have a fixed format; token formats use the checked inputs.
      const form = event.currentTarget;
      form.querySelector('[data-export-zip]')?.remove();
      if (value(form, 'content') === 'implementation') {
        const input = document.createElement('input'); input.type = 'hidden'; input.name = 'format'; input.value = 'zip'; input.dataset.exportZip = ''; form.append(input);
      }
    });
    panel.addEventListener('change', () => { panel.querySelector('[data-export-zip]')?.remove(); if (panel.open) update(panel); });
    panel.addEventListener('close', () => { pending.get(panel)?.abort(); busy(panel, false); openers.get(panel)?.focus({preventScroll: true}); });
  }
  document.addEventListener('htmx:afterSwap', initialize);
  document.addEventListener('htmx:historyRestore', initialize);
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', initialize); else initialize();
})();
