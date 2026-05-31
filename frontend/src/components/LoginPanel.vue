<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { loadRememberUsername } from '../api/auth'

const emit = defineEmits<{
  submit: [payload: { username: string; remember: boolean }]
}>()

defineProps<{
  loading?: boolean
  appName: string
}>()

const username = ref('')
const rememberUsername = ref(false)
const error = ref('')

onMounted(() => {
  const remembered = loadRememberUsername()
  if (remembered) {
    username.value = remembered.username
    rememberUsername.value = true
  }
})

function handleSubmit() {
  error.value = ''
  const value = username.value.trim()
  if (!value) {
    error.value = '请输入用户名'
    return
  }
  emit('submit', { username: value, remember: rememberUsername.value })
}

defineExpose({
  setError(message: string) {
    error.value = message
  },
})
</script>

<template>
  <section class="login-panel">
    <div class="login-panel__glow" />
    <div class="login-panel__card">
      <p class="login-panel__tag">无需注册 · 输入即玩</p>
      <h1>{{ appName }}</h1>
      <p class="login-panel__desc">输入用户名即可进入大厅，首次登录会自动创建账号。</p>

      <form class="login-panel__form" @submit.prevent="handleSubmit">
        <label for="username">用户名</label>
        <input
          id="username"
          v-model="username"
          maxlength="16"
          autocomplete="username"
          placeholder="2-16 位，支持中文/英文/数字/下划线"
        />

        <label class="login-panel__remember">
          <input v-model="rememberUsername" type="checkbox" />
          <span>保存用户名，下次直接进入</span>
        </label>

        <p v-if="error" class="login-panel__error">{{ error }}</p>
        <button type="submit" :disabled="loading">
          {{ loading ? '进入中...' : '进入游戏' }}
        </button>
      </form>
    </div>
  </section>
</template>
