const DashboardPage = {
  async render() {
    return `<div id="dashboard-root"><div class="empty-state">Loading...</div></div>`;
  },

  async afterRender() {
    await this.load();
  },

  async load() {
    const root = document.getElementById("dashboard-root");
    if (!root) return;

    try {
      const data = await API.listTweets({ limit: 200 });
      const tweets = data.tweets || [];

      const counts = { draft: 0, reviewing: 0, approved: 0, scheduled: 0, posted: 0, archived: 0 };
      for (const t of tweets) {
        if (counts[t.status] !== undefined) counts[t.status]++;
      }
      const total = tweets.length;
      const totalCost = tweets.reduce((s, t) => s + (t.cost_usd || 0), 0);
      const totalTokens = tweets.reduce((s, t) => s + (t.input_tokens || 0) + (t.output_tokens || 0), 0);

      const recent = tweets
        .filter(t => t.status !== "archived")
        .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
        .slice(0, 8);

      const toReview = tweets
        .filter(t => t.status === "reviewing")
        .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
        .slice(0, 5);

      root.innerHTML = `
        <div class="stats-grid">
          ${this.statCard(total, "Total", "total")}
          ${this.statCard(counts.draft, "Draft", "draft")}
          ${this.statCard(counts.reviewing, "Review", "review")}
          ${this.statCard(counts.approved, "Approved", "approved")}
          ${this.statCard(counts.scheduled, "Scheduled", "scheduled")}
          ${this.statCard(counts.posted, "Posted", "posted")}
        </div>
        <div class="dashboard-grid">
          <div class="dashboard-section">
            <h2>Recent Activity</h2>
            ${recent.length === 0 ? '<div class="dashboard-tweet-item" style="color:#484f58;padding:20px 0;text-align:center;justify-content:center">No tweets yet</div>' : ''}
            ${recent.map(t => `
              <div class="dashboard-tweet-item" onclick="App.navigate('edit', '${t.id}')">
                <span class="status-badge ${t.status}">${t.status}</span>
                <span class="dash-tweet-content">${escHtml(t.content)}</span>
                <span class="dash-tweet-time">${fmtDate(t.updated_at)}</span>
              </div>
            `).join('')}
          </div>
          <div class="dashboard-section">
            <h2>Needs Review</h2>
            ${toReview.length === 0 ? '<div class="dashboard-tweet-item" style="color:#484f58;padding:20px 0;text-align:center;justify-content:center">All caught up</div>' : ''}
            ${toReview.map(t => `
              <div class="dashboard-tweet-item" onclick="App.navigate('edit', '${t.id}')">
                <span class="dash-tweet-content" style="font-weight:500">${escHtml(t.content)}</span>
                <span class="dash-tweet-time">${fmtDate(t.updated_at)}</span>
              </div>
            `).join('')}
          </div>
        </div>
        <div class="stats-grid" style="margin-top:20px;grid-template-columns:repeat(auto-fill,minmax(200px,1fr))">
          <div class="stat-card">
            <div class="stat-value" style="font-size:1.2em;color:#8b949e">${totalCost < 0.001 ? '<$0.001' : '$' + totalCost.toFixed(4)}</div>
            <div class="stat-label">Total Cost</div>
          </div>
          <div class="stat-card">
            <div class="stat-value" style="font-size:1.2em;color:#8b949e">${totalTokens.toLocaleString()}</div>
            <div class="stat-label">Total Tokens</div>
          </div>
        </div>
      `;
    } catch (err) {
      root.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },

  statCard(value, label, cls) {
    return `<div class="stat-card stat-${cls}"><div class="stat-value">${value}</div><div class="stat-label">${label}</div></div>`;
  },
};
