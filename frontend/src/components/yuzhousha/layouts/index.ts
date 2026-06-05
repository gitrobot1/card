import type { Component } from 'vue'
import YzsLayout1v1 from '../YzsLayout1v1.vue'
import YzsLayout2v2 from '../YzsLayout2v2.vue'
import YzsLayout3pChain from '../YzsLayout3pChain.vue'
import YzsLayout3v3 from '../YzsLayout3v3.vue'
import YzsLayoutIdentity5 from '../YzsLayoutIdentity5.vue'
import YzsLayoutIdentity8 from '../YzsLayoutIdentity8.vue'

export const YZS_LAYOUTS: Record<string, Component> = {
  solo_1v1: YzsLayout1v1,
  cross_2v2: YzsLayout2v2,
  triangle_3p: YzsLayout3pChain,
  hex_3v3: YzsLayout3v3,
  pentagon_5: YzsLayoutIdentity5,
  octagon_8: YzsLayoutIdentity8,
}

export function resolveYzsLayout(layoutKey?: string | null): Component {
  if (layoutKey && YZS_LAYOUTS[layoutKey]) {
    return YZS_LAYOUTS[layoutKey]
  }
  return YzsLayout1v1
}

/** Fallback subtitle when state has no layout_key (legacy saves). */
export function yzsLayoutSubtitle(layoutKey?: string | null, mode?: string | null): string {
  if (layoutKey === 'cross_2v2' || mode === '2v2') {
    return '2v2 十字阵 · 1 真人 + 3 电脑'
  }
  if (layoutKey === 'triangle_3p' || mode === '3p_chain') {
    return '3 人链式 · 杀上家保下家'
  }
  if (mode === '3p_ddz') {
    return '3 人斗地主 · 地主 vs 两农民'
  }
  if (layoutKey === 'hex_3v3' || mode === '3v3') {
    return '3v3 竞技 · 暖色 vs 冷色'
  }
  if (layoutKey === 'pentagon_5' || mode === 'identity_5') {
    return '5 人身份局 · 主公 vs 反贼'
  }
  if (layoutKey === 'octagon_8' || mode === 'identity_8') {
    return '8 人身份局 · 主公 vs 反贼'
  }
  return '1v1 单机 · 基础牌验证'
}
