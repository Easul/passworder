window.PassworderTranslatorMethods = {
  syncAndroidOnlyVisibility() {
    const isAndroid = window.PassworderShared.isAndroidApp();
    document.querySelectorAll('.android-only').forEach(element => {
      element.style.display = isAndroid ? '' : 'none';
    });
    if (!isAndroid) this.hideTranslator();
  },

  async saveTranslatorSettings() {
    const fields = {
      'translator.base_url': document.getElementById('setting-translator-base-url').value.trim(),
      'translator.api_key': document.getElementById('setting-translator-key').value.trim(),
      'translator.model': document.getElementById('setting-translator-model').value.trim()
    };

    try {
      await Promise.all(Object.entries(fields).map(([key, value]) => this.api(`/settings/${key}`, {
        method: 'PUT', body: { value }
      })));
      this.senderSettings = { ...(this.senderSettings || {}), ...fields };
      this.showToast('success', '翻译配置已保存');
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async showTranslator() {
    if (!window.PassworderShared.isAndroidApp()) return;
    this.syncAndroidOnlyVisibility();
    const settings = await this.ensureTranslatorSettings();
    if (!settings) return;
    if (typeof window.PassworderAndroid?.showTranslatorOverlay === 'function') {
      window.PassworderAndroid.showTranslatorOverlay(settings.baseUrl, settings.apiKey, settings.model);
      return;
    }
    const panel = document.getElementById('translator-panel');
    panel.classList.remove('hidden');
    document.getElementById('translator-input')?.focus();
  },

  hideTranslator() {
    this.cancelPendingTranslations?.();
    document.getElementById('translator-panel')?.classList.add('hidden');
  },

  clearTranslator() {
    document.getElementById('translator-input').value = '';
    document.getElementById('translator-output').textContent = '';
  },

  async translateText() {
    const input = document.getElementById('translator-input').value.trim();
    if (!input) {
      this.showToast('error', '请输入要翻译的内容');
      return;
    }

    const settings = await this.ensureTranslatorSettings();
    if (!settings) return;

    const output = document.getElementById('translator-output');
    output.textContent = '翻译中...';

    try {
      const translated = await this.requestTranslation(settings, input);
      output.textContent = translated;
    } catch (e) {
      output.textContent = '';
      this.showToast('error', e.message || '翻译失败');
    }
  },

  async ensureTranslatorSettings() {
    if (!this.senderSettings || !Object.keys(this.senderSettings).length) {
      await this.loadSenderSettings();
    }
    const settings = {
      baseUrl: (this.senderSettings?.['translator.base_url'] || '').trim(),
      apiKey: (this.senderSettings?.['translator.api_key'] || '').trim(),
      model: (this.senderSettings?.['translator.model'] || '').trim()
    };
    if (!settings.baseUrl || !settings.apiKey || !settings.model) {
      this.showToast('error', '请先在设置中填写翻译配置');
      this.showSettings();
      return null;
    }
    return settings;
  },

  async requestTranslation(settings, text) {
    if (window.PassworderShared.isAndroidApp() && typeof window.PassworderAndroid?.translate === 'function') {
      return this.requestAndroidTranslation(settings, text);
    }
    throw new Error('当前平台暂不支持翻译入口');
  },

  requestAndroidTranslation(settings, text) {
    return new Promise((resolve, reject) => {
      const requestId = `translate-${Date.now()}-${Math.random().toString(16).slice(2)}`;
      this.pendingTranslations = this.pendingTranslations || {};
      this.pendingTranslations[requestId] = { resolve, reject };
      window.PassworderAndroid.translate(requestId, settings.baseUrl, settings.apiKey, settings.model, text);
    });
  },

  receiveTranslationResult(requestId, content, error) {
    const pending = this.pendingTranslations?.[requestId];
    if (!pending) return;
    delete this.pendingTranslations[requestId];
    if (error) {
      pending.reject(new Error(error));
      return;
    }
    if (!content) {
      pending.reject(new Error('翻译接口返回为空'));
      return;
    }
    pending.resolve(content.trim());
  },

  cancelPendingTranslations() {
    Object.values(this.pendingTranslations || {}).forEach(pending => {
      pending.reject(new Error('翻译已取消'));
    });
    this.pendingTranslations = {};
  },

  copyTranslatorResult() {
    const result = document.getElementById('translator-output').textContent.trim();
    if (!result || result === '翻译中...') {
      this.showToast('error', '没有可复制的翻译结果');
      return;
    }
    this.copyText(result, '翻译结果已复制');
  }
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderTranslatorMethods);
}
