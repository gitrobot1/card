import type { YuzhoushaState } from '../../../types/yuzhousha'
import type { PendingContext } from './context'

/** 当前 pending 的 actor 是否为本座（统一用 actor_seat，P5 清理完毕） */
export function isMyPendingActor(state: YuzhoushaState | null | undefined, mySeat: number): boolean {
  const p = state?.pending
  if (!p || state?.phase !== 'response') return false
  return p.actor_seat === mySeat
}

/** 被操作座位（统一用 subject_seat，P5 清理完毕） */
export function pickFromSeat(state: YuzhoushaState | null | undefined): number {
  return state?.pending?.subject_seat ?? -1
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
