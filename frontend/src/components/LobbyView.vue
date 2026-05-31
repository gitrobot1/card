<script setup lang="ts">
import CardStagePanel from './CardStagePanel.vue'
import type { AppConfig } from '../config/loadConfig'
import type { User } from '../types/auth'

defineProps<{
  appName: string
  config: AppConfig
  user: User
}>()

const emit = defineEmits<{
  logout: []
}>()
</script>

<template>
  <main class="app">
    <section class="hero">
      <div class="hero__top">
        <div>
          <p class="hero__tag">欢迎回来</p>
          <h1>{{ appName }}</h1>
          <p class="hero__desc">
            当前玩家：<strong>{{ user.nickname }}</strong>
            <span class="hero__id">#{{ user.id }}</span>
          </p>
        </div>
        <button class="hero__logout" type="button" @click="emit('logout')">切换账号</button>
      </div>
      <dl class="hero__meta">
        <div>
          <dt>用户名</dt>
          <dd>{{ user.username }}</dd>
        </div>
        <div>
          <dt>最近登录</dt>
          <dd>{{ user.last_login }}</dd>
        </div>
        <div>
          <dt>API</dt>
          <dd>{{ config.apiBaseUrl }}</dd>
        </div>
      </dl>
    </section>

    <CardStagePanel :app-name="appName" :config="config" />
  </main>
</template>
