import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { createI18n } from 'vue-i18n';
import { createRouter, createWebHistory } from 'vue-router';
import { VueCookies } from 'vue3-cookies';

import App from './App.vue';
import { routes } from './router/routes';

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      login: 'Login',
      register: 'Register',
      username: 'Username',
      password: 'Password',
      message: 'Message',
      send: 'Send',
      logout: 'Logout'
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

