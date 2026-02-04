<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useMessagesStore } from '../stores/messages';

const router = useRouter();
const auth = useAuthStore();
const messages = useMessagesStore();
const { t } = useI18n();

const text = ref('');

async function send() {
  await messages.send(text.value);
  text.value = '';
}

async function remove(id: number) {
  await messages.remove(id);
}

async function logout() {
  await auth.logout();
  await router.push('/login');
}

onMounted(async () => {
  try {
    await messages.loadRecent(50);
  } catch {
    await router.push('/login');
    return;
  }
  messages.connect();
});

onBeforeUnmount(() => {
  messages.disconnect();
});
</script>

<template>
  <main class="wrap">
    <header class="bar">
      <strong>{{ t('chat.title') }}</strong>
      <button class="btn btn-ghost" @click="logout">{{ t('auth.logout') }}</button>
    </header>

    <section class="composer">
      <input v-model="text" :placeholder="t('chat.placeholder')" @keydown.enter="send" />
      <button class="btn btn-primary" @click="send">{{ t('chat.send') }}</button>
    </section>

    <section class="list">
      <article v-for="m in messages.items" :key="m.id" class="msg surface" :class="{ deleted: !!m.deleted_at }">
        <div class="meta">
          <span>#{{ m.id }}</span>
          <span>u{{ m.user_id }}</span>
          <span>{{ new Date(m.created_at).toLocaleTimeString() }}</span>
          <button v-if="!m.deleted_at" class="btn btn-ghost" @click="remove(m.id)">{{ t('chat.delete') }}</button>
        </div>
        <p>{{ m.text }}</p>
        <small v-if="m.deleted_at">{{ t('chat.deleted') }}</small>
      </article>
    </section>
  </main>
</template>

<style scoped>
.wrap { max-width: 820px; margin: 24px auto; padding: 0 16px; }
.bar { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; }
.composer { display: flex; gap: 8px; margin: 12px 0; }
.composer input { flex: 1; }
.list { display: grid; gap: 10px; }
.msg { border-radius: 16px; padding: 12px; }
.msg.deleted { opacity: 0.6; }
.meta { display: flex; gap: 10px; align-items: center; font-size: 12px; }
.meta button { padding: 6px 10px; }
p { margin: 8px 0 0 0; }
</style>
