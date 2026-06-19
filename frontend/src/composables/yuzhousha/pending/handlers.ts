import {
  finishYuzhoushaPeekDeck,
  playYuzhoushaCard,
  respondYuzhoushaCard,
  useYuzhoushaSkill,
} from '../../../api/games'
import { YZS_CARD_LABELS } from '../../../types/yuzhousha'
import { isBusy, responseAnyMode, responseMode } from './helpers'
import { makeTakeWindowHandler } from './templates/takeWindow'
import type { PendingHandler } from './types'

function cardLabel(kind: string | undefined) {
  if (!kind) return '牌'
  return YZS_CARD_LABELS[kind] ?? kind
}

const peekDeckHandler: PendingHandler = {
  modes: ['peek_deck'],
  match: (state) =>
    state.phase === 'response' &&
    state.pending?.response_mode === 'peek_deck' &&
    state.pending.actor_seat === state.human_player,
  allowsCancel: false,
  onEnter(ctx) {
    const ids = ctx.state.pending?.revealed_cards?.map((c) => c.id) ?? []
    ctx.peekDeckTopIds.value = [...ids]
    ctx.peekDeckBottomIds.value = []
  },
  canSubmitPlay(ctx) {
    if (!ctx.canUsePeekDeckUI || isBusy(ctx)) return false
    const total = ctx.state.pending?.revealed_cards?.length ?? 0
    return ctx.peekDeckTopIds.value.length + ctx.peekDeckBottomIds.value.length === total
  },
  async submitPlay(ctx) {
    await ctx.act(() =>
      finishYuzhoushaPeekDeck(ctx.state.id, {
        top_card_ids: [...ctx.peekDeckTopIds.value],
        bottom_card_ids: [...ctx.peekDeckBottomIds.value],
      }),
    )
  },
  hint(ctx) {
    const skillName =
      ctx.myCharacterSkills.value.find((s) => s.id === ctx.peekDeckSkillId.value)?.name ?? '看牌'
    return `【${skillName}】：拖拽调序，最左为牌堆顶（下次判定先亮）；拖至下方归入牌堆底`
  },
}

const fankuiHandler = makeTakeWindowHandler({
  modes: ['skill_fankui'],
  skillId: 'fankui',
  hint: (ctx) => ctx.centerMessage.value || '【反馈】：选择来源的一张牌，再点「反馈」',
})

const tuxiHandler = makeTakeWindowHandler({
  modes: ['skill_tuxi'],
  skillId: 'tuxi',
  hint: (ctx) => ctx.centerMessage.value || '【突袭】：选择获得对手的一张牌，再点「突袭」',
})

const qixiHandler = makeTakeWindowHandler({
  modes: ['skill_qixi'],
  skillId: 'qixi',
  hint: (ctx) => ctx.centerMessage.value || '【奇袭】：选择一张黑色牌（手牌或装备区）',
  zoneFilter: ['hand'],
})

const pojunHandler: PendingHandler = {
  modes: ['skill_pojun'],
  match: (state) => responseMode(state, 'skill_pojun'),
  skillOnly: true,
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'pojun') return false
    if (isBusy(ctx)) return false
    // 支持批量 cardIds（由 YuzhoushaView 在 onPojunConfirm 中设置）
    if (ctx.pojunCardIds?.value?.length) return true
    // 兼容单选
    return !!ctx.selectedTargetZone.value
  },
  async submitSkill(ctx) {
    const cardIds = ctx.pojunCardIds?.value
    if (cardIds && cardIds.length > 0) {
      // 批量提交：一次性发送所有选中的牌
      await ctx.act(() =>
        useYuzhoushaSkill(ctx.state.id, 'pojun', {
          cardIds: [...cardIds],
        }),
      )
      ctx.pojunCardIds!.value = []
    } else {
      const zone = ctx.selectedTargetZone.value || 'hand'
      await ctx.act(() =>
        useYuzhoushaSkill(ctx.state.id, 'pojun', {
          targetZone: zone,
          targetCardId: ctx.selectedTargetCardId.value,
        }),
      )
      ctx.selectedTargetZone.value = ''
      ctx.selectedTargetCardId.value = ''
    }
  },
  hint(ctx) {
    const left = Math.max(
      0,
      (ctx.state.pending?.pojun_max ?? 0) - (ctx.state.pending?.pojun_placed ?? 0),
    )
    return ctx.centerMessage.value || `【破军】：选择目标至多 ${left} 张牌置于「营」，或「取消」结束`
  },
}

