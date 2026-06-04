import type { YzsEvent } from '../../../types/yuzhousha'
import type { EventReplayContext } from './context'

export interface EventReplayHandler {
  /** Primary event.type values this handler owns. */
  types: string[]
  match: (event: YzsEvent) => boolean
  replay: (ctx: EventReplayContext) => Promise<void>
}
