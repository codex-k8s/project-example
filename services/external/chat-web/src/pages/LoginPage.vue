<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const auth = useAuthStore();

const username = ref('');
const password = ref('');
const mode = ref<'login' | 'register'>('login');

async function submit() {
  if (mode.value === 'register') {
    await auth.register(username.value, password.value);
  } else {
    await auth.login(username.value, password.value);
  }
  await router.push('/chat');
}
</script>

<template>
  <main class="wrap">
    <h1>Chat</h1>
    <div class="card">
      <div class="tabs">
        <button :class="{ active: mode === 'login' }" @click="mode = 'login'">Login</button>
        <button :class="{ active: mode === 'register' }" @click="mode = 'register'">Register</button>
      </div>

      <form @submit.prevent="submit">
        <label>
          Username
          <input v-model="username" autocomplete="username" />
        </label>
        <label>
          Password
          <input v-model="password" type="password" autocomplete="current-password" />
        </label>
        <p v-if="auth.error" class="err">{{ auth.error }}</p>
        <button type="submit" :disabled="auth.loading">{{ mode === 'register' ? 'Register' : 'Login' }}</button>
      </form>
    </div>
  </main>
</template>

<style scoped>
.wrap { max-width: 520px; margin: 48px auto; padding: 0 16px; font-family: ui-sans-serif, system-ui; }
.card { border: 1px solid #243040; border-radius: 16px; padding: 16px; background: #0f1622; color: #e7eefc; }
.tabs { display: flex; gap: 8px; margin-bottom: 12px; }
button { background: #1b2a3b; color: #e7eefc; border: 1px solid #243040; padding: 10px 12px; border-radius: 12px; cursor: pointer; }
button.active { background: #2a3f59; }
label { display: grid; gap: 6px; margin: 10px 0; }
input { padding: 10px 12px; border-radius: 12px; border: 1px solid #243040; background: #0b0f14; color: #e7eefc; }
.err { color: #ffb4b4; }
</style>