const pojunDiscardHandler: PendingHandler = {
  modes: ['skill_pojun_discard'],
  match: (state) => responseMode(state, 'skill_pojun_discard'),
  skillOnly: true,
  canPlayCard() {
    return true
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'pojun') return false
    return !!ctx.selectedId.value && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'pojun') return
    const cardId = ctx.selectedId.value
    if (!cardId) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'pojun', {
        cardIds: [cardId],
      }),
    )
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【破军】：选择一张手牌弃置，或「取消」跳过'
  },
}

const guoheHandler: PendingHandler = {
  modes: ['guohe'],
  match: (state) => responseMode(state, 'guohe'),
  skillOnly: true,
  canPlayCard() {
    return false
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== '') return false
    if (isBusy(ctx)) return false
    // 装备/判定区需要选具体 zone；手牌选背面牌即可（后端总是取第一张）
    return !!ctx.selectedTargetZone.value
  },
  async submitSkill(ctx, _skillId) {
    const zone = ctx.selectedTargetZone.value || 'hand'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, '', {
        targetZone: zone,
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【过河拆桥】：选择要拆掉的一张牌，或「取消」'
  },
}

const tannangHandler: PendingHandler = {
  modes: ['tannang'],
  match: (state) => responseMode(state, 'tannang'),
  skillOnly: true,
  canPlayCard() {
    return false
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== '') return false
    if (isBusy(ctx)) return false
    // 装备/判定区需要选具体 zone；手牌选背面牌即可（后端总是取第一张）
    return !!ctx.selectedTargetZone.value
  },
  async submitSkill(ctx, _skillId) {
    const zone = ctx.selectedTargetZone.value || 'hand'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, '', {
        targetZone: zone,
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【顺手牵羊】：选择要获得的一张牌，或「取消」'
  },
}

const guicaiHandler: PendingHandler = {
  modes: ['skill_guicai'],
  match: (state) => responseMode(state, 'skill_guicai'),
  skillOnly: true,
  suppressPlaySubmit: true,
  canSubmitSkill(ctx, skillId) {
    return skillId === 'guicai' && !!ctx.selectedId.value && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'guicai' || !ctx.selectedId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'guicai', {
        cardIds: [ctx.selectedId.value],
      }),
    )
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    const judge = ctx.state.pending?.judge_card?.label ?? '判定牌'
    return ctx.centerMessage.value || `【鬼才】：选手牌代替判定 ${judge}，或「取消」`
  },
}

const guidaoHandler: PendingHandler = {
  modes: ['skill_guidao'],
  match: (state) => responseMode(state, 'skill_guidao'),
  skillOnly: true,
  suppressPlaySubmit: true,
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'guidao' || isBusy(ctx)) return false
    const card = ctx.selectedCard.value
    return !!card && ctx.isBlackCard(card)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'guidao' || !ctx.selectedId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'guidao', {
        cardIds: [ctx.selectedId.value],
      }),
    )
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    const judge = ctx.state.pending?.judge_card?.label ?? '判定牌'
    return ctx.centerMessage.value || `【鬼道】：选黑色手牌代替判定 ${judge}，或「取消」`
  },
}

const leijiHandler: PendingHandler = {
  modes: ['skill_leiji_offer'],
  match: (state) => responseMode(state, 'skill_leiji_offer'),
  skillOnly: true,
  canSubmitSkill(ctx, skillId) {
    return skillId === 'leiji' && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'leiji') return
    await ctx.act(() => useYuzhoushaSkill(ctx.state.id, 'leiji'))
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【雷击】：点技能进行判定，或「取消」跳过'
  },
}

const ganglieOfferHandler: PendingHandler = {
  modes: ['skill_ganglie_offer'],
  match: (state) => responseMode(state, 'skill_ganglie_offer'),
  skillOnly: true,
  hint(ctx) {
    return ctx.centerMessage.value || '【刚烈】：点技能进行判定，或「取消」跳过'
  },
}

