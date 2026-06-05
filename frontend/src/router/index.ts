import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import DoudizhuModeView from '../views/doudizhu/DoudizhuModeView.vue'
import DoudizhuRoomView from '../views/doudizhu/DoudizhuRoomView.vue'
import DoudizhuView from '../views/doudizhu/DoudizhuView.vue'
import ZhajinhuaModeView from '../views/zhajinhua/ZhajinhuaModeView.vue'
import ZhajinhuaRoomView from '../views/zhajinhua/ZhajinhuaRoomView.vue'
import ZhajinhuaView from '../views/zhajinhua/ZhajinhuaView.vue'
import UnoModeView from '../views/uno/UnoModeView.vue'
import UnoRoomView from '../views/uno/UnoRoomView.vue'
import UnoView from '../views/uno/UnoView.vue'
import DouNiuModeView from '../views/douniu/DouNiuModeView.vue'
import DouNiuRoomView from '../views/douniu/DouNiuRoomView.vue'
import DouNiuView from '../views/douniu/DouNiuView.vue'
import YuzhoushaModeView from '../views/yuzhousha/YuzhoushaModeView.vue'
import YuzhoushaPickView from '../views/yuzhousha/YuzhoushaPickView.vue'
import YuzhoushaView from '../views/yuzhousha/YuzhoushaView.vue'
import YuzhoushaRoomView from '../views/yuzhousha/YuzhoushaRoomView.vue'
import DiceDemoView from '../views/dev/DiceDemoView.vue'

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
    { path: '/games/douniu', name: 'douniu-mode', component: DouNiuModeView },
    { path: '/games/douniu/solo', name: 'douniu-solo', component: DouNiuView },
    { path: '/games/douniu/online', name: 'douniu-online', component: DouNiuRoomView },
    { path: '/games/douniu/play/:gameId', name: 'douniu-play', component: DouNiuView },
    { path: '/games/yuzhousha', name: 'yuzhousha-mode', component: YuzhoushaModeView },
    { path: '/games/yuzhousha/solo/pick', name: 'yuzhousha-pick', component: YuzhoushaPickView },
    { path: '/games/yuzhousha/online', name: 'yuzhousha-online', component: YuzhoushaRoomView },
    { path: '/games/yuzhousha/play/:gameId', name: 'yuzhousha-play', component: YuzhoushaView },
    { path: '/dev/dice', name: 'dice-demo', component: DiceDemoView },
  ],
})

export default router
