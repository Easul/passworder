window.PassworderPreviewMethods = {
  pendingFiles: [],
  noteAttachments: {},

  async loadNoteAttachments(noteId) {
    try {
      const attachments = await this.api(`/files/${noteId}/attachments`) || [];
      this.noteAttachments[noteId] = attachments;
      this.renderNoteAttachmentList(this.notes.find(n => n.id === noteId));
    } catch (e) {
      console.error('Failed to load attachments:', e);
    }
  },

  renderNoteAttachmentList(note) {
    const container = document.getElementById('note-attachment-list');
    const attInfo = document.getElementById('note-attachment-info');
    if (!container) return;

    const attachments = note ? (this.noteAttachments[note.id] || []) : [];
    const pending = this.pendingFiles || [];
    const totalCount = attachments.length + pending.length + (note && note.storedName && attachments.length === 0 && pending.length === 0 ? 1 : 0);

    if (totalCount === 0) {
      container.innerHTML = '';
      attInfo.style.display = 'none';
      return;
    }

    let html = '<ul class="attachment-list">';
    attachments.forEach(att => {
      const icon = att.fileType === 'image' ? '🖼️' : att.fileType === 'archive' ? '📦' : att.fileType === 'document' ? '📄' : '📎';
      html += `
        <li class="attachment-item">
          <span class="attachment-icon">${icon}</span>
          <span class="attachment-name">${this.escape(att.originalName)}</span>
          <span class="attachment-size">${this.formatSize(att.sizeBytes)}</span>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.previewNoteAttachment(${att.id}, '${att.fileType}')" title="预览">👁️</button>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.downloadNoteAttachment(${att.id})" title="下载">⬇️</button>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.deleteNoteAttachment(${att.id})" title="删除">🗑️</button>
        </li>`;
    });

    pending.forEach((file, idx) => {
      const ext = file.name.split('.').pop().toLowerCase();
      const icon = ['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(ext) ? '🖼️' : ['zip'].includes(ext) ? '📦' : ['pdf', 'doc', 'docx', 'txt', 'csv', 'xls', 'xlsx'].includes(ext) ? '📄' : '📎';
      html += `
        <li class="attachment-item">
          <span class="attachment-icon">${icon}</span>
          <span class="attachment-name">${this.escape(file.name)}</span>
          <span class="attachment-size">${this.formatSize(file.size)}</span>
          <span style="color:var(--text-secondary);font-size:0.75rem">待上传</span>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.removePendingFile(${idx})" title="移除">❌</button>
        </li>`;
    });

    if (note && note.storedName && attachments.length === 0 && pending.length === 0) {
      html += `
        <li class="attachment-item">
          <span class="attachment-icon">🖼️</span>
          <span class="attachment-name">${this.escape(note.originalName || note.storedName)}</span>
          <span class="attachment-size">${this.formatSize(note.sizeBytes)}</span>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.previewAttachment(${note.id})" title="预览">👁️</button>
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.downloadAttachment(${note.id})" title="下载">⬇️</button>
        </li>`;
    }

    html += '</ul>';
    container.innerHTML = html;
    attInfo.textContent = `共 ${totalCount} 个附件`;
    attInfo.style.display = 'block';
  },

  removePendingFile(idx) {
    this.pendingFiles.splice(idx, 1);
    this.renderNoteAttachmentList(this.notes.find(n => n.id === this.currentNoteId));
  },

  getNoteAttachment(attachmentId) {
    for (const noteId in this.noteAttachments) {
      const att = this.noteAttachments[noteId].find(a => a.id === attachmentId);
      if (att) return att;
    }
    return null;
  },

  async previewNoteAttachment(attachmentId, fileType) {
    try {
      let att = this.getNoteAttachment(attachmentId);
      if (!att) {
        const noteId = this.currentPreviewNote?.id || this.currentNoteId;
        if (!noteId) return this.showToast('error', '无法获取附件信息');
        try {
          const attachments = await this.api(`/files/${noteId}/attachments`);
          if (attachments) {
            att = attachments.find(a => a.id === attachmentId);
            if (att && noteId) {
              this.noteAttachments[noteId] = attachments;
            }
          }
        } catch (e) {
          console.error('Failed to fetch attachment info:', e);
        }
      }
      if (!att) return this.showToast('error', '附件信息不存在，请刷新页面');
      const type = (fileType || '').toLowerCase();
      if (type.includes('image')) {
        const res = await fetch(`/api/note-attachments/${attachmentId}/preview`, { headers: this.token ? { 'Authorization': this.token } : {} });
        if (!res.ok) throw new Error('预览失败');
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const img = document.getElementById('preview-image');
        if (this.currentPreviewUrl) URL.revokeObjectURL(this.currentPreviewUrl);
        this.currentPreviewUrl = url;
        img.src = url;
        this.openModal('image-preview-modal');
      } else if (type.includes('zip') || type.includes('rar') || type.includes('7z')) {
        await this.previewArchive(att);
      } else if (type.includes('pdf')) {
        await this.previewPDF(attachmentId, att.originalName);
      } else if (type.includes('word') || type.includes('document') || type.includes('text') || type.includes('excel') || type.includes('sheet')) {
        await this.previewDocument(att);
      } else {
        this.downloadNoteAttachment(attachmentId);
      }
    } catch (e) {
      this.showToast('error', '预览失败');
    }
  },

  async previewPDF(attachmentId, originalName) {
    try {
      const res = await fetch(`/api/note-attachments/${attachmentId}/preview`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('预览失败');
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const iframe = document.getElementById('pdf-preview-frame');
      iframe.src = url;
      document.getElementById('pdf-preview-title').textContent = `📄 ${this.escape(originalName || 'PDF')}`;
      this.openModal('pdf-preview-modal');
    } catch (e) {
      this.showToast('error', 'PDF预览失败');
    }
  },

  async downloadNoteAttachment(attachmentId) {
    try {
      const res = await fetch(`/api/note-attachments/${attachmentId}`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('下载失败');
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      const att = this.getNoteAttachment(attachmentId);
      a.download = att ? att.originalName : 'attachment';
      a.click();
      URL.revokeObjectURL(url);
      this.showToast('success', '下载已开始');
    } catch (e) {
      this.showToast('error', '下载失败');
    }
  },

  async deleteNoteAttachment(attachmentId) {
    if (!confirm('确定删除此附件?')) return;
    try {
      await this.api(`/note-attachments/${attachmentId}`, { method: 'DELETE' });
      this.showToast('success', '删除成功');
      this.loadNoteAttachments(this.currentNoteId);
    } catch (e) {
      this.showToast('error', '删除失败');
    }
  },

  async previewArchive(att) {
    await this.loadJSZip();
    const res = await fetch(`/api/note-attachments/${att.id}`, { headers: this.token ? { 'Authorization': this.token } : {} });
    if (!res.ok) throw new Error('下载失败');
    const blob = await res.blob();
    const zip = await JSZip.loadAsync(blob);
    let treeHtml = '<ul class="archive-tree">';
    zip.forEach((relativePath, zipEntry) => {
      const isDir = zipEntry.dir;
      const icon = isDir ? '📁' : '📄';
      treeHtml += `<li>${icon} <span>${this.escape(relativePath)}</span> ${!isDir ? `<span class="archive-size">${this.formatSize(zipEntry._data.uncompressedSize)}</span>` : ''}</li>`;
    });
    treeHtml += '</ul>';
    document.getElementById('archive-preview-content').innerHTML = treeHtml;
    document.getElementById('archive-preview-title').textContent = `📦 ${this.escape(att.originalName)}`;
    this.openModal('archive-preview-modal');
  },

  loadJSZip() { return window.PassworderShared.loadJSZip(); },

  async previewDocument(att) {
    const ext = att.originalName ? att.originalName.split('.').pop().toLowerCase() : '';
    if (ext === 'pdf') {
      await this.previewPDF(att.id, att.originalName);
      return;
    }
    const res = await fetch(`/api/note-attachments/${att.id}/preview`, { headers: this.token ? { 'Authorization': this.token } : {} });
    if (!res.ok) throw new Error('预览失败');
    if (['doc', 'docx'].includes(ext)) {
      await this.loadMammoth();
      const arrayBuffer = await res.arrayBuffer();
      const result = await mammoth.convertToHtml({ arrayBuffer });
      document.getElementById('document-preview-content').innerHTML = result.value;
      document.getElementById('document-preview-title').textContent = `📝 ${this.escape(att.originalName)}`;
      this.openModal('document-preview-modal');
    } else if (['xls', 'xlsx'].includes(ext)) {
      await this.loadSheetJS();
      const arrayBuffer = await res.arrayBuffer();
      const workbook = XLSX.read(arrayBuffer, { type: 'array' });
      let html = '<div class="excel-preview">';
      workbook.SheetNames.forEach(sheetName => {
        const worksheet = workbook.Sheets[sheetName];
        const csv = XLSX.utils.sheet_to_csv(worksheet);
        html += `<h4>📊 ${this.escape(sheetName)}</h4><pre class="excel-sheet">${this.escape(csv)}</pre>`;
      });
      html += '</div>';
      document.getElementById('document-preview-content').innerHTML = html;
      document.getElementById('document-preview-title').textContent = `📊 ${this.escape(att.originalName)}`;
      this.openModal('document-preview-modal');
    } else {
      const text = await res.text();
      document.getElementById('document-preview-content').innerHTML = `<pre class="text-preview">${this.escape(text)}</pre>`;
      document.getElementById('document-preview-title').textContent = `📖 ${this.escape(att.originalName)}`;
      this.openModal('document-preview-modal');
    }
  },

  loadMammoth() { return window.PassworderShared.loadMammoth(); },
  loadSheetJS() { return window.PassworderShared.loadSheetJS(); },

  async previewAttachment(id) {
    const note = this.notes.find(n => n.id === id);
    if (!note || !note.storedName) return this.showToast('error', '没有附件可预览');
    try {
      const res = await fetch(`/api/files/${note.id}`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('下载失败');
      const blob = await res.blob();
      const ext = note.originalName ? note.originalName.split('.').pop().toLowerCase() : '';
      if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg'].includes(ext)) {
        const url = URL.createObjectURL(blob);
        const img = document.getElementById('preview-image');
        if (this.currentPreviewUrl) URL.revokeObjectURL(this.currentPreviewUrl);
        this.currentPreviewUrl = url;
        this.currentPreviewNote = note;
        img.src = url;
        this.openModal('image-preview-modal');
      } else if (ext === 'zip') {
        await this.loadJSZip();
        const zip = await JSZip.loadAsync(blob);
        let treeHtml = '<ul class="archive-tree">';
        zip.forEach((relativePath, zipEntry) => {
          const isDir = zipEntry.dir;
          const icon = isDir ? '📁' : '📄';
          treeHtml += `<li>${icon} <span>${this.escape(relativePath)}</span> ${!isDir ? `<span class="archive-size">${this.formatSize(zipEntry._data.uncompressedSize)}</span>` : ''}</li>`;
        });
        treeHtml += '</ul>';
        document.getElementById('archive-preview-content').innerHTML = treeHtml;
        document.getElementById('archive-preview-title').textContent = `📦 ${this.escape(note.originalName || '压缩包')}`;
        this.openModal('archive-preview-modal');
      } else if (ext === 'pdf') {
        const url = URL.createObjectURL(blob);
        const iframe = document.getElementById('pdf-preview-frame');
        iframe.src = url;
        document.getElementById('pdf-preview-title').textContent = `📄 ${this.escape(note.originalName || 'PDF')}`;
        this.openModal('pdf-preview-modal');
      } else if (['doc', 'docx'].includes(ext)) {
        await this.loadMammoth();
        const arrayBuffer = await blob.arrayBuffer();
        const result = await mammoth.convertToHtml({ arrayBuffer });
        document.getElementById('document-preview-content').innerHTML = result.value;
        document.getElementById('document-preview-title').textContent = `📝 ${this.escape(note.originalName || 'Word文档')}`;
        this.openModal('document-preview-modal');
      } else if (['xls', 'xlsx'].includes(ext)) {
        await this.loadSheetJS();
        const arrayBuffer = await blob.arrayBuffer();
        const workbook = XLSX.read(arrayBuffer, { type: 'array' });
        let html = '<div class="excel-preview">';
        workbook.SheetNames.forEach(sheetName => {
          const worksheet = workbook.Sheets[sheetName];
          const csv = XLSX.utils.sheet_to_csv(worksheet);
          html += `<h4>📊 ${this.escape(sheetName)}</h4><pre class="excel-sheet">${this.escape(csv)}</pre>`;
        });
        html += '</div>';
        document.getElementById('document-preview-content').innerHTML = html;
        document.getElementById('document-preview-title').textContent = `📊 ${this.escape(note.originalName || 'Excel')}`;
        this.openModal('document-preview-modal');
      } else if (['txt', 'csv', 'json', 'xml', 'md', 'log'].includes(ext)) {
        const text = await blob.text();
        document.getElementById('document-preview-content').innerHTML = `<pre class="text-preview">${this.escape(text)}</pre>`;
        document.getElementById('document-preview-title').textContent = `📖 ${this.escape(note.originalName || '文本文件')}`;
        this.openModal('document-preview-modal');
      } else {
        this.showToast('error', '不支持预览的格式，请下载后查看');
      }
    } catch (e) {
      this.showToast('error', '预览失败: ' + e.message);
    }
  },

  async downloadAttachment(id) {
    const note = this.notes.find(n => n.id === id);
    if (!note || !note.storedName) return;
    try {
      const res = await fetch(`/api/files/${note.id}`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('下载失败');
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = note.originalName || 'attachment';
      a.click();
      URL.revokeObjectURL(url);
      this.showToast('success', '下载已开始');
    } catch (e) {
      this.showToast('error', '下载失败');
    }
  },

  loadMarkdownLibs() {
    return window.PassworderShared.loadMarkdownLibs().then(() => {
      this.markdownLibLoaded = !!window.marked;
      this.dompurifyLoaded = !!window.DOMPurify;
    });
  },

  async viewNoteBody(id) {
    const note = this.notes.find(n => n.id === id);
    if (!note) return;
    this.currentPreviewNote = note;
    document.getElementById('note-view-title').textContent = note.title || '笔记内容';
    const remarksEl = document.getElementById('note-view-remarks');
    if (remarksEl) {
      const textSpan = remarksEl.querySelector('span:last-child');
      if (textSpan) textSpan.textContent = note.remarks || '';
      remarksEl.style.display = note.remarks ? 'flex' : 'none';
    }
    const contentDiv = document.getElementById('note-view-content');
    if (note.bodyFormat === 'markdown') {
      await this.loadMarkdownLibs();
      let html = window.marked.parse(note.body || '');
      if (this.dompurifyLoaded && window.DOMPurify) html = window.DOMPurify.sanitize(html);
      contentDiv.innerHTML = html;
    } else {
      contentDiv.innerHTML = `<pre>${this.escape(note.body || '')}</pre>`;
    }
    await this.renderNoteViewAttachments(note.id);
    this.openModal('note-view-modal');
  },

  copyNoteContent() {
    const note = this.currentPreviewNote;
    if (!note) return;
    this.copyText(note.body, '内容已复制');
  },

  async renderNoteViewAttachments(noteId) {
    const attachmentsContainer = document.getElementById('note-view-attachments');
    const attachmentsList = document.getElementById('note-view-attachments-list');
    if (!attachmentsContainer || !attachmentsList) return;

    try {
      const attachments = await this.api(`/files/${noteId}/attachments`);
      this.noteAttachments[noteId] = attachments || [];
      if (!attachments || attachments.length === 0) {
        attachmentsContainer.style.display = 'none';
        return;
      }

      attachmentsContainer.style.display = 'block';
      attachmentsList.innerHTML = attachments.map(att => {
        const icon = this.getAttachmentIcon(att.fileType);
        const size = this.formatFileSize(att.fileSize);
        return `
          <div class="note-view-attachment-item">
            <span class="note-view-attachment-icon">${icon}</span>
            <div class="note-view-attachment-info">
              <div class="note-view-attachment-name">${this.escape(att.originalName)}</div>
              <div class="note-view-attachment-meta">${size} · ${this.formatDate(att.createdAt)}</div>
            </div>
            <div class="note-view-attachment-actions">
              <button class="btn btn-sm btn-icon" onclick="Passworder.previewNoteAttachment(${att.id}, '${this.escapeAttr(att.fileType)}')" title="预览">👁️</button>
              <button class="btn btn-sm btn-icon" onclick="Passworder.downloadNoteAttachment(${att.id}, '${this.escapeAttr(att.originalName)}')" title="下载">⬇️</button>
            </div>
          </div>
        `;
      }).join('');
    } catch (e) {
      attachmentsContainer.style.display = 'none';
    }
  },

  getAttachmentIcon(fileType) {
    if (!fileType) return '📄';
    const type = fileType.toLowerCase();
    if (type.includes('image')) return '🖼️';
    if (type.includes('pdf')) return '📃';
    if (type.includes('zip') || type.includes('rar') || type.includes('7z')) return '📦';
    if (type.includes('word') || type.includes('document')) return '📝';
    if (type.includes('excel') || type.includes('sheet')) return '📊';
    if (type.includes('text')) return '📖';
    if (type.includes('code') || type.includes('json') || type.includes('xml')) return '💻';
    return '📄';
  },

  formatFileSize(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  async openMarkdownEditor(id) {
    const note = this.notes.find(n => n.id === id);
    if (!note) return;
    await this.loadMarkdownLibs();
    this.currentMarkdownNote = note;
    document.getElementById('markdown-editor').value = note.body || '';
    this.renderMarkdownPreview();
    this.openModal('markdown-modal');
  },

  renderMarkdownPreview() {
    const text = document.getElementById('markdown-editor').value || '';
    const preview = document.getElementById('markdown-preview');
    if (!this.markdownLibLoaded || !window.marked) {
      preview.innerHTML = '<p style="color:var(--text-secondary)">正在加载编辑器...</p>';
      return;
    }
    let html = window.marked.parse(text);
    if (this.dompurifyLoaded && window.DOMPurify) html = window.DOMPurify.sanitize(html);
    preview.innerHTML = html;
  },

  async saveMarkdown() {
    const body = document.getElementById('markdown-editor').value;
    if (!this.currentMarkdownNote) return;
    try {
      await this.api(`/files/${this.currentMarkdownNote.id}`, {
        method: 'PUT',
        body: { title: this.currentMarkdownNote.title, body: body, bodyFormat: 'markdown' }
      });
      this.showToast('success', '保存成功');
      this.closeModal('markdown-modal');
      await this.loadNotes();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  downloadPreviewImage() {
    if (this.currentPreviewNote) {
      this.downloadAttachment(this.currentPreviewNote.id);
    } else {
      this.closeModal('image-preview-modal');
    }
  },
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderAuthMethods);
  Object.assign(window.Passworder, window.PassworderPreviewMethods);
}
