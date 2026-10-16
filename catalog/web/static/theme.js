// Bind theme controls after parsing; saved settings are applied inline in the head.
(() => {
  const root = document.documentElement;
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
