// The backdrop targets the dialog too; its bounds distinguish it from padding.
document.querySelectorAll('dialog').forEach(dialog => {
  dialog.addEventListener('click', event => {
    if (event.target !== dialog) return;
    const bounds = dialog.getBoundingClientRect();
    if (event.clientX >= bounds.left && event.clientX <= bounds.right &&
        event.clientY >= bounds.top && event.clientY <= bounds.bottom) return;
    if (dialog.dispatchEvent(new Event('cancel', {cancelable: true}))) dialog.close();
  });
});

document.querySelectorAll('.tokens-token-download').forEach(download => {
  const dialog = download.querySelector('dialog');
  if (!dialog) return;
  download.querySelector('[data-download-open]').addEventListener('click', () => dialog.showModal());
  download.querySelector('[data-download-cancel]').addEventListener('click', () => dialog.close());
  dialog.querySelector('form').addEventListener('submit', () => dialog.close());
});

document.querySelectorAll('[data-site-download-open]').forEach(button => {
  button.addEventListener('click', () => {
    const dialog = document.querySelector('#tokens-site-download');
    dialog.querySelector(`input[value="${button.dataset.siteDownloadOpen}"]`).checked = true;
    dialog.showModal();
  });
});

document.querySelectorAll('.tokens-token-browser').forEach(browser => {
  const form = browser.querySelector('form');
  const dialog = browser.querySelector('dialog');
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
    status.textContent = 'Checking available contexts…';
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
      dialog.querySelector(`input[name="${CSS.escape(changedField)}"]:checked`)?.focus();
      status.textContent = '';
      save.disabled = false;
    } catch (error) {
      if (!request.signal.aborted) status.textContent = 'Could not check contexts. Change a selection to retry, or cancel.';
    }
  });

  const cancelContext = () => {
    contextRequest?.abort();
    dialog.querySelector('.tokens-context-fields').innerHTML = contextFields;
    dialog.querySelector('[role="status"]').textContent = '';
    dialog.querySelector('[type="submit"]').disabled = false;
  };

  const picker = browser.querySelector('.tokens-token-set');
  const menu = browser.querySelector('.tokens-token-set-menu');
  const placeMenu = () => {
    const bounds = picker.getBoundingClientRect();
    const edge = 16;
    const below = innerHeight - bounds.bottom - edge;
    const above = bounds.top - edge;
    const upwards = below < 200 && above > below;
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
  addEventListener('resize', () => { if (menu.matches(':popover-open')) placeMenu(); });
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
    popup.addEventListener('toggle', event => {
      if (event.newState === 'open') placeReference(popup);
    });
  });
  const reposition = () => browser.querySelectorAll('.tokens-reference-popover:popover-open').forEach(placeReference);
  addEventListener('resize', reposition);
  addEventListener('scroll', reposition, true);
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
        break;
      case 'cancel':
        cancelContext();
        dialog.close();
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
  if (dialog) dialog.addEventListener('cancel', cancelContext);
});
