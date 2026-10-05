document.querySelectorAll('.sl-token-browser').forEach(browser => {
  const form = browser.querySelector('form');
  const dialog = browser.querySelector('dialog');
  const groups = [...browser.querySelectorAll('.sl-token-folder')];
  let openGroups = [];
  let contextValues = [];

  form.querySelector('[name="set"]').addEventListener('change', () => form.requestSubmit());
  browser.addEventListener('click', event => {
    const button = event.target.closest('button[data-action]');
    if (!button) return;
    switch (button.dataset.action) {
      case 'context':
        contextValues = [...dialog.querySelectorAll('input')].map(input => input.checked);
        dialog.showModal();
        break;
      case 'cancel':
        dialog.querySelectorAll('input').forEach((input, index) => { input.checked = contextValues[index]; });
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
  if (dialog) dialog.addEventListener('cancel', () => {
    dialog.querySelectorAll('input').forEach((input, index) => { input.checked = contextValues[index]; });
  });
});
