window.PassworderNotesMethods = {
  async loadNotes() {
    try {
      this.notes = await this.api('/files') || [];
      this.renderNotes();
    } catch (e) {
      this.showToast('error', '加载笔记失败');
    }
  },

  async loadTrash() {
    try {
      this.trashNotes = await this.api('/notes/trash') || [];
      this.renderTrash();
      this.openModal('trash-modal');
    } catch (e) {
      this.showToast('error', '加载回收站失败');
    }
  },

  renderNotes() {
    const container = document.getElementById('note-list');
    const paginationContainer = document.getElementById('note-pagination');
    let filtered = this.notes;

    if (this.noteSearchQuery) {
      filtered = filtered.filter(n =>
        (n.title && n.title.toLowerCase().includes(this.noteSearchQuery)) ||
        (n.body && n.body.toLowerCase().includes(this.noteSearchQuery))
      );
    }

    if (this.currentNoteFilter !== 'all') {
      filtered = filtered.filter(n => n.bodyFormat === this.currentNoteFilter);
    }

    if (filtered.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">📭</div>
          <div class="empty-title">暂无笔记</div>
          <p>点击右上角新建按钮创建</p>
        </div>
      `;
      if (paginationContainer) paginationContainer.innerHTML = '';
      return;
    }

    const total = filtered.length;
    const start = (this.notePage - 1) * this.notePageSize;
    const pageItems = filtered.slice(start, start + this.notePageSize);

    container.innerHTML = pageItems.map(n => {
      const excerpt = n.body ? (n.body.length > 120 ? n.body.slice(0, 120) + '...' : n.body) : '';
      const formatBadge = n.bodyFormat === 'markdown' ? '📝 Markdown' : '📄 文本';
      const hasAttachment = (n.storedName || (n.attachmentCount || 0) > 0) ? '📎' : '';
      const hasPrimaryFile = !!n.storedName;
      return `
        <div class="note-card" data-id="${n.id}">
          <div class="note-card-header">
            <div class="note-card-title">${this.escape(n.title)}</div>
            <div class="note-card-actions" onclick="event.stopPropagation()">
              <button class="btn btn-icon btn-sm" onclick="Passworder.viewNoteBody(${n.id})" title="浏览内容">👁️</button>
              ${hasPrimaryFile ? `<button class="btn btn-icon btn-sm" onclick="Passworder.previewAttachment(${n.id})" title="预览附件">📎</button>` : ''}
              ${hasPrimaryFile ? `<button class="btn btn-icon btn-sm" onclick="Passworder.downloadAttachment(${n.id})" title="下载附件">⬇️</button>` : ''}
              <button class="btn btn-icon btn-sm" onclick="Passworder.showNoteModal(${n.id})" title="编辑">✏️</button>
              <button class="btn btn-icon btn-sm" onclick="Passworder.deleteNote(${n.id})" title="删除">🗑️</button>
            </div>
          </div>
          <div class="note-card-body">${this.escape(excerpt)}</div>
          <div class="note-card-meta">
            <span class="note-format-badge">${formatBadge}</span>
            <span>${hasAttachment}</span>
            <span>${this.formatDate(n.createdAt)}</span>
          </div>
        </div>
      `;
    }).join('');

    this.renderPagination('note-pagination', total, this.notePage, this.notePageSize, (p) => {
      this.notePage = p;
      this.renderNotes();
    });
  },

  renderTrash() {
    const container = document.getElementById('trash-list');
    if (!container) return;
    if (this.trashNotes.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">🗑️</div>
          <div class="empty-title">回收站为空</div>
          <p>已删除的笔记会显示在这里</p>
        </div>
      `;
      return;
    }
    container.innerHTML = this.trashNotes.map(note => {
      const deletedAt = note.deletedAt ? this.formatDate(note.deletedAt) : this.formatDate(note.updatedAt);
      const hasAttachment = (note.storedName || (note.attachmentCount || 0) > 0) ? '📎' : '';
      return `
        <div class="note-card">
          <div class="note-card-header">
            <div class="note-card-title">${this.escape(note.title)}</div>
            <div class="note-card-actions">
              <button class="btn btn-icon btn-sm" onclick="Passworder.restoreNote(${note.id})" title="恢复">↩️</button>
            </div>
          </div>
          <div class="note-card-body">${this.escape(note.body ? (note.body.length > 120 ? note.body.slice(0, 120) + '...' : note.body) : '')}</div>
          <div class="note-card-meta">
            <span>${note.bodyFormat === 'markdown' ? '📝 Markdown' : '📄 文本'}</span>
            <span>${hasAttachment}</span>
            <span>删除于 ${deletedAt}</span>
          </div>
        </div>
      `;
    }).join('');
  }
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderNotesMethods);
}
