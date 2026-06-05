/** Online lobby modes (full-human; player_count from backend mode registry). */
export const YZS_ONLINE_MODES = {
  '1v1': { label: '1v1 真人对战', playerCount: 2 },
  '2v2': { label: '2v2 十字阵', playerCount: 4 },
  '3p_chain': { label: '杀上保下', playerCount: 3 },
} as const

export type YzsOnlineModeId = keyof typeof YZS_ONLINE_MODES

export function normalizeOnlineMode(mode?: string): YzsOnlineModeId {
  switch (mode) {
    case '2v2':
    case '2V2':
      return '2v2'
    case '3p_chain':
    case '3p':
    case '杀上保下':
      return '3p_chain'
    default:
      return '1v1'
  }
}

export function onlineModeMeta(mode?: string) {
  return YZS_ONLINE_MODES[normalizeOnlineMode(mode)]
}
