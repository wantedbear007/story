const ResourcesPage = {
  entryId: null,

  async render(params) {
    this.entryId = (params && params.id) || null;
    return `
      <div class="section-title">
        <span>${this.entryId ? 'Entry Details' : 'Resources'}</span>
        <button class="btn btn-secondary btn-sm" onclick="App.navigate('drafts')">← Back</button>
      </div>
      <div id="resource-content"><div class="empty-state">Loading...</div></div>
    `;
  },

  async afterRender() {
    if (this.entryId) {
      await this.loadEntry();
    } else {
      await this.loadEntries();
    }
  },

  async loadEntry() {
    const el = document.getElementById('resource-content');
    if (!el) return;

    try {
      const entry = await API.getEntry(this.entryId);
      el.innerHTML = `
        <div class="tweet-detail">
          <h2>${escHtml(entry.title)}</h2>
          <div class="tweet-detail-meta" style="margin-bottom:12px">
            <span>${entry.type}</span>
            <span>${fmtDate(entry.created_at)}</span>
            ${entry.tags && entry.tags.length ? `<span>Tags: ${entry.tags.join(', ')}</span>` : ''}
          </div>
          <div style="white-space:pre-wrap;margin-bottom:16px;font-size:0.9em">${escHtml(entry.content)}</div>
          ${entry.resources && entry.resources.length ? `
            <h3 style="margin-bottom:8px;color:#8b949e">Attached Resources (${entry.resources.length})</h3>
            <div class="resource-grid">
              ${entry.resources.map(r => `
                <div class="resource-card">
                  <div class="type-badge">${r.type}</div>
                  <h3>${escHtml(r.title)}</h3>
                  <div class="url"><a href="${escHtml(r.url)}" target="_blank">${escHtml(r.url)}</a></div>
                  ${r.description ? `<div class="desc">${escHtml(r.description)}</div>` : ''}
                </div>
              `).join('')}
            </div>
          ` : '<p style="color:#8b949e;font-size:0.9em">No resources attached.</p>'}
        </div>
      `;
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  async loadEntries() {
    const el = document.getElementById('resource-content');
    if (!el) return;

    try {
      const data = await API.listEntries({ page_size: 20 });
      if (!data.entries || data.entries.length === 0) {
        el.innerHTML = '<div class="empty-state"><h2>No entries yet</h2></div>';
        return;
      }
      el.innerHTML = `
        <input type="text" class="filter-input" placeholder="Search entries..." oninput="ResourcesPage.search(this.value)" style="margin-bottom:16px">
        <div class="resource-grid">
          ${data.entries.map(e => `
            <div class="resource-card" onclick="App.navigate('resources', null, '${e.id}')">
              <div class="type-badge">${e.type}</div>
              <h3>${escHtml(e.title)}</h3>
              ${e.tags && e.tags.length ? `<div class="desc">Tags: ${e.tags.join(', ')}</div>` : ''}
              <div class="desc" style="margin-top:4px">${e.resources ? e.resources.length : 0} resources</div>
            </div>
          `).join('')}
        </div>
      `;
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  async search(q) {
    if (q.length < 2) return;
    const el = document.getElementById('resource-content');
    if (!el) return;
    try {
      const data = await API.listEntries({ q, page_size: 20 });
      if (!data.entries || data.entries.length === 0) {
        el.innerHTML = '<p>No results</p>';
        return;
      }
      el.innerHTML = `
        <input type="text" class="filter-input" placeholder="Search entries..." oninput="ResourcesPage.search(this.value)" style="margin-bottom:16px" value="${escHtml(q)}">
        <div class="resource-grid">
          ${data.entries.map(e => `
            <div class="resource-card" onclick="App.navigate('resources', null, '${e.id}')">
              <div class="type-badge">${e.type}</div>
              <h3>${escHtml(e.title)}</h3>
              ${e.tags && e.tags.length ? `<div class="desc">Tags: ${e.tags.join(', ')}</div>` : ''}
            </div>
          `).join('')}
        </div>
      `;
    } catch { /* ignore */ }
  },
};
