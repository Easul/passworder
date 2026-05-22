window.PassworderNoteEditorMethods = {
  showNoteModal(id) {
    const note = id ? this.notes.find(n => n.id === id) : null;
    document.getElementById('note-modal-title').textContent = note ? '编辑笔记' : '新建笔记';
    document.getElementById('note-title').value = note ? note.title : '';
    document.getElementById('note-remarks').value = note ? (note.remarks || '') : '';
    const format = note ? note.bodyFormat : 'text';
    document.querySelectorAll('input[name="note-format"]').forEach(r => { r.checked = r.value === format; });
    this.currentNoteId = id || null;
    this.setNoteEditFullscreen(false);
    this.openModal('note-modal');
    if (format === 'markdown') {
      this.initVditor(note ? note.body : '');
    } else {
      this.destroyVditor();
      document.getElementById('note-body-vditor').classList.add('hidden');
      document.getElementById('note-body').classList.remove('hidden');
      document.getElementById('note-body').value = note ? note.body : '';
    }
    document.getElementById('note-file').value = '';
    this.pendingFiles = [];
    this.renderNoteAttachmentList(note);
    if (note) this.loadNoteAttachments(note.id);
  },

  setNoteEditFullscreen(enabled) {
    const modal = document.querySelector('#note-modal .modal');
    const toggle = document.getElementById('note-edit-fullscreen-toggle');
    if (!modal) return;
    modal.classList.toggle('modal-fullscreen', !!enabled);
    if (toggle) {
      toggle.textContent = enabled ? '🗗' : '⛶';
      toggle.title = enabled ? '恢复原状' : '撑满编辑区域';
    }
  },

  toggleNoteEditFullscreen() {
    const modal = document.querySelector('#note-modal .modal');
    if (!modal) return;
    this.setNoteEditFullscreen(!modal.classList.contains('modal-fullscreen'));
  },

  async initVditor(initialValue) {
    await this.loadVditor();
    const textarea = document.getElementById('note-body');
    const mount = document.getElementById('note-body-vditor');
    textarea.classList.add('hidden');
    mount.classList.remove('hidden');
    if (this.vditor) {
      this.vditor.setValue(initialValue || '');
      this.applyToolbarVisibility();
      return;
    }
    this.vditor = new window.Vditor('note-body-vditor', {
      mode: 'ir', height: 320, value: initialValue || '', cache: { enable: false }, counter: { enable: false },
      toolbarConfig: { pin: false, hide: true },
      toolbar: ['headings','bold','italic','strike','|','list','ordered-list','check','|','quote','link','table','|','code','inline-code','|','undo','redo','|','fullscreen'],
      preview: { delay: 0, markdown: { toc: false } }
    });
    this.vditorToolbarVisible = false;
    this.applyToolbarVisibility();
  },

  toggleVditorToolbar() {
    this.vditorToolbarVisible = !this.vditorToolbarVisible;
    this.applyToolbarVisibility();
    const btn = document.getElementById('vditor-toolbar-toggle');
    if (btn) {
      btn.style.opacity = this.vditorToolbarVisible ? '1' : '0.6';
      btn.title = this.vditorToolbarVisible ? '隐藏工具栏' : '显示工具栏';
    }
  },

  applyToolbarVisibility() {
    const container = document.getElementById('note-body-container');
    if (container) container.classList.toggle('vditor-toolbar-hidden', !this.vditorToolbarVisible);
    if (this.vditor && this.vditor.vditor && this.vditor.vditor.toolbar) {
      const toolbarEl = this.vditor.vditor.toolbar.element;
      if (toolbarEl) toolbarEl.style.display = this.vditorToolbarVisible ? 'flex' : 'none';
    }
  },

  destroyVditor() {
    if (this.vditor) {
      try { this.vditor.destroy(); } catch (e) {}
      this.vditor = null;
    }
  },

  onNoteFormatChange() {
    const format = document.querySelector('input[name="note-format"]:checked').value;
    const note = this.currentNoteId ? this.notes.find(n => n.id === this.currentNoteId) : null;
    if (format === 'markdown') {
      const currentBody = note ? note.body : document.getElementById('note-body').value;
      this.initVditor(currentBody || '');
    } else {
      const currentBody = this.vditor ? this.vditor.getValue() : (note ? note.body : '');
      this.destroyVditor();
      document.getElementById('note-body-vditor').classList.add('hidden');
      document.getElementById('note-body').classList.remove('hidden');
      document.getElementById('note-body').value = currentBody || '';
    }
  },

  loadVditor() { return window.PassworderShared.loadVditor(); },

  async saveNote() {
    const title = document.getElementById('note-title').value.trim();
    const remarks = document.getElementById('note-remarks').value.trim();
    const bodyFormat = document.querySelector('input[name="note-format"]:checked').value;
    let body = bodyFormat === 'markdown' && this.vditor ? this.vditor.getValue() : document.getElementById('note-body').value;
    if (!title) return this.showToast('error', '标题不能为空');
    try {
      let noteId = this.currentNoteId;
      if (this.currentNoteId) {
        await this.api(`/files/${this.currentNoteId}`, { method: 'PUT', body: { title, remarks, body, bodyFormat } });
        this.showToast('success', '保存成功');
      } else {
        const formData = new FormData();
        formData.append('title', title);
        formData.append('remarks', remarks);
        formData.append('body', body);
        formData.append('bodyFormat', bodyFormat);
        const res = await fetch('/api/files', { method: 'POST', headers: this.token ? { 'Authorization': this.token } : {}, body: formData });
        if (!res.ok) {
          const data = await res.json().catch(() => null);
          throw new Error(data?.message || '创建失败');
        }
        const result = await res.json();
        noteId = result.data ? result.data.id : null;
        this.showToast('success', '创建成功');
      }
      if (noteId && this.pendingFiles.length > 0) {
        const attFormData = new FormData();
        for (let i = 0; i < this.pendingFiles.length; i++) attFormData.append('files', this.pendingFiles[i]);
        const attRes = await fetch(`/api/files/${noteId}/attachments`, { method: 'POST', headers: this.token ? { 'Authorization': this.token } : {}, body: attFormData });
        if (!attRes.ok) console.error('Attachment upload failed');
        this.pendingFiles = [];
      }
      this.destroyVditor();
      this.closeModal('note-modal');
      await this.loadNotes();
      this.showNotes(false);
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async deleteNote(id) {
    if (!confirm('确定要将此笔记移入回收站吗？')) return;
    try {
      await this.api(`/notes/${id}`, { method: 'DELETE' });
      this.showToast('success', '已移入回收站');
      await this.loadNotes();
    } catch (e) {
      this.showToast('error', '删除失败');
    }
  },

  async restoreNote(id) {
    try {
      await this.api(`/notes/${id}/restore`, { method: 'POST' });
      this.showToast('success', '笔记已恢复');
      await Promise.all([this.loadNotes(), this.loadTrash()]);
    } catch (e) {
      this.showToast('error', e.message || '恢复失败');
    }
  },

  async emptyTrash() {
    if (!confirm('确定清空回收站吗？此操作会永久删除笔记及其文件，无法恢复。')) return;
    try {
      await this.api('/notes/trash', { method: 'DELETE' });
      this.showToast('success', '回收站已清空');
      this.trashNotes = [];
      this.renderTrash();
      this.closeModal('trash-modal');
      await this.loadNotes();
    } catch (e) {
      this.showToast('error', e.message || '清空失败');
    }
  }
};

if (window.Passworder) Object.assign(window.Passworder, window.PassworderNoteEditorMethods);
