(() => {
const initialized = new WeakSet();
let tableDisplay;
function folderKey(group) {
  const path=[];
  for(let node=group;node;node=node.parentElement.closest('details.tokens-token-folder')) path.unshift(node.querySelector('summary').textContent);
  return JSON.stringify(path);
}
document.addEventListener('htmx:before:swap', event => {
  if (event.detail.ctx.target?.id !== 'tokens-token-browser') return;
  const browser = event.detail.ctx.target;
  tableDisplay = {grid:browser.classList.contains('tokens-grid-view'),hidden:browser.classList.contains('tokens-no-previews'),groups:new Map([...browser.querySelectorAll('details.tokens-token-folder')].map(group=>[folderKey(group),group.open]))};
});

function unbound(selector) {
  return [...document.querySelectorAll(selector)].filter(element => {
    if (initialized.has(element)) return false;
    initialized.add(element);
    return true;
  });
}
function initialize() {
// Cached pages can still contain the old inline no-script fallback rules.
document.querySelectorAll('noscript style').forEach(style => style.remove());
// The backdrop targets the dialog too; its bounds distinguish it from padding.
// Sheets own backdrop gestures and wait for their exit animation.
unbound('dialog:not(.md-sheet)').forEach(dialog => {
  dialog.addEventListener('click', event => {
    if (event.target !== dialog) return;
    const bounds = dialog.getBoundingClientRect();
    if (event.clientX >= bounds.left && event.clientX <= bounds.right &&
        event.clientY >= bounds.top && event.clientY <= bounds.bottom) return;
    if (dialog.dispatchEvent(new Event('cancel', {cancelable: true}))) dialog.close();
  });
});

unbound('.tokens-token-download').forEach(download => {
  const dialog = download.querySelector('dialog');
  if (!dialog) return;
  download.querySelector('[data-download-open]').addEventListener('click', () => dialog.showModal());
  download.querySelector('[data-download-cancel]').addEventListener('click', () => dialog.close());
  dialog.querySelector('form').addEventListener('submit', () => dialog.close());
});

unbound('[data-site-download-open]').forEach(button => {
  button.addEventListener('click', () => {
    const dialog = document.querySelector('#tokens-site-download');
    dialog.querySelector(`input[value="${button.dataset.siteDownloadOpen}"]`).checked = true;
    dialog.showModal();
  });
});

unbound('.tokens-token-browser').forEach(browser => {
  if (tableDisplay) {
    browser.classList.toggle('tokens-grid-view',tableDisplay.grid);
    browser.classList.toggle('tokens-no-previews',tableDisplay.hidden);
    for(const group of browser.querySelectorAll('details.tokens-token-folder')){const open=tableDisplay.groups.get(folderKey(group));if(open!==undefined)group.open=open;}
    tableDisplay = null;
  }
  const lifecycle = new AbortController();
  const signal = lifecycle.signal;
  const form = browser.querySelector('form');
  // Playground swaps update the URL independently of the table. Capture its
  // current settings before htmx (or a native GET) serializes this form.
  form.addEventListener('submit', () => {
    form.querySelectorAll('input[name^="preview."]').forEach(input => input.remove());
    for (const [name, value] of new URLSearchParams(location.search)) {
      if (!name.startsWith('preview.')) continue;
      const input = document.createElement('input');
      input.type = 'hidden'; input.name = name; input.value = value; form.append(input);
    }
  }, {capture:true});
  const dialog = browser.querySelector('.tokens-context-dialog');
  const groups = [...browser.querySelectorAll('.tokens-token-folder')];
  let contextFields = '';
  let contextRequest, changedField;
  // htmx owns validation requests and replacement; these hooks own dialog feedback.
  dialog?.addEventListener('change', event => { changedField = event.target.name; });
  dialog?.addEventListener('htmx:before:request', event => {
    if (event.detail.ctx.sourceElement !== dialog) return;
    contextRequest = event.detail.ctx;
    dialog.querySelector('[type="submit"]').disabled = true;
    dialog.querySelector('[role="status"]').textContent = '';
    dialog.setAttribute('aria-busy', 'true');
  });
  dialog?.addEventListener('htmx:before:swap', event => {
    if (event.detail.ctx !== contextRequest) return;
    const current = dialog.querySelector('.tokens-context-advanced');
    const open = current?.open && current.dataset.expanding !== 'false';
    for (const task of event.detail.tasks) {
      if (task.target?.id !== 'tokens-context-fields') continue;
      const advanced = task.fragment.querySelector('.tokens-context-advanced');
      if (advanced) advanced.open = open;
    }
  });
  dialog?.addEventListener('htmx:after:swap', event => {
    if (event.detail.ctx !== contextRequest || event.detail.ctx.response.status !== 200) return;
    dialog.querySelector(`input[name="${CSS.escape(changedField)}"]:checked`)?.focus({preventScroll:true});
    dialog.querySelector('[type="submit"]').disabled = false;
  });
  dialog?.addEventListener('htmx:finally:request', event => {
    if (event.detail.ctx !== contextRequest) return;
    dialog.removeAttribute('aria-busy');
    if (dialog.querySelector('[type="submit"]').disabled) dialog.querySelector('[role="status"]').textContent = 'Could not check contexts. Change a selection to retry, or cancel.';
  });
  const cancelContext = () => {
    contextRequest = null;
    htmx.trigger(dialog, 'htmx:abort');
    dialog.removeAttribute('aria-busy');
    dialog.querySelector('.tokens-context-fields').innerHTML = contextFields;
    htmx.process(dialog.querySelector('.tokens-context-fields'));
    dialog.querySelector('[role="status"]').textContent = '';
    dialog.querySelector('[type="submit"]').disabled = false;
  };

  const contextButton = browser.querySelector('[data-action="context"]');
  if (dialog) {
    dialog.addEventListener('close', () => {
      contextButton.setAttribute('aria-expanded', 'false');
      cancelContext();
      contextButton.focus({preventScroll:true});
    });
  }

  const menu = browser.querySelector('.tokens-token-set-menu');
  menu.addEventListener('toggle', event => {
    if (event.newState === 'open') menu.querySelector('[aria-current="page"]')?.focus();
  });
  const previewToggle = browser.querySelector('[data-action="previews"]');
  const viewToggle = browser.querySelector('[data-action="view"]');
  const expandToggle = browser.querySelector('[data-action="expand"]');
  const showIcon = (button, alternate) => {
    button.querySelector('[data-icon="on"]').hidden = alternate;
    button.querySelector('[data-icon="off"]').hidden = !alternate;
  };
  const syncControls = () => {
    const grid = browser.classList.contains('tokens-grid-view');
    const hidden = browser.classList.contains('tokens-no-previews');
    previewToggle.hidden = grid;
    previewToggle.setAttribute('aria-pressed', String(!hidden));
    previewToggle.setAttribute('aria-label', hidden ? 'Show previews' : 'Hide previews');
    viewToggle.setAttribute('aria-pressed', String(grid));
    viewToggle.setAttribute('aria-label', grid ? 'Show table view' : 'Show grid view');
    showIcon(previewToggle, hidden);
    showIcon(viewToggle, grid);
    const expanded = groups.length > 0 && groups.every(group => group.open);
    expandToggle.disabled = groups.length === 0;
    expandToggle.setAttribute('aria-expanded', String(expanded));
    expandToggle.setAttribute('aria-label', expanded ? 'Collapse all groups' : 'Expand all groups');
    showIcon(expandToggle, expanded);
    for (const button of [previewToggle, viewToggle, expandToggle]) {
      const tooltip = button.closest('.tokens-table-tooltip');
      if (tooltip) { tooltip.hidden = button.hidden; tooltip.querySelector('[role="tooltip"] p').textContent = button.getAttribute('aria-label'); }
    }
  };
  browser.addEventListener('toggle', event => {
    if (event.target.classList.contains('tokens-token-folder')) syncControls();
  }, true);

  // Native popovers provide Escape, outside-click dismissal and invoker focus.
  // Position beside the active value where space allows, then clamp to the viewport.
  let invoker;
  const prepareReference = popup => {
    if (popup.querySelector('header')) return;
    const row = popup.closest('tr');
    const chrome = browser.querySelector('.tokens-reference-chrome').content.cloneNode(true);
    const title = chrome.querySelector('h3');
    title.id = `${popup.id}-title`;
    title.textContent = row.querySelector('.tokens-token-description').textContent;
    title.title = title.textContent;
    popup.setAttribute('aria-labelledby', title.id);
    popup.removeAttribute('aria-label');
    chrome.querySelector('.tokens-popover-close').setAttribute('popovertarget', popup.id);
    const preview = chrome.querySelector('.tokens-popup-value');
    for (const child of row.querySelector('.tokens-large-preview').children) {
      if (!child.classList.contains('tokens-preview-name')) preview.append(child.cloneNode(true));
    }
    const resolved = chrome.querySelector('.tokens-reference-resolved');
    const compact = row.querySelector('.tokens-compact-value');
    if (!compact.querySelector('.tokens-color-dot')) resolved.querySelector('.tokens-reference-type').remove();
    resolved.append(compact.cloneNode(true));
    const copyButton = chrome.querySelector('[data-copy-reference]');
    for (const node of popup.querySelectorAll('.tokens-reference-node')) {
      const button = copyButton.cloneNode(true);
      const name = node.querySelector('code').textContent;
      button.dataset.copyReference = name;
      button.setAttribute('aria-label', `Copy ${name}`);
      node.append(button);
    }
    popup.prepend(chrome.querySelector('header'), preview);
    popup.querySelector('.tokens-reference-chain').append(resolved);
    popup.append(chrome.querySelector('[role="status"]'));
  };
  const placeReference = popup => {
    if (!invoker || !popup.matches(':popover-open')) return;
    const bounds = invoker.getBoundingClientRect();
    const gap = parseFloat(getComputedStyle(browser).getPropertyValue('--md-sys-measurement-space200'));
    const width = popup.offsetWidth;
    const height = popup.offsetHeight;
    let left = bounds.right + gap;
    if (left + width > innerWidth - gap) left = bounds.left - width - gap;
    if (left < gap) left = Math.max(gap, (innerWidth - width) / 2);
    popup.style.left = `${left}px`;
    popup.style.top = `${Math.max(gap, Math.min(bounds.top, innerHeight - height - gap))}px`;
  };
  browser.querySelectorAll('.tokens-reference-popover').forEach(popup => {
    popup.addEventListener('beforetoggle', event => {
      if (event.newState === 'open') prepareReference(popup);
    });
    popup.addEventListener('toggle', event => {
      if (event.newState === 'open') placeReference(popup);
    });
  });
  const reposition = () => browser.querySelectorAll('.tokens-reference-popover:popover-open').forEach(placeReference);
  addEventListener('resize', reposition, {signal});
  addEventListener('scroll', reposition, {capture: true, signal});
  browser.addEventListener('click', async event => {
    const popupButton = event.target.closest('button[popovertarget]');
    if (popupButton && popupButton.getAttribute('popovertargetaction') !== 'hide') invoker = popupButton;
    const copy = event.target.closest('[data-copy-reference]');
    if (copy) {
      const status = copy.closest('.tokens-reference-popover').querySelector('[role="status"]');
      try {
        await navigator.clipboard.writeText(copy.dataset.copyReference);
        status.textContent = 'Copied.';
      } catch {
        const range = document.createRange();
        range.selectNodeContents(copy.previousElementSibling);
        const selection = getSelection();
        selection.removeAllRanges();
        selection.addRange(range);
        status.textContent = 'Reference selected. Copy it with your keyboard.';
      }
      return;
    }
    const button = event.target.closest('button[data-action]');
    if (!button) return;
    browser.querySelectorAll('.tokens-reference-popover:popover-open').forEach(popup => popup.hidePopover());
    switch (button.dataset.action) {
      case 'context':
        contextFields = dialog.querySelector('.tokens-context-fields').innerHTML;
        dialog.showModal();
        contextButton.setAttribute('aria-expanded', 'true');
        dialog.querySelector('input:checked')?.focus({preventScroll:true});
        break;
      case 'cancel':
        dialog.close();
        contextButton.focus({preventScroll:true});
        break;
      case 'previews':
        browser.classList.toggle('tokens-no-previews');
        break;
      case 'view':
        browser.classList.toggle('tokens-grid-view');
        break;
      case 'expand': {
        const open = button.getAttribute('aria-expanded') !== 'true';
        groups.forEach(group => { group.open = open; });
        break;
      }
    }
    syncControls();
  });
  syncControls();
  document.addEventListener('htmx:beforeCleanupElement', event => {
    if ((event.detail?.elt || event.target) === browser || (event.detail?.elt || event.target).contains(browser)) {
      if (dialog) htmx.trigger(dialog, 'htmx:abort');
      lifecycle.abort();
    }
  }, {signal});
});
}
document.addEventListener('htmx:afterSwap', initialize);
document.addEventListener('htmx:historyRestore', initialize);
initialize();
})();
