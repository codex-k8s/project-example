import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import { createRouter, createWebHistory } from 'vue-router';
import VueCookies from 'vue3-cookies';

import App from './App.vue';
import { routes } from './router/routes';

import './shared/styles/tokens.css';
import './shared/styles/base.css';
import './shared/styles/ui.css';

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      app: { title: 'Chat' },
      auth: {
        login: 'Login',
        register: 'Register',
        logout: 'Logout',
        username: 'Username',
        password: 'Password'
      },
      chat: {
        title: 'Chat',
        placeholder: 'Type a message...',
        send: 'Send',
        delete: 'Delete',
        deleted: 'Deleted'
      },
      errors: {
        login_failed: 'Login failed',
        register_failed: 'Register failed',
        logout_failed: 'Logout failed',
        load_failed: 'Failed to load'
      }
    }
  }
});

const router = createRouter({
  history: createWebHistory(),
  routes
});

createApp(App)
  .use(createPinia())
  .use(router)
  .use(i18n)
  .use(VueCookies, { expireTimes: '1d' })
  .mount('#app');