const ganglieChoiceHandler: PendingHandler = {
  modes: ['skill_ganglie_choice'],
  match: (state) => responseMode(state, 'skill_ganglie_choice'),
  skillOnly: true,
  allowsCancel: false,
  canSubmitSkill(ctx, skillId) {
    return (
      skillId === 'ganglie' &&
      ctx.ganglieDiscardIds.value.length === 2 &&
      !isBusy(ctx)
    )
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'ganglie' || ctx.ganglieDiscardIds.value.length < 2) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'ganglie', {
        cardIds: ctx.ganglieDiscardIds.value.slice(0, 2),
      }),
    )
    ctx.ganglieDiscardIds.value = []
  },
  canSubmitAction(ctx, action) {
    return action === 'ganglie_take_damage' && !isBusy(ctx)
  },
  async submitAction(ctx, action) {
    if (action !== 'ganglie_take_damage') return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'ganglie', { targetZone: 'take_damage' }),
    )
    ctx.ganglieDiscardIds.value = []
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【刚烈】：弃 2 张手牌，或点「受1点伤害」'
  },
}

const ddzJudgeCancelHandler: PendingHandler = {
  modes: ['ddz_judge_cancel'],
  match: (state) => responseMode(state, 'ddz_judge_cancel'),
  skillOnly: true,
  canSubmitSkill(ctx, skillId) {
    return (
      skillId === 'ddz_judge_cancel' &&
      ctx.ddzCancelDiscardIds.value.length === 2 &&
      !isBusy(ctx)
    )
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'ddz_judge_cancel' || ctx.ddzCancelDiscardIds.value.length < 2) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'ddz_judge_cancel', {
        cardIds: ctx.ddzCancelDiscardIds.value.slice(0, 2),
      }),
    )
    ctx.ddzCancelDiscardIds.value = []
  },
  hint(ctx) {
    return ctx.centerMessage.value || '选择两张手牌弃置，取消此次判定'
  },
}

const yijiOfferHandler: PendingHandler = {
  modes: ['skill_yiji_offer'],
  match: (state) => responseMode(state, 'skill_yiji_offer'),
  skillOnly: true,
  canSubmitSkill(ctx, skillId) {
    return skillId === 'yiji' && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'yiji') return
    await ctx.act(() => useYuzhoushaSkill(ctx.state.id, 'yiji'))
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【遗计】：点技能摸 2 张并分配手牌，或「取消」跳过'
  },
}

const yijiGiveHandler: PendingHandler = {
  modes: ['skill_yiji_give'],
  match: (state) => responseMode(state, 'skill_yiji_give'),
  onEnter(ctx) {
    ctx.shaTarget.value = ctx.opponentSeat
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'yiji' || isBusy(ctx)) return false
    return (
      ctx.yijiSelectedIds.value.length > 0 &&
      ctx.yijiSelectedIds.value.length <= ctx.yijiGiveRemaining.value &&
      ctx.shaTarget.value != null
    )
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'yiji' || ctx.shaTarget.value == null) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'yiji', {
        targetIndex: ctx.shaTarget.value!,
        cardIds: [...ctx.yijiSelectedIds.value],
      }),
    )
    ctx.yijiSelectedIds.value = []
    ctx.shaTarget.value = null
  },
  canSubmitAction(ctx, action) {
    return action === 'yiji_pass_give' && !isBusy(ctx)
  },
  async submitAction(ctx, action) {
    if (action !== 'yiji_pass_give') return
    await ctx.act(() => useYuzhoushaSkill(ctx.state.id, 'yiji', { cardIds: [] }))
    ctx.yijiSelectedIds.value = []
    ctx.shaTarget.value = null
  },
  hint(ctx) {
    const left = ctx.yijiGiveRemaining.value
    return ctx.centerMessage.value || `【遗计】：选至多 ${left} 张手牌交给对手，点「给出」或「完成」`
  },
}

