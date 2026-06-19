import { useYuzhoushaSkill } from '../../../../api/games'
import { isBusy, pickFirstTarget, responseMode } from '../helpers'
import type { PendingHandler } from '../types'
import type { PendingContext } from '../context'

/**
 * 工厂函数：为「从 Subject 的 Zone 取牌」类技能生成 PendingHandler。
 *
 * 适用技能：反馈、突袭、奇袭、破军（take 类）。
 * 前提：后端已写入 window_kind=Take，actor_seat / subject_seat 已输出到 JSON。
 *
 * @param opts.modes        - 对应的 response_mode 列表
 * @param opts.skillId      - 技能 ID，如 'fankui'
 * @param opts.hint         - 提示文案（固定字符串或根据 ctx 动态计算）
 * @param opts.zoneFilter   - 可选，限制可选区；默认全部（hand/weapon/armor/plus_horse/minus_horse/judge）
 * @param opts.maxKey       - 可选，Pending 中表示剩余可拿数的字段名（如 'pojun_max'），用于 hint 动态显示剩余
 * @param opts.placedKey    - 可选，已拿数的字段名（如 'pojun_placed'），与 maxKey 配合计算剩余
 * @param opts.targetOptionsKey - 可选，ctx 上目标选项字段名；默认 `${skillId}TargetOptions`
 */
export function makeTakeWindowHandler(opts: {
  modes: string[]
  skillId: string
  hint?: string | ((ctx: PendingContext) => string)
  zoneFilter?: string[]
  maxKey?: string
  placedKey?: string
  targetOptionsKey?: string
}): PendingHandler {
  const {
    modes,
    skillId,
    hint,
    zoneFilter,
    maxKey,
    placedKey,
    targetOptionsKey,
  } = opts

  const optionsKey = targetOptionsKey ?? `${skillId}TargetOptions`

  function getTargetOptions(ctx: PendingContext): { zone: string; cardId: string }[] {
    const val = (ctx as unknown as Record<string, unknown>)[optionsKey]
    if (Array.isArray(val)) return val as { zone: string; cardId: string }[]
    return []
  }

  function hasHandOption(options: { zone: string }[]): boolean {
    return options.some(o => o.zone === 'hand')
  }

  function defaultHint(ctx: PendingContext): string {
    const skillName = ctx.myCharacterSkills.value.find(s => s.id === skillId)?.name ?? skillId
    if (maxKey && placedKey) {
      const pending = ctx.state.pending as unknown as Record<string, number>
      const max = pending?.[maxKey] ?? 0
      const placed = pending?.[placedKey] ?? 0
      const left = Math.max(0, max - placed)
      return `【${skillName}】：选择目标至多 ${left} 张牌，或「取消」结束`
    }
    return `【${skillName}】：选择来源的一张牌，再点「${skillName}」`
  }

  return {
    modes,
    match: (state) => responseMode(state, modes[0]),
    skillOnly: true,

    onEnter(ctx) {
      pickFirstTarget(ctx, getTargetOptions(ctx))
    },

    canSubmitSkill(ctx, sid) {
      if (sid !== skillId || isBusy(ctx)) return false
      if (ctx.selectedTargetZone.value) return true
      const options = getTargetOptions(ctx)
      if (zoneFilter) {
        return options.some(o => zoneFilter.includes(o.zone))
      }
      return hasHandOption(options)
    },

    async submitSkill(ctx) {
      const zone = ctx.selectedTargetZone.value || 'hand'
      await ctx.act(() =>
        useYuzhoushaSkill(ctx.state.id, skillId, {
          targetZone: zone,
          targetCardId: ctx.selectedTargetCardId.value,
        }),
      )
      ctx.selectedTargetZone.value = ''
      ctx.selectedTargetCardId.value = ''
    },

    hint(ctx) {
      if (!hint) return defaultHint(ctx)
      return typeof hint === 'function' ? hint(ctx) : hint
    },
  }
}
