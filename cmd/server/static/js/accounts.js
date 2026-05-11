window.PassworderAccountsMethods = {
  async loadCategories() {
    this.categories = await this.api('/categories') || [];
  },

  async loadAccounts() {
    this.accounts = await this.api('/accounts') || [];
  },

  renderCategoryFilter() {
    const select = document.getElementById('filter-category');
    const options = this.categories.map(c => `<option value="${c.id}">${c.name}</option>`).join('');
    select.innerHTML = `<option value="">所有分类</option>${options}`;
  },

  renderAccounts() {
    const search = document.getElementById('search-input').value.toLowerCase();
    const categoryId = document.getElementById('filter-category').value;
    const reminderFilter = document.getElementById('filter-reminder')?.value;

    let filtered = this.accounts;
    if (search) {
      filtered = filtered.filter(a =>
        a.title.toLowerCase().includes(search) ||
        a.username.toLowerCase().includes(search) ||
        a.website.toLowerCase().includes(search) ||
        (a.notes && a.notes.toLowerCase().includes(search)) ||
        (a.registrationNotes && a.registrationNotes.toLowerCase().includes(search))
      );
    }
    if (categoryId) {
      filtered = filtered.filter(a => a.categoryId === Number.parseInt(categoryId, 10));
    }
    if (reminderFilter) {
      filtered = filtered.filter(a => a.reminderStatus === reminderFilter);
    }

    filtered.sort((a, b) => {
      const field = this.accountSortField === 'createdAt' ? 'createdAt' : 'registrationTime';
      const valA = a[field] || 0;
      const valB = b[field] || 0;
      return this.accountSortOrder === 'desc' ? valB - valA : valA - valB;
    });

    const container = document.getElementById('account-list');
    const paginationContainer = document.getElementById('account-pagination');
    if (filtered.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">📭</div>
          <div class="empty-title">暂无账号</div>
          <p>点击右上角添加按钮创建</p>
        </div>
      `;
      if (paginationContainer) paginationContainer.innerHTML = '';
      return;
    }

    const total = filtered.length;
    const start = (this.accountPage - 1) * this.accountPageSize;
    const pageItems = filtered.slice(start, start + this.accountPageSize);

    const statusLabels = { sent: '✅ 已提醒', pending: '⏳ 待提醒', none: '❌ 无提醒' };
    const statusColors = { sent: 'var(--success)', pending: 'var(--warning)', none: 'var(--text-secondary)' };

    container.innerHTML = pageItems.map(acc => {
      const cat = this.categories.find(c => c.id === acc.categoryId);
      const statusLabel = statusLabels[acc.reminderStatus] || '';
      const statusColor = statusColors[acc.reminderStatus] || 'transparent';
      const accountStatusLabel = acc.status === 'active' ? '激活' : '禁用';
      const accountStatusColor = acc.status === 'active' ? '#4caf50' : '#9e9e9e';
      return `
        <div class="account-item" onclick="Passworder.viewAccount(${acc.id})">
          <div class="account-info">
            <div class="account-title">
              ${this.escape(acc.title)}${acc.isFavorite ? ' ⭐' : ''}
              <span class="status-badge" style="background:${accountStatusColor};color:#fff;padding:2px 6px;border-radius:4px;font-size:0.7rem;margin-left:6px">${accountStatusLabel}</span>
            </div>
            <div class="account-meta">
              ${cat ? `<span class="category-tag">${this.escape(cat.name)}</span>` : ''}
              <span>${this.escape(acc.username || '-')}</span>
              ${statusLabel ? `<span class="reminder-tag" style="background:${statusColor};color:#fff;padding:2px 6px;border-radius:4px;font-size:0.75rem">${statusLabel}</span>` : ''}
            </div>
          </div>
          <div class="account-actions" onclick="event.stopPropagation()">
            <button class="btn btn-icon btn-sm" onclick="Passworder.copyPassword(${acc.id})" title="复制密码">📋</button>
            <button class="btn btn-icon btn-sm" onclick="Passworder.editAccount(${acc.id})" title="编辑">✏️</button>
            <button class="btn btn-icon btn-sm" onclick="Passworder.deleteAccount(${acc.id})" title="删除">🗑️</button>
          </div>
        </div>
      `;
    }).join('');

    this.renderPagination('account-pagination', total, this.accountPage, this.accountPageSize, (p) => {
      this.accountPage = p;
      this.renderAccounts();
    });
  },

  onAccountSortChange() {
    const sortSelect = document.getElementById('account-sort');
    if (!sortSelect) return;
    const [field, order] = sortSelect.value.split('-');
    this.accountSortField = field;
    this.accountSortOrder = order;
    this.accountPage = 1;
    this.renderAccounts();
  },

  async showNewAccount() {
    await this.loadSenderSettings();
    this.currentAccount = null;
    this.renderAccountForm();
    this.openModal('account-modal');
  },

  async editAccount(id) {
    await this.loadSenderSettings();
    this.currentAccount = await this.api(`/accounts/${id}`);
    try {
      const data = await this.api(`/accounts/${id}/password`);
      this.currentAccount.password = data.password;
    } catch (e) {
      this.showToast('error', '密码加载失败');
    }
    this.renderAccountForm();
    this.openModal('account-modal');
  },

  viewAccount(id) {
    this.currentAccount = this.accounts.find(a => a.id === id);
    this.renderAccountView();
    this.openModal('view-modal');
  },

  renderAccountForm() {
    const acc = this.currentAccount || {};
    const selectedCategoryId = acc.categoryId || (this.categories[0] && this.categories[0].id) || "";
    const catOptions = this.categories.length
      ? this.categories.map(c => `<option value="${c.id}" ${selectedCategoryId === c.id ? 'selected' : ''}>${c.name}</option>`).join('')
      : '<option value="">请先创建分类</option>';

    document.getElementById('account-modal-title').textContent = acc.id ? '编辑账号' : '新建账号';
    document.getElementById('account-form').innerHTML = `
      <div class="form-group"><label class="form-label">分类</label><select class="form-select" id="acc-category">${catOptions}</select></div>
      <div class="form-group"><label class="form-label">标题 *</label><input type="text" class="form-input" id="acc-title" value="${this.escapeAttr(acc.title)}" placeholder="例如：GitHub"></div>
      <div class="form-group"><label class="form-label">网站</label><input type="text" class="form-input" id="acc-website" value="${this.escapeAttr(acc.website)}" placeholder="https://..."></div>
      <div class="form-group"><label class="form-label">用户名</label><input type="text" class="form-input" id="acc-username" value="${this.escapeAttr(acc.username)}"></div>
      <div class="form-group"><label class="form-label">密码</label><div class="input-group"><input type="password" class="form-input password-input" id="acc-password" value="${this.escapeAttr(acc.password)}" autocomplete="new-password"><button type="button" class="btn btn-secondary" onclick="Passworder.togglePassword('acc-password')">👁️</button><button type="button" class="btn btn-secondary" onclick="Passworder.openGenerator()">🎲</button></div></div>
      <div class="form-group"><label class="form-label">邮箱</label><input type="email" class="form-input" id="acc-email" value="${this.escapeAttr(acc.email)}"></div>
      <div class="form-group"><label class="form-label">注册时间</label><div style="display:flex;gap:8px;align-items:center"><input type="date" class="form-input" id="acc-registration-date" value="${acc.registrationTime ? this.formatDateInput(acc.registrationTime) : ''}" style="flex:1"><input type="time" class="form-input" id="acc-registration-time" value="${acc.registrationTime ? this.formatTimeInput(acc.registrationTime) : ''}" step="60" style="flex:1"></div></div>
      <div class="form-group"><label class="form-label">注册备注</label><textarea class="form-textarea" id="acc-registration-notes"></textarea></div>
      <div class="form-group"><label class="form-label">手机</label><input type="tel" class="form-input" id="acc-phone" value="${this.escapeAttr(acc.phone)}"></div>
      <div class="form-group"><label class="form-label">备注</label><textarea class="form-textarea" id="acc-notes"></textarea></div>
      <div class="form-group"><label class="form-label">⏰ 登录提醒</label><div style="display:flex;gap:8px;align-items:center"><input type="date" class="form-input" id="acc-remind-date" value="${acc.remindAt ? this.formatDateInput(acc.remindAt) : ''}" style="flex:1"><input type="time" class="form-input" id="acc-remind-at" value="${acc.remindAt ? this.formatTimeInput(acc.remindAt) : ''}" step="60" style="flex:1"></div><p class="form-hint" style="color:var(--text-secondary);font-size:0.8125rem;margin-top:4px">设置日期和时间后，系统会提醒您定期登录此账号</p></div>
      <div class="form-group" id="reminder-period-group" style="display:none"><label class="form-label">重复周期</label><div style="display:flex;gap:8px;align-items:center"><select class="form-select" id="acc-period-type" style="flex:1" onchange="Passworder.onPeriodTypeChange()"><option value="">不重复</option><option value="yearly">每年</option><option value="monthly">每月</option><option value="weekly">每周</option><option value="daily">每天</option><option value="hourly">每小时</option><option value="days">每N天</option></select><input type="number" class="form-input" id="acc-period-value" value="1" min="1" max="365" style="width:80px;display:none" placeholder="间隔"></div><p class="form-hint" id="period-hint" style="color:var(--text-secondary);font-size:0.8125rem;margin-top:4px"></p></div>
      <div class="form-group"><label class="form-label">提醒邮箱</label><div style="display:flex;gap:8px;align-items:center"><select class="form-select" id="acc-reminder-email-select" style="flex:1" onchange="Passworder.onReminderEmailChange()"><option value="">自定义</option><option value="${this.escapeAttr(acc.email || '')}">${this.escape(acc.email || '') || '账号邮箱'}</option></select><input type="email" class="form-input" id="acc-reminder-email" value="${this.escapeAttr(acc.reminderEmail || acc.email || '')}" placeholder="接收提醒的邮箱地址" style="flex:2"></div></div>
      <div class="form-group"><label class="form-label">账号状态</label><select class="form-select" id="acc-status"><option value="active" ${acc.status === 'active' || !acc.status ? 'selected' : ''}>有效</option><option value="inactive" ${acc.status === 'inactive' ? 'selected' : ''}>无效</option></select></div>
      <div class="form-group"><label class="option-checkbox"><input type="checkbox" id="acc-favorite" ${acc.isFavorite ? 'checked' : ''}><span>收藏</span></label></div>
    `;

    document.getElementById('account-modal-footer').innerHTML = `
      <button class="btn btn-secondary" onclick="Passworder.closeModal('account-modal')">取消</button>
      <button class="btn btn-primary" onclick="Passworder.saveAccount()">保存</button>
    `;

    const periodTypeSelect = document.getElementById('acc-period-type');
    const periodValueInput = document.getElementById('acc-period-value');
    const registrationNotesInput = document.getElementById('acc-registration-notes');
    const notesInput = document.getElementById('acc-notes');
    if (registrationNotesInput) registrationNotesInput.value = acc.registrationNotes || '';
    if (notesInput) notesInput.value = acc.notes || '';
    if (periodTypeSelect && acc.reminderPeriodType) periodTypeSelect.value = acc.reminderPeriodType;
    if (periodValueInput && acc.reminderPeriodValue) periodValueInput.value = acc.reminderPeriodValue;
    this.onPeriodTypeChange();

    const remindAtInput = document.getElementById('acc-remind-at');
    const periodGroup = document.getElementById('reminder-period-group');
    if (remindAtInput && periodGroup) {
      const showPeriod = () => { periodGroup.style.display = remindAtInput.value ? 'block' : 'none'; };
      showPeriod();
      remindAtInput.addEventListener('change', showPeriod);
    }
    this.populateReminderEmailSelect(acc);
  },

  populateReminderEmailSelect(acc) {
    const select = document.getElementById('acc-reminder-email-select');
    const input = document.getElementById('acc-reminder-email');
    if (!select) return;
    const emails = [];
    if (acc.email) emails.push(acc.email);
    const settingEmails = (this.senderSettings['reminder_emails'] || '').split('\n').map(s => s.trim()).filter(Boolean);
    emails.push(...settingEmails);
    const uniqueEmails = [...new Set(emails)];
    const currentValue = acc.reminderEmail || '';
    const defaultValue = settingEmails[0] || acc.email || '';
    select.innerHTML = '<option value="">自定义</option>' + uniqueEmails.map(e => `<option value="${this.escapeAttr(e)}" ${e === currentValue ? 'selected' : ''}>${this.escape(e)}</option>`).join('');
    if (!acc.id && defaultValue) {
      input.value = defaultValue;
      if (uniqueEmails.includes(defaultValue)) {
        select.value = defaultValue;
        input.style.display = 'none';
      } else {
        input.style.display = 'block';
      }
    } else if (currentValue && uniqueEmails.includes(currentValue)) {
      input.style.display = 'none';
    } else {
      input.style.display = 'block';
    }
  },

  onReminderEmailChange() {
    const select = document.getElementById('acc-reminder-email-select');
    const input = document.getElementById('acc-reminder-email');
    if (!select || !input) return;
    if (select.value) {
      input.value = select.value;
      input.style.display = 'none';
    } else {
      input.style.display = 'block';
      input.focus();
    }
  },

  renderAccountView() {
    const acc = this.currentAccount;
    const cat = this.categories.find(c => c.id === acc.categoryId);
    const accountStatusLabel = acc.status === 'active' ? '激活' : '禁用';
    const accountStatusColor = acc.status === 'active' ? '#4caf50' : '#9e9e9e';
    document.getElementById('view-modal-body').innerHTML = `
      <div class="form-group"><label class="form-label">标题</label><div class="form-input" style="background:var(--bg)">${this.escape(acc.title)}</div></div>
      <div class="form-group"><label class="form-label">状态</label><div class="form-input" style="background:var(--bg)"><span class="status-badge" style="background:${accountStatusColor};color:#fff;padding:4px 8px;border-radius:4px;font-size:0.85rem">${accountStatusLabel}</span></div></div>
      <div class="form-group"><label class="form-label">分类</label><div class="form-input" style="background:var(--bg)">${cat ? this.escape(cat.name) : '-'}</div></div>
      ${acc.website ? `<div class="form-group"><label class="form-label">网站</label><div class="input-group"><div class="form-input" style="flex:1;background:var(--bg)">${this.escape(acc.website)}</div><a href="${acc.website}" target="_blank" class="btn btn-secondary">打开</a></div></div>` : ''}
      <div class="form-group"><label class="form-label">用户名</label><div class="input-group"><div class="form-input" style="flex:1;background:var(--bg)">${this.escape(acc.username || '-')}</div>${acc.username ? `<button class="btn btn-secondary" onclick="Passworder.copyText('${this.escape(acc.username)}')">复制</button>` : ''}</div></div>
      <div class="form-group"><label class="form-label">密码</label><div class="input-group"><input type="password" class="form-input password-input" id="view-password" value="${this.escapeAttr(acc.password)}" readonly><button class="btn btn-secondary" onclick="Passworder.togglePassword('view-password')">👁️</button><button class="btn btn-secondary" onclick="Passworder.copyText('${this.escape(acc.password || '')}')">复制</button></div></div>
      ${acc.email ? `<div class="form-group"><label class="form-label">邮箱</label><div class="input-group"><div class="form-input" style="flex:1;background:var(--bg)">${this.escape(acc.email)}</div><button class="btn btn-secondary" onclick="Passworder.copyText('${this.escape(acc.email)}')">复制</button></div></div>` : ''}
      ${acc.reminderEmail ? `<div class="form-group"><label class="form-label">默认收信邮箱</label><div class="form-input" style="background:var(--bg)">${this.escape(acc.reminderEmail)}</div></div>` : ''}
      ${acc.phone ? `<div class="form-group"><label class="form-label">手机</label><div class="input-group"><div class="form-input" style="flex:1;background:var(--bg)">${this.escape(acc.phone)}</div><button class="btn btn-secondary" onclick="Passworder.copyText('${this.escape(acc.phone)}')">复制</button></div></div>` : ''}
      ${acc.notes ? `<div class="form-group"><label class="form-label">备注</label><div class="form-textarea" style="background:var(--bg);min-height:auto">${this.escape(acc.notes)}</div></div>` : ''}
      ${acc.remindAt ? `<div class="form-group"><label class="form-label">⏰ 登录提醒</label><div class="form-input" style="background:var(--bg)">${this.formatDate(acc.remindAt)}</div></div>` : ''}
    `;
  },

  async saveAccount() {
    const categoryValue = document.getElementById('acc-category').value;
    const body = {
      categoryId: Number.parseInt(categoryValue, 10),
      title: document.getElementById('acc-title').value.trim(),
      website: document.getElementById('acc-website').value.trim(),
      username: document.getElementById('acc-username').value.trim(),
      password: document.getElementById('acc-password').value,
      email: document.getElementById('acc-email').value.trim(),
      reminderEmail: document.getElementById('acc-reminder-email').value.trim(),
      remindAt: 0,
      reminderPeriodType: '',
      reminderPeriodValue: 0,
      registrationTime: 0,
      registrationNotes: document.getElementById('acc-registration-notes').value.trim(),
      phone: document.getElementById('acc-phone').value.trim(),
      notes: document.getElementById('acc-notes').value.trim(),
      isFavorite: document.getElementById('acc-favorite').checked ? 1 : 0,
      status: document.getElementById('acc-status').value || 'active'
    };
    const remindAtValue = this.combineDateAndTime(document.getElementById('acc-remind-date')?.value, document.getElementById('acc-remind-at')?.value);
    body.remindAt = remindAtValue ? Math.floor(new Date(remindAtValue).getTime() / 1000) : 0;
    const registrationTimeValue = this.combineDateAndTime(document.getElementById('acc-registration-date')?.value, document.getElementById('acc-registration-time')?.value);
    body.registrationTime = registrationTimeValue ? Math.floor(new Date(registrationTimeValue).getTime() / 1000) : 0;
    const periodType = document.getElementById('acc-period-type');
    const periodValue = document.getElementById('acc-period-value');
    if (periodType && periodValue) {
      body.reminderPeriodType = periodType.value || '';
      body.reminderPeriodValue = parseInt(periodValue.value, 10) || 0;
    }
    if (!body.categoryId || Number.isNaN(body.categoryId)) return this.showToast('error', '请选择分类');
    if (!body.title) return this.showToast('error', '请输入标题');
    if (!this.currentAccount && !body.password) return this.showToast('error', '请输入密码');
    if (body.website && !/^https?:\/\/.+/i.test(body.website)) return this.showToast('error', '网站格式错误，请以 http:// 或 https:// 开头');
    if (body.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(body.email)) return this.showToast('error', '邮箱格式错误');
    if (body.reminderEmail && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(body.reminderEmail)) return this.showToast('error', '默认收信邮箱格式错误');
    if (body.phone && !/^\+?\d{7,15}$/.test(body.phone)) return this.showToast('error', '手机格式错误，请输入7-15位数字，可带+号');
    try {
      if (this.currentAccount) {
        await this.api(`/accounts/${this.currentAccount.id}`, { method: 'PUT', body });
      } else {
        await this.api('/accounts', { method: 'POST', body });
      }
      this.closeModal('account-modal');
      this.showToast('success', '保存成功');
      await this.loadAccounts();
      this.renderAccounts();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async deleteAccount(id) {
    if (!confirm('确定要删除此账号吗？')) return;
    try {
      await this.api(`/accounts/${id}`, { method: 'DELETE' });
      this.showToast('success', '删除成功');
      await this.loadAccounts();
      this.renderAccounts();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async copyPassword(id) {
    try {
      const data = await this.api(`/accounts/${id}/password`);
      await this.copyText(data.password, '密码已复制');
    } catch (e) {
      this.showToast('error', e.message || '复制失败');
    }
  },

  showCategories() {
    this.renderCategoryList();
    this.openModal('category-modal');
  },

  renderCategoryList() {
    const list = document.getElementById('category-list');
    if (this.categories.length === 0) {
      list.innerHTML = '<p style="color:var(--text-secondary);text-align:center;padding:24px">暂无分类</p>';
      return;
    }
    list.innerHTML = this.categories.map(c => `
      <div class="account-item" style="margin-bottom:8px">
        <div class="account-info"><div class="account-title">${this.escape(c.name)}</div></div>
        <div class="account-actions"><button class="btn btn-icon btn-sm" onclick="Passworder.editCategory(${c.id})">✏️</button><button class="btn btn-icon btn-sm" onclick="Passworder.deleteCategory(${c.id})">🗑️</button></div>
      </div>
    `).join('');
  },

  showNewCategory() {
    const name = prompt('分类名称：');
    if (!name) return;
    this.saveCategory({ name });
  },

  editCategory(id) {
    const cat = this.categories.find(c => c.id === id);
    const name = prompt('分类名称：', cat.name);
    if (!name || name === cat.name) return;
    this.saveCategory({ id, name });
  },

  async saveCategory(cat) {
    try {
      if (cat.id) {
        await this.api(`/categories/${cat.id}`, { method: 'PUT', body: { name: cat.name, icon: cat.icon || '' } });
      } else {
        await this.api('/categories', { method: 'POST', body: { name: cat.name, icon: '' } });
      }
      this.showToast('success', '保存成功');
      await this.loadCategories();
      this.renderCategoryList();
      this.renderCategoryFilter();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async deleteCategory(id) {
    if (!confirm('确定删除此分类？')) return;
    try {
      await this.api(`/categories/${id}`, { method: 'DELETE' });
      this.showToast('success', '删除成功');
      await this.loadCategories();
      this.renderCategoryList();
      this.renderCategoryFilter();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderAccountsMethods);
}
