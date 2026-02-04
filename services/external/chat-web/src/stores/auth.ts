import { defineStore } from 'pinia';
import { http } from '../shared/api/http';

type User = { id: number; username: string; created_at: string };

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    loading: false as boolean,
    errorMessage: '' as string,
    errorKey: '' as string
  }),
  actions: {
    async register(username: string, password: string) {
      this.loading = true;
      this.errorMessage = '';
      this.errorKey = '';
      try {
        const { data } = await http.post<User>('/auth/register', { username, password });
        this.user = data;
      } catch (e: any) {
        this.errorMessage = e?.response?.data?.message ?? '';
        this.errorKey = this.errorMessage ? '' : 'errors.register_failed';
        throw e;
      } finally {
        this.loading = false;
      }
    },
    async login(username: string, password: string) {
      this.loading = true;
      this.errorMessage = '';
      this.errorKey = '';
      try {
        const { data } = await http.post<User>('/auth/login', { username, password });
        this.user = data;
      } catch (e: any) {
        this.errorMessage = e?.response?.data?.message ?? '';
        this.errorKey = this.errorMessage ? '' : 'errors.login_failed';
        throw e;
      } finally {
        this.loading = false;
      }
    },
    async logout() {
      this.loading = true;
      this.errorMessage = '';
      this.errorKey = '';
      try {
        await http.post('/auth/logout');
        this.user = null;
      } catch (e: any) {
        this.errorMessage = e?.response?.data?.message ?? '';
        this.errorKey = this.errorMessage ? '' : 'errors.logout_failed';
      } finally {
        this.loading = false;
      }
    }
  }
});
