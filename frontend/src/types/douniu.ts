import type { Card } from './doudizhu'

export type { Card, GameMeta } from './doudizhu'

export interface DouNiuHandLayout {
  head_ids: string[]
  niu_ids?: string[]
}

export interface DouNiuPlayer {
  index: number
  name: string
  is_ai: boolean
  chips: number
  grab_mult: number
  bet_mult: number
  hand_type?: string
  hand_label?: string
  hand_multiplier?: number
  hand_layout?: DouNiuHandLayout
  round_delta?: number
  hand?: Card[]
  card_count: number
  grab_done: boolean
  bet_done: boolean
}

export interface DouNiuEvent {
  type: string
  player_index: number
  player_name: string
  target_index?: number
  target_name?: string
  amount?: number
  grab_mult?: number
  bet_mult?: number
  hand_type?: string
  hand_label?: string
  multiplier?: number
  message?: string
}

export interface DouNiuRoom {
  id: string
  status: string
  game_id?: string
  host_user_id: number
  players: { user_id: number; username: string; ready: boolean }[]
}

export interface DouNiuState {
  id: string
  phase: 'grab_banker' | 'betting' | 'finished'
  players: DouNiuPlayer[]
  human_player: number
  banker_index: number
  base_ante: number
  message: string
  my_hand?: Card[]
  my_hand_label?: string
  my_hand_type?: string
  my_hand_multiplier?: number
  my_hand_layout?: DouNiuHandLayout
  hand_multipliers?: Record<string, number>
  grab_options?: number[]
  bet_options?: number[]
  turn_deadline_unix?: number
  events?: DouNiuEvent[]
}

export const DOUNIU_HAND_LABELS: Record<string, string> = {
  five_small: '五小牛',
  bomb: '炸弹牛',
  five_flower: '五花牛',
  niu_niu: '牛牛',
  niu_9: '牛九',
  niu_8: '牛八',
  niu_7: '牛七',
  niu_6: '牛六',
  niu_5: '牛五',
  niu_4: '牛四',
  niu_3: '牛三',
  niu_2: '牛二',
  niu_1: '牛一',
  none: '没牛',
}
