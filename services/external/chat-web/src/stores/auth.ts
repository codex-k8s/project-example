import { defineStore } from 'pinia';
import { http } from '../shared/api/http';

type User = { id: number; username: string; created_at: string };

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    loading: false as boolean,
    error: '' as string
  }),
  actions: {
    async register(username: string, password: string) {
      this.loading = true;
      this.error = '';
      try {
        const { data } = await http.post<User>('/auth/register', { username, password });
        this.user = data;
      } catch (e: any) {
        this.error = e?.response?.data?.message ?? 'register failed';
        throw e;
      } finally {
        this.loading = false;
      }
    },
    async login(username: string, password: string) {
      this.loading = true;
      this.error = '';
      try {
        const { data } = await http.post<User>('/auth/login', { username, password });
        this.user = data;
      } catch (e: any) {
        this.error = e?.response?.data?.message ?? 'login failed';
        throw e;
      } finally {
        this.loading = false;
      }
    },
    async logout() {
      this.loading = true;
      this.error = '';
      try {
        await http.post('/auth/logout');
        this.user = null;
      } catch (e: any) {
        this.error = e?.response?.data?.message ?? 'logout failed';
      } finally {
        this.loading = false;
      }
    }
  }
});

