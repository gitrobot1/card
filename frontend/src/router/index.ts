import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DoudizhuModeView from '../views/DoudizhuModeView.vue'
import DoudizhuRoomView from '../views/DoudizhuRoomView.vue'
import DoudizhuView from '../views/DoudizhuView.vue'
import ZhajinhuaModeView from '../views/ZhajinhuaModeView.vue'
import ZhajinhuaRoomView from '../views/ZhajinhuaRoomView.vue'
import ZhajinhuaView from '../views/ZhajinhuaView.vue'
import UnoModeView from '../views/UnoModeView.vue'
import UnoRoomView from '../views/UnoRoomView.vue'
import UnoView from '../views/UnoView.vue'
import DiceDemoView from '../views/DiceDemoView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/games/doudizhu', name: 'doudizhu-mode', component: DoudizhuModeView },
    { path: '/games/doudizhu/solo', name: 'doudizhu-solo', component: DoudizhuView },
    { path: '/games/doudizhu/online', name: 'doudizhu-online', component: DoudizhuRoomView },
    { path: '/games/doudizhu/play/:gameId', name: 'doudizhu-play', component: DoudizhuView },
    { path: '/games/zhajinhua', name: 'zhajinhua-mode', component: ZhajinhuaModeView },
    { path: '/games/zhajinhua/solo', name: 'zhajinhua-solo', component: ZhajinhuaView },
    { path: '/games/zhajinhua/online', name: 'zhajinhua-online', component: ZhajinhuaRoomView },
    { path: '/games/zhajinhua/play/:gameId', name: 'zhajinhua-play', component: ZhajinhuaView },
    { path: '/games/uno', name: 'uno-mode', component: UnoModeView },
    { path: '/games/uno/solo', name: 'uno-solo', component: UnoView },
    { path: '/games/uno/online', name: 'uno-online', component: UnoRoomView },
    { path: '/games/uno/play/:gameId', name: 'uno-play', component: UnoView },
    { path: '/dev/dice', name: 'dice-demo', component: DiceDemoView },
  ],
})

export default router
