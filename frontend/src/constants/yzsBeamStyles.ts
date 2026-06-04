export const YZS_BEAM_STYLE_STORAGE_KEY = 'yuzhousha.beamStyle'

export type YzsBeamStyleId =
  | 'dash-flow'
  | 'solid-glow'
  | 'laser'
  | 'dot-pulse'
  | 'neon'
  | 'arrow'

export interface YzsBeamStyle {
  id: YzsBeamStyleId
  label: string
  flyClassName: string
}

export const YZS_BEAM_STYLES: YzsBeamStyle[] = [
  { id: 'dash-flow', label: '流光', flyClassName: 'yzs-fly-bolt--dash-flow' },
  { id: 'solid-glow', label: '光晕', flyClassName: 'yzs-fly-bolt--solid-glow' },
  { id: 'laser', label: '激光', flyClassName: 'yzs-fly-bolt--laser' },
  { id: 'dot-pulse', label: '点迹', flyClassName: 'yzs-fly-bolt--dot-pulse' },
  { id: 'neon', label: '霓虹', flyClassName: 'yzs-fly-bolt--neon' },
  { id: 'arrow', label: '箭矢', flyClassName: 'yzs-fly-bolt--arrow' },
]

export function getYzsBeamStyle(id: string): YzsBeamStyle {
  return YZS_BEAM_STYLES.find((s) => s.id === id) ?? YZS_BEAM_STYLES[0]
}

export function loadSavedYzsBeamStyle(): YzsBeamStyleId {
  try {
    const saved = localStorage.getItem(YZS_BEAM_STYLE_STORAGE_KEY)
    if (saved && YZS_BEAM_STYLES.some((s) => s.id === saved)) {
      return saved as YzsBeamStyleId
    }
  } catch {
    // ignore
  }
  return 'dash-flow'
}
