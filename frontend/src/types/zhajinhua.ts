import type { Card } from './doudizhu'

export type { Card, GameMeta } from './doudizhu'

export interface ZhajinhuaPlayer {
  index: number
  name: string
  is_ai: boolean
  looked: boolean
  folded: boolean
  chips: number
  bet_round: number
  total_bet: number
  hand_type?: string
  hand_label?: string
  multiplier?: number
  hand?: Card[]
  card_count: number
}

export interface ZhajinhuaEvent {
  type: string
  player_index: number
  player_name: string
  target_index?: number
  target_name?: string
  amount?: number
  hand_type?: string
  hand_label?: string
  multiplier?: number
  message?: string
}

export interface ZhajinhuaState {
  id: string
  phase: 'betting' | 'finished'
  players: ZhajinhuaPlayer[]
  human_player: number
  dealer_index: number
  current_turn: number
  pot: number
  current_bet: number
  base_ante: number
  min_raise: number
  compare_cost: number
  winner_index?: number
  win_multiplier?: number
  win_hand_label?: string
  message: string
  my_hand?: Card[]
  hand_multipliers: Record<string, number>
  turn_deadline_unix: number
  events?: ZhajinhuaEvent[]
}

export interface ZhajinhuaRoom {
  id: string
  status: 'waiting' | 'playing'
  game_id?: string
  host_user_id: number
  players: { user_id: number; username: string; ready: boolean }[]
}

export const HAND_TYPE_LABELS: Record<string, string> = {
  '235': '235',
  leopard: '豹子',
  straight_flush: '顺金',
  flush: '金花',
  straight: '顺子',
  pair: '对子',
  high_card: '单牌',
}
