// Static hosts use the Catalog's presentation controllers with local filtering.
(() => {
  for (const table of document.querySelectorAll('.tokens-static-table')) {
    const form = table.querySelector('form');
    const search = form.elements.q;
    const picker = table.querySelector('.tokens-token-set');
    const options = [...table.querySelectorAll('[data-static-set-option]')];
    const sets = [...table.querySelectorAll('.tokens-token-groups > details')];
    const groups = [...table.querySelectorAll('details.tokens-token-folder')];
    const rows = [...table.querySelectorAll('tbody tr')].map(row => ({
      row,
      text: row.textContent.toLocaleLowerCase(),
      groups: groups.filter(group => group.contains(row))
    }));
    const empty = table.querySelector('.tokens-empty');
    let selected = 'all';
    let beforeSearch;

    function refresh() {
      const query = search.value.trim().toLocaleLowerCase();
      const terms = query.split(/\s+/u);
      const matchingGroups = new Set();
      let matched = false;
      if (query && !beforeSearch) {
        beforeSearch = new Map(groups.map(group => [group, group.open]));
      }
      for (const { row, text, groups: ancestors } of rows) {
        const matches = terms.every(term => text.includes(term));
        row.hidden = !matches;
        if (matches && (selected === 'all' || ancestors[0] === sets[Number(selected)])) {
          matched = true;
          for (const group of ancestors) matchingGroups.add(group);
        }
      }
      for (const group of groups) {
        group.hidden = !matchingGroups.has(group);
        if (query) group.open = matchingGroups.has(group);
        else if (beforeSearch) group.open = beforeSearch.get(group);
      }
      if (!query) beforeSearch = undefined;
      empty.hidden = matched;
    }

    for (const option of options) {
      option.addEventListener('click', event => {
        event.preventDefault();
        selected = option.dataset.staticSetOption;
        const label = option.querySelector('.md-menu-item-text').textContent.trim();
        picker.querySelector('.tokens-context-label').textContent = label;
        picker.setAttribute('aria-label', `${picker.dataset.setLabel}: ${label}`);
        for (const item of options) {
          item.setAttribute('aria-current', item === option ? 'page' : 'false');
        }
        refresh();
        if (selected !== 'all' && !search.value.trim()) {
          sets[Number(selected)].open = true;
        }
      });
    }
    form.addEventListener('submit', event => event.preventDefault());
    search.addEventListener('input', refresh);
    table.querySelector('.tokens-token-toolbar').hidden = false;
    refresh();
  }
})();
