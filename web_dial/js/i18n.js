/**
 * Polyglot Web Dial i18n Engine
 * Mengatur peralihan bahasa global (id / en) secara dinamis,
 * otomatis menerjemahkan elemen DOM ber-atribut data-i18n,
 * dan menyediakan fungsi pembantu global t(key, params).
 */

const I18n = {
  currentLang: localStorage.getItem('dial_lang') || 'id',

  init() {
    this.setLanguage(this.currentLang, false);
  },

  getLanguage() {
    return this.currentLang;
  },

  setLanguage(lang, triggerRerender = true) {
    if (lang !== 'en' && lang !== 'id') lang = 'id';
    this.currentLang = lang;
    try {
      localStorage.setItem('dial_lang', lang);
    } catch {
      // Ignore storage restrictions
    }
    document.documentElement.lang = lang;
    this.applyDOM();

    if (triggerRerender && typeof Layout !== 'undefined' && Layout.rerender) {
      const currentHref = (window.location.pathname.split('/').pop() || 'index.html') + (window.location.hash || '');
      Layout.rerender(currentHref);
    }

    // Trigger custom event agar view controller dapat memperbarui string dinamis
    window.dispatchEvent(new CustomEvent('languagechange', { detail: { lang } }));
  },

  toggleLanguage() {
    const nextLang = this.currentLang === 'id' ? 'en' : 'id';
    this.setLanguage(nextLang, true);
  },

  t(key, params = {}) {
    const dict = (typeof DialLocales !== 'undefined' && DialLocales[this.currentLang]) || (typeof DialLocales !== 'undefined' && DialLocales.id) || {};
    let text = dict[key];
    if (text === undefined && typeof DialLocales !== 'undefined' && DialLocales.en) {
      text = DialLocales.en[key];
    }
    if (text === undefined) {
      return key;
    }

    if (params && typeof params === 'object') {
      Object.entries(params).forEach(([pKey, pVal]) => {
        text = text.replaceAll(`{${pKey}}`, pVal);
      });
    }
    return text;
  },

  applyDOM() {
    document.querySelectorAll('[data-i18n]').forEach((el) => {
      const key = el.getAttribute('data-i18n');
      if (key) {
        el.textContent = this.t(key);
      }
    });

    document.querySelectorAll('[data-i18n-placeholder]').forEach((el) => {
      const key = el.getAttribute('data-i18n-placeholder');
      if (key) {
        el.setAttribute('placeholder', this.t(key));
      }
    });

    document.querySelectorAll('[data-i18n-title]').forEach((el) => {
      const key = el.getAttribute('data-i18n-title');
      if (key) {
        el.setAttribute('title', this.t(key));
      }
    });
  }
};

// Expose helper global t
window.t = (key, params) => I18n.t(key, params);

// Auto-init saat dokumen dimuat
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => I18n.init());
} else {
  I18n.init();
}
