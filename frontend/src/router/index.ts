import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DoudizhuModeView from '../views/DoudizhuModeView.vue'
import DoudizhuRoomView from '../views/DoudizhuRoomView.vue'
import DoudizhuView from '../views/DoudizhuView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/games/doudizhu', name: 'doudizhu-mode', component: DoudizhuModeView },
    { path: '/games/doudizhu/solo', name: 'doudizhu-solo', component: DoudizhuView },
    { path: '/games/doudizhu/online', name: 'doudizhu-online', component: DoudizhuRoomView },
    { path: '/games/doudizhu/play/:gameId', name: 'doudizhu-play', component: DoudizhuView },
  ],
})

export default router
