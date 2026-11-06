(() => {
  if (!navigator.clipboard?.writeText) return;
  let timer, copiedStatus;
  function clearCopied() {
    clearTimeout(timer);
    if (copiedStatus) copiedStatus.textContent = '';
    copiedStatus = undefined;
  }
  document.addEventListener('pointerdown', clearCopied, true);
  document.addEventListener('keydown', clearCopied, true);
  for (const button of document.querySelectorAll('[data-copy-target]')) {
    const field = document.getElementById(button.dataset.copyTarget);
    const status = button.closest('section').querySelector('[data-copy-status]');
    if (!field || !status) continue;
    button.hidden = false;
    button.addEventListener('click', async () => {
      clearCopied();
      try {
        await navigator.clipboard.writeText(field.textContent);
        status.textContent = 'Copied.';
        copiedStatus = status;
        timer = setTimeout(clearCopied, 1500);
      } catch {
        field.focus();
        window.getSelection().selectAllChildren(field);
        status.textContent = 'Copy failed. Copy the selected text.';
      }
    });
  }
})();
