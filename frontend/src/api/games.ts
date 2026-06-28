import { getAppConfig } from '../config/loadConfig'
import { loadSession } from './auth'
import { parseResponse, readApiError } from './client'
import type { DouDizhuHint, DouDizhuRoom, DouDizhuState, GameMeta } from '../types/doudizhu'
import type { ZhajinhuaRoom, ZhajinhuaState } from '../types/zhajinhua'
import type { DouNiuRoom, DouNiuState } from '../types/douniu'
import type { UnoRoom, UnoState } from '../types/uno'
import type { YuzhoushaState, YzsModeMeta, YzsPackMeta, YzsHeroesPage, YzsHeroesQuery, YuzhoushaRoom } from '../types/yuzhousha'

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { apiBaseUrl } = getAppConfig()
  const session = loadSession()
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string>),
  }

  const hasBody = init?.body !== undefined
  if (hasBody) {
    headers['Content-Type'] = 'application/json'
  }

  if (session?.token) {
    headers.Authorization = `Bearer ${session.token}`
  }

  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}${path}`, { ...init, headers })
  } catch {
    throw new Error('无法连接后端，请确认 backend 已启动')
  }

  const data = await parseResponse<T & { error?: string }>(response)
  if (!response.ok) {
    throw new Error(readApiError(data, `请求失败 (${response.status})`))
  }
  return data
}

export function fetchGameCatalog() {
  return apiFetch<{ games: GameMeta[] }>('/api/games/catalog')
}

export function startDouDizhuGame() {
  return apiFetch<DouDizhuState>('/api/games/doudizhu/start', { method: 'POST' })
}

export function getDouDizhuState(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}`)
}

export function callLandlord(gameId: string, call: boolean) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/call`, {
    method: 'POST',
    body: JSON.stringify({ call }),
  })
}

export function playCards(gameId: string, cardIds: string[]) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/play`, {
    method: 'POST',
    body: JSON.stringify({ card_ids: cardIds }),
  })
}

export function passTurn(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/pass`, {
    method: 'POST',
  })
}

export function fetchDouDizhuHint(gameId: string) {
  return apiFetch<DouDizhuHint>(`/api/games/doudizhu/${gameId}/hint`)
}

export function tickDouDizhuGame(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/tick`, { method: 'POST' })
}

export function joinDouDizhuRoom(roomId?: string) {
  return apiFetch<DouDizhuRoom>('/api/games/doudizhu/rooms/join', {
    method: 'POST',
    body: JSON.stringify(roomId ? { room_id: roomId } : {}),
  })
}

export function fetchDouDizhuRoom(roomId: string) {
  return apiFetch<DouDizhuRoom>(`/api/games/doudizhu/rooms/${roomId}`)
}

export function leaveDouDizhuRoom(roomId: string) {
  return apiFetch<DouDizhuRoom | { left: true }>(`/api/games/doudizhu/rooms/${roomId}/leave`, {
    method: 'POST',
  })
}

export function readyDouDizhuRoom(roomId: string, ready: boolean) {
  return apiFetch<DouDizhuRoom>(`/api/games/doudizhu/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function nextDouDizhuRoom(roomId: string, gameId: string, ready: boolean) {
  return apiFetch<DouDizhuRoom>(
    `/api/games/doudizhu/rooms/${roomId}/next?game_id=${encodeURIComponent(gameId)}`,
    {
      method: 'POST',
      body: JSON.stringify({ ready }),
    },
  )
}

export function startZhajinhuaGame(botCount: number) {
  return apiFetch<ZhajinhuaState>('/api/games/zhajinhua/start', {
    method: 'POST',
    body: JSON.stringify({ bot_count: botCount }),
  })
}

export function getZhajinhuaState(gameId: string) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}`)
}

export function zhajinhuaLook(gameId: string) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/look`, { method: 'POST' })
}

export function zhajinhuaFold(gameId: string) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/fold`, { method: 'POST' })
}

export function zhajinhuaFollow(gameId: string) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/follow`, { method: 'POST' })
}

export function zhajinhuaRaise(gameId: string, amount: number) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/raise`, {
    method: 'POST',
    body: JSON.stringify({ amount }),
  })
}

export function zhajinhuaCompare(gameId: string, targetIndex: number) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/compare`, {
    method: 'POST',
    body: JSON.stringify({ target_index: targetIndex }),
  })
}

export function tickZhajinhuaGame(gameId: string) {
  return apiFetch<ZhajinhuaState>(`/api/games/zhajinhua/${gameId}/tick`, { method: 'POST' })
}

export function joinZhajinhuaRoom(roomId?: string) {
  return apiFetch<ZhajinhuaRoom>('/api/games/zhajinhua/rooms/join', {
    method: 'POST',
    body: JSON.stringify(roomId ? { room_id: roomId } : {}),
  })
}

export function fetchZhajinhuaRoom(roomId: string) {
  return apiFetch<ZhajinhuaRoom>(`/api/games/zhajinhua/rooms/${roomId}`)
}

export function leaveZhajinhuaRoom(roomId: string) {
  return apiFetch<ZhajinhuaRoom>(`/api/games/zhajinhua/rooms/${roomId}/leave`, { method: 'POST' })
}

export function readyZhajinhuaRoom(roomId: string, ready: boolean) {
  return apiFetch<ZhajinhuaRoom>(`/api/games/zhajinhua/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startZhajinhuaRoom(roomId: string) {
  return apiFetch<ZhajinhuaRoom>(`/api/games/zhajinhua/rooms/${roomId}/start`, { method: 'POST' })
}

