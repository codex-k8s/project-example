import type { RouteRecordRaw } from 'vue-router';
import LoginPage from '../pages/LoginPage.vue';
import ChatPage from '../pages/ChatPage.vue';

export const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/chat' },
  { path: '/login', component: LoginPage },
  { path: '/chat', component: ChatPage }
];

