/** Standard hidden-role identity modes (5p / 8p). */
export function isIdentityMode(mode?: string): boolean {
  return mode === 'identity_5' || mode === 'identity_8'
}

/** Modes where lord skills (激将/护驾/救援) are active. */
export function lordSkillsActiveInMode(mode?: string): boolean {
  return mode === '2v2' || isIdentityMode(mode)
}

/** Whether a skill marked inactive_in_1v1 is actually blocked in the current mode. */
export function skillBlockedInMode(
  skill: { inactive_in_1v1?: boolean },
  mode?: string,
): boolean {
  if (!skill.inactive_in_1v1) return false
  return !lordSkillsActiveInMode(mode)
}
