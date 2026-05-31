export type UnoColor = 'red' | 'yellow' | 'green' | 'blue' | 'wild'

export interface UnoCard {
  id: string
  color: UnoColor
  value: string
  label: string
}

export interface UnoPlayer {
  index: number
  name: string
  is_ai: boolean
  hand_count: number
}

export interface UnoEvent {
  type: string
  player_index: number
  player_name: string
  card?: UnoCard
  color?: UnoColor
  amount?: number
  message?: string
}

export interface UnoState {
  id: string
  phase: 'playing' | 'finished'
  players: UnoPlayer[]
  human_player: number
  current_turn: number
  direction: number
  current_color: UnoColor
  top_card: UnoCard
  draw_count: number
  discard_count: number
  winner_index?: number
  message: string
  pending_draw_penalty?: number
  draw_stack_wild4_only?: boolean
  must_play_after_stack?: boolean
  turn_deadline_unix?: number
  my_hand?: UnoCard[]
  events?: UnoEvent[]
}

export const UNO_COLOR_LABELS: Record<UnoColor, string> = {
  red: '红',
  yellow: '黄',
  green: '绿',
  blue: '蓝',
  wild: '万能',
}

export const UNO_PLAY_COLORS: UnoColor[] = ['red', 'yellow', 'green', 'blue']

export function unoColorClass(color: UnoColor) {
  return `uno-card--${color}`
}

export const UNO_ACTION_VALUES = new Set(['skip', 'reverse', 'draw2', 'wild', 'wild4'])

export function isUnoActionValue(value: string) {
  return UNO_ACTION_VALUES.has(value)
}

export function canPlayUnoCard(card: UnoCard, state: UnoState, hand: UnoCard[]): boolean {
  if (hand.length === 1 && isUnoActionValue(card.value)) return false
  const pending = state.pending_draw_penalty ?? 0
  if (pending > 0) {
    if (card.value === 'wild4') return true
    if (card.value === 'draw2' && !state.draw_stack_wild4_only) return true
    return false
  }
  if (card.color === 'wild' || card.value === 'wild' || card.value === 'wild4') {
    return true
  }
  if (card.color === state.current_color) return true
  return card.value === state.top_card.value
}
