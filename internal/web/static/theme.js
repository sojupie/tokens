// Apply saved settings before styles load, without changing token-browser filters.
(() => {
  const root = document.documentElement;
  const themes = ['light', 'dark'];
  const contrasts = ['default', 'medium', 'high'];
  let saved = {};
  try { saved = JSON.parse(localStorage.getItem('source-loom-theme')) || {}; } catch {}
  root.dataset.theme = themes.includes(saved.theme) ? saved.theme : 'light';
  root.dataset.contrast = contrasts.includes(saved.contrast) ? saved.contrast : 'default';

  document.addEventListener('DOMContentLoaded', () => {
    const theme = document.querySelector('[data-app-theme]');
    const contrast = document.querySelector('[data-app-contrast]');
    theme.value = root.dataset.theme;
    contrast.value = root.dataset.contrast;
    theme.disabled = contrast.disabled = false;
    const update = () => {
      root.dataset.theme = theme.value;
      root.dataset.contrast = contrast.value;
      try {
        localStorage.setItem('source-loom-theme', JSON.stringify({theme: theme.value, contrast: contrast.value}));
      } catch {}
    };
    theme.addEventListener('change', update);
    contrast.addEventListener('change', update);
  });
})();
