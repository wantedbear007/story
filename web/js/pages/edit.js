const EditPage = {
  tweetId: null,
  tweet: null,
  currentContent: '',

  async render(id) {
    this.tweetId = id;
    return `<div class="tweet-detail" id="tweet-detail"><div class="empty-state">Loading...</div></div>`;
  },

  async afterRender() {
    await this.load();
  },

  async load() {
    const el = document.getElementById('tweet-detail');
    if (!el) return;

    try {
      this.tweet = await API.getTweet(this.tweetId);
      this.currentContent = this.tweet.content;
      this.renderDetail(el);
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  renderDetail(el) {
    const t = this.tweet;
    el.innerHTML = `
      <div class="tweet-detail-header">
        <div>
          <h2>${t.status} (v${t.version})</h2>
          <div class="tweet-detail-meta">
            <span>Provider: ${t.provider_name || 'N/A'}</span>
            <span>Model: ${t.model_name || 'N/A'}</span>
            <span>Tokens: ${t.input_tokens || 0} in / ${t.output_tokens || 0} out</span>
            <span>Cost: $${t.cost_usd ? t.cost_usd.toFixed(6) : '0'}</span>
            <span>Latency: ${t.latency_ms || 0}ms</span>
          </div>
        </div>
        <div>
          <button class="btn btn-icon" onclick="EditPage.copy()" title="Copy content">📋</button>
          <button class="btn btn-secondary btn-sm" onclick="App.navigate('drafts')">← Back</button>
        </div>
      </div>
      <textarea id="tweet-content" oninput="EditPage.onContentChange(this.value)">${escHtml(t.content)}</textarea>
      <div class="${t.content.length > 280 ? 'char-count over' : 'char-count'}" id="char-count">
        ${t.content.length} / 280 characters
      </div>
      <div class="tweet-actions">
        <button class="btn btn-primary btn-sm" onclick="EditPage.save()" id="btn-save">Save</button>
        <button class="btn btn-secondary btn-sm" onclick="EditPage.regenerate()" id="btn-regen">Regenerate</button>
        <button class="btn btn-secondary btn-sm" onclick="EditPage.review()" id="btn-review">Send to Review</button>
        <button class="btn btn-primary btn-sm" onclick="EditPage.approve()" id="btn-approve">Approve</button>
        <button class="btn btn-secondary btn-sm" onclick="EditPage.reject()" id="btn-reject">Reject</button>
        <button class="btn btn-danger btn-sm" onclick="EditPage.archive()" id="btn-archive">Archive</button>
      </div>
      <div class="tweet-detail-meta">
        <span>Entry: <a href="#/resources?id=${t.entry_id}" onclick="App.navigate('resources', null, '${t.entry_id}')">${t.entry_id}</a></span>
        <span>Prompt: ${t.prompt_name || 'N/A'} (v${t.prompt_version || '?'})</span>
      </div>
      <div class="tweet-audit" id="audit-section">
        <h3>Audit Trail</h3>
        <div id="audit-list">Loading...</div>
      </div>
    `;

    this.updateButtons();
    this.loadAudits();
  },

  onContentChange(val) {
    this.currentContent = val;
    const cc = document.getElementById('char-count');
    if (cc) {
      cc.textContent = `${val.length} / 280 characters`;
      cc.className = val.length > 280 ? 'char-count over' : 'char-count';
    }
  },

  updateButtons() {
    const status = this.tweet.status;
    setVisible('btn-review', status === 'draft');
    setVisible('btn-approve', status === 'draft' || status === 'reviewing');
    setVisible('btn-reject', status === 'reviewing' || status === 'approved');
    setVisible('btn-archive', status !== 'archived' && status !== 'posted');
    setVisible('btn-regen', status !== 'posted');
    setVisible('btn-save', true);
  },

  async save() {
    if (!this.currentContent.trim()) { showToast('Content cannot be empty', 'error'); return; }
    try {
      const result = await API.updateTweet(this.tweetId, { content: this.currentContent });
      this.tweet = result;
      showToast('Saved', 'success');
    } catch (err) {
      showToast(err.message, 'error');
    }
  },

  async regenerate() {
    try {
      const result = await API.regenerateTweet(this.tweetId, {});
      showToast('Regenerated!', 'success');
      App.navigate('edit', result.id);
    } catch (err) {
      showToast(err.message, 'error');
    }
  },

  async review() {
    try {
      await API.reviewTweet(this.tweetId);
      showToast('Sent to review', 'success');
      this.tweet.status = 'reviewing';
      this.updateButtons();
    } catch (err) { showToast(err.message, 'error'); }
  },

  async approve() {
    try {
      await API.approveTweet(this.tweetId);
      showToast('Approved!', 'success');
      this.tweet.status = 'approved';
      this.updateButtons();
    } catch (err) { showToast(err.message, 'error'); }
  },

  async reject() {
    try {
      await API.rejectTweet(this.tweetId);
      showToast('Rejected, returned to draft', 'success');
      this.tweet.status = 'draft';
      this.updateButtons();
    } catch (err) { showToast(err.message, 'error'); }
  },

  async archive() {
    if (!confirm('Archive this tweet?')) return;
    try {
      await API.archiveTweet(this.tweetId);
      showToast('Archived', 'success');
      this.tweet.status = 'archived';
      this.updateButtons();
    } catch (err) { showToast(err.message, 'error'); }
  },

  async copy() {
    try {
      await navigator.clipboard.writeText(this.currentContent);
      showToast('Copied to clipboard', 'success');
    } catch { showToast('Failed to copy', 'error'); }
  },

  async loadAudits() {
    const el = document.getElementById('audit-list');
    if (!el) return;
    try {
      const data = await API.getAudits(this.tweetId);
      const audits = data.audits || [];
      if (audits.length === 0) {
        el.innerHTML = '<div class="audit-item">No audit records</div>';
        return;
      }
      el.innerHTML = audits.map(a => `
        <div class="audit-item">
          <span class="time">${fmtDate(a.created_at)}</span>
          <span class="action">${escHtml(a.action)}</span>
          ${a.previous_status ? `<span>${escHtml(a.previous_status)} → ${escHtml(a.new_status)}</span>` : ''}
        </div>
      `).join('');
    } catch {
      el.innerHTML = '<div class="audit-item">Failed to load audits</div>';
    }
  },
};
