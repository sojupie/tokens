(() => {
  if (!navigator.clipboard?.writeText) return;
  for (const button of document.querySelectorAll('[data-copy-target]')) {
    const field = document.getElementById(button.dataset.copyTarget);
    const status = button.closest('section').querySelector('[data-copy-status]');
    if (!field || !status) continue;
    button.hidden = false;
    button.addEventListener('click', async () => {
      try {
        await navigator.clipboard.writeText(field.textContent);
        status.textContent = 'Copied.';
      } catch {
        field.focus();
        window.getSelection().selectAllChildren(field);
        status.textContent = 'Copy failed. Copy the selected text.';
      }
    });
  }
})();
