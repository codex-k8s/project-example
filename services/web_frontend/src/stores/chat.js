import { defineStore } from "pinia";

function loadFromStorage() {
  try {
    const raw = window.localStorage.getItem("chat-session");
    if (!raw) {
      return null;
    }
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function saveToStorage(state) {
  try {
    window.localStorage.setItem(
      "chat-session",
      JSON.stringify({
        token: state.token,
        userId: state.userId,
        nickname: state.nickname
      })
    );
  } catch {
    // ignore storage errors
  }
}

export const useChatStore = defineStore("chat", {
  state: () => ({
    token: null,
    userId: null,
    nickname: "",
    messages: [],
    lastMessageId: 0,
    loading: false,
    error: null
  }),
  actions: {
    initFromStorage() {
      const data = loadFromStorage();
      if (data && data.token && data.nickname) {
        this.token = data.token;
        this.userId = data.userId;
        this.nickname = data.nickname;
      }
    },
    async register(nickname, password) {
      this.error = null;
      const payload = { nickname, password };
      const resp = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        this.error = data.error || "Registration failed";
        throw new Error(this.error);
      }
      return data;
    },
    async login(nickname, password) {
      this.error = null;
      const payload = { nickname, password };
      const resp = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        this.error = data.error || "Login failed";
        throw new Error(this.error);
      }
      this.token = data.token;
      this.userId = data.userId;
      this.nickname = data.nickname;
      saveToStorage(this.$state);
      return data;
    },
    logout() {
      this.token = null;
      this.userId = null;
      this.nickname = "";
      this.messages = [];
      this.lastMessageId = 0;
      this.error = null;
      try {
        window.localStorage.removeItem("chat-session");
      } catch {
        // ignore storage errors
      }
    },
    async fetchMessages() {
      const params = new URLSearchParams();
      if (this.lastMessageId) {
        params.set("after_id", String(this.lastMessageId));
      }
      const resp = await fetch(`/api/messages?${params.toString()}`);
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        this.error = data.error || "Failed to load messages";
        throw new Error(this.error);
      }
      if (Array.isArray(data) && data.length > 0) {
        this.messages.push(...data);
        this.lastMessageId = data[data.length - 1].id;
      }
    },
    async sendMessage(text) {
      if (!this.token) {
        this.error = "Not authenticated";
        throw new Error(this.error);
      }
      const payload = { text };
      const resp = await fetch("/api/messages", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.token}`
        },
        body: JSON.stringify(payload)
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        this.error = data.error || "Failed to send message";
        throw new Error(this.error);
      }
      this.messages.push(data);
      this.lastMessageId = data.id;
    }
  }
});

