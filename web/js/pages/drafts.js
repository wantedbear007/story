const DraftsPage = {
  currentStatus: '',
  currentFilter: '',

  async render() {
    return `
      <div class="status-bar" id="status-bar">
        <button class="status-btn ${this.currentStatus === '' ? 'active' : ''}" onclick="DraftsPage.setStatus('')">All</button>
        <button class="status-btn ${this.currentStatus === 'draft' ? 'active' : ''}" onclick="DraftsPage.setStatus('draft')">Draft</button>
        <button class="status-btn ${this.currentStatus === 'reviewing' ? 'active' : ''}" onclick="DraftsPage.setStatus('reviewing')">Review</button>
        <button class="status-btn ${this.currentStatus === 'approved' ? 'active' : ''}" onclick="DraftsPage.setStatus('approved')">Approved</button>
        <button class="status-btn ${this.currentStatus === 'scheduled' ? 'active' : ''}" onclick="DraftsPage.setStatus('scheduled')">Scheduled</button>
        <button class="status-btn ${this.currentStatus === 'posted' ? 'active' : ''}" onclick="DraftsPage.setStatus('posted')">Posted</button>
        <button class="status-btn ${this.currentStatus === 'archived' ? 'active' : ''}" onclick="DraftsPage.setStatus('archived')">Archived</button>
      </div>
      <div class="filters">
        <input type="text" class="filter-input" id="draft-filter" placeholder="Search by entry ID..." value="${escHtml(this.currentFilter)}" oninput="DraftsPage.setFilter(this.value)">
      </div>
      <div id="tweet-list" class="tweet-list"></div>
      <div id="generate-modal-overlay" class="modal-overlay" style="display:none" onclick="DraftsPage.closeGenerateModal(event)">
        <div class="modal" onclick="event.stopPropagation()">
          <h2>Generate Tweet</h2>
          <input type="text" id="gen-entry-id" placeholder="Entry ID" class="filter-input">
          <button class="btn btn-primary" onclick="DraftsPage.doGenerate()">Generate</button>
          <button class="btn btn-secondary" onclick="DraftsPage.closeGenerateModal()">Cancel</button>
        </div>
      </div>
    `;
  },

  async afterRender() {
    await this.loadTweets();
  },

  async loadTweets() {
    const el = document.getElementById('tweet-list');
    if (!el) return;
    el.innerHTML = '<div class="empty-state">Loading...</div>';

    try {
      const params = { limit: 200 };
      if (this.currentStatus) params.status = this.currentStatus;
      const data = await API.listTweets(params);

      if (!data.tweets || data.tweets.length === 0) {
        el.innerHTML = '<div class="empty-state"><h2>No tweets found</h2><p>Generate one from an entry.</p></div>';
        return;
      }

      el.innerHTML = data.tweets.map(t => `
        <div class="tweet-card" onclick="App.navigate('edit', '${t.id}')">
          <div class="tweet-card-header">
            <span class="status-badge ${t.status}">${t.status} v${t.version}</span>
            <span style="color:#484f58;font-size:0.8em">${fmtDate(t.created_at)}</span>
          </div>
          <div class="tweet-card-content">${escHtml(t.content)}</div>
          <div class="tweet-card-footer">
            <span class="${t.content.length > 280 ? 'char-count over' : 'char-count'}">${t.content.length}/280</span>
            <span>${t.provider_name || 'N/A'}</span>
            ${t.cost_usd ? `<span>$${t.cost_usd.toFixed(6)}</span>` : ''}
            <button class="btn-icon copy-btn" onclick="event.stopPropagation(); DraftsPage.copyTweet('${escHtml(t.content)}')" title="Copy">📋</button>
          </div>
        </div>
      `).join('');
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  setStatus(status) {
    this.currentStatus = status;
    this.loadTweets();
    document.querySelectorAll('.status-btn').forEach(b => b.classList.remove('active'));
  },

  setFilter(val) {
    this.currentFilter = val;
  },

  showGenerateModal() {
    document.getElementById('gen-entry-id').value = '';
    document.getElementById('generate-modal-overlay').style.display = 'flex';
    document.getElementById('gen-entry-id').focus();
  },

  showGenerateWithEntry(entryId) {
    document.getElementById('gen-entry-id').value = entryId;
    document.getElementById('generate-modal-overlay').style.display = 'flex';
    document.getElementById('gen-entry-id').focus();
  },

  closeGenerateModal(e) {
    if (e && e.target !== e.currentTarget) return;
    document.getElementById('generate-modal-overlay').style.display = 'none';
  },

  async doGenerate() {
    const entryId = document.getElementById('gen-entry-id').value.trim();
    if (!entryId) { showToast('Enter an entry ID', 'error'); return; }

    try {
      const result = await API.generateTweet({ entry_id: entryId });
      showToast('Tweet generated!', 'success');
      this.closeGenerateModal();
      App.navigate('edit', result.id);
    } catch (err) {
      showToast(err.message, 'error');
    }
  },

  async copyTweet(content) {
    try {
      await navigator.clipboard.writeText(content);
      showToast('Copied to clipboard', 'success');
    } catch {
      showToast('Failed to copy', 'error');
    }
  },
};
