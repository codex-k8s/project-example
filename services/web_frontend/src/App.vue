<template>
  <div class="app">
    <header class="app-header">
      <h1>Project Example Chat</h1>
      <p v-if="store.nickname">Вы вошли как: <strong>{{ store.nickname }}</strong></p>
    </header>

    <main class="app-main">
      <section v-if="!store.token" class="auth">
        <h2>Регистрация</h2>
        <form @submit.prevent="onRegister">
          <label>
            Ник:
            <input v-model="nickname" type="text" name="nickname" autocomplete="username" required maxlength="64" />
          </label>
          <label>
            Пароль:
            <input v-model="password" type="password" name="password" autocomplete="new-password" required />
          </label>
          <button type="submit">📝 Зарегистрироваться</button>
        </form>

        <h2>Вход</h2>
        <form @submit.prevent="onLogin">
          <label>
            Ник:
            <input v-model="nickname" type="text" name="nickname" autocomplete="username" required maxlength="64" />
          </label>
          <label>
            Пароль:
            <input v-model="password" type="password" name="password" autocomplete="current-password" required />
          </label>
          <button type="submit">🔑 Войти</button>
        </form>
      </section>

      <section v-else class="chat">
        <div class="chat-header">
          <h2>Общий чат</h2>
          <button type="button" @click="onLogout">🚪 Выйти</button>
        </div>

        <div class="messages" ref="messagesRef">
          <div
            v-for="msg in store.messages"
            :key="msg.id"
            :class="['message', { 'message--own': msg.userId === store.userId }]"
          >
            <div class="meta-row">
              <div class="meta">
                <span class="author">{{ msg.nickname }}</span>
                <span class="time">{{ formatTime(msg.createdAt) }}</span>
              </div>
              <button
                v-if="msg.userId === store.userId"
                type="button"
                class="delete-btn"
                title="Удалить сообщение"
                @click="onDeleteMessage(msg.id)"
              >
                🗑️
              </button>
            </div>
            <div class="bubble">{{ msg.text }}</div>
          </div>
          <p v-if="store.messages.length === 0" class="empty">
            Сообщений пока нет — станьте первым!
          </p>
        </div>

        <form class="composer" @submit.prevent="onSend">
          <input
            v-model="messageText"
            placeholder="Введите сообщение..."
            maxlength="2000"
          />
          <button type="submit">📨 Отправить</button>
        </form>
      </section>

      <p v-if="store.error" class="error">
        {{ store.error }}
      </p>
    </main>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from "vue";
import { useChatStore } from "./stores/chat";

const store = useChatStore();
const nickname = ref("");
const password = ref("");
const messageText = ref("");
const messagesRef = ref(null);

let intervalId = null;

function formatTime(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  return date.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit"
  });
}

async function onRegister() {
  await store.register(nickname.value, password.value);
  await store.fetchMessages();
  scrollToBottom();
}

async function onLogin() {
  await store.login(nickname.value, password.value);
  await store.fetchMessages();
}

function onLogout() {
  store.logout();
}

async function onSend() {
  const text = messageText.value.trim();
  if (!text) {
    return;
  }
  await store.sendMessage(text);
  messageText.value = "";
  scrollToBottom();
}

async function onDeleteMessage(messageId) {
  await store.deleteMessage(messageId);
}

function scrollToBottom() {
  if (messagesRef.value) {
    const el = messagesRef.value;
    requestAnimationFrame(() => {
      el.scrollTop = el.scrollHeight;
    });
  }
}

onMounted(async () => {
  store.initFromStorage();
  if (store.token) {
    await store.fetchMessages();
    scrollToBottom();
  }
  intervalId = window.setInterval(async () => {
    try {
      await store.fetchMessages();
      scrollToBottom();
    } catch {
      // errors are already stored in state
    }
  }, 3000);
});

onUnmounted(() => {
  if (intervalId) {
    window.clearInterval(intervalId);
  }
});
</script>

<style scoped>
.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 16px;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.app-header {
  margin-bottom: 16px;
}

.app-main {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth form,
.composer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

input {
  padding: 6px 8px;
  border-radius: 4px;
  border: 1px solid #ccc;
}

button {
  padding: 6px 10px;
  border-radius: 4px;
  border: none;
  background-color: #2563eb;
  color: white;
  cursor: pointer;
}

button:hover {
  background-color: #1d4ed8;
}

.chat {
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.messages {
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 8px;
  height: 320px;
  overflow-y: auto;
  background-color: #fafafa;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.message {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.message--own {
  align-items: flex-end;
}

.message .meta {
  font-size: 12px;
  color: #555;
  display: flex;
  gap: 8px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.message .author {
  font-weight: 600;
}

.message .bubble {
  font-size: 14px;
  padding: 8px 10px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  background-color: #ffffff;
  max-width: 75%;
  word-break: break-word;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
}

.message--own .bubble {
  background-color: #e0f2fe;
  border-color: #bae6fd;
}

.delete-btn {
  padding: 2px 6px;
  border-radius: 4px;
  border: none;
  background-color: #dc2626;
  color: white;
  cursor: pointer;
}

.delete-btn:hover {
  background-color: #b91c1c;
}

.empty {
  text-align: center;
  color: #777;
}

.error {
  color: #b91c1c;
}
</style>
