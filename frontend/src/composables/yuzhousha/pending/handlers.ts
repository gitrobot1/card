import {
  finishYuzhoushaPeekDeck,
  playYuzhoushaCard,
  respondYuzhoushaCard,
  useYuzhoushaSkill,
} from '../../../api/games'
import { YZS_CARD_LABELS } from '../../../types/yuzhousha'
import { isBusy, pickFirstTarget, responseAnyMode, responseMode } from './helpers'
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
    state.pending.target_index === state.human_player,
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

const fankuiHandler: PendingHandler = {
  modes: ['skill_fankui'],
  match: (state) => responseMode(state, 'skill_fankui'),
  skillOnly: true,
  onEnter(ctx) {
    pickFirstTarget(ctx, ctx.fankuiTargetOptions.value)
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'fankui' || isBusy(ctx)) return false
    if (ctx.selectedTargetZone.value) return true
    return ctx.fankuiTargetOptions.value.some((o) => o.zone === 'hand')
  },
  async submitSkill(ctx) {
    const zone = ctx.selectedTargetZone.value || 'hand'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'fankui', {
        targetZone: zone,
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【反馈】：选择来源的一张牌，再点「反馈」'
  },
}

const tuxiHandler: PendingHandler = {
  modes: ['skill_tuxi'],
  match: (state) => responseMode(state, 'skill_tuxi'),
  skillOnly: true,
  onEnter(ctx) {
    pickFirstTarget(ctx, ctx.tuxiTargetOptions.value)
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'tuxi' || isBusy(ctx)) return false
    if (ctx.selectedTargetZone.value) return true
    return ctx.tuxiTargetOptions.value.some((o) => o.zone === 'hand')
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'tuxi') return
    const zone = ctx.selectedTargetZone.value || 'hand'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'tuxi', {
        targetZone: zone,
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【突袭】：选择获得对手的一张牌，再点「突袭」'
  },
}

const qixiHandler: PendingHandler = {
  modes: ['skill_qixi'],
  match: (state) => responseMode(state, 'skill_qixi'),
  skillOnly: true,
  onEnter(ctx) {
    pickFirstTarget(ctx, ctx.qixiTargetOptions.value)
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'qixi' || isBusy(ctx)) return false
    return ctx.qixiTargetOptions.value.length > 0
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'qixi') return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'qixi', {
        targetZone: 'hand',
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【奇袭】：选择获得对手的一张手牌'
  },
}

const pojunHandler: PendingHandler = {
  modes: ['skill_pojun'],
  match: (state) => responseMode(state, 'skill_pojun'),
  skillOnly: true,
  onEnter(ctx) {
    pickFirstTarget(ctx, ctx.pojunTargetOptions.value)
  },
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'pojun' || isBusy(ctx)) return false
    if (ctx.selectedTargetZone.value) return true
    return ctx.pojunTargetOptions.value.length > 0
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'pojun') return
    const zone = ctx.selectedTargetZone.value || 'hand'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'pojun', {
        targetZone: zone,
        targetCardId: ctx.selectedTargetCardId.value,
      }),
    )
    ctx.selectedTargetZone.value = ''
    ctx.selectedTargetCardId.value = ''
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
  canSubmitSkill(ctx, skillId) {
    if (skillId !== 'pojun' || isBusy(ctx)) return false
    return !!ctx.selectedTargetCardId.value
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'pojun' || !ctx.selectedTargetCardId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'pojun', {
        cardIds: [ctx.selectedTargetCardId.value],
      }),
    )
    ctx.selectedTargetCardId.value = ''
  },
  hint(ctx) {
    const need = ctx.state.pending?.pojun_remaining ?? 1
    return ctx.centerMessage.value || `【破军】：弃置「营」中 ${need} 张牌`
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
      (action === 'yinghun_draw_both' || action === 'yinghun_draw_two_discard') && !isBusy(ctx)
    )
  },
  async submitAction(ctx, action) {
    const option =
      action === 'yinghun_draw_two_discard' ? 'draw_two_discard' : 'draw_both'
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'yinghun', {
        targetZone: option,
      }),
    )
  },
  hint(ctx) {
    return ctx.centerMessage.value || '【英魂】：选择一项（双方各摸一张 / 令孙坚摸两张并弃一张手牌）'
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
    return skillId === 'yinghun' && !!ctx.selectedId.value && !isBusy(ctx)
  },
  async submitSkill(ctx, skillId) {
    if (skillId !== 'yinghun' || !ctx.selectedId.value) return
    await ctx.act(() =>
      useYuzhoushaSkill(ctx.state.id, 'yinghun', {
        cardIds: [ctx.selectedId.value],
      }),
    )
    ctx.selectedId.value = ''
  },
  hint(ctx) {
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
  allowsCancel: false,
  canPlayCard(ctx) {
    return (
      !!ctx.selectedId.value &&
      (ctx.state.pending?.revealed_cards?.some((c) => c.id === ctx.selectedId.value) ?? false)
    )
  },
  canSubmitPlay(ctx) {
    return (
      !!ctx.selectedId.value &&
      (ctx.state.pending?.revealed_cards?.some((c) => c.id === ctx.selectedId.value) ??
        false) &&
      !isBusy(ctx)
    )
  },
  async submitPlay(ctx) {
    if (!ctx.selectedId.value) return
    await ctx.act(() => playYuzhoushaCard(ctx.state.id, ctx.selectedId.value, ctx.mySeat))
    ctx.selectedId.value = ''
  },
  hint(ctx) {
    return ctx.centerMessage.value || '请选择【五谷丰登】亮出的一张牌'
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
