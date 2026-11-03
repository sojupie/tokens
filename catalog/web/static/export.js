(() => {
  const initialized = new WeakSet(), openers = new WeakMap();
  const inputs = (form, name) => Array.from(form.querySelectorAll('input[name]')).filter(input => input.name === name);
  function set(form, name, next) {
    for (const input of inputs(form, name)) {
      if (input.type === 'radio' || input.type === 'checkbox') input.checked = input.value === next;
      else input.value = next;
    }
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
  document.addEventListener('htmx:config:request', event => {
    const form = event.detail.ctx.sourceElement.closest('#tokens-export-form');
    if (!form) return;
    // A second parent choice can arrive while htmx has disabled its children.
    // Those draft values still belong to the selection, unlike unavailable inputs.
    for (const input of form.querySelectorAll('fieldset:disabled input[name]:not([disabled])')) {
      if (!['checkbox','radio'].includes(input.type) || input.checked)
        event.detail.ctx.request.body.append(input.name, input.value);
    }
  });
  document.addEventListener('htmx:before:request', event => {
    const source = event.detail.ctx.sourceElement;
    if (!source.closest('#tokens-export-form') && source.id !== 'tokens-export-summary') return;
    const panel = source.closest('dialog');
    // This selection is unconfirmed. htmx owns temporary request disabling;
    // only the server's summary response may confirm the Download action.
    panel.querySelector('[data-export-confirmation]').disabled = true;
    panel.querySelector('[data-export-error]').hidden = true;
  });
  document.addEventListener('htmx:before:swap', event => {
    const ctx = event.detail.ctx;
    if (!ctx.sourceElement.closest('#tokens-export-form') && ctx.sourceElement.id !== 'tokens-export-summary') return;
    const panel = ctx.sourceElement.closest('dialog');
    // Apply local disclosure state before insertion: after:swap runs after
    // settling, so restoring it there briefly paints a collapsed modal.
    const disclosures = Array.from(panel.querySelectorAll('.tokens-export-disclosure'), detail => detail.open && detail.dataset.expanding !== 'false');
    for (const task of event.detail.tasks) {
      if (task.target?.id !== 'tokens-export-fields') continue;
      task.fragment.querySelectorAll('.tokens-export-disclosure').forEach((detail, i) => { detail.open = disclosures[i]; });
    }
  });
  function requestFailed(event) {
    const source = event.detail.ctx.sourceElement;
    if (!source.closest('#tokens-export-form') && source.id !== 'tokens-export-summary') return;
    const panel = source.closest('dialog');
    const error = panel.querySelector('[data-export-error]');
    error.textContent = 'Could not check selection. Change a selection to retry, or close the dialog.';
    error.hidden = false;
  }
  document.addEventListener('htmx:error', requestFailed);
  document.addEventListener('htmx:response:error', requestFailed);
  function initialize() {
    const panel = document.getElementById('tokens-site-download'); if (!panel || initialized.has(panel)) return;
    initialized.add(panel);
    const guard = panel.querySelector('#tokens-export-submit-pending'), form = panel.querySelector('form');
    let downloadToken = '', downloadTimer;
    const clearCookie = token => { document.cookie = `tokens-download-${token}=; Max-Age=0; Path=/; SameSite=Strict`; };
    function finishDownload(error) {
      clearInterval(downloadTimer);
      if (downloadToken) clearCookie(downloadToken);
      downloadToken = '';
      form.querySelector('input[name="_download"]').value = '';
      reflectLoading();
      if (error) {
        const notice = panel.querySelector('[data-export-error]');
        notice.textContent = error; notice.hidden = false;
      }
    }
    const reflectLoading = () => {
      const button = panel.querySelector('[data-export-submit]');
      button.disabled = !!downloadToken;
      button.setAttribute('aria-busy', String(!!downloadToken || guard.classList.contains('htmx-request')));
    };
    // The component owns the animation. htmx tracks option requests; the response
    // cookie tracks preparation of the native attachment.
    new MutationObserver(reflectLoading).observe(guard, {attributes:true, attributeFilter:['class']});
    reflectLoading();
    form.addEventListener('submit', event => {
      if (downloadToken) { event.preventDefault(); return; }
      downloadToken = Date.now().toString(36) + '-' + Math.random().toString(36).slice(2);
      form.querySelector('input[name="_download"]').value = downloadToken;
      form.target = 'tokens-export-transfer';
      panel.querySelector('[data-export-error]').hidden = true;
      reflectLoading();
      const deadline = Date.now() + 300000;
      downloadTimer = setInterval(() => {
        const cookie = document.cookie.split('; ').find(value => value.startsWith(`tokens-download-${downloadToken}=`));
        if (cookie) finishDownload(cookie.endsWith('=ready') ? '' : 'Could not prepare download. Try again.');
        else if (Date.now() > deadline) finishDownload('Download preparation has not responded. Try again.');
      }, 150);
    });
    panel.addEventListener('close', () => {
      finishDownload('');
      htmx.trigger(panel.querySelector('form'), 'htmx:abort');
      openers.get(panel)?.focus({preventScroll:true});
    });
  }
  document.addEventListener('htmx:afterSwap', initialize);
  document.addEventListener('htmx:historyRestore', initialize);
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', initialize); else initialize();
})();
