import type { YzsCard } from '../../../types/yuzhousha'
import type { PendingContext } from './context'
import { pendingHandlers } from './handlers'
import type { PendingHandler } from './types'
import type { YuzhoushaState } from '../../../types/yuzhousha'

// 保留兼容：window_kind 为 take 的 pending 也需要清空目标选择状态
const TARGET_PICK_WINDOW_KINDS = new Set(['take'])

export function findPendingHandler(
  state: YuzhoushaState | null | undefined,
): PendingHandler | undefined {
  if (!state) return undefined
  return pendingHandlers.find((h) => h.match(state))
}

export function findPendingHandlerByMode(mode: string | undefined): PendingHandler | undefined {
  if (!mode) return undefined
  return pendingHandlers.find((h) => h.modes.includes(mode))
}

export function pendingIsSkillOnly(state: YuzhoushaState | null | undefined): boolean {
  return findPendingHandler(state)?.skillOnly ?? false
}

export function pendingSuppressPlaySubmit(state: YuzhoushaState | null | undefined): boolean {
  return findPendingHandler(state)?.suppressPlaySubmit ?? false
}

export function pendingAllowsCancel(ctx: PendingContext): boolean | undefined {
  const handler = findPendingHandler(ctx.state)
  if (!handler) return undefined
  if (handler.allowsCancel === false) return false
  return true
}

export function pendingHint(ctx: PendingContext): string | null {
  const handler = findPendingHandler(ctx.state)
  return handler?.hint?.(ctx) ?? null
}

export function pendingOnEnter(ctx: PendingContext, mode: string | undefined) {
  const handler = findPendingHandlerByMode(mode)
  handler?.onEnter?.(ctx)
}

export function pendingOnModeChange(
  ctx: PendingContext,
  mode: string | undefined,
  prevMode: string | undefined,
  opts: { isMyPlay: boolean; selectedCardNeedsTargetCard: () => boolean },
) {
  if (prevMode && prevMode !== mode) {
    findPendingHandlerByMode(prevMode)?.onModeLeave?.(ctx)
  }
  const windowKind = ctx.state.pending?.window_kind
  if (!TARGET_PICK_WINDOW_KINDS.has(windowKind ?? '')) {
    if (!opts.isMyPlay || !opts.selectedCardNeedsTargetCard()) {
      ctx.selectedTargetZone.value = ''
      ctx.selectedTargetCardId.value = ''
    }
  }
  if (mode !== 'skill_ganglie_choice') {
    ctx.ganglieDiscardIds.value = []
  }
  if (mode !== 'ddz_judge_cancel') {
    ctx.ddzCancelDiscardIds.value = []
  }
  if (mode !== 'skill_yiji_give') {
    ctx.yijiSelectedIds.value = []
  }
  pendingOnEnter(ctx, mode)
}

export function pendingCanPlayCard(
  ctx: PendingContext,
  card: YzsCard,
): boolean | undefined {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.canPlayCard) return undefined
  return handler.canPlayCard(ctx, card)
}

export function pendingCanSubmitPlay(ctx: PendingContext): boolean | undefined {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.canSubmitPlay) return undefined
  return handler.canSubmitPlay(ctx)
}

export async function pendingSubmitPlay(ctx: PendingContext): Promise<boolean> {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.submitPlay) return false
  await handler.submitPlay(ctx)
  return true
}

export function pendingCanSubmitSkill(
  ctx: PendingContext,
  skillId: string,
): boolean | undefined {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.canSubmitSkill) return undefined
  return handler.canSubmitSkill(ctx, skillId)
}

export async function pendingSubmitSkill(
  ctx: PendingContext,
  skillId: string,
): Promise<boolean> {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.submitSkill) return false
  await handler.submitSkill(ctx, skillId)
  return true
}

export function pendingCanSubmitAction(
  ctx: PendingContext,
  action: string,
): boolean | undefined {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.canSubmitAction) return undefined
  return handler.canSubmitAction(ctx, action)
}

export async function pendingSubmitAction(
  ctx: PendingContext,
  action: string,
): Promise<boolean> {
  const handler = findPendingHandler(ctx.state)
  if (!handler?.submitAction) return false
  await handler.submitAction(ctx, action)
  return true
}

export { pendingHandlers } from './handlers'
