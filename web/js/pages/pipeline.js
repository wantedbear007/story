const PipelinePage = {
  async render() {
    return `<div id="pipeline-root"></div>`;
  },

  async afterRender() {
    await this.load();
  },

  async load() {
    const root = document.getElementById('pipeline-root');
    if (!root) return;
    root.innerHTML = '<div class="empty-state" style="padding:40px 20px"><p>Loading...</p></div>';

    try {
      const [config, rawData, tweetData] = await Promise.all([
        API.request('GET', '/api/config'),
        API.request('GET', '/api/raw-entries?limit=200'),
        API.listTweets({ limit: 200 }),
      ]);

      const rawEntries = rawData.entries || [];
      const tweets = tweetData.tweets || [];

      const raw = rawEntries.filter(e => e.status === 'raw').length;
      const processing = rawEntries.filter(e => e.status === 'processing').length;
      const structured = rawEntries.filter(e => e.status === 'structured').length;
      const llmOk = config.llm_configured;

      root.innerHTML = `
        ${!llmOk ? `
          <div class="pipeline-warning">
            <strong>LLM not configured</strong>
            <p>Set up an LLM provider (OpenAI, Gemini, or Anthropic) in your config or environment to enable automatic tweet generation from captures.</p>
          </div>
        ` : `
          <div class="pipeline-info">
            <strong>LLM configured</strong>
            <p>Auto-summarization is active. New captures are processed every 30 seconds.</p>
          </div>
        `}

        <div class="pipeline-stages">
          <div class="pipeline-stage stage-raw">
            <div class="pipeline-stage-value">${raw}</div>
            <div class="pipeline-stage-label">Captured</div>
            <div class="pipeline-stage-desc">Pending processing</div>
          </div>
          <div class="pipeline-arrow">→</div>
          <div class="pipeline-stage stage-processing">
            <div class="pipeline-stage-value">${processing}</div>
            <div class="pipeline-stage-label">Processing</div>
            <div class="pipeline-stage-desc">Summarizing</div>
          </div>
          <div class="pipeline-arrow">→</div>
          <div class="pipeline-stage stage-structured">
            <div class="pipeline-stage-value">${structured}</div>
            <div class="pipeline-stage-label">Summarized</div>
            <div class="pipeline-stage-desc">Tweets ready</div>
          </div>
        </div>

        <div class="pipeline-summary">
          <div class="stat-card">
            <div class="stat-value">${rawEntries.length}</div>
            <div class="stat-label">Total Captures</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">${tweets.length}</div>
            <div class="stat-label">Tweets Generated</div>
          </div>
        </div>

        ${!llmOk && raw > 0 ? `
          <p class="caption" style="margin-top:20px;text-align:center">
            ${raw} capture${raw > 1 ? 's are' : ' is'} waiting · configure LLM to auto-summarize
          </p>
        ` : ''}
      `;
    } catch (err) {
      root.innerHTML = `<div class="empty-state"><h2>Error</h2><p>${escHtml(err.message)}</p></div>`;
    }
  },
};
