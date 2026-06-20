const App = {
  async init() {
    await this.handleUrlCode();

    const token = API.getToken();
    if (token) {
      try {
        await API.me();
        this.showMain();
      } catch {
        API.clearToken();
        this.showLogin();
      }
    } else {
      this.showLogin();
    }

    document.getElementById('login-btn').addEventListener('click', () => this.login());
    document.getElementById('code-input').addEventListener('keydown', e => {
      if (e.key === 'Enter') this.login();
    });
    document.getElementById('logout-link').addEventListener('click', e => {
      e.preventDefault();
      API.clearToken();
      this.showLogin();
    });

    window.addEventListener('hashchange', () => this.route());
    this.route();
  },

  async handleUrlCode() {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    if (!code) return;
    try {
      const token = await API.exchangeCode(code);
      API.setToken(token);
      window.history.replaceState({}, document.title, window.location.pathname);
      showToast('Connected', 'success');
    } catch (err) {
      showToast(err.message || 'Invalid code', 'error');
    }
  },

  async login() {
    const code = document.getElementById('code-input').value.trim();
    if (!code) return;
    try {
      const token = await API.exchangeCode(code);
      API.setToken(token);
      this.showMain();
      this.route();
    } catch (err) {
      showToast(err.message || 'Invalid code', 'error');
    }
  },

  showLogin() {
    document.getElementById('login-view').style.display = 'flex';
    document.getElementById('main-view').style.display = 'none';
    document.getElementById('sidebar').style.display = 'none';
  },

  showMain() {
    document.getElementById('login-view').style.display = 'none';
    document.getElementById('main-view').style.display = 'block';
    document.getElementById('sidebar').style.display = 'flex';
  },

  route() {
    const hash = window.location.hash.slice(1) || '/captures';
    const page = hash.split('/').filter(Boolean)[0] || 'captures';
    this.renderPage(page);
  },

  async navigate(page) {
    window.location.hash = `#/${page}`;
    this.renderPage(page);
  },

  pages: {
    captures: { title: 'Captures', render: () => CapturesPage.render(), after: () => CapturesPage.afterRender() },
    created: { title: 'Created', render: () => CreatedPage.render(), after: () => CreatedPage.afterRender() },
    pipeline: { title: 'Pipeline', render: () => PipelinePage.render(), after: () => PipelinePage.afterRender() },
  },

  async renderPage(page) {
    const content = document.getElementById('page-content');
    const pageTitle = document.getElementById('page-title');
    const pageActions = document.getElementById('page-actions');
    if (!content) return;

    document.querySelectorAll('.sidebar-link[data-route]').forEach(l => {
      l.classList.toggle('active', l.getAttribute('href') === `#/${page}`);
    });

    const p = this.pages[page];
    if (!p) {
      content.innerHTML = '<div class="empty-state"><h2>Not found</h2></div>';
      return;
    }

    if (pageTitle) pageTitle.textContent = p.title;
    if (pageActions) pageActions.innerHTML = '';
    content.innerHTML = await p.render();
    await p.after();
  },
};

function escHtml(s) {
  if (!s) return '';
  const d = document.createElement('div');
  d.textContent = s;
  return d.innerHTML;
}

function fmtDate(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const now = new Date();
  const diff = (now - d) / 1000;
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function showToast(msg, type) {
  const existing = document.querySelector('.toast');
  if (existing) existing.remove();
  const t = document.createElement('div');
  t.className = `toast toast-${type}`;
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 3000);
}

document.addEventListener('DOMContentLoaded', () => App.init());
