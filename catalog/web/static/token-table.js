(() => {
const initialized = new WeakSet();
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
  const lifecycle = new AbortController();
  const signal = lifecycle.signal;
  const form = browser.querySelector('form');
  const dialog = browser.querySelector('.tokens-context-dialog');
  const groups = [...browser.querySelectorAll('.tokens-token-folder')];
  let contextFields = '';
  let contextRequest;

  // Ask the server for supported choices; keep one source of context rules.
  if (dialog) dialog.addEventListener('change', async event => {
    const changedField = event.target.name;
    contextRequest?.abort();
    const request = new AbortController();
    contextRequest = request;
    const save = dialog.querySelector('[type="submit"]');
    const status = dialog.querySelector('[role="status"]');
    save.disabled = true;
    status.textContent = '';
    dialog.setAttribute('aria-busy', 'true');
    const url = new URL(form.action);
    url.search = new URLSearchParams(new FormData(form)).toString();
    try {
      const response = await fetch(url, {signal: request.signal, cache: 'no-store'});
      if (!response.ok) throw new Error('Context request failed');
      const page = new DOMParser().parseFromString(await response.text(), 'text/html');
      const fields = page.querySelector('.tokens-context-fields');
      if (!fields) throw new Error('Missing context choices');
      const advancedOpen = dialog.querySelector('.tokens-context-advanced')?.open;
      dialog.querySelector('.tokens-context-fields').replaceWith(fields);
      const advanced = dialog.querySelector('.tokens-context-advanced');
      if (advanced) advanced.open = advancedOpen;
      dialog.querySelector(`input[name="${CSS.escape(changedField)}"]:checked`)?.focus({preventScroll:true});
      status.textContent = '';
      save.disabled = false;
      placeContext();
    } catch (error) {
      if (!request.signal.aborted) status.textContent = 'Could not check contexts. Change a selection to retry, or cancel.';
    } finally {
      if (contextRequest === request) dialog.removeAttribute('aria-busy');
    }
  });

  const cancelContext = () => {
    contextRequest?.abort();
    dialog.removeAttribute('aria-busy');
    dialog.querySelector('.tokens-context-fields').innerHTML = contextFields;
    dialog.querySelector('[role="status"]').textContent = '';
    dialog.querySelector('[type="submit"]').disabled = false;
  };

  const contextButton = browser.querySelector('[data-action="context"]');
  const placeContext = () => {
    if (!dialog?.matches(':popover-open')) return;
    const r = contextButton.getBoundingClientRect(), gap = 8;
    const width = dialog.offsetWidth, height = dialog.offsetHeight;
    let left, top;
    if (r.left - width - gap >= gap) { left = r.left - width - gap; top = r.top; }
    else if (r.right + gap + width <= innerWidth - gap) { left = r.right + gap; top = r.top; }
    else {
      left = r.right - width;
      const below = r.bottom + gap;
      top = below + height <= innerHeight - gap ? below : r.top - height - gap;
    }
    dialog.style.left = `${Math.max(gap, Math.min(left, innerWidth - width - gap))}px`;
    dialog.style.top = `${Math.max(gap, Math.min(top, innerHeight - height - gap))}px`;
  };
  if (dialog) {
    dialog.addEventListener('beforetoggle', event => {
      if (event.newState === 'open') {
        contextFields = dialog.querySelector('.tokens-context-fields').innerHTML;
        // Position before focusing so opening cannot scroll the document.
        dialog.style.visibility = 'hidden';
      } else contextRequest?.abort();
    });
    dialog.addEventListener('toggle', () => {
      contextButton.setAttribute('aria-expanded', String(dialog.matches(':popover-open')));
      placeContext();
      dialog.style.visibility = '';
      if (dialog.matches(':popover-open')) dialog.querySelector('input:checked')?.focus({preventScroll:true});
      else cancelContext();
    });
    dialog.addEventListener('keydown', event => {
      if (event.key !== 'Escape') return;
      event.preventDefault();
      event.stopPropagation();
      dialog.hidePopover();
      contextButton.focus({preventScroll:true});
    });
    dialog.addEventListener('toggle', placeContext, true);
    addEventListener('resize', placeContext, {signal});
    addEventListener('scroll', placeContext, {capture: true, signal});
  }

  const picker = browser.querySelector('.tokens-token-set');
  const menu = browser.querySelector('.tokens-token-set-menu');
  const placeMenu = () => {
    const bounds = picker.getBoundingClientRect();
    const edge = 16;
    const below = innerHeight - bounds.bottom - edge;
    const above = bounds.top - edge;
    const upwards = below < 200 && above > below;
    menu.dataset.menuAbove = String(upwards);
    menu.style.left = `${Math.max(edge, Math.min(bounds.left, innerWidth - menu.offsetWidth - edge))}px`;
    menu.style.maxHeight = `${Math.max(0, upwards ? above : below)}px`;
    menu.style.top = `${upwards ? bounds.top - menu.offsetHeight : bounds.bottom}px`;
  };
  menu.addEventListener('toggle', event => {
    if (event.newState === 'open') {
      placeMenu();
      menu.querySelector('[aria-current="page"]')?.focus();
    }
  });
  menu.addEventListener('keydown', event => {
    if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return;
    const links = [...menu.querySelectorAll('a')];
    let index = links.indexOf(document.activeElement);
    if (event.key === 'Home') index = 0;
    else if (event.key === 'End') index = links.length - 1;
    else index = (index + (event.key === 'ArrowDown' ? 1 : -1) + links.length) % links.length;
    links[index]?.focus();
    event.preventDefault();
  });
  addEventListener('resize', () => { if (menu.matches(':popover-open')) placeMenu(); }, {signal});
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
        // The native popover invoker handles opening, dismissal and focus.
        break;
      case 'cancel':
        dialog.hidePopover();
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
    if (event.detail.elt === browser || event.detail.elt.contains(browser)) {
      contextRequest?.abort();
      lifecycle.abort();
    }
  }, {signal});
});
}
document.addEventListener('htmx:afterSwap', initialize);
document.addEventListener('htmx:historyRestore', initialize);
initialize();
})();
