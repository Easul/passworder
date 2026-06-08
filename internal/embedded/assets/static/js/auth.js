window.PassworderAuthMethods = {
  async checkAuth() {
    try {
      const res = await fetch('/api/auth/check').then(r => r.json());
      const status = res.data || res;
      this.hideLoadingMask();
      if (status.initialized) {
        if (this.token) {
          this.showMain();
          this.loadData();
        } else {
          this.showLogin();
        }
      } else {
        this.showSetup();
      }
    } catch (e) {
      this.showToast('error', '连接失败');
    }
  },

  hideLoadingMask() {
    const mask = document.getElementById('loading-mask');
    const app = document.getElementById('app');
    if (mask) {
      mask.style.opacity = '0';
      setTimeout(() => { mask.style.display = 'none'; }, 300);
    }
    if (app) {
      app.style.visibility = 'visible';
    }
  },

  showSetup() {
    this.showPage('setup-page');
  },

  showLogin() {
    this.showPage('login-page');
    const input = document.getElementById('login-password');
    if (input) {
      input.focus();
      input.onkeydown = (e) => {
        if (e.key === 'Enter') this.login();
      };
    }
  },

  async setup() {
    const password = document.getElementById('setup-password').value;
    const confirm = document.getElementById('setup-confirm').value;

    if (password.length < 8) {
      this.showToast('error', '密码至少需要8位');
      return;
    }
    if (password !== confirm) {
      this.showToast('error', '两次输入的密码不一致');
      return;
    }

    try {
      const data = await this.api('/auth/setup', {
        method: 'POST',
        body: { password }
      });
      this.token = data.token;
      localStorage.setItem('token', this.token);
      this.showToast('success', '设置完成');
      this.showMain();
      this.loadData();
    } catch (e) {
      this.showToast('error', e.message);
    }
  },

  async login() {
    const password = document.getElementById('login-password').value;
    try {
      const data = await this.api('/auth/login', {
        method: 'POST',
        body: { password }
      });
      this.token = data.token;
      localStorage.setItem('token', this.token);
      this.showToast('success', '登录成功');
      this.showMain();
      this.loadData();
    } catch (e) {
      this.showToast('error', '密码错误');
    }
  },

  logout() {
    this.token = null;
    localStorage.removeItem('token');
    this.api('/auth/logout', { method: 'POST' }).catch(() => {});
    this.showLogin();
  },

  checkPasswordStrength() {
    const pwd = document.getElementById('setup-password').value;
    const bar = document.getElementById('strength-bar');
    let strength = 0;
    if (pwd.length >= 8) strength++;
    if (pwd.length >= 12) strength++;
    if (/[A-Z]/.test(pwd)) strength++;
    if (/[0-9]/.test(pwd)) strength++;
    if (/[^A-Za-z0-9]/.test(pwd)) strength++;

    bar.className = 'password-strength-bar';
    if (strength <= 2) bar.classList.add('weak');
    else if (strength <= 4) bar.classList.add('medium');
    else bar.classList.add('strong');
  },
};

if (window.Passworder) {
  Object.assign(window.Passworder, window.PassworderAuthMethods);
}