export function nextZhajinhuaRoom(roomId: string, ready: boolean) {
  return apiFetch<ZhajinhuaRoom>(`/api/games/zhajinhua/rooms/${roomId}/next`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startUnoGame(botCount: number) {
  return apiFetch<UnoState>('/api/games/uno/start', {
    method: 'POST',
    body: JSON.stringify({ bot_count: botCount }),
  })
}

export function getUnoState(gameId: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}`)
}

export function playUnoCard(gameId: string, cardId: string, color?: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}/play`, {
    method: 'POST',
    body: JSON.stringify(color ? { card_id: cardId, color } : { card_id: cardId }),
  })
}

export function drawUnoCard(gameId: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}/draw`, { method: 'POST' })
}

export function voteEndUno(gameId: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}/vote-end`, { method: 'POST' })
}

export function rollUnoFirst(gameId: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}/roll-first`, { method: 'POST' })
}

export function tickUnoGame(gameId: string) {
  return apiFetch<UnoState>(`/api/games/uno/${gameId}/tick`, { method: 'POST' })
}

export function joinUnoRoom(roomId?: string) {
  return apiFetch<UnoRoom>('/api/games/uno/rooms/join', {
    method: 'POST',
    body: JSON.stringify(roomId ? { room_id: roomId } : {}),
  })
}

export function fetchUnoRoom(roomId: string) {
  return apiFetch<UnoRoom>(`/api/games/uno/rooms/${roomId}`)
}

export function leaveUnoRoom(roomId: string) {
  return apiFetch<UnoRoom>(`/api/games/uno/rooms/${roomId}/leave`, { method: 'POST' })
}

export function readyUnoRoom(roomId: string, ready: boolean) {
  return apiFetch<UnoRoom>(`/api/games/uno/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startUnoRoom(roomId: string) {
  return apiFetch<UnoRoom>(`/api/games/uno/rooms/${roomId}/start`, { method: 'POST' })
}

