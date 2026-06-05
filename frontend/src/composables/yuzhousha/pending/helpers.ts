import type { YuzhoushaState } from '../../../types/yuzhousha'
import type { PendingContext } from './context'

/** 当前 pending 的 actor 是否为本座（优先 actor_seat，fallback 旧 target_index 规则） */
export function isMyPendingActor(state: YuzhoushaState | null | undefined, mySeat: number): boolean {
  const p = state?.pending
  if (!p || state?.phase !== 'response') return false
  if (p.actor_seat != null && p.actor_seat >= 0) return p.actor_seat === mySeat
  if (p.response_mode === 'skill_pojun') return p.source_index === mySeat
  if (p.response_mode === 'dying_rescue') return p.source_index === mySeat
  if (p.response_mode === 'wugu_pick') return p.wugu_pick_seat === mySeat
  return p.target_index === mySeat
}

/** 被操作座位；take 类窗口优先 subject_seat */
export function pickFromSeat(state: YuzhoushaState | null | undefined): number {
  const p = state?.pending
  if (!p) return -1
  if (p.window_kind === 'take' && p.subject_seat != null && p.subject_seat >= 0) {
    return p.subject_seat
  }
  return p.source_index
}

export function isBusy(ctx: PendingContext) {
  return ctx.loading || ctx.isAnimating
}

export function responseMode(state: YuzhoushaState | null | undefined, mode: string) {
  return state?.phase === 'response' && state.pending?.response_mode === mode
}

export function responseAnyMode(state: YuzhoushaState | null | undefined, modes: string[]) {
  const mode = state?.pending?.response_mode
  return state?.phase === 'response' && !!mode && modes.includes(mode)
}

export function pickFirstTarget(
  ctx: PendingContext,
  options: { zone: string; cardId: string }[],
) {
  const first = options[0]
  if (first) {
    ctx.selectedTargetZone.value = first.zone
    ctx.selectedTargetCardId.value = first.cardId
  }
}
