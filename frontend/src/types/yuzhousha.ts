export interface YzsSkillMeta {
  id: string
  name: string
  desc: string
  kind: 'passive' | 'active' | 'lord' | 'awakening' | string
  inactive_in_1v1?: boolean
}

export interface YzsHeroDisplay {
  skin_id: string
  portrait_url?: string
  accent_color?: string
}

export const YZS_KINGDOM_LABELS: Record<string, string> = {
  shu: '蜀',
  wei: '魏',
  wu: '吴',
  qun: '群',
}

export interface YzsCharacter {
  id: string
  name: string
  max_hp: number
  kingdom?: string
  skill_ids?: string[]
  skills?: YzsSkillMeta[]
  accent_color?: string
  portrait_url?: string
  pack?: string
  default_skin_id?: string
  skin_id?: string
  display?: YzsHeroDisplay
}

export interface YzsHeroesPage {
  heroes: YzsCharacter[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface YzsHeroesQuery {
  mode?: string
  kingdom?: string
  pack?: string
  page?: number
  page_size?: number
}

export interface YzsPackMeta {
  id: string
  name: string
  description?: string
  hero_pack: string
  skin_pack?: string
}

export interface YzsCard {
  id: string
  kind: 'sha' | 'shan' | 'tao' | string
  suit?: string
  rank?: number
  label?: string
  name: string
}

export interface YzsPlayer {
  index: number
  name: string
  is_ai: boolean
  team?: number
  identity?: 'lord' | 'loyalist' | 'spy' | 'rebel' | string
  identity_revealed?: boolean
  character: YzsCharacter
  hp: number
  max_hp: number
  hand_count: number
  sha_used_this_turn: boolean
  sha_extra_used_this_turn?: boolean
  skip_play?: boolean
  skip_draw?: boolean
  drunk?: boolean
  weapon?: YzsCard
  armor?: YzsCard
  plus_horse?: YzsCard
  minus_horse?: YzsCard
  judge_area?: YzsCard[]
  camp_cards?: YzsCard[]
  hand?: YzsCard[]
  skill_counters?: Record<string, number>
}

export interface YuzhoushaRoomPlayer {
  user_id: number
  username: string
  ready: boolean
  character_id?: string
}

export interface YuzhoushaRoom {
  id: string
  mode?: string
  status: 'waiting' | 'playing' | string
  game_id?: string
  host_user_id: number
  players: YuzhoushaRoomPlayer[]
}

export interface YzsPendingCombat {
  source_index: number
  target_index: number
  /** 当前应操作的座位；-1 表示无 */
  actor_seat?: number
  /** 被操作座位（拿牌/选目标区等） */
  subject_seat?: number
  /** 事件来源（resume/伤害链） */
  origin_seat?: number
  /** respond | take | discard | choice | peek | pick */
  window_kind?: string
  card: YzsCard
  required_kind?: string
  response_mode?: string
  allow_wuxiek?: boolean
  bagua_used?: boolean
  ignore_armor?: boolean
  effect_target?: number
  revealed_cards?: YzsCard[]
  wugu_pick_seat?: number
  skill_id?: string
  jijiang_lord?: number
  jijiang_use?: boolean
  judge_card?: YzsCard
  ganglie_owner?: number
  yiji_give_remaining?: number
  pojun_max?: number
  pojun_placed?: number
  pojun_remaining?: number
}

export interface YzsEvent {
  type: string
  player_index?: number
  target_index?: number
  card?: YzsCard
  message?: string
  damage?: number
  heal?: number
  amount?: number
  skill_id?: string
}

export interface YzsSeatSlot {
  seat: number
  area: string
  placement: 'top' | 'left' | 'right' | string
  is_teammate: boolean
  seat_role?: 'protect' | 'mark' | 'landlord' | 'farmer' | string
}

export interface YuzhoushaState {
  id: string
  phase: 'playing' | 'response' | 'finished' | string
  turn_step: 'draw' | 'play' | 'discard' | string
  current_turn: number
  human_player: number
  players: YzsPlayer[]
  pending?: YzsPendingCombat
  message: string
  winner_index?: number
  winner_team?: number
  mode?: '1v1' | '2v2' | '3p_chain' | '3p_ddz' | '3v3' | 'identity_5' | 'identity_8' | string
  layout_key?: string
  seat_map?: YzsSeatSlot[]
  landlord_seat?: number
  lord_seat?: number
  draw_count: number
  discard_count: number
  my_hand?: YzsCard[]
  turn_deadline_unix: number
  events?: YzsEvent[]
  activatable_skills?: YzsSkillMeta[]
}

export interface YzsModeMeta {
  id: string
  name: string
  tag?: string
  description: string
  hint?: string
  subtitle?: string
  layout_key: string
  tags: string[]
  rules?: string[]
  player_count: number
  human_seats: number[]
  seat_map?: YzsSeatSlot[]
  hero_pool?: { packs?: string[]; kingdoms?: string[] }
}

export const YZS_CARD_LABELS: Record<string, string> = {
  sha: '杀',
  shan: '闪',
  tao: '桃',
  jiu: '酒',
  guohe: '过河拆桥',
  tannang: '顺手牵羊',
  nanman: '南蛮入侵',
  wanjian: '万箭齐发',
  juedou: '决斗',
  lebu: '乐不思蜀',
  bingliang: '兵粮寸断',
  shandian: '闪电',
  wugu: '五谷丰登',
  taoyuan: '桃园结义',
  wuzhong: '无中生有',
  wuxiek: '无懈可击',
  weapon_1: '诸葛连弩',
  weapon_2: '青釭剑',
  weapon_3: '青龙偃月刀',
  weapon_4: '方天画戟',
  weapon_5: '麒麟弓',
  weapon_6: '古锭刀',
  armor: '八卦阵',
  armor_vine: '藤甲',
  huogong: '火攻',
  tiesuo: '铁索连环',
  plus_horse: '+1马',
  minus_horse: '-1马',
}
