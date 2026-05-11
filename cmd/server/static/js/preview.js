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
          <button type="button" class="btn btn-icon btn-sm" onclick="Passworder.removePendingFile(${idx})" title="移除">✕</button>
        </li>`;
    });

    if (note && note.storedName && attachments.length === 0 && pending.length === 0) {
      const icon = note.fileType === 'image' ? '🖼️' : note.fileType === 'archive' ? '📦' : note.fileType === 'document' ? '📄' : '📎';
      html += `
        <li class="attachment-item">
          <span class="attachment-icon">${icon}</span>
          <span class="attachment-name">${this.escape(note.originalName)}</span>
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
      const att = this.getNoteAttachment(attachmentId);
      if (!att) return this.showToast('error', '附件信息不存在，请刷新页面');
      if (fileType === 'image') {
        const res = await fetch(`/api/note-attachments/${attachmentId}/preview`, { headers: this.token ? { 'Authorization': this.token } : {} });
        if (!res.ok) throw new Error('预览失败');
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const img = document.getElementById('preview-image');
        if (this.currentPreviewUrl) URL.revokeObjectURL(this.currentPreviewUrl);
        this.currentPreviewUrl = url;
        img.src = url;
        this.openModal('image-preview-modal');
      } else if (fileType === 'archive') {
        await this.previewArchive(att);
      } else if (fileType === 'document') {
        const ext = att.originalName ? att.originalName.split('.').pop().toLowerCase() : '';
        if (ext === 'pdf') await this.previewPDF(attachmentId, att.originalName);
        else await this.previewDocument(att);
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
      const disposition = res.headers.get('content-disposition');
      let filename = `attachment-${attachmentId}`;
      if (disposition) {
        const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/);
        if (utf8Match) filename = decodeURIComponent(utf8Match[1]);
        else {
          const asciiMatch = disposition.match(/filename="([^"]+)"/);
          if (asciiMatch) filename = asciiMatch[1];
        }
      }
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      this.showToast('error', '下载失败');
    }
  },

  async deleteNoteAttachment(attachmentId) {
    if (!confirm('确定要删除此附件吗？')) return;
    try {
      await this.api(`/note-attachments/${attachmentId}`, { method: 'DELETE' });
      this.showToast('success', '附件已删除');
      if (this.currentNoteId) await this.loadNoteAttachments(this.currentNoteId);
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async previewAttachment(id) {
    const note = this.notes.find(n => n.id === id);
    if (!note || !note.storedName) return;
    this.currentPreviewNote = note;
    if (note.fileType === 'image') {
      try {
        const res = await fetch(`/api/files/${note.id}/preview`, { headers: this.token ? { 'Authorization': this.token } : {} });
        if (!res.ok) throw new Error('预览失败');
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const img = document.getElementById('preview-image');
        if (this.currentPreviewUrl) URL.revokeObjectURL(this.currentPreviewUrl);
        this.currentPreviewUrl = url;
        img.src = url;
        this.openModal('image-preview-modal');
      } catch (e) {
        this.showToast('error', '图片加载失败');
      }
    } else if (note.fileType === 'archive') {
      await this.previewArchive(note);
    } else if (note.fileType === 'document') {
      await this.previewDocument(note);
    }
  },

  async previewArchive(att) {
    try {
      const res = await fetch(`/api/note-attachments/${att.id}`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('下载失败');
      const blob = await res.blob();
      const ext = att.originalName ? att.originalName.split('.').pop().toLowerCase() : '';
      if (ext === 'zip') {
        await this.loadJSZip();
        const zip = await JSZip.loadAsync(blob);
        let treeHtml = '<ul class="archive-tree">';
        zip.forEach((relativePath, zipEntry) => {
          const isDir = zipEntry.dir;
          const icon = isDir ? '📁' : '📄';
          const name = relativePath.split('/').pop() || relativePath;
          treeHtml += `<li>${icon} <span>${this.escape(name)}</span> ${!isDir ? `<span class="archive-size">${this.formatSize(zipEntry._data.uncompressedSize)}</span>` : ''}</li>`;
        });
        treeHtml += '</ul>';
        document.getElementById('archive-preview-content').innerHTML = treeHtml;
      } else {
        document.getElementById('archive-preview-content').innerHTML = '<p style="color:var(--text-secondary)">暂仅支持 .zip 格式预览</p>';
      }
      document.getElementById('archive-preview-title').textContent = `📦 ${this.escape(att.originalName || '压缩包')}`;
      this.openModal('archive-preview-modal');
    } catch (e) {
      this.showToast('error', '压缩包解析失败');
    }
  },

  async previewDocument(att) {
    try {
      const res = await fetch(`/api/note-attachments/${att.id}`, { headers: this.token ? { 'Authorization': this.token } : {} });
      if (!res.ok) throw new Error('下载失败');
      const blob = await res.blob();
      const ext = att.originalName ? att.originalName.split('.').pop().toLowerCase() : '';
      const container = document.getElementById('document-preview-content');
      const titleEl = document.getElementById('document-preview-title');
      titleEl.textContent = `📄 ${this.escape(att.originalName || '文档')}`;
      if (ext === 'docx') {
        await this.loadMammoth();
        const arrayBuffer = await blob.arrayBuffer();
        const result = await mammoth.convertToHtml({ arrayBuffer });
        container.innerHTML = `<div class="document-html-content">${result.value}</div>`;
      } else if (ext === 'xlsx' || ext === 'xls') {
        await this.loadSheetJS();
        const arrayBuffer = await blob.arrayBuffer();
        const workbook = XLSX.read(arrayBuffer, { type: 'array' });
        let html = '';
        workbook.SheetNames.forEach(sheetName => {
          const sheet = workbook.Sheets[sheetName];
          html += `<h4 class="sheet-name">📊 ${this.escape(sheetName)}</h4>`;
          html += XLSX.utils.sheet_to_html(sheet, { id: `sheet-${sheetName}` });
        });
        container.innerHTML = `<div class="document-html-content">${html}</div>`;
      } else if (ext === 'txt' || ext === 'csv') {
        const text = await blob.text();
        container.innerHTML = `<pre class="document-text-content">${this.escape(text)}</pre>`;
      } else {
        container.innerHTML = '<p style="color:var(--text-secondary)">该格式暂不支持预览，请下载后查看</p>';
      }
      this.openModal('document-preview-modal');
    } catch (e) {
      this.showToast('error', '文档解析失败');
    }
  },

  loadJSZip() { return window.PassworderShared.loadJSZip(); },
  loadMammoth() { return window.PassworderShared.loadMammoth(); },
  loadSheetJS() { return window.PassworderShared.loadSheetJS(); },

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
    document.getElementById('note-view-title').textContent = note.title || '笔记内容';
    const contentDiv = document.getElementById('note-view-content');
    if (note.bodyFormat === 'markdown') {
      await this.loadMarkdownLibs();
      let html = window.marked.parse(note.body || '');
      if (this.dompurifyLoaded && window.DOMPurify) html = window.DOMPurify.sanitize(html);
      contentDiv.innerHTML = html;
    } else {
      contentDiv.innerHTML = `<pre style="white-space:pre-wrap;word-break:break-word;">${this.escape(note.body || '')}</pre>`;
    }
    this.openModal('note-view-modal');
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
