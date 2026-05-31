<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterView } from 'vue-router'
import LoginPanel from './components/LoginPanel.vue'
import { useAuth } from './composables/useAuth'
import {
  clearRememberUsername,
  loadRememberUsername,
  saveRememberUsername,
} from './api/auth'
import { loadAppConfig } from './config/loadConfig'
import type { AppConfig } from './config/loadConfig'

const ready = ref(false)
const appName = ref('Card Hub')
const config = ref<AppConfig | null>(null)
const loggingIn = ref(false)
const autoLoggingIn = ref(false)
const loginPanel = ref<InstanceType<typeof LoginPanel> | null>(null)

const { user, isLoggedIn, bootstrapping, bootstrap, login, logout } = useAuth()

onMounted(async () => {
  const loaded = await loadAppConfig()
  config.value = loaded
  appName.value = loaded.appName
  ready.value = true
  await bootstrap()

  if (!isLoggedIn.value) {
    await tryAutoLogin()
  }
})

async function tryAutoLogin() {
  const remembered = loadRememberUsername()
  if (!remembered?.enabled || !remembered.username) {
    return
  }

  autoLoggingIn.value = true
  try {
    await performLogin(remembered.username, true)
  } catch {
    // fall back to login form
  } finally {
    autoLoggingIn.value = false
  }
}

async function performLogin(username: string, remember: boolean) {
  loggingIn.value = true
  try {
    await login(username)
    if (remember) {
      saveRememberUsername(username)
    } else {
      clearRememberUsername()
    }
  } catch (error) {
    loginPanel.value?.setError(error instanceof Error ? error.message : '登录失败')
    throw error
  } finally {
    loggingIn.value = false
  }
}

async function handleLogin(payload: { username: string; remember: boolean }) {
  await performLogin(payload.username, payload.remember)
}
</script>

<template>
  <div v-if="!ready || bootstrapping || autoLoggingIn" class="boot-screen">
    {{ autoLoggingIn ? '正在进入...' : '加载中...' }}
  </div>

  <LoginPanel
    v-else-if="!isLoggedIn && config"
    ref="loginPanel"
    :app-name="appName"
    :loading="loggingIn"
    @submit="handleLogin"
  />

  <RouterView v-else-if="isLoggedIn && config && user" v-slot="{ Component }">
    <component
      :is="Component"
      :app-name="appName"
      :config="config"
      :user="user"
      @logout="logout"
    />
  </RouterView>
</template>
