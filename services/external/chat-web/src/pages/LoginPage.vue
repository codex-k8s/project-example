<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const auth = useAuthStore();
const { t } = useI18n();

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
    <h1 class="title">{{ t('app.title') }}</h1>
    <div class="surface">
      <div class="tabs">
        <button class="btn btn-ghost" :class="{ active: mode === 'login' }" @click="mode = 'login'">
          {{ t('auth.login') }}
        </button>
        <button class="btn btn-ghost" :class="{ active: mode === 'register' }" @click="mode = 'register'">
          {{ t('auth.register') }}
        </button>
      </div>

      <form @submit.prevent="submit">
        <label>
          {{ t('auth.username') }}
          <input v-model="username" autocomplete="username" />
        </label>
        <label>
          {{ t('auth.password') }}
          <input v-model="password" type="password" autocomplete="current-password" />
        </label>
        <p v-if="auth.errorMessage || auth.errorKey" class="err">{{ auth.errorMessage || t(auth.errorKey) }}</p>
        <button class="btn btn-primary" type="submit" :disabled="auth.loading">
          {{ mode === 'register' ? t('auth.register') : t('auth.login') }}
        </button>
      </form>
    </div>
  </main>
</template>

<style scoped>
.wrap { max-width: 520px; margin: 48px auto; padding: 0 16px; }
.title { margin: 0 0 16px 0; }
.tabs { display: flex; gap: 8px; margin-bottom: 12px; }
.tabs .btn.active { border-color: var(--c-accent); }
label { display: grid; gap: 6px; margin: 10px 0; }
.err { color: #ffb4b4; }
</style>
