window.PassworderShared = {
  escape(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  },

  escapeAttr(text) {
    return this.escape(text ?? '').replace(/"/g, '&quot;');
  },

  formatDate(timestamp) {
    if (!timestamp) return '-';
    const d = new Date(timestamp * 1000);
    return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  },

  formatDateTimeLocal(timestamp) {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  },

  formatDateInput(timestamp) {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  },

  formatTimeInput(timestamp) {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
  },

  combineDateAndTime(dateValue, timeValue) {
    if (!dateValue) return '';
    return `${dateValue}T${timeValue || '00:00'}`;
  },

  loadStylesheetOnce(id, href) {
    if (document.getElementById(id)) return;
    const link = document.createElement('link');
    link.id = id;
    link.rel = 'stylesheet';
    link.href = href;
    document.head.appendChild(link);
  },

  loadScriptOnce(key, src, globalName) {
    if (globalName && window[globalName]) {
      return Promise.resolve();
    }
    this._scriptPromises = this._scriptPromises || {};
    if (this._scriptPromises[key]) {
      return this._scriptPromises[key];
    }

    this._scriptPromises[key] = new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = src;
      script.dataset.sharedKey = key;
      script.onload = resolve;
      script.onerror = reject;
      document.head.appendChild(script);
    });

    return this._scriptPromises[key];
  },

  loadVditor() {
    this.loadStylesheetOnce('vditor-style', '/vendor/vditor/index.min.css');
    return this.loadScriptOnce('vditor', '/vendor/vditor/index.min.js', 'Vditor');
  },

  loadJSZip() {
    return this.loadScriptOnce('jszip', '/vendor/jszip/jszip.min.js', 'JSZip');
  },

  loadMammoth() {
    return this.loadScriptOnce('mammoth', '/vendor/mammoth/mammoth.browser.min.js', 'mammoth');
  },

  loadSheetJS() {
    return this.loadScriptOnce('xlsx', '/vendor/xlsx/xlsx.full.min.js', 'XLSX');
  },

  loadMarkdownLibs() {
    return Promise.all([
      this.loadScriptOnce('marked', '/vendor/marked/marked.min.js', 'marked'),
      this.loadScriptOnce('dompurify', '/vendor/dompurify/purify.min.js', 'DOMPurify'),
    ]);
  },
};