export function nextUnoRoom(roomId: string, ready: boolean) {
  return apiFetch<UnoRoom>(`/api/games/uno/rooms/${roomId}/next`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startDouNiuGame(botCount: number, previousGameId?: string) {
  return apiFetch<DouNiuState>('/api/games/douniu/start', {
    method: 'POST',
    body: JSON.stringify({
      bot_count: botCount,
      previous_game_id: previousGameId || undefined,
    }),
  })
}

export function getDouNiuState(gameId: string) {
  return apiFetch<DouNiuState>(`/api/games/douniu/${gameId}`)
}

export function grabDouNiuBanker(gameId: string, multiplier: number) {
  return apiFetch<DouNiuState>(`/api/games/douniu/${gameId}/grab`, {
    method: 'POST',
    body: JSON.stringify({ multiplier }),
  })
}

export function betDouNiu(gameId: string, multiplier: number) {
  return apiFetch<DouNiuState>(`/api/games/douniu/${gameId}/bet`, {
    method: 'POST',
    body: JSON.stringify({ multiplier }),
  })
}

export function tickDouNiuGame(gameId: string) {
  return apiFetch<DouNiuState>(`/api/games/douniu/${gameId}/tick`, { method: 'POST' })
}

export function joinDouNiuRoom(roomId?: string) {
  return apiFetch<DouNiuRoom>('/api/games/douniu/rooms/join', {
    method: 'POST',
    body: JSON.stringify(roomId ? { room_id: roomId } : {}),
  })
}

export function fetchDouNiuRoom(roomId: string) {
  return apiFetch<DouNiuRoom>(`/api/games/douniu/rooms/${roomId}`)
}

export function leaveDouNiuRoom(roomId: string) {
  return apiFetch<DouNiuRoom>(`/api/games/douniu/rooms/${roomId}/leave`, { method: 'POST' })
}

export function readyDouNiuRoom(roomId: string, ready: boolean) {
  return apiFetch<DouNiuRoom>(`/api/games/douniu/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startDouNiuRoom(roomId: string) {
  return apiFetch<DouNiuRoom>(`/api/games/douniu/rooms/${roomId}/start`, { method: 'POST' })
}

export function nextDouNiuRoom(roomId: string, ready: boolean) {
  return apiFetch<DouNiuRoom>(`/api/games/douniu/rooms/${roomId}/next`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function fetchYuzhoushaModes() {
  return apiFetch<{ modes: YzsModeMeta[] }>('/api/games/yuzhousha/modes')
}

export function fetchYuzhoushaPacks() {
  return apiFetch<{ packs: YzsPackMeta[] }>('/api/games/yuzhousha/packs')
}

export function fetchYuzhoushaHeroes(query: YzsHeroesQuery = {}) {
  const params = new URLSearchParams()
  if (query.mode) params.set('mode', query.mode)
  if (query.kingdom) params.set('kingdom', query.kingdom)
  if (query.pack) params.set('pack', query.pack)
  if (query.page) params.set('page', String(query.page))
  if (query.page_size) params.set('page_size', String(query.page_size))
  const qs = params.toString()
  return apiFetch<YzsHeroesPage>(`/api/games/yuzhousha/heroes${qs ? `?${qs}` : ''}`)
}

export function startYuzhoushaGame(characterId: string, mode = '1v1') {
  return apiFetch<YuzhoushaState>('/api/games/yuzhousha/start', {
    method: 'POST',
    body: JSON.stringify({ character_id: characterId, mode }),
  })
}

export interface YuzhoushaSkillPayload {
  targetIndex?: number
  cardIds?: string[]
  targetZone?: string
  targetCardId?: string
}

export function useYuzhoushaSkill(
  gameId: string,
  skillId: string,
  payload: YuzhoushaSkillPayload = {},
) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/skill`, {
    method: 'POST',
    body: JSON.stringify({
      skill_id: skillId,
      target_index: payload.targetIndex ?? 0,
      card_ids: payload.cardIds ?? [],
      target_zone: payload.targetZone ?? '',
      target_card_id: payload.targetCardId ?? '',
    }),
  })
}

export function getYuzhoushaState(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}`)
}

export interface YuzhoushaPlayTarget {
  targetIndex: number
  secondTargetIndex?: number
  targetZone?: string
  targetCardId?: string
  zhangbaSecondCardId?: string   // 丈八蛇矛：第二张手牌ID
  fangtianExtraTargets?: number[] // 方天画戟：额外目标列表
}

export function playYuzhoushaCard(
  gameId: string,
  cardId: string,
  target: number | YuzhoushaPlayTarget,
) {
  const body =
    typeof target === 'number'
      ? { card_id: cardId, target_index: target }
      : {
          card_id: cardId,
          target_index: target.targetIndex,
          second_target_index: target.secondTargetIndex,
          target_zone: target.targetZone,
          target_card_id: target.targetCardId,
          zhangba_second_card_id: target.zhangbaSecondCardId,
          fangtian_extra_targets: target.fangtianExtraTargets,
        }
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/play`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function respondYuzhoushaShan(gameId: string, cardId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/shan`, {
    method: 'POST',
    body: JSON.stringify({ card_id: cardId }),
  })
}

export function respondYuzhoushaCard(gameId: string, cardId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/respond`, {
    method: 'POST',
    body: JSON.stringify({ card_id: cardId }),
  })
}

export function passYuzhoushaResponse(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/pass`, { method: 'POST' })
}

export function passAllWuxiek(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/pass-all-wuxiek`, { method: 'POST' })
}

export function baguaYuzhoushaJudge(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/bagua`, { method: 'POST' })
}

export function endYuzhoushaPlay(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/end`, { method: 'POST' })
}

export function discardYuzhoushaCards(gameId: string, cardIds: string[]) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/discard`, {
    method: 'POST',
    body: JSON.stringify({ card_ids: cardIds }),
  })
}

export function respondZhangbaSha(gameId: string, cardIDs: [string, string]) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/respond-zhangba`, {
    method: 'POST',
    body: JSON.stringify({ card_ids: cardIDs }),
  })
}

export function passYuzhoushaPrepare(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/prepare/pass`, {
    method: 'POST',
  })
}

export function passYuzhoushaDraw(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/draw/pass`, {
    method: 'POST',
  })
}

export function finishYuzhoushaPeekDeck(
  gameId: string,
  payload: { top_card_ids: string[]; bottom_card_ids: string[] },
) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/peek-deck`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

/** @deprecated 使用 finishYuzhoushaPeekDeck */
export function finishYuzhoushaGuanxing(
  gameId: string,
  payload: { top_card_ids: string[]; bottom_card_ids: string[] },
) {
  return finishYuzhoushaPeekDeck(gameId, payload)
}

export function tickYuzhoushaGame(gameId: string) {
  return apiFetch<YuzhoushaState>(`/api/games/yuzhousha/${gameId}/tick`, { method: 'POST' })
}

export function joinYuzhoushaRoom(payload: { room_id?: string; mode?: string } = {}) {
  return apiFetch<YuzhoushaRoom>('/api/games/yuzhousha/rooms/join', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function fetchYuzhoushaRoom(roomId: string) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}`)
}

export function leaveYuzhoushaRoom(roomId: string) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}/leave`, { method: 'POST' })
}

export function setYuzhoushaRoomHero(roomId: string, characterId: string) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}/hero`, {
    method: 'POST',
    body: JSON.stringify({ character_id: characterId }),
  })
}

export function readyYuzhoushaRoom(roomId: string, ready = true) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function startYuzhoushaRoom(roomId: string) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}/start`, { method: 'POST' })
}

export function nextYuzhoushaRoom(roomId: string, ready = true) {
  return apiFetch<YuzhoushaRoom>(`/api/games/yuzhousha/rooms/${roomId}/next`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}
