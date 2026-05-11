const Passworder = {
  token: localStorage.getItem('token') || null,
  currentAccount: null,
  categories: [],
  accounts: [],
  notes: [],
  trashNotes: [],
  currentNoteFilter: 'all',
  currentPreviewNote: null,
  currentMarkdownNote: null,
  currentPreviewUrl: null,
  currentNoteId: null,
  senderSettings: {},
  markdownLibLoaded: false,
  dompurifyLoaded: false,
  vditor: null,
  vditorToolbarVisible: false,
  accountPage: 1,
  accountPageSize: 20,
  notePage: 1,
  notePageSize: 20,
  noteSearchQuery: '',
  accountSortField: 'createdAt',
  accountSortOrder: 'desc',

  init() {
    this.setupEventListeners();
    this.checkAuth();
  },

  async api(path, options = {}) {
    const url = `/api${path}`;
    const opts = {
      headers: {
        'Content-Type': 'application/json',
        ...(this.token && { 'Authorization': this.token }),
        ...options.headers
      },
      ...options
    };
    if (opts.body && typeof opts.body === 'object') {
      opts.body = JSON.stringify(opts.body);
    }
    const res = await fetch(url, opts);
    const data = await res.json().catch(() => null);
    if (res.status === 401) {
      this.token = null;
      localStorage.removeItem('token');
      this.showLogin();
      this.showToast('error', '会话已过期，请重新登录');
      throw new Error('会话过期');
    }
    if (!res.ok) {
      throw new Error(data?.message || `HTTP ${res.status}`);
    }
    return data.data !== undefined ? data.data : data;
  },

  escape(text) {
    return window.PassworderShared.escape(text);
  },

  escapeAttr(text) {
    return window.PassworderShared.escapeAttr(text);
  },

  async loadData() {
    await Promise.all([this.loadCategories(), this.loadAccounts()]);
    this.renderCategoryFilter();
    this.renderAccounts();
  },

  showNotes(shouldLoad = true) {
    this.showPage('notes-page');
    this.renderHeaderInto('notes-header', 'notes');
    if (shouldLoad) {
      this.loadNotes();
    }
  },

  async copyText(text, successMessage = '已复制') {
    try {
      await navigator.clipboard.writeText(text || '');
      this.showToast('success', successMessage);
    } catch (e) {
      const input = document.createElement('textarea');
      input.value = text || '';
      document.body.appendChild(input);
      input.select();
      document.execCommand('copy');
      input.remove();
      this.showToast('success', successMessage);
    }
  },

  togglePassword(id) {
    const input = document.getElementById(id);
    if (!input) return;
    input.type = input.type === 'password' ? 'text' : 'password';
  },

  openGenerator() {
    this.openModal('generator-modal');
    this.generatePassword();
  },

  generatePassword() {
    const length = parseInt(document.getElementById('gen-length')?.value || '16', 10);
    const upper = document.getElementById('gen-upper')?.checked;
    const lower = document.getElementById('gen-lower')?.checked;
    const numbers = document.getElementById('gen-numbers')?.checked;
    const symbols = document.getElementById('gen-symbols')?.checked;

    let chars = '';
    if (upper) chars += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    if (lower) chars += 'abcdefghijklmnopqrstuvwxyz';
    if (numbers) chars += '0123456789';
    if (symbols) chars += '!@#$%^&*()_+-=[]{}|;:,.<>?';
    if (!chars) {
      this.showToast('error', '请至少选择一种字符类型');
      return;
    }

    const bytes = new Uint32Array(length);
    window.crypto.getRandomValues(bytes);
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars[bytes[i] % chars.length];
    }
    const output = document.getElementById('generator-result');
    if (output) output.value = result;
  },

  useGeneratedPassword() {
    const generated = document.getElementById('generator-result')?.value || '';
    const target = document.getElementById('acc-password');
    if (target) target.value = generated;
    this.closeModal('generator-modal');
    if (target) target.focus();
  },




  formatSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  },

  formatDate(timestamp) {
    return window.PassworderShared.formatDate(timestamp);
  },

  formatDateTimeLocal(timestamp) {
    return window.PassworderShared.formatDateTimeLocal(timestamp);
  },

  formatDateInput(timestamp) {
    return window.PassworderShared.formatDateInput(timestamp);
  },

  formatTimeInput(timestamp) {
    return window.PassworderShared.formatTimeInput(timestamp);
  },

  combineDateAndTime(dateValue, timeValue) {
    return window.PassworderShared.combineDateAndTime(dateValue, timeValue);
  },
};

window.Passworder = Passworder;
document.addEventListener('DOMContentLoaded', () => Passworder.init());
