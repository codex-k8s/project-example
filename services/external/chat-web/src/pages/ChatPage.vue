<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue';
import { useRouter } from 'vue-router';
import { http } from '../shared/api/http';
import { connectWs, type WsEnvelope } from '../shared/ws/ws';
import { useAuthStore } from '../stores/auth';

type Message = { id: number; user_id: number; text: string; created_at: string; deleted_at: string | null };

const router = useRouter();
const auth = useAuthStore();

const text = ref('');
const messages = ref<Message[]>([]);
let ws: WebSocket | null = null;

async function load() {
  const { data } = await http.get<{ messages: Message[] }>('/messages', { params: { limit: 50 } });
  messages.value = data.messages;
}

async function send() {
  if (!text.value.trim()) return;
  await http.post('/messages', { text: text.value });
  text.value = '';
}

async function remove(id: number) {
  await http.delete(`/messages/${id}`);
}

async function logout() {
  await auth.logout();
  await router.push('/login');
}

function onWs(msg: WsEnvelope) {
  if (msg.type === 'message.created') {
    const m: Message = msg.payload.message;
    messages.value.unshift(m);
  }
  if (msg.type === 'message.deleted') {
    const id = msg.payload.message_id as number;
    messages.value = messages.value.map((m) =>
      m.id === id ? { ...m, deleted_at: msg.payload.deleted_at } : m
    );
  }
}

onMounted(async () => {
  try {
    await load();
  } catch {
    await router.push('/login');
    return;
  }
  ws = connectWs(onWs);
});

onBeforeUnmount(() => {
  ws?.close();
});
</script>

<template>
  <main class="wrap">
    <header class="bar">
      <strong>Chat</strong>
      <button @click="logout">Logout</button>
    </header>

    <section class="composer">
      <input v-model="text" placeholder="Type a message..." @keydown.enter="send" />
      <button @click="send">Send</button>
    </section>

    <section class="list">
      <article v-for="m in messages" :key="m.id" class="msg" :class="{ deleted: !!m.deleted_at }">
        <div class="meta">
          <span>#{{ m.id }}</span>
          <span>u{{ m.user_id }}</span>
          <span>{{ new Date(m.created_at).toLocaleTimeString() }}</span>
          <button v-if="!m.deleted_at" @click="remove(m.id)">delete</button>
        </div>
        <p>{{ m.text }}</p>
        <small v-if="m.deleted_at">deleted</small>
      </article>
    </section>
  </main>
</template>

<style scoped>
.wrap { max-width: 820px; margin: 24px auto; padding: 0 16px; font-family: ui-sans-serif, system-ui; color: #e7eefc; }
.bar { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; }
button { background: #1b2a3b; color: #e7eefc; border: 1px solid #243040; padding: 10px 12px; border-radius: 12px; cursor: pointer; }
.composer { display: flex; gap: 8px; margin: 12px 0; }
input { flex: 1; padding: 10px 12px; border-radius: 12px; border: 1px solid #243040; background: #0b0f14; color: #e7eefc; }
.list { display: grid; gap: 10px; }
.msg { border: 1px solid #243040; border-radius: 16px; padding: 12px; background: #0f1622; }
.msg.deleted { opacity: 0.6; }
.meta { display: flex; gap: 10px; align-items: center; font-size: 12px; color: #b8c7e6; }
.meta button { padding: 6px 10px; }
p { margin: 8px 0 0 0; }
</style>

