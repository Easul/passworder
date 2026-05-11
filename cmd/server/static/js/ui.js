window.PassworderUiMethods = {
  showPage(pageId) {
    document.querySelectorAll('.page').forEach(p => {
      p.classList.remove('active');
    });
    document.getElementById(pageId).classList.add('active');
  },

  showMain() {
    this.showPage('main-page');
    this.renderHeader('accounts');
  },

  renderHeader(activeTab) {
    this.renderHeaderInto('main-header', activeTab);
  },

  renderHeaderInto(containerId, activeTab) {
    const header = document.getElementById(containerId);
    if (!header) return;
    header.innerHTML = `
      <div class="header-content">
        <div class="logo">🔐 Passworder</div>
        <div class="nav-actions">
          <button class="btn btn-secondary btn-sm nav-btn ${activeTab === 'accounts' ? 'active' : ''}" onclick="Passworder.showMain()">🔑 <span>账号</span></button>
          <button class="btn btn-secondary btn-sm nav-btn ${activeTab === 'notes' ? 'active' : ''}" onclick="Passworder.showNotes()">📁 <span>笔记</span></button>
          <button class="btn btn-secondary btn-sm nav-btn" onclick="Passworder.loadTrash()">🗑️ <span>回收站</span></button>
          <button class="btn btn-secondary btn-sm nav-btn" onclick="Passworder.showSettings()">⚙️ <span>设置</span></button>
          <button class="btn btn-secondary btn-sm nav-btn" onclick="Passworder.logout()">🚪 <span>退出</span></button>
        </div>
      </div>
    `;
  },

  openModal(id) {
    document.getElementById(id).classList.add('active');
    document.body.style.overflow = 'hidden';
  },

  closeModal(id) {
    document.getElementById(id).classList.remove('active');
    document.body.style.overflow = '';
    if (id === 'image-preview-modal') {
      const img = document.getElementById('preview-image');
      if (img) img.src = '';
      if (this.currentPreviewUrl) {
        URL.revokeObjectURL(this.currentPreviewUrl);
        this.currentPreviewUrl = null;
      }
      this.currentPreviewNote = null;
    }
    if (id === 'markdown-modal') this.currentMarkdownNote = null;
    if (id === 'note-modal') {
      this.currentNoteId = null;
      this.destroyVditor?.();
    }
  },

  showToast(type, message) {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `<span>${type === 'success' ? '✓' : type === 'error' ? '✗' : 'ℹ'}</span><span>${message}</span>`;
    container.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
    return toast;
  },

  renderPagination(containerId, total, currentPage, pageSize, onPageChange) {
    const container = document.getElementById(containerId);
    if (!container) return;
    const totalPages = Math.ceil(total / pageSize) || 1;
    if (totalPages <= 1) {
      container.innerHTML = '';
      return;
    }
    container.innerHTML = '';
    const prevBtn = document.createElement('button');
    prevBtn.textContent = '上一页';
    prevBtn.disabled = currentPage === 1;
    prevBtn.onclick = () => onPageChange(currentPage - 1);
    container.appendChild(prevBtn);
    for (let i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= currentPage - 1 && i <= currentPage + 1)) {
        const btn = document.createElement('button');
        btn.textContent = i;
        if (i === currentPage) btn.classList.add('active');
        btn.onclick = () => onPageChange(i);
        container.appendChild(btn);
      } else if (i === currentPage - 2 || i === currentPage + 2) {
        const span = document.createElement('span');
        span.textContent = '...';
        span.style.cssText = 'padding:6px 8px;color:var(--text-secondary)';
        container.appendChild(span);
      }
    }
    const nextBtn = document.createElement('button');
    nextBtn.textContent = '下一页';
    nextBtn.disabled = currentPage === totalPages;
    nextBtn.onclick = () => onPageChange(currentPage + 1);
    container.appendChild(nextBtn);
  },

  filterNotes(type) {
    this.currentNoteFilter = type;
    this.notePage = 1;
    document.querySelectorAll('.file-filter-tab').forEach(tab => {
      tab.classList.toggle('active', tab.dataset.filter === type);
    });
    this.renderNotes();
  },

  setupEventListeners() {
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        document.querySelectorAll('.modal-overlay.active').forEach(m => {
          m.classList.remove('active');
        });
        document.body.style.overflow = '';
      }
    });

    const searchInput = document.getElementById('search-input');
    if (searchInput) searchInput.addEventListener('input', () => { this.accountPage = 1; this.renderAccounts(); });

    const filterCategory = document.getElementById('filter-category');
    if (filterCategory) filterCategory.addEventListener('change', () => { this.accountPage = 1; this.renderAccounts(); });

    const noteSearchInput = document.getElementById('note-search-input');
    if (noteSearchInput) {
      noteSearchInput.addEventListener('input', (e) => {
        this.noteSearchQuery = e.target.value.toLowerCase();
        this.notePage = 1;
        this.renderNotes();
      });
    }

    const noteFilterTabs = document.getElementById('note-filter-tabs');
    if (noteFilterTabs) {
      noteFilterTabs.addEventListener('click', (e) => {
        if (e.target.classList.contains('file-filter-tab')) this.filterNotes(e.target.dataset.filter);
      });
    }

    const markdownEditor = document.getElementById('markdown-editor');
    if (markdownEditor) {
      let debounceTimer;
      markdownEditor.addEventListener('input', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => this.renderMarkdownPreview(), 300);
      });
    }

    const formatRadios = document.querySelectorAll('input[name="note-format"]');
    formatRadios.forEach(r => {
      r.addEventListener('change', (e) => {
        if (e.target.value === 'markdown') {
          const currentBody = document.getElementById('note-body').value;
          this.initVditor(currentBody);
        } else {
          const currentBody = this.vditor ? this.vditor.getValue() : document.getElementById('note-body').value;
          this.destroyVditor();
          document.getElementById('note-body-vditor').classList.add('hidden');
          document.getElementById('note-body').classList.remove('hidden');
          document.getElementById('note-body').value = currentBody;
        }
      });
    });

    const noteFileInput = document.getElementById('note-file');
    if (noteFileInput) {
      noteFileInput.addEventListener('change', () => {
        const files = Array.from(noteFileInput.files);
        this.pendingFiles.push(...files);
        noteFileInput.value = '';
        this.renderNoteAttachmentList(this.notes.find(n => n.id === this.currentNoteId));
      });
    }
  },

  initPeriodControls() {
    const remindAtInput = document.getElementById('acc-remind-at');
    const periodGroup = document.getElementById('reminder-period-group');
    if (remindAtInput && periodGroup) {
      const showPeriod = !!remindAtInput.value;
      periodGroup.style.display = showPeriod ? 'block' : 'none';
    }
  },

  onPeriodTypeChange() {
    const type = document.getElementById('acc-period-type')?.value || '';
    const valueInput = document.getElementById('acc-period-value');
    const hint = document.getElementById('period-hint');
    if (!type || !valueInput || !hint) return;
    const needsValue = type === 'monthly' || type === 'hourly' || type === 'days';
    valueInput.style.display = needsValue ? 'block' : 'none';
    const hints = {
      yearly: '每年在这一天提醒您',
      monthly: '每月提醒您',
      weekly: '每周这一天提醒您',
      daily: '每天提醒您',
      hourly: '每隔几小时提醒您',
      days: '每隔几天提醒您'
    };
    hint.textContent = hints[type] || '';
  }
};

if (window.Passworder) Object.assign(window.Passworder, window.PassworderUiMethods);
