import { computed } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import type { YuzhoushaState, YzsCard, YzsPlayer, YzsSkillMeta } from '../../types/yuzhousha'

export interface YzsHintsDeps {
  state: Ref<YuzhoushaState | null>
  centerMessage: Ref<string>
  tableActionHint: Ref<string>
  selectedDiscardIds: Ref<string[]>
  selectedId: Ref<string>
  shaTarget: Ref<number | null>
  selectedTargetZone: Ref<string>
  peekDeckSkillId: ComputedRef<string>
  rendeMode: Ref<boolean>
  zhihengMode: Ref<boolean>
  jieyinMode: Ref<boolean>
  fanjianMode: Ref<boolean>
  qixiMode: Ref<boolean>
  wushengMode: Ref<boolean>
  isDealing: Ref<boolean>
  isFinished: ComputedRef<boolean>
  isMyDiscard: ComputedRef<boolean>
  isMyPrepare: ComputedRef<boolean>
  isMyDraw: ComputedRef<boolean>
  isPeekDeck: ComputedRef<boolean>
  isMyPlay: ComputedRef<boolean>
  isMyResponse: ComputedRef<boolean>
  hasTeamMode: ComputedRef<boolean>
  isGuanYuFollow: ComputedRef<boolean>
  isQilinBow: ComputedRef<boolean>
  isJijiangRespond: ComputedRef<boolean>
  isJianxiong: ComputedRef<boolean>
  isYijiOffer: ComputedRef<boolean>
  isYijiGive: ComputedRef<boolean>
  isGanglieOffer: ComputedRef<boolean>
  isGanglieChoice: ComputedRef<boolean>
  isFankui: ComputedRef<boolean>
  isGuicai: ComputedRef<boolean>
  isGuidao: ComputedRef<boolean>
  isLeijiOffer: ComputedRef<boolean>
  isFanjianSuit: ComputedRef<boolean>
  isTianxiangOffer: ComputedRef<boolean>
  isQixiTake: ComputedRef<boolean>
  isYinghunChoice: ComputedRef<boolean>
  isYinghunDiscard: ComputedRef<boolean>
  isWuxiekOffer: ComputedRef<boolean>
  isWuguPick: ComputedRef<boolean>
  discardNeeded: ComputedRef<number>
  activatableSkillIds: ComputedRef<Set<string>>
  myCharacterSkills: ComputedRef<YzsSkillMeta[]>
  selectedCard: ComputedRef<YzsCard | null>
  yijiGiveRemaining: ComputedRef<number>
  responseRequiredKind: ComputedRef<string>
  canSubmitBagua: ComputedRef<boolean>
  enemySeats: ComputedRef<number[]>
  cardPlaysAsSha: (card: YzsCard | null | undefined) => boolean
  needsOpponentTarget: (card: YzsCard | null | undefined) => boolean
  canTargetOpponentWith: (card: YzsCard | null | undefined) => boolean
  canPlayCard: (card: YzsCard | null | undefined) => boolean
  selectedCardNeedsTargetCard: (card?: YzsCard | null) => boolean
  distanceToSeat: (seat: number) => number
  attackRangeOf: (player?: YzsPlayer) => number
  isKongchengProtected: (player?: YzsPlayer) => boolean
  seatAt: (seat: number) => YzsPlayer | undefined
  cardLabel: (kind: string | undefined) => string
  resolvePendingHint?: () => string | null
}

