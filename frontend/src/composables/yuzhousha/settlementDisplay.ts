import type { YuzhoushaState } from '../../types/yuzhousha'

/** 与后端 engine/mode/identity.go 阵营编号一致 */
export const YZS_IDENTITY_TEAM = {
  lordFaction: 0,
  rebel: 1,
  spy: 2,
} as const

export interface SettlementDisplay {
  isFinished: boolean
  centerHint: string
  showRestart: boolean
  winnerIndex: number | null
  winnerTeam: number | null
}

/** 前端实际展示的结算文案（对局页中心 + 结束态） */
export function getSettlementDisplay(state: YuzhoushaState | null | undefined): SettlementDisplay {
  const isFinished = state?.phase === 'finished'
  return {
    isFinished,
    centerHint: isFinished ? (state?.message?.trim() ?? '') : '',
    showRestart: isFinished,
    winnerIndex: state?.winner_index ?? null,
    winnerTeam: state?.winner_team ?? null,
  }
}

function isIdentityMode(mode: string | undefined): boolean {
  return mode === 'identity_5' || mode === 'identity_8'
}

function hasGameOverEvent(state: YuzhoushaState): boolean {
  return (state.events ?? []).some((e) => e.type === 'game_over')
}

function messageMatchesIdentityTeam(message: string, team: number): boolean {
  switch (team) {
    case YZS_IDENTITY_TEAM.rebel:
      return /反贼/.test(message)
    case YZS_IDENTITY_TEAM.spy:
      return /内奸/.test(message)
    case YZS_IDENTITY_TEAM.lordFaction:
      return /主公/.test(message) || /己方/.test(message)
    default:
      return false
  }
}

/**
 * 校验后端 PublicView 快照在前端能否正确展示结算。
 * 返回空数组表示通过。
 */
export function validateSettlementState(state: YuzhoushaState): string[] {
  const errors: string[] = []
  const display = getSettlementDisplay(state)
  const mode = state.mode ?? '1v1'

  if (state.phase !== 'finished') {
    errors.push(`phase=${state.phase}, want finished`)
    return errors
  }

  if (!display.centerHint) {
    errors.push('finished but message/centerHint empty')
  }

  if (state.winner_index == null) {
    errors.push('winner_index missing')
  } else if (state.winner_index < 0 || state.winner_index >= state.players.length) {
    errors.push(`winner_index ${state.winner_index} out of range`)
  }

  if (isIdentityMode(mode) && state.winner_team == null) {
    errors.push('identity mode winner_team missing')
  }

  if (!hasGameOverEvent(state)) {
    errors.push('events missing game_over')
  }

  if (isIdentityMode(mode)) {
    if (state.winner_team != null && display.centerHint) {
      if (!messageMatchesIdentityTeam(display.centerHint, state.winner_team)) {
        errors.push(
          `identity message "${display.centerHint}" does not match winner_team=${state.winner_team}`,
        )
      }
    }
    const lordSeat = state.lord_seat ?? 0
    const lord = state.players[lordSeat]
    if (lord && state.winner_team === YZS_IDENTITY_TEAM.rebel && lord.hp > 0) {
      errors.push('rebel win but lord still alive')
    }
  }

  if (mode === '1v1' && display.centerHint && !/获胜/.test(display.centerHint)) {
    errors.push(`1v1 message should contain 获胜: "${display.centerHint}"`)
  }

  if ((mode === '2v2' || mode === '3v3' || mode === '3p_ddz') && display.centerHint) {
    if (!/获胜/.test(display.centerHint)) {
      errors.push(`team mode message should contain 获胜: "${display.centerHint}"`)
    }
  }

  if (mode === 'identity_8' && state.layout_key && state.layout_key !== 'octagon_8') {
    errors.push(`identity_8 layout_key=${state.layout_key}, want octagon_8`)
  }
  if (mode === 'identity_5' && state.layout_key && state.layout_key !== 'pentagon_5') {
    errors.push(`identity_5 layout_key=${state.layout_key}, want pentagon_5`)
  }

  return errors
}
