/** Online lobby modes (full-human; player_count from backend mode registry). */
export const YZS_ONLINE_MODES = {
  '1v1': { label: '1v1 真人对战', playerCount: 2 },
  '2v2': { label: '2v2 十字阵', playerCount: 4 },
  '3p_chain': { label: '杀上保下', playerCount: 3 },
  '3v3': { label: '3v3 竞技', playerCount: 6 },
  identity_5: { label: '5 人身份局', playerCount: 5 },
  identity_8: { label: '8 人身份局', playerCount: 8 },
} as const

/** Join order → seat role labels for online 3v3. */
export const YZS_ONLINE_3V3_SEAT_ROLES = [
  '暖色主帅',
  '冷前锋',
  '冷色主帅',
  '冷前锋',
  '暖前锋',
  '暖前锋',
] as const

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
    case '3v3':
    case '3V3':
      return '3v3'
    case 'identity_5':
    case 'identity':
    case '身份局':
      return 'identity_5'
    case 'identity_8':
    case '8人身份局':
      return 'identity_8'
    default:
      return '1v1'
  }
}

export function onlineModeMeta(mode?: string) {
  return YZS_ONLINE_MODES[normalizeOnlineMode(mode)]
}
