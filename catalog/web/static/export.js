(() => {
  const initialized = new WeakSet(), openers = new WeakMap(), pending = new WeakMap();
  const inputs = (form, name) => Array.from(form.querySelectorAll('input[name]')).filter(input => input.name === name);
  function set(form, name, next) {
    for (const input of inputs(form, name)) {
      if (input.type === 'radio' || input.type === 'checkbox') input.checked = input.value === next;
      else input.value = next;
    }
  }
  function busy(panel, active) {
    panel.dataset.exportBusy = String(active);
    panel.querySelector('form').setAttribute('aria-busy', String(active));
    const button = panel.querySelector('[data-export-submit]');
    button.setAttribute('aria-busy', String(active));
    if (active) button.disabled = true;
  }
  function sync(panel, trigger) {
    const form = panel.querySelector('form'), table = document.querySelector('.tokens-token-filters');
    if (['tokens', 'implementation'].includes(trigger.dataset.exportContent)) set(form, 'content', trigger.dataset.exportContent);
    for (const detail of panel.querySelectorAll('.tokens-export-disclosure')) detail.open = trigger.dataset.exportContent === 'custom';
    const params = new URLSearchParams(location.search);
    if (table) for (const [key, val] of new FormData(table)) params.set(key, val);
    for (const [key, val] of params) if (key === 'set' || key === 'q' || key.startsWith('context.')) {
      set(form, key, val); if (key.startsWith('context.')) set(form, '_' + key, val);
    }
    set(form, 'scope', trigger.dataset.exportScope || panel.dataset.defaultScope);
  }
  document.addEventListener('click', event => {
    const panel = document.getElementById('tokens-site-download'); if (!panel) return;
    const trigger = event.target.closest?.('[data-export-open]');
    if (trigger) {
      const menu = trigger.closest?.('.md-split-menu');
      openers.set(panel, menu ? menu.closest('.md-split-demo').querySelector('.md-split-trailing') : trigger);
      menu?.hidePopover(); sync(panel, trigger); panel.showModal();
      htmx.trigger(panel.querySelector('form'), 'export-refresh');
    }
    if (event.target.closest?.('[data-export-cancel]')) panel.close();
  });
  document.addEventListener('htmx:before:request', event => {
    const ctx = event.detail.ctx;
    if (ctx.sourceElement.id !== 'tokens-export-form') return;
    const panel = ctx.sourceElement.closest('dialog');
    pending.set(panel, ctx); busy(panel, true);
    panel.querySelector('[data-export-counts]').textContent = 'Checking selection…';
    panel.querySelector('[data-export-error]').hidden = true;
  });
  document.addEventListener('htmx:before:swap', event => {
    const ctx = event.detail.ctx;
    if (ctx.sourceElement.id !== 'tokens-export-form') return;
    const panel = ctx.sourceElement.closest('dialog');
    ctx.exportDisclosures = Array.from(panel.querySelectorAll('.tokens-export-disclosure'), detail => detail.open);
    const input = document.activeElement;
    if (panel.contains(input) && input.matches('input')) ctx.exportFocus = {name:input.name,value:input.value};
  });
  document.addEventListener('htmx:after:swap', event => {
    const ctx = event.detail.ctx;
    if (ctx.sourceElement.id !== 'tokens-export-form') return;
    const panel = ctx.sourceElement.closest('dialog');
    if (pending.get(panel) !== ctx || ctx.response.status !== 200) return;
    panel.querySelectorAll('.tokens-export-disclosure').forEach((detail, i) => { detail.open = ctx.exportDisclosures[i]; });
    const fields = panel.querySelector('#tokens-export-fields'), button = panel.querySelector('[data-export-submit]');
    button.disabled = fields.dataset.exportReady !== 'true';
    button.querySelector('[data-button-label]').textContent = fields.dataset.exportLabel;
    const focus = ctx.exportFocus;
    const input = focus && inputs(ctx.sourceElement, focus.name).find(input => input.value === focus.value && !input.disabled);
    (input || panel.querySelector('input:checked:not(:disabled)'))?.focus({preventScroll:true});
  });
  document.addEventListener('htmx:finally:request', event => {
    const ctx = event.detail.ctx;
    if (ctx.sourceElement.id !== 'tokens-export-form') return;
    const panel = ctx.sourceElement.closest('dialog');
    if (pending.get(panel) !== ctx) return;
    busy(panel, false);
    if (ctx.status !== 'swapped' || ctx.response?.status !== 200) {
      panel.querySelector('[data-export-counts]').textContent = '';
      const error = panel.querySelector('[data-export-error]');
      error.textContent = 'Could not check selection. Change a selection to retry, or close the dialog.'; error.hidden = false;
    }
  });
  function initialize() {
    const panel = document.getElementById('tokens-site-download'); if (!panel || initialized.has(panel)) return;
    initialized.add(panel);
    panel.addEventListener('close', () => {
      pending.delete(panel); htmx.trigger(panel.querySelector('form'), 'htmx:abort');
      busy(panel, false); openers.get(panel)?.focus({preventScroll:true});
    });
  }
  document.addEventListener('htmx:afterSwap', initialize);
  document.addEventListener('htmx:historyRestore', initialize);
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', initialize); else initialize();
})();