const jianxiongHandler: PendingHandler = {
  modes: ['skill_jianxiong'],
  match: (state) => responseMode(state, 'skill_jianxiong'),
  skillOnly: true,
  hint(ctx) {
    const cardName = ctx.state.pending?.card?.name ?? '伤害牌'
    return ctx.centerMessage.value || `【奸雄】：可获得 ${cardName}，点技能或「取消」跳过`
  },
}

const tianxiangHandler: PendingHandler = {
  modes: ['skill_tianxiang'],
  match: (state) => responseMode(state, 'skill_tianxiang'),
  skillOnly: true,
  canPlayCard(ctx, card) {
    return ctx.isRedCard(card)
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'tianxiang' || isBusy(ctx)) return false
    const card = ctx.selectedCard.value
    return !!card && ctx.isRedCard(card)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'tianxiang' || !ctx.selectedId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'tianxiang', {
        cardIds: [ctx.selectedId.value],
      }),
    )
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【天香】：选一张红色手牌转移伤害，或「取消」承受'
  },
}

const liuliHandler: PendingHandler = {
  modes: ['skill_liuli'],
  match: (state) => responseMode(state, 'skill_liuli'),
  skillOnly: true,
  canSubmitSkill(ctx, skillId) {
    return skillId === 'liuli' && ctx.liuliSelectedId.value !== '' && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'liuli' || !ctx.liuliSelectedId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'liuli', {
        targetIndex: ctx.opponentSeat,
        cardIds: [ctx.liuliSelectedId.value],
      }),
    )
    ctx.liuliSelectedId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【流离】：选择一张手牌交给对手，再点「流离」'
  },
}

const fanjianSuitHandler: PendingHandler = {
  modes: ['skill_fanjian_suit'],
  match: (state) => responseMode(state, 'skill_fanjian_suit'),
  skillOnly: true,
  canSubmitAction(ctx, action) {
    return action.startsWith('fanjian_suit:') && !isBusy(ctx)
  },
  async submitAction(ctx, action) {
    const suit = action.slice('fanjian_suit:'.length)
    if (!suit) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'fanjian', {
        targetZone: suit,
      }),
    )
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【反间】：选择一种花色（猜中则受到 1 点伤害）'
  },
}

const yinghunChoiceHandler: PendingHandler = {
  modes: ['skill_yinghun'],
  match: (state) => responseMode(state, 'skill_yinghun'),
  skillOnly: true,
  canSubmitAction(ctx, action) {
    return (
      (action === 'yinghun_opp_draw_x_discard_1' || action === 'yinghun_opp_draw_1_discard_x') && !isBusy(ctx)
    )
  },
  async submitAction(ctx, action) {
    const option = action === 'yinghun_opp_draw_1_discard_x' ? 'opp_draw_1_discard_x' : 'opp_draw_x_discard_1'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'yinghun', {
        targetZone: option,
      }),
    )
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【英魂】：选择一项'
  },
}

const yinghunDiscardHandler: PendingHandler = {
  modes: ['skill_yinghun_discard'],
  match: (state) => responseMode(state, 'skill_yinghun_discard'),
  skillOnly: true,
  canPlayCard() {
    return true
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'yinghun') return false
    const extra = ctx.state.pending?.extra
    const need = (extra?.['yinghun_discard_need'] ?? 1) as number
    const done = (extra?.['yinghun_discard_done'] ?? 0) as number
    const remaining = need - done
    // 选项1：弃1张，使用 selectedId；选项2：弃X张，使用 selectedDiscardIds
    if (remaining > 1) {
      return ctx.selectedDiscardIds.value.length > 0 && ctx.selectedDiscardIds.value.length <= remaining && !isBusy(ctx)
    }
    return !!ctx.selectedId.value && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'yinghun') return
    const extra = ctx.state.pending?.extra
    const need = (extra?.['yinghun_discard_need'] ?? 1) as number
    const done = (extra?.['yinghun_discard_done'] ?? 0) as number
    const remaining = need - done

    let cardIds: string[]
    if (remaining > 1) {
      // 多选模式：使用 selectedDiscardIds
      if (ctx.selectedDiscardIds.value.length === 0) return
      cardIds = [...ctx.selectedDiscardIds.value]
      ctx.selectedDiscardIds.value = []
    } else {
      // 单选模式：使用 selectedId
      if (!ctx.selectedId.value) return
      cardIds = [ctx.selectedId.value]
      ctx.selectedId.value = ''
    }

    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'yinghun', {
        cardIds,
      }),
    )
  },
  hint(ctx) {
    const extra = ctx.state.pending?.extra
    const need = (extra?.['yinghun_discard_need'] ?? 1) as number
    const done = (extra?.['yinghun_discard_done'] ?? 0) as number
    const remaining = need - done
    if (remaining > 1) {
      const selected = ctx.selectedDiscardIds.value.length
      return ctx.centerMessage.value || `【英魂】：请选择 ${remaining} 张手牌弃置（已选 ${selected}/${remaining}）`
    }
    return ctx.centerMessage.value || '【英魂】：请选择一张手牌弃置'
  },
}

