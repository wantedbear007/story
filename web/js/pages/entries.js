const EntriesPage = {
  entryId: null,

  async render(params) {
    this.entryId = (params && params.id) || null;
    return `<div id="entry-content"><div class="empty-state">Loading...</div></div>`;
  },

  async afterRender() {
    if (this.entryId) {
      await this.loadEntry();
    } else {
      await this.loadEntries();
    }
  },

  async loadEntry() {
    const el = document.getElementById('entry-content');
    if (!el) return;

    try {
      const entry = await API.getEntry(this.entryId);
      el.innerHTML = `
        <div class="tweet-detail">
          <div class="tweet-detail-header">
            <div>
              <span class="status-badge approved" style="margin-bottom:8px;display:inline-block">${entry.type}</span>
              <h2 style="font-size:1.3em">${escHtml(entry.title)}</h2>
              <div class="tweet-detail-meta" style="margin-top:8px">
                <span>${fmtDate(entry.created_at)}</span>
                ${entry.tags && entry.tags.length ? `<span>Tags: ${entry.tags.map(t => escHtml(t)).join(', ')}</span>` : ''}
              </div>
            </div>
            <button class="btn btn-secondary btn-sm" onclick="App.navigate('entries')">← Back</button>
          </div>
          <div style="white-space:pre-wrap;margin-bottom:20px;font-size:0.9em;line-height:1.7;color:#c9d1d9">${escHtml(entry.content)}</div>
          ${entry.resources && entry.resources.length ? `
            <h3 style="margin-bottom:12px;font-size:0.9em;color:#8b949e;font-weight:600">Attached Resources (${entry.resources.length})</h3>
            <div class="resource-grid">
              ${entry.resources.map(r => `
                <div class="resource-card" onclick="event.stopPropagation()">
                  <div class="type-badge">${r.type}</div>
                  <h3>${escHtml(r.title)}</h3>
                  <div class="url"><a href="${escHtml(r.url)}" target="_blank">${escHtml(r.url)}</a></div>
                  ${r.description ? `<div class="desc">${escHtml(r.description)}</div>` : ''}
                </div>
              `).join('')}
            </div>
          ` : '<p style="color:#484f58;font-size:0.9em">No resources attached.</p>'}
        </div>
      `;
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  async loadEntries() {
    const el = document.getElementById('entry-content');
    if (!el) return;

    try {
      const data = await API.listEntries({ page_size: 20 });
      if (!data.entries || data.entries.length === 0) {
        el.innerHTML = '<div class="empty-state"><h2>No entries yet</h2><p>Create entries using the CLI or capture page.</p></div>';
        return;
      }
      el.innerHTML = `
        <input type="text" class="filter-input" placeholder="Search entries..." oninput="EntriesPage.search(this.value)" style="max-width:100%;margin-bottom:20px">
        <div class="resource-grid">
          ${data.entries.map(e => `
            <div class="resource-card" onclick="App.navigate('entries', '${e.id}')">
              <div class="type-badge">${e.type}</div>
              <h3>${escHtml(e.title)}</h3>
              ${e.tags && e.tags.length ? `<div class="desc">Tags: ${e.tags.join(', ')}</div>` : ''}
              <div class="entry-meta">
                <span>${fmtDate(e.created_at)}</span>
                ${e.resources ? `<span>${e.resources.length} resource${e.resources.length !== 1 ? 's' : ''}</span>` : ''}
              </div>
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
    const el = document.getElementById('entry-content');
    if (!el) return;
    try {
      const data = await API.listEntries({ q, page_size: 20 });
      if (!data.entries || data.entries.length === 0) {
        el.innerHTML = `
          <input type="text" class="filter-input" placeholder="Search entries..." oninput="EntriesPage.search(this.value)" style="max-width:100%;margin-bottom:20px" value="${escHtml(q)}">
          <div class="empty-state"><p>No results for "${escHtml(q)}"</p></div>
        `;
        return;
      }
      el.innerHTML = `
        <input type="text" class="filter-input" placeholder="Search entries..." oninput="EntriesPage.search(this.value)" style="max-width:100%;margin-bottom:20px" value="${escHtml(q)}">
        <div class="resource-grid">
          ${data.entries.map(e => `
            <div class="resource-card" onclick="App.navigate('entries', '${e.id}')">
              <div class="type-badge">${e.type}</div>
              <h3>${escHtml(e.title)}</h3>
              ${e.tags && e.tags.length ? `<div class="desc">Tags: ${e.tags.join(', ')}</div>` : ''}
              <div class="entry-meta">
                <span>${fmtDate(e.created_at)}</span>
              </div>
            </div>
          `).join('')}
        </div>
      `;
    } catch { /* ignore */ }
  },
};
