export interface Card {
  id: string
  suit: string
  rank: number
  label: string
}

export interface PlayerSeat {
  index: number
  name: string
  is_human: boolean
  is_landlord: boolean
  hand_count: number
}

export interface PlayRecord {
  player_index: number
  player_name: string
  cards: Card[]
  pattern: string
}

export interface GameEvent {
  type: 'play' | 'pass' | 'call' | 'timeout' | 'game_over'
  player_index: number
  player_name: string
  cards?: Card[]
  call?: boolean
}

export interface RevealedHand {
  index: number
  name: string
  is_landlord: boolean
  cards: Card[]
}

export interface DouDizhuState {
  id: string
  phase: 'calling' | 'playing' | 'finished'
  players: PlayerSeat[]
  bottom_cards: Card[]
  current_turn: number
  calling_index: number
  last_caller: number | null
  last_play: PlayRecord | null
  leader_index: number
  pass_count: number
  winner_index: number | null
  winner_role: string
  revealed_hands?: RevealedHand[]
  message: string
  human_player: number
  online?: boolean
  my_hand: Card[]
  events?: GameEvent[]
  turn_deadline_unix?: number
  turn_seconds_left?: number
}

export interface DouDizhuRoom {
  id: string
  status: 'waiting' | 'playing'
  game_id?: string
  players: DouDizhuRoomPlayer[]
}

export interface DouDizhuRoomPlayer {
  user_id: number
  username: string
  ready: boolean
}

export interface GameMeta {
  type: string
  name: string
  description: string
  enabled: boolean
}

export interface DouDizhuHint {
  action: 'play' | 'pass' | 'none'
  card_ids: string[]
  message: string
}