const luanwuHandler: PendingHandler = {
  modes: ['skill_luanwu'],
  match: (state) => responseMode(state, 'skill_luanwu'),
  canPlayCard(ctx, card) {
    return ctx.cardPlaysAsSha(card)
  },
  canSubmitPlay(ctx) {
    const card = ctx.selectedCard.value
    return !!card && ctx.cardPlaysAsSha(card) && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    const card = ctx.selectedCard.value
    if (!card || !ctx.cardPlaysAsSha(card)) return
    await ctx.act(() => playYuzhoushaCard(ctx.state.id, card.id, ctx.mySeat))
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【乱武】：出【杀】，或点「取消」承受 1 点伤害'
  },
}

const guanyuFollowHandler: PendingHandler = {
  modes: ['guanyu_follow'],
  match: (state) => responseMode(state, 'guanyu_follow'),
  canPlayCard(ctx, card) {
    return card.kind === 'sha' || ctx.cardPlaysAsSha(card)
  },
  canSubmitPlay(ctx) {
    const card = ctx.selectedCard.value
    return !!card && (card.kind === 'sha' || ctx.cardPlaysAsSha(card)) && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    const card = ctx.selectedCard.value
    if (!card) return
    await ctx.act(() =>
      playYuzhoushaCard(
        ctx.state.id,
        card.id,
        ctx.state.pending?.effect_target ?? ctx.opponentSeat,
      ),
    )
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【青龙偃月刀】追击：出【杀】或点「取消」'
  },
}

const qilinBowHandler: PendingHandler = {
  modes: ['qilin_bow'],
  match: (state) => responseMode(state, 'qilin_bow'),
  canPlayCard() {
    return false
  },
  canSubmitPlay(ctx) {
    return ctx.selectedQilinZone.value !== '' && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    if (!ctx.selectedQilinZone.value) return
    await ctx.act(() =>
      playYuzhoushaCard(ctx.state.id, '', {
        targetIndex: ctx.state.pending?.effect_target ?? ctx.opponentSeat,
        targetZone: ctx.selectedQilinZone.value,
      }),
    )
    ctx.selectedQilinZone.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【麒麟弓】：选择要弃置的坐骑，或点「取消」'
  },
}

const wuguPickHandler: PendingHandler = {
  modes: ['wugu_pick'],
  match: (state) => responseMode(state, 'wugu_pick'),
  // 选牌者不能取消，必须选牌
  get allowsCancel() {
    return false
  },
  canPlayCard(ctx, _card) {
    // 只有当前选牌者可以选牌
    if (ctx.state.pending?.actor_seat !== ctx.mySeat) return false
    return (
      !!ctx.selectedId.value &&
      (ctx.state.pending?.revealed_cards?.some((c) => c.id === ctx.selectedId.value) ?? false)
    )
  },
  canSubmitPlay(ctx) {
    // 只有当前选牌者可以提交
    if (ctx.state.pending?.actor_seat !== ctx.mySeat) return false
    return (
      !!ctx.selectedId.value &&
      (ctx.state.pending?.revealed_cards?.some((c) => c.id === ctx.selectedId.value) ?? false) &&
      !isBusy(ctx)
    )
  },
  async submitPlay(ctx) {
    if (!ctx.selectedId.value) return
    await ctx.act(() => playYuzhoushaCard(ctx.state.id, ctx.selectedId.value, ctx.mySeat))
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    if (ctx.state.pending?.actor_seat === ctx.mySeat) {
      return ctx.centerMessage.value || '请选择【五谷丰登】亮出的一张牌'
    }
    return ctx.centerMessage.value || '等待选牌中...'
  },
}

