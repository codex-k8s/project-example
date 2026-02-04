import { defineStore } from 'pinia';

import { http } from '../shared/api/http';
import { connectWs, type WsEnvelope } from '../shared/ws/ws';

export type Message = {
  id: number;
  user_id: number;
  text: string;
  created_at: string;
  deleted_at: string | null;
};

let ws: WebSocket | null = null;

export const useMessagesStore = defineStore('messages', {
  state: () => ({
    items: [] as Message[],
    loading: false as boolean,
    errorMessage: '' as string,
    errorKey: '' as string,
    connected: false as boolean
  }),
  actions: {
    async loadRecent(limit = 50) {
      this.loading = true;
      this.errorMessage = '';
      this.errorKey = '';
      try {
        const { data } = await http.get<{ messages: Message[] }>('/messages', { params: { limit } });
        this.items = data.messages;
      } catch (e: any) {
        this.errorMessage = e?.response?.data?.message ?? '';
        this.errorKey = this.errorMessage ? '' : 'errors.load_failed';
        throw e;
      } finally {
        this.loading = false;
      }
    },

    async send(text: string) {
      const trimmed = text.trim();
      if (!trimmed) return;
      await http.post('/messages', { text: trimmed });
    },

    async remove(id: number) {
      await http.delete(`/messages/${id}`);
    },

    connect() {
      if (ws) return;
      this.connected = true;
      ws = connectWs((msg) => this.onWs(msg));
      ws.onclose = () => {
        ws = null;
        this.connected = false;
      };
    },

    disconnect() {
      ws?.close();
      ws = null;
      this.connected = false;
    },

    onWs(msg: WsEnvelope) {
      if (msg.type === 'message.created') {
        const m: Message = msg.payload.message;
        this.items.unshift(m);
      }
      if (msg.type === 'message.deleted') {
        const id = msg.payload.message_id as number;
        this.items = this.items.map((m) => (m.id === id ? { ...m, deleted_at: msg.payload.deleted_at } : m));
      }
    }
  }
});
