const CreatedPage = {
  async render() {
    return `<div id="created-list" class="capture-list"></div>`;
  },

  async afterRender() {
    await this.load();
  },

  async load() {
    const el = document.getElementById('created-list');
    if (!el) return;
    el.innerHTML = '<div class="empty-state" style="padding:40px 20px"><p>Loading...</p></div>';

    try {
      const data = await API.listTweets({ limit: 100 });
      const tweets = data.tweets || [];

      if (tweets.length === 0) {
        el.innerHTML = '<div class="empty-state"><h2>No tweets yet</h2><p>Captures are auto-summarized when LLM is configured.</p></div>';
        return;
      }

      el.innerHTML = tweets.map(t => `
        <div class="capture-card">
          <div class="capture-card-header">
            <span class="mat-chip ${t.status}">${t.status}</span>
            <span class="mat-chip">v${t.version}</span>
            ${t.provider_name ? `<span class="mat-chip">${t.provider_name}</span>` : ''}
            <span class="caption" style="margin-left:auto">${fmtDate(t.created_at)}</span>
          </div>
          <div class="capture-card-content">${escHtml(t.content)}</div>
          <div class="capture-card-footer">
            <span class="caption" style="${t.content.length > 280 ? 'color:#f28b82' : ''}">${t.content.length}/280</span>
            ${t.cost_usd ? `<span class="meta-sep">·</span><span class="caption">$${t.cost_usd.toFixed(6)}</span>` : ''}
            <button class="btn-icon" style="margin-left:auto" onclick="CreatedPage.copy('${escHtml(t.content)}')" title="Copy">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </button>
          </div>
        </div>
      `).join('');
    } catch (err) {
      el.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  async copy(content) {
    try {
      await navigator.clipboard.writeText(content);
      showToast('Copied', 'success');
    } catch {
      showToast('Failed to copy', 'error');
    }
  },
};
