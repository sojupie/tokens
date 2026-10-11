(() => {
  const context = document.modelContext || navigator.modelContext;
  if (!context || typeof context.registerTool !== 'function') return;
  const base = document.currentScript ? new URL('../', document.currentScript.src).pathname.replace(/\/$/, '') : '';
  const lifecycle = new AbortController();
  const options = {signal: lifecycle.signal};
  async function read(path) {
    const response = await fetch(path, {headers: {Accept: 'application/json'}, signal: lifecycle.signal});
    if (!response.ok) throw new Error(`Tokens request failed (${response.status})`);
    return response.json();
  }
  function result(value) {
    return {content: [{type: 'text', text: JSON.stringify(value)}]};
  }
  const kinds = ['component', 'style', 'foundation', 'reference-token-set', 'system-token-set'];
  const tools = [
    {
      name: 'list_subjects',
      description: 'Find Material Design components and styles with exact token download URLs.',
      inputSchema: {type: 'object', properties: {query: {type: 'string', description: 'Optional name or slug filter.'}}, additionalProperties: false},
      execute: async ({query = ''} = {}) => {
        const index = await read(base + '/api/v1/subjects');
        const needle = query.toLowerCase();
        index.subjects = index.subjects.filter(subject => `${subject.name} ${subject.slug}`.toLowerCase().includes(needle));
        return result(index);
      },
    },
    {
      name: 'get_tokens',
      description: 'Read every table, token and supported context for one subject returned by list_subjects. Ignores browser search and filters. Use the HTTP download API for filtered tables.',
      inputSchema: {type: 'object', properties: {kind: {type: 'string', enum: kinds}, slug: {type: 'string', description: 'Exact subject slug.'}}, required: ['kind', 'slug'], additionalProperties: false},
      execute: async ({kind, slug}) => {
        const index = await read(base + '/api/v1/subjects');
        const subject = index.subjects.find(item => item.kind === kind && item.slug === slug);
        if (!subject) throw new Error('Subject not found; use kind and slug from list_subjects.');
        // Use the known local route from the index, never a caller-supplied URL.
        const endpoint = new URL(subject.jsonUrl, location.origin);
        if (!endpoint.pathname.startsWith(base + '/downloads/')) throw new Error('Unexpected token endpoint.');
        return result(await read(endpoint.pathname + endpoint.search));
      },
    },
    {
      name: 'list_changes',
      description: 'Read observed token changes; dates show detection, not upstream release dates.',
      inputSchema: {type: 'object', properties: {subject: {type: 'string', description: 'Optional exact subject slug.'}}, additionalProperties: false},
      execute: async ({subject = ''} = {}) => {
        const data = await read(base + '/api/v1/changes');
        data.changes = data.changes || [];
        if (subject) data.changes = data.changes.filter(change => change.subject === subject);
        return result(data);
      },
    },
  ];
  function register() {
    for (const tool of tools) {
      try { Promise.resolve(context.registerTool(tool, options)).catch(error => console.warn('WebMCP registration failed:', error)); }
      catch (error) { console.warn('WebMCP registration failed:', error); }
    }
  }
  register();
  // Preserve tools on a bfcache round trip; unregister when the document is destroyed.
  window.addEventListener('pagehide', event => { if (!event.persisted) lifecycle.abort(); });
})();
