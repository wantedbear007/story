const CapturesPage = {
  async render() {
    return `
      <div id="capture-toolbar" style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <span class="caption" id="capture-count"></span>
        <button class="btn btn-primary btn-sm" onclick="CapturesPage.toggleForm()">+ New</button>
      </div>
      <div id="create-form" class="mat-card" style="display:none;margin-bottom:12px">
        <textarea id="new-capture-content" class="mat-textarea" placeholder="What did you learn or do?" style="margin-bottom:10px"></textarea>
        <div style="display:flex;gap:6px">
          <button class="btn btn-primary btn-sm" onclick="CapturesPage.createCapture()">Save</button>
          <button class="btn btn-secondary btn-sm" onclick="CapturesPage.toggleForm()">Cancel</button>
        </div>
      </div>
      <div id="capture-list" class="capture-list"></div>
    `;
  },

  async afterRender() {
    await this.load();
  },

  async load() {
    const el = document.getElementById('capture-list');
    if (!el) return;
    el.innerHTML = '<div class="empty-state" style="padding:40px 20px"><p>Loading...</p></div>';

    try {
      const data = await API.request('GET', '/api/raw-entries?limit=100');
      const entries = data.entries || [];

      const countEl = document.getElementById('capture-count');
      if (countEl) countEl.textContent = `${entries.length} capture${entries.length !== 1 ? 's' : ''}`;

      if (entries.length === 0) {
        el.innerHTML = '<div class="empty-state"><h2>Nothing yet</h2><p>Captures from notifications, CLI, or the web appear here.</p></div>';
        return;
      }

      el.innerHTML = entries.map(e => this.card(e)).join('');
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  card(e) {
    const words = e.content.split(/\s+/).filter(Boolean).length;
    const ready = e.status === 'structured';
    const isRaw = e.status === 'raw';
    const isFailed = e.status === 'failed';
    const isProcessing = e.status === 'processing';
    const src = e.source === 'notification_capture' ? 'Notification' : e.source === 'cli' ? 'CLI' : 'Web';

    return `
      <div class="capture-card" id="capture-${e.id}">
        <div class="capture-card-header">
          <span class="mat-chip ${isProcessing ? 'processing' : isFailed ? 'failed' : ready ? 'ready' : isRaw ? 'pending' : ''}">${isProcessing ? 'Processing' : isFailed ? 'Failed' : ready ? 'Ready' : isRaw ? 'Pending' : e.status}</span>
          <span class="mat-chip">${src}</span>
          <span class="caption" style="margin-left:auto">${fmtDate(e.created_at)}</span>
        </div>
        <div class="capture-card-content">${escHtml(e.content)}</div>
        <div class="capture-card-footer">
          <span class="caption">${words} words</span>
          <span class="meta-sep">·</span>
          <span class="caption">${isProcessing ? 'Summarizing...' : isFailed ? 'Conversion failed' : ready ? 'Tweet ready' : 'Waiting for LLM'}</span>
          <button class="btn-icon" onclick="CapturesPage.del('${e.id}')" title="Delete">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>
    `;
  },

  toggleForm() {
    const f = document.getElementById('create-form');
    const hidden = f.style.display === 'none' || f.style.display === '';
    f.style.display = hidden ? 'block' : 'none';
    if (hidden) document.getElementById('new-capture-content').focus();
  },

  async createCapture() {
    const content = document.getElementById('new-capture-content').value.trim();
    if (!content) { showToast('Content is required', 'error'); return; }

    try {
      await API.request('POST', '/api/raw-entries', { content, source: 'web' });
      showToast('Capture saved', 'success');
      document.getElementById('new-capture-content').value = '';
      document.getElementById('create-form').style.display = 'none';
      this.load();
    } catch (err) {
      showToast(err.message, 'error');
    }
  },

  async del(id) {
    if (!confirm('Delete this capture?')) return;
    try {
      await API.request('DELETE', `/api/raw-entries/${id}`);
      showToast('Deleted', 'success');
      this.load();
    } catch (err) {
      showToast(err.message, 'error');
    }
  },
};
