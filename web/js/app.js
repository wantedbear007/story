const App = {
  currentPage: null,
  currentParams: null,

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
      showToast('Connected!', 'success');
    } catch (err) {
      showToast(err.message || 'Invalid or expired code', 'error');
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
      showToast(err.message || 'Invalid or expired code', 'error');
    }
  },

  showLogin() {
    document.getElementById('login-view').style.display = 'flex';
    document.getElementById('main-view').style.display = 'none';
    document.getElementById('nav-bar').style.display = 'none';
  },

  showMain() {
    document.getElementById('login-view').style.display = 'none';
    document.getElementById('main-view').style.display = 'block';
    document.getElementById('nav-bar').style.display = 'flex';
  },

  route() {
    const hash = window.location.hash.slice(1) || '/drafts';
    const parts = hash.split('/').filter(Boolean);

    let page = parts[0] || 'drafts';
    const params = {};

    if (parts.length > 1) {
      if (page === 'edit') {
        this.navigate('edit', parts[1]);
        return;
      }
      if (page === 'resources' && parts[1]) {
        params.id = parts[1];
      }
    }

    this.renderPage(page, params);
  },

  async navigate(page, id, extra) {
    let hash = `#/${page}`;
    if (id) hash += `/${id}`;
    if (extra && page === 'resources') hash = `#/resources?id=${extra}`;
    window.location.hash = hash;
    this.renderPage(page, { id, extra });
  },

  async renderPage(page, params) {
    const content = document.getElementById('page-content');
    if (!content) return;

    document.querySelectorAll('.nav-link[data-route]').forEach(l => {
      l.classList.toggle('active', l.getAttribute('href') === `#/${page}` || (page === 'edit' && l.getAttribute('href') === '#/drafts'));
    });

    switch (page) {
      case 'drafts':
        content.innerHTML = await DraftsPage.render();
        await DraftsPage.afterRender();
        break;
      case 'edit':
        content.innerHTML = await EditPage.render(params.id);
        await EditPage.afterRender();
        break;
      case 'resources':
        content.innerHTML = await ResourcesPage.render(params);
        await ResourcesPage.afterRender();
        break;
      default:
        content.innerHTML = '<div class="empty-state"><h2>Page not found</h2></div>';
    }
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
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function showToast(msg, type) {
  const existing = document.querySelector('.toast');
  if (existing) existing.remove();
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.textContent = msg;
  document.body.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

function setVisible(id, visible) {
  const el = document.getElementById(id);
  if (el) el.style.display = visible ? '' : 'none';
}

document.addEventListener('DOMContentLoaded', () => App.init());
