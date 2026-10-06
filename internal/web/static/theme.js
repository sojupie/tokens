// Apply saved settings before styles load, without changing token-browser filters.
(() => {
  const root = document.documentElement;
  const themes = ['light', 'dark'];
  const contrasts = ['default', 'medium', 'high'];
  let saved = {};
  try { saved = JSON.parse(localStorage.getItem('tokens-theme')) || {}; } catch {}
  root.dataset.theme = themes.includes(saved.theme) ? saved.theme : 'light';
  root.dataset.contrast = contrasts.includes(saved.contrast) ? saved.contrast : 'default';

  document.addEventListener('DOMContentLoaded', () => {
    const theme = document.querySelector('[data-app-theme]');
    const contrastButtons = document.querySelectorAll('[data-app-contrast]');
    const update = () => {
      theme.setAttribute('aria-pressed', String(root.dataset.theme === 'dark'));
      theme.title = root.dataset.theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode';
      document.querySelectorAll('[data-theme-icon]').forEach(icon => {
        icon.hidden = icon.dataset.themeIcon !== root.dataset.theme;
      });
      contrastButtons.forEach(button => {
        button.setAttribute('aria-pressed', String(button.dataset.appContrast === root.dataset.contrast));
      });
      try {
        localStorage.setItem('tokens-theme', JSON.stringify({theme: root.dataset.theme, contrast: root.dataset.contrast}));
      } catch {}
    };
    theme.disabled = false;
    theme.addEventListener('click', () => {
      root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark';
      update();
    });
    contrastButtons.forEach(button => {
      button.disabled = false;
      button.addEventListener('click', () => {
        root.dataset.contrast = button.dataset.appContrast;
        update();
      });
    });
    update();
  });
})();
