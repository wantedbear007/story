const API = {
  token: null,

  setToken(t) { this.token = t; localStorage.setItem('story_token', t); },
  getToken() {
    if (this.token) return this.token;
    const saved = localStorage.getItem('story_token');
    if (saved) this.token = saved;
    return this.token;
  },
  clearToken() { this.token = null; localStorage.removeItem('story_token'); },

  async request(method, path, body) {
    const token = this.getToken();
    if (!token) throw new Error('Not authenticated');

    const opts = {
      method,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    };
    if (body !== undefined) opts.body = JSON.stringify(body);

    const res = await fetch(path, opts);
    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || `HTTP ${res.status}`);
    }
    return data;
  },

  get(path) { return this.request('GET', path); },
  post(path, body) { return this.request('POST', path, body); },
  put(path, body) { return this.request('PUT', path, body); },

  // Tweets
  listTweets(params) {
    const q = new URLSearchParams();
    if (params.status) q.set('status', params.status);
    if (params.entry_id) q.set('entry_id', params.entry_id);
    if (params.limit) q.set('limit', params.limit);
    if (params.offset) q.set('offset', params.offset);
    return this.get(`/api/tweets?${q}`);
  },
  getTweet(id) { return this.get(`/api/tweets/${id}`); },
  generateTweet(data) { return this.post('/api/tweets/generate', data); },
  regenerateTweet(id, data) { return this.post(`/api/tweets/${id}/regenerate`, data); },
  updateTweet(id, data) { return this.put(`/api/tweets/${id}`, data); },
  approveTweet(id) { return this.post(`/api/tweets/${id}/approve`); },
  reviewTweet(id) { return this.post(`/api/tweets/${id}/review`); },
  rejectTweet(id) { return this.post(`/api/tweets/${id}/reject`); },
  scheduleTweet(id, data) { return this.post(`/api/tweets/${id}/schedule`, data); },
  archiveTweet(id) { return this.post(`/api/tweets/${id}/archive`); },
  getAudits(id) { return this.get(`/api/tweets/${id}/audits`); },

  // Entries
  listEntries(params) {
    const q = new URLSearchParams();
    if (params.q) q.set('q', params.q);
    if (params.page) q.set('page', params.page);
    if (params.page_size) q.set('page_size', params.page_size);
    return this.get(`/api/entries?${q}`);
  },
  getEntry(id) { return this.get(`/api/entries/${id}`); },

  // Prompts
  listPrompts() { return this.get('/api/prompts'); },

  // Me
  me() { return this.get('/api/me'); },

  // Auth - exchange a short code for a JWT token
  async exchangeCode(code) {
    const res = await fetch(`/api/exchange/${encodeURIComponent(code)}`);
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Invalid code');
    }
    return data.token;
  },
};
