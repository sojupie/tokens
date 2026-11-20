(() => {
  function filter(query: string) {
    const list = document.getElementById('subject-list');
    if (!list) return;
    const term = query.trim().toLowerCase();
    let found = false;
    for (const card of list.querySelectorAll<HTMLElement>('[data-subject-name][data-subject-slug]')) {
      const matches = !term ||
        card.dataset.subjectName!.toLowerCase().includes(term) ||
        card.dataset.subjectSlug!.toLowerCase().includes(term);
      card.hidden = !matches;
      found ||= matches;
    }
    const empty = list.querySelector<HTMLElement>('[data-subject-empty]');
    if (empty) empty.hidden = found;
  }

  function update(input: HTMLInputElement) {
    filter(input.value);
    const url = new URL(location.href);
    const query = input.value.trim();
    if (query) url.searchParams.set('q', query);
    else url.searchParams.delete('q');
    if (url.href === location.href) return;

    const state = { ...history.state, htmx: true };
    if (state.subjectSearch) {
      history.replaceState(state, '', url);
    } else {
      // Keep a Back entry for the gallery as it was before this search.
      history.replaceState(state, '', location.href);
      history.pushState({ ...state, subjectSearch: true }, '', url);
    }
  }

  function restore(path: string) {
    const input = document.querySelector<HTMLInputElement>('#subject-search');
    const url = new URL(path, location.href);
    if (!input?.form?.matches('[data-subject-search]') ||
        new URL(input.form.action).pathname !== url.pathname) return false;
    input.value = url.searchParams.get('q') ?? '';
    filter(input.value);
    const clear = input.form.querySelector<HTMLElement>('[data-search-clear-target]');
    if (clear) clear.hidden = !input.value;
    return true;
  }

  document.addEventListener('input', event => {
    const input = event.target;
    if (input instanceof HTMLInputElement && input.matches('[data-subject-search] #subject-search')) update(input);
  });
  document.addEventListener('submit', event => {
    const form = event.target;
    if (!(form instanceof HTMLFormElement) || !form.matches('[data-subject-search]')) return;
    const input = form.querySelector<HTMLInputElement>('#subject-search');
    if (!input) return;
    event.preventDefault();
    update(input);
  });
  window.addEventListener('popstate', () => restore(location.href));
  document.addEventListener('htmx:before:history:restore', event => {
    if (restore((event as CustomEvent<{path: string}>).detail.path)) event.preventDefault();
  });
})();
