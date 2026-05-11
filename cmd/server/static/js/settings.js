window.PassworderSettingsMethods = {
  async showSettings() {
    await Promise.all([this.loadSenderSettings(), this.loadServerConfig()]);
    this.openModal('settings-modal');
  },

  async loadServerConfig() {
    try {
      const config = await this.api('/server-config');
      if (config) {
        document.getElementById('setting-server-host').value = config.host || '';
        document.getElementById('setting-server-port').value = config.port || '';
        document.getElementById('setting-server-db').value = config.dbPath || '';
        document.getElementById('setting-server-storage').value = config.storageDir || '';
        document.getElementById('setting-server-reminder-interval').value = config.reminderCheckInterval || '';
      }
    } catch (e) {
      console.error('Failed to load server config:', e);
    }
  },

  async saveServerConfig() {
    const config = {
      host: document.getElementById('setting-server-host').value.trim(),
      port: parseInt(document.getElementById('setting-server-port').value) || 0,
      dbPath: document.getElementById('setting-server-db').value.trim(),
      storageDir: document.getElementById('setting-server-storage').value.trim(),
      reminderCheckInterval: parseInt(document.getElementById('setting-server-reminder-interval').value) || 0
    };

    try {
      await this.api('/server-config', { method: 'PUT', body: config });
      this.showToast('success', '服务器配置已保存，重启后生效');
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async loadSenderSettings() {
    const settings = await this.api('/settings');
    this.senderSettings = (settings || []).reduce((acc, item) => {
      acc[item.key] = item.value;
      return acc;
    }, {});

    const mappings = {
      'mail.smtp_host': 'setting-smtp-host',
      'mail.smtp_port': 'setting-smtp-port',
      'mail.smtp_username': 'setting-smtp-username',
      'mail.smtp_password': 'setting-smtp-password',
      'mail.from_address': 'setting-from-address',
      'mail.from_name': 'setting-from-name'
    };

    Object.entries(mappings).forEach(([key, id]) => {
      const input = document.getElementById(id);
      if (input) input.value = this.senderSettings[key] || '';
    });

    const reminderEmailsInput = document.getElementById('setting-reminder-emails');
    if (reminderEmailsInput) {
      reminderEmailsInput.value = this.senderSettings['reminder_emails'] || '';
    }
  },

  async saveSenderSettings() {
    const fields = {
      'mail.smtp_host': document.getElementById('setting-smtp-host').value.trim(),
      'mail.smtp_port': document.getElementById('setting-smtp-port').value.trim(),
      'mail.smtp_username': document.getElementById('setting-smtp-username').value.trim(),
      'mail.smtp_password': document.getElementById('setting-smtp-password').value,
      'mail.from_address': document.getElementById('setting-from-address').value.trim(),
      'mail.from_name': document.getElementById('setting-from-name').value.trim(),
      'reminder_emails': document.getElementById('setting-reminder-emails').value.trim()
    };

    try {
      await Promise.all(Object.entries(fields).map(([key, value]) => this.api(`/settings/${key}`, {
        method: 'PUT', body: { value }
      })));
      this.showToast('success', '邮件设置已保存');
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async sendDueReminders() {
    try {
      const result = await this.api('/reminders/send-due', { method: 'POST' });
      const sent = result.sentReminders || 0;
      const failed = result.failedGroups || 0;
      const due = result.dueAccounts || 0;
      if (failed > 0 && sent === 0) {
        this.showToast('error', `检测到 ${due} 条到期提醒，但发送失败 ${failed} 组`);
      } else if (failed > 0) {
        this.showToast('warning', `已发送 ${sent} 条提醒，失败 ${failed} 组`);
      } else {
        this.showToast('success', `已发送 ${sent} 条提醒`);
      }
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async exportData() {
    const loadingToast = this.showToast('info', '正在导出数据...');
    try {
      const res = await fetch('/api/export', {
        headers: this.token ? { 'Authorization': this.token } : {}
      });
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `passworder-export-${new Date().toISOString().slice(0,10)}.zip`;
      a.click();
      URL.revokeObjectURL(url);
      if (loadingToast) loadingToast.remove();
      this.showToast('success', '导出成功');
    } catch (e) {
      if (loadingToast) loadingToast.remove();
      this.showToast('error', '导出失败');
    }
  },

  async importData() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.zip';
    input.onchange = async (e) => {
      const file = e.target.files[0];
      if (!file) return;
      const formData = new FormData();
      formData.append('file', file);
      try {
        const res = await fetch('/api/import', {
          method: 'POST',
          headers: this.token ? { 'Authorization': this.token } : {},
          body: formData
        });
        if (!res.ok) {
          const data = await res.json().catch(() => null);
          throw new Error(data?.message || '导入失败');
        }
        const result = await res.json();
        this.showToast('success', `导入成功：账号${result.accountsImported || 0}个，笔记${result.notesImported || 0}条`);
        window.location.reload();
      } catch (err) {
        this.showToast('error', err.message || '导入失败');
      }
    };
    input.click();
  },
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderSettingsMethods);
}
