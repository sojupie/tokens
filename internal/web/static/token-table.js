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

document.querySelectorAll('.sl-token-download').forEach(download => {
  const dialog = download.querySelector('dialog');
  download.querySelector('[data-download-open]').addEventListener('click', () => dialog.showModal());
  download.querySelector('[data-download-cancel]').addEventListener('click', () => dialog.close());
  dialog.querySelector('form').addEventListener('submit', () => dialog.close());
});

document.querySelectorAll('.sl-token-browser').forEach(browser => {
  const form = browser.querySelector('form');
  const dialog = browser.querySelector('dialog');
  const groups = [...browser.querySelectorAll('.sl-token-folder')];
  let openGroups = [];
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
      const fields = page.querySelector('.sl-context-fields');
      if (!fields) throw new Error('Missing context choices');
      const advancedOpen = dialog.querySelector('.sl-context-advanced')?.open;
      dialog.querySelector('.sl-context-fields').replaceWith(fields);
      const advanced = dialog.querySelector('.sl-context-advanced');
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
    dialog.querySelector('.sl-context-fields').innerHTML = contextFields;
    dialog.querySelector('[role="status"]').textContent = '';
    dialog.querySelector('[type="submit"]').disabled = false;
  };

  const picker = browser.querySelector('.sl-token-set');
  const menu = browser.querySelector('.sl-token-set-menu');
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
  browser.addEventListener('click', event => {
    const button = event.target.closest('button[data-action]');
    if (!button) return;
    switch (button.dataset.action) {
      case 'context':
        contextFields = dialog.querySelector('.sl-context-fields').innerHTML;
        dialog.showModal();
        break;
      case 'cancel':
        cancelContext();
        dialog.close();
        break;
      case 'previews': {
        const hidden = browser.classList.toggle('sl-no-previews');
        button.setAttribute('aria-pressed', String(!hidden));
        break;
      }
      case 'view': {
        const flat = browser.classList.toggle('sl-flat-view');
        button.setAttribute('aria-pressed', String(flat));
        if (flat) openGroups = groups.map(group => group.open);
        groups.forEach((group, index) => { group.open = flat || openGroups[index]; });
        break;
      }
      case 'expand': {
        const open = button.getAttribute('aria-expanded') !== 'true';
        groups.forEach(group => { group.open = open; });
        browser.classList.remove('sl-flat-view');
        browser.querySelector('[data-action="view"]').setAttribute('aria-pressed', 'false');
        button.setAttribute('aria-expanded', String(open));
        button.setAttribute('aria-label', open ? 'Collapse all groups' : 'Expand all groups');
        break;
      }
    }
  });
  if (dialog) dialog.addEventListener('cancel', cancelContext);
});