const dyingRescueHandler: PendingHandler = {
  modes: ['dying_rescue'],
  match: (state) => responseMode(state, 'dying_rescue'),
  canPlayCard(ctx, card) {
    return ctx.cardPlaysAsTao(card)
  },
  canSubmitPlay(ctx) {
    const card = ctx.selectedCard.value
    return !!card && ctx.cardPlaysAsTao(card) && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    const card = ctx.selectedCard.value
    if (!card || !ctx.cardPlaysAsTao(card)) return
    await ctx.act(() => respondYuzhoushaCard(ctx.state.id, card.id))
  },
}

const jijiangHandler: PendingHandler = {
  modes: ['skill_jijiang'],
  match: (state) => responseMode(state, 'skill_jijiang'),
  canPlayCard(ctx, card) {
    return ctx.cardPlaysAsSha(card)
  },
  canSubmitPlay(ctx) {
    const card = ctx.selectedCard.value
    return !!card && ctx.cardPlaysAsSha(card) && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    const card = ctx.selectedCard.value
    if (!card || !ctx.cardPlaysAsSha(card)) return
    await ctx.act(() => respondYuzhoushaCard(ctx.state.id, card.id))
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【激将】：请出【杀】或点「取消」'
  },
}

const wuxiekHandler: PendingHandler = {
  modes: ['wuxiek_trick', 'wuxiek_lebu', 'wuxiek_bingliang', 'wuxiek_shandian'],
  match: (state) =>
    responseAnyMode(state, [
      'wuxiek_trick',
      'wuxiek_lebu',
      'wuxiek_bingliang',
      'wuxiek_shandian',
    ]),
  canPlayCard(_ctx, card) {
    return card.kind === 'wuxiek'
  },
  canSubmitPlay(ctx) {
    const card = ctx.selectedCard.value
    return !!card && card.kind === 'wuxiek' && !isBusy(ctx)
  },
  async submitPlay(ctx) {
    const card = ctx.selectedCard.value
    if (!card || card.kind !== 'wuxiek') return
    await ctx.act(() => respondYuzhoushaCard(ctx.state.id, card.id))
  },
  hint(ctx) {
    const mode = ctx.state.pending?.response_mode
    if (mode === 'wuxiek_lebu') {
      return ctx.centerMessage.value || '判定前可出【无懈可击】抵消【乐不思蜀】，或点「取消」进行判定'
    }
    if (mode === 'wuxiek_bingliang') {
      return ctx.centerMessage.value || '判定前可出【无懈可击】抵消【兵粮寸断】，或点「取消」跳过摸牌'
    }
    if (mode === 'wuxiek_shandian') {
      return ctx.centerMessage.value || '判定前可出【无懈可击】抵消【闪电】，或点「取消」进行判定'
    }
    return ctx.centerMessage.value || '可出【无懈可击】抵消该锦囊，或点「取消」让效果生效'
  },
}

/** Registered pending handlers for all response_mode flows. */
export const pendingHandlers: PendingHandler[] = [
  peekDeckHandler,
  fankuiHandler,
  tuxiHandler,
  qixiHandler,
  pojunHandler,
  pojunDiscardHandler,
  guoheHandler,
  tannangHandler,
  guicaiHandler,
  guidaoHandler,
  leijiHandler,
  ganglieOfferHandler,
  ganglieChoiceHandler,
  ddzJudgeCancelHandler,
  yijiOfferHandler,
  yijiGiveHandler,
  jianxiongHandler,
  tianxiangHandler,
  liuliHandler,
  fanjianSuitHandler,
  yinghunChoiceHandler,
  yinghunDiscardHandler,
  luanwuHandler,
  guanyuFollowHandler,
  qilinBowHandler,
  wuguPickHandler,
  dyingRescueHandler,
  jijiangHandler,
  wuxiekHandler,
]

export function cardLabelForPending(kind: string | undefined) {
  return cardLabel(kind)
}
