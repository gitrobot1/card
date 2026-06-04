import type { YzsEvent } from '../../../types/yuzhousha'
import type { EventReplayContext } from './context'
import { eventReplayerHandlers } from './handlers'
import type { EventReplayHandler } from './types'

export type { EventReplayHandler } from './types'

export function findEventReplayer(event: YzsEvent): EventReplayHandler | undefined {
  return eventReplayerHandlers.find((h) => h.match(event))
}

export function shouldPrefetchEventMessage(event: YzsEvent): boolean {
  return !!event.message && event.type !== 'discard' && event.type !== 'discard_phase'
}

export async function replayRegisteredEvent(ctx: EventReplayContext): Promise<boolean> {
  const handler = findEventReplayer(ctx.event)
  if (!handler) return false
  await handler.replay(ctx)
  return true
}

export { eventReplayerHandlers } from './handlers'