export function useYzsHints(deps: YzsHintsDeps) {
  const centerHint = computed(() => {
    const registryHint = deps.resolvePendingHint?.()
    if (registryHint) return registryHint

    if (deps.isDealing.value) return '发牌中…'
    if (deps.isFinished.value) return deps.state.value?.message ?? ''
    if (deps.isMyDiscard.value) {
      if (deps.tableActionHint.value) return deps.tableActionHint.value
      if (deps.discardNeeded.value <= 0) {
        return deps.centerMessage.value || deps.state.value?.message || '弃牌阶段'
      }
      const picked = deps.selectedDiscardIds.value.length
      if (picked === 0) {
        return `请一次选择 ${deps.discardNeeded.value} 张牌，选满后点「弃牌」`
      }
      if (picked < deps.discardNeeded.value) {
        return `已选 ${picked}/${deps.discardNeeded.value} 张，请继续选择`
      }
      return `已选满 ${deps.discardNeeded.value} 张，点「弃牌」一起丢到牌桌`
    }
    if (deps.isMyPrepare.value) {
      const parts: string[] = []
      if (deps.activatableSkillIds.value.has('guanxing')) parts.push('【观星】')
      if (deps.activatableSkillIds.value.has('luoshen')) parts.push('【洛神】')
      if (deps.activatableSkillIds.value.has('yinghun')) parts.push('【英魂】')
      if (deps.activatableSkillIds.value.has('hunzi')) parts.push('【魂姿·觉醒】')
      if (parts.length) {
        return deps.centerMessage.value || `准备阶段：可发动${parts.join('、')}，或点「跳过」`
      }
      return deps.centerMessage.value || '准备阶段：点「跳过」进入判定/摸牌'
    }
    if (deps.isMyDraw.value) {
      return deps.centerMessage.value || '摸牌阶段：可发动【裸衣】放弃摸牌，或点「摸牌」正常摸 2 张'
    }
    if (deps.isMyPlay.value && deps.rendeMode.value) {
      return '【仁德】：选择要给出的手牌，点击敌方头像确定目标，再点「发动仁德」'
    }
    if (deps.isMyPlay.value && deps.zhihengMode.value) {
      return '【制衡】：选择要弃置的手牌（至少一张），再点「发动制衡」'
    }
    if (deps.isMyPlay.value && deps.jieyinMode.value) {
      return '【结姻】：选择恰好 2 张手牌，点击敌方头像确定目标，再点「发动结姻」'
    }
    if (deps.isMyPlay.value && deps.fanjianMode.value) {
      return '【反间】：选择一张手牌交给对手，再点「发动反间」'
    }
    if (deps.isMyPlay.value && deps.qixiMode.value) {
      if (deps.selectedCard.value && deps.selectedCard.value.suit && (deps.selectedCard.value.suit === 'S' || deps.selectedCard.value.suit === 'C')) {
        if (deps.shaTarget.value == null) {
          return '【奇袭】：黑色牌将当过河拆桥打出，点击敌方头像选定目标，再点「出牌」'
        }
        return '【奇袭】：已锁定目标，点「出牌」将黑色牌当【过河拆桥】打出'
      }
      return '【奇袭】：选择一张黑色牌（手牌或装备区），选目标后点「出牌」即可当过河拆桥打出'
    }
    if (deps.isMyPlay.value && deps.wushengMode.value) {
      if (deps.selectedCard.value && deps.cardPlaysAsSha(deps.selectedCard.value) && deps.selectedCard.value.kind !== 'sha') {
        if (deps.shaTarget.value == null) {
          return '【武圣】：点击敌方头像选定目标，再点「出牌」'
        }
        return '【武圣】：已锁定目标，点「出牌」将红色牌当【杀】打出'
      }
      return '【武圣】已发动：选一张红色牌（♥♦），再指定敌方'
    }
    if (deps.isMyPlay.value && deps.needsOpponentTarget(deps.selectedCard.value) && deps.shaTarget.value == null) {
      if (deps.selectedCard.value?.kind === 'bingliang') {
        const inRange = deps.enemySeats.value.filter((s) => deps.distanceToSeat(s) <= 1)
        if (inRange.length === 0) {
          return '【兵粮寸断】只能对距离 1 及以内的敌方使用'
        }
      }
      if (deps.selectedCard.value?.kind === 'sha' && !deps.canTargetOpponentWith(deps.selectedCard.value)) {
        return `攻击距离不足（攻击范围 ${deps.attackRangeOf()}），点击可攻击的敌方头像`
      }
      if (
        deps.selectedCard.value?.kind === 'juedou' &&
        deps.enemySeats.value.every((s) => deps.isKongchengProtected(deps.seatAt(s)))
      ) {
        return '【空城】：敌方无手牌时不能成为【决斗】的目标'
      }
      if (
        deps.cardPlaysAsSha(deps.selectedCard.value) &&
        deps.enemySeats.value.every((s) => deps.isKongchengProtected(deps.seatAt(s)))
      ) {
        return '【空城】：敌方无手牌时不能成为【杀】的目标'
      }
      if (
        !deps.hasTeamMode.value &&
        deps.isKongchengProtected() &&
        (deps.selectedCard.value?.kind === 'juedou' || deps.cardPlaysAsSha(deps.selectedCard.value))
      ) {
        return '【空城】：对方无手牌时不能成为【杀】或【决斗】的目标'
      }
      return `选中【${deps.selectedCard.value?.name ?? '牌'}】后点击敌方头像锁定目标，再点「出牌」`
    }
    if (deps.isMyPlay.value && deps.needsOpponentTarget(deps.selectedCard.value) && deps.shaTarget.value != null) {
      const target = deps.seatAt(deps.shaTarget.value)
      if (deps.selectedCard.value?.kind === 'bingliang' && deps.distanceToSeat(deps.shaTarget.value) > 1) {
        return `【兵粮寸断】只能对距离 1 及以内的角色使用（距 ${target?.name} 为 ${deps.distanceToSeat(deps.shaTarget.value)}）`
      }
      if (deps.cardPlaysAsSha(deps.selectedCard.value) && deps.distanceToSeat(deps.shaTarget.value) > deps.attackRangeOf()) {
        return `攻击距离不足：距 ${target?.name} 为 ${deps.distanceToSeat(deps.shaTarget.value)}，攻击范围 ${deps.attackRangeOf()}`
      }
    }
    if (deps.isMyPlay.value && deps.selectedCardNeedsTargetCard() && deps.shaTarget.value != null && deps.selectedTargetZone.value === '') {
      return `请选择要${deps.selectedCard.value?.kind === 'tannang' ? '拿取' : '拆掉'}的对手牌`
    }
    if (deps.isMyPlay.value && deps.needsOpponentTarget(deps.selectedCard.value) && deps.shaTarget.value != null) {
      return `已锁定 ${deps.seatAt(deps.shaTarget.value)?.name ?? '敌方'}，点「出牌」使用【${deps.selectedCard.value?.name ?? '牌'}】`
    }
    if (deps.isMyResponse.value) {
      if (deps.state.value?.pending?.allow_wuxiek) {
        return (
          deps.centerMessage.value ||
          `出【${deps.cardLabel(deps.responseRequiredKind.value)}】或【无懈可击】仅抵消对自己的效果，或点「取消」`
        )
      }
      if (deps.canSubmitBagua.value) {
        return (
          deps.centerMessage.value ||
          '可点「八卦判定」、出【闪】，或点「取消」承受伤害'
        )
      }
      return deps.centerMessage.value || `选中【${deps.cardLabel(deps.responseRequiredKind.value)}】后点「出牌」，或点「取消」`
    }
    if (deps.isMyPlay.value && !deps.selectedId.value) {
      return '选中手牌后点「出牌」，或点「结束出牌」'
    }
    if (deps.isMyPlay.value && deps.selectedCard.value && !deps.canPlayCard(deps.selectedCard.value)) {
      return '当前选中的牌无法打出，请换牌或点「结束出牌」'
    }
    return deps.centerMessage.value || deps.state.value?.message || '\u00a0'
  })

  return { centerHint }
}
