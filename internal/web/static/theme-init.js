// Apply saved settings before styles load, without changing token-browser filters.
(() => {
  const root = document.documentElement;
  const themes = ['light', 'dark'];
  const contrasts = ['default', 'medium', 'high'];
  let saved = {};
  try { saved = JSON.parse(localStorage.getItem('tokens-theme')) || {}; } catch {}
  root.dataset.theme = themes.includes(saved.theme) ? saved.theme : 'light';
  root.dataset.contrast = contrasts.includes(saved.contrast) ? saved.contrast : 'default';

})();
