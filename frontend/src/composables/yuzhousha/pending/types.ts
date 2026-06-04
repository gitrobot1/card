import type { YuzhoushaState, YzsCard } from '../../../types/yuzhousha'
import type { PendingContext } from './context'

export interface PendingHandler {
  /** Primary response_mode values this handler owns. */
  modes: string[]
  match: (state: YuzhoushaState) => boolean
  /** Disables generic play button; skill / custom actions only. */
  skillOnly?: boolean
  /** Blocks canSubmitPlay even when a card is selected (鬼才/鬼道). */
  suppressPlaySubmit?: boolean
  /** Default true when handler is active; false for wugu/peek/ganglie choice. */
  allowsCancel?: boolean
  onEnter?: (ctx: PendingContext) => void
  onModeLeave?: (ctx: PendingContext) => void
  canPlayCard?: (ctx: PendingContext, card: YzsCard) => boolean
  canSubmitPlay?: (ctx: PendingContext) => boolean
  submitPlay?: (ctx: PendingContext) => Promise<void>
  canSubmitSkill?: (ctx: PendingContext, skillId: string) => boolean
  submitSkill?: (ctx: PendingContext, skillId: string) => Promise<void>
  canSubmitAction?: (ctx: PendingContext, action: string) => boolean
  submitAction?: (ctx: PendingContext, action: string) => Promise<void>
  hint?: (ctx: PendingContext) => string | null
}
