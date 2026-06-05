import {
  animateYzsBaguaJudge,
  animateYzsPlayEvent,
} from '../useYzsPlayAnimation'
import {
  equipSlotOf,
  judgeAreaCards,
  removeKnownCardFromPlayer,
  trickStaysInJudge,
} from '../playerCardHelpers'
import type { YzsPlayer } from '../../../types/yuzhousha'
import type { EventReplayHandler } from './types'

function typeIs(type: string, ...types: string[]) {
  return types.includes(type)
}

const drawNoopHandler: EventReplayHandler = {
  types: ['draw'],
  match: (e) => e.type === 'draw',
  async replay() {
    /* batched in applyState */
  },
}

const discardPhaseHandler: EventReplayHandler = {
  types: ['discard_phase'],
  match: (e) => e.type === 'discard_phase',
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    const need = event.amount ?? 0
    const hint =
      event.message ??
      (need > 0 ? `请一次选择 ${need} 张牌，选满后一起弃牌` : '进入弃牌阶段')
    tableActionHint.value = hint
    centerMessage.value = hint
    if (state.value) {
      state.value = {
        ...state.value,
        turn_step: 'discard',
        message: hint,
      }
    }
    await sleep(450)
    tableActionHint.value = ''
  },
}

const playShaHandler: EventReplayHandler = {
  types: ['play_sha'],
  match: (e) => e.type === 'play_sha' && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, playAreaRef, displayedHand, setTableCard, nextTick: tick, runShaFlyBolt } = ctx
    const source = event.player_index
    const target = event.target_index
    if (source != null && target != null) {
      await tick()
      await runShaFlyBolt(source, target)
    }
    await animateYzsPlayEvent(event, playAreaRef.value, mySeat, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value) {
        state.value = {
          ...state.value,
          discard_count: state.value.discard_count + 1,
          players: state.value.players.map((p) =>
            p.index === event.player_index
              ? { ...p, hand_count: Math.max(0, p.hand_count - 1), sha_used_this_turn: true, drunk: false }
              : p,
          ),
        }
      }
    })
  },
}

const playTrickHandler: EventReplayHandler = {
  types: ['play_trick'],
  match: (e) => e.type === 'play_trick' && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, playAreaRef, displayedHand, setTableCard } = ctx
    await animateYzsPlayEvent(event, playAreaRef.value, mySeat, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value && event.player_index != null) {
        state.value = {
          ...state.value,
          discard_count:
            trickStaysInJudge(event.card?.kind ?? '') ? state.value.discard_count : state.value.discard_count + 1,
          players: state.value.players.map((p) => {
            if (p.index === event.player_index) {
              return { ...p, hand_count: Math.max(0, p.hand_count - 1) }
            }
            if (event.target_index == null || !event.card || !trickStaysInJudge(event.card.kind)) {
              return p
            }
            const patch: Partial<YzsPlayer> = {
              judge_area: [...judgeAreaCards(p), event.card],
            }
            if (event.card.kind === 'lebu') patch.skip_play = true
            if (event.card.kind === 'bingliang') patch.skip_draw = true
            return { ...p, ...patch }
          }),
        }
      }
    })
  },
}

const playJiuEquipHandler: EventReplayHandler = {
  types: ['play_jiu', 'equip'],
  match: (e) => (e.type === 'play_jiu' || e.type === 'equip') && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, playAreaRef, displayedHand, setTableCard } = ctx
    await animateYzsPlayEvent(event, playAreaRef.value, mySeat, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value && event.player_index != null) {
        state.value = {
          ...state.value,
          discard_count: event.type === 'play_jiu' ? state.value.discard_count + 1 : state.value.discard_count,
          players: state.value.players.map((p) => {
            if (p.index !== event.player_index) return p
            const next = { ...p, hand_count: Math.max(0, p.hand_count - 1) }
            if (event.type === 'play_jiu') {
              next.drunk = true
              return next
            }
            const slot = equipSlotOf(event.card!)
            if (slot === 'weapon') next.weapon = event.card
            if (slot === 'armor') next.armor = event.card
            if (slot === 'plus_horse') next.plus_horse = event.card
            if (slot === 'minus_horse') next.minus_horse = event.card
            return next
          }),
        }
      }
    })
  },
}

const respondCardHandler: EventReplayHandler = {
  types: ['play_tao', 'skill_jiji', 'respond_shan', 'respond_sha'],
  match: (e) =>
    typeIs(e.type, 'play_tao', 'skill_jiji', 'respond_shan', 'respond_sha') && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, playAreaRef, displayedHand, setTableCard } = ctx
    await animateYzsPlayEvent(event, playAreaRef.value, mySeat, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value && event.player_index != null) {
        const healSeat =
          (event.type === 'play_tao' || event.type === 'skill_jiji') && event.heal
            ? (event.target_index ?? event.player_index)
            : null
        state.value = {
          ...state.value,
          discard_count: state.value.discard_count + 1,
          players: state.value.players.map((p) => {
            let next = p
            if (p.index === event.player_index) {
              next = { ...p, hand_count: Math.max(0, p.hand_count - 1) }
            }
            if (healSeat != null && p.index === healSeat) {
              next = { ...next, hp: Math.min(p.max_hp, p.hp + (event.heal ?? 0)) }
            }
            return next
          }),
        }
      }
    })
  },
}

const trickEffectHandler: EventReplayHandler = {
  types: ['trick_effect'],
  match: (e) => e.type === 'trick_effect',
  async replay(ctx) {
    const { event, state, mySeat, centerMessage, tableActionHint, displayedHand, appendDrawnCards, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    if (state.value && event.player_index != null && event.target_index != null) {
      if (event.amount === 1 && event.player_index === mySeat && event.card) {
        appendDrawnCards([event.card])
      }
      if (event.target_index === mySeat && event.card) {
        displayedHand.value = displayedHand.value.filter((card) => card.id !== event.card!.id)
      }
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (event.amount === 1) {
            if (p.index === event.player_index) return { ...p, hand_count: p.hand_count + 1 }
            if (p.index === event.target_index) return removeKnownCardFromPlayer(p, event.card)
          } else if (event.card && p.index === event.target_index) {
            return removeKnownCardFromPlayer(p, event.card)
          }
          return p
        }),
      }
    }
    await sleep(360)
    tableActionHint.value = ''
  },
}

const trickHealHandler: EventReplayHandler = {
  types: ['trick_heal'],
  match: (e) => e.type === 'trick_heal',
  async replay(ctx) {
    const { event, state, sleep } = ctx
    if (event.target_index != null && event.heal && state.value) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) =>
          p.index === event.target_index ? { ...p, hp: Math.min(p.max_hp, p.hp + event.heal!) } : p,
        ),
      }
    }
    await sleep(280)
  },
}

const lebuSkipHandler: EventReplayHandler = {
  types: ['lebu_skip'],
  match: (e) => e.type === 'lebu_skip' && e.player_index != null,
  async replay(ctx) {
    const { event, state, sleep } = ctx
    if (!state.value) return
    state.value = {
      ...state.value,
      players: state.value.players.map((p) => {
        if (p.index !== event.player_index) return p
        const area = p.judge_area?.filter((j) => j.kind !== 'lebu') ?? []
        return { ...p, skip_play: false, judge_area: area.length ? area : undefined }
      }),
    }
    await sleep(360)
  },
}

const bingliangSkipHandler: EventReplayHandler = {
  types: ['bingliang_skip'],
  match: (e) => e.type === 'bingliang_skip' && e.player_index != null,
  async replay(ctx) {
    const { event, state, sleep } = ctx
    if (!state.value) return
    state.value = {
      ...state.value,
      players: state.value.players.map((p) => {
        if (p.index !== event.player_index) return p
        const area = p.judge_area?.filter((j) => j.kind !== 'bingliang') ?? []
        return { ...p, judge_area: area.length ? area : undefined, skip_draw: false }
      }),
    }
    await sleep(360)
  },
}

const wuguPickHandler: EventReplayHandler = {
  types: ['wugu_pick'],
  match: (e) => e.type === 'wugu_pick' && e.player_index != null && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, appendDrawnCards, sleep } = ctx
    if (!state.value) return
    if (event.player_index === mySeat) {
      appendDrawnCards([event.card!])
    }
    state.value = {
      ...state.value,
      players: state.value.players.map((p) =>
        p.index === event.player_index ? { ...p, hand_count: p.hand_count + 1 } : p,
      ),
    }
    await sleep(320)
  },
}

const shandianJudgeHandler: EventReplayHandler = {
  types: ['shandian_judge'],
  match: (e) => e.type === 'shandian_judge' && e.player_index != null,
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    if (event.amount === 1 && state.value) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (p.index !== event.player_index) return p
          const area = p.judge_area?.filter((j) => j.kind !== 'shandian') ?? []
          const hp = Math.max(0, p.hp - 3)
          return { ...p, judge_area: area.length ? area : undefined, hp }
        }),
      }
    }
    await sleep(420)
    tableActionHint.value = ''
  },
}

const trickResponseHandler: EventReplayHandler = {
  types: ['trick_response', 'wuxiek_offer'],
  match: (e) => typeIs(e.type, 'trick_response', 'wuxiek_offer'),
  async replay(ctx) {
    await ctx.sleep(360)
  },
}

const wuxiekCancelHandler: EventReplayHandler = {
  types: ['play_wuxiek', 'trick_cancelled'],
  match: (e) => typeIs(e.type, 'play_wuxiek', 'trick_cancelled'),
  async replay(ctx) {
    const { event, mySeat, displayedHand, sleep } = ctx
    if (event.card && event.player_index === mySeat) {
      displayedHand.value = displayedHand.value.filter((card) => card.id !== event.card!.id)
    }
    await sleep(320)
  },
}

const phaseMessageHandler: EventReplayHandler = {
  types: ['prepare_phase', 'draw_phase'],
  match: (e) => typeIs(e.type, 'prepare_phase', 'draw_phase'),
  async replay(ctx) {
    if (ctx.event.message) ctx.centerMessage.value = ctx.event.message
  },
}

const peekDeckHandler: EventReplayHandler = {
  types: ['peek_deck_reveal', 'peek_deck_show', 'peek_deck_finish'],
  match: (e) => typeIs(e.type, 'peek_deck_reveal', 'peek_deck_show', 'peek_deck_finish'),
  async replay(ctx) {
    if (ctx.event.message && ctx.event.type === 'peek_deck_reveal') {
      ctx.centerMessage.value = ctx.event.message
    }
  },
}

const guanxingTriggerHandler: EventReplayHandler = {
  types: ['skill_trigger'],
  match: (e) => e.type === 'skill_trigger' && e.skill_id === 'guanxing',
  async replay(ctx) {
    if (ctx.event.message) ctx.centerMessage.value = ctx.event.message
  },
}

const skillAwakenHandler: EventReplayHandler = {
  types: ['skill_awaken'],
  match: (e) => e.type === 'skill_awaken' && !!e.message,
  async replay(ctx) {
    ctx.centerMessage.value = ctx.event.message!
    ctx.tableActionHint.value = ctx.event.message!
    await ctx.sleep(520)
    ctx.tableActionHint.value = ''
  },
}

const skillJiangHandler: EventReplayHandler = {
  types: ['skill_jiang'],
  match: (e) => e.type === 'skill_jiang' && !!e.message,
  async replay(ctx) {
    ctx.centerMessage.value = ctx.event.message!
    await ctx.sleep(320)
  },
}

const judgeFlipHandler: EventReplayHandler = {
  types: ['judge_flip'],
  match: (e) => e.type === 'judge_flip',
  async replay(ctx) {
    if (!ctx.event.message) return
    ctx.centerMessage.value = ctx.event.message
    ctx.tableActionHint.value = ctx.event.message
    await ctx.sleep(420)
    ctx.tableActionHint.value = ''
  },
}

const luoshenGainHandler: EventReplayHandler = {
  types: ['luoshen_gain'],
  match: (e) => e.type === 'luoshen_gain' && !!e.card,
  async replay(ctx) {
    if (ctx.event.player_index !== ctx.mySeat) return
    ctx.appendDrawnCards([ctx.event.card!])
    if (ctx.event.message) {
      ctx.tableActionHint.value = ctx.event.message
      ctx.centerMessage.value = ctx.event.message
    }
    await ctx.sleep(420)
    ctx.tableActionHint.value = ''
  },
}

const luoshenStopHandler: EventReplayHandler = {
  types: ['luoshen_stop'],
  match: (e) => e.type === 'luoshen_stop' && !!e.message,
  async replay(ctx) {
    ctx.tableActionHint.value = ctx.event.message!
    ctx.centerMessage.value = ctx.event.message!
    await ctx.sleep(420)
    ctx.tableActionHint.value = ''
  },
}

const skillTriggerHealHandler: EventReplayHandler = {
  types: ['skill_trigger', 'skill_heal'],
  match: (e) =>
    e.type === 'skill_heal' || (e.type === 'skill_trigger' && e.skill_id !== 'guanxing'),
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    if (event.type === 'skill_heal' && event.player_index != null && event.heal && state.value) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) =>
          p.index === event.player_index ? { ...p, hp: Math.min(p.max_hp, p.hp + event.heal!) } : p,
        ),
      }
    }
    await sleep(380)
    tableActionHint.value = ''
  },
}

function gainCardSkillHandler(type: string): EventReplayHandler {
  return {
    types: [type],
    match: (e) => e.type === type && !!e.card,
    async replay(ctx) {
      if (ctx.event.player_index !== ctx.mySeat) return
      ctx.appendDrawnCards([ctx.event.card!])
      if (ctx.event.message) {
        ctx.tableActionHint.value = ctx.event.message
        ctx.centerMessage.value = ctx.event.message
      }
      await ctx.sleep(360)
      ctx.tableActionHint.value = ''
    },
  }
}

const guicaiReplaceHandler: EventReplayHandler = {
  types: ['guicai_replace'],
  match: (e) => e.type === 'guicai_replace' && !!e.message,
  async replay(ctx) {
    ctx.tableActionHint.value = ctx.event.message!
    ctx.centerMessage.value = ctx.event.message!
    if (ctx.event.card && ctx.event.player_index === ctx.mySeat) {
      ctx.displayedHand.value = ctx.displayedHand.value.filter((c) => c.id !== ctx.event.card!.id)
    }
    await ctx.sleep(360)
    ctx.tableActionHint.value = ''
  },
}

const skillGiveCardHandler: EventReplayHandler = {
  types: ['skill_give_card'],
  match: (e) => e.type === 'skill_give_card' && e.target_index != null && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, appendDrawnCards, displayedHand, sleep } = ctx
    if (!state.value) return
    if (event.target_index === mySeat) {
      appendDrawnCards([event.card!])
    } else if (event.player_index === mySeat) {
      displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
    }
    await sleep(260)
  },
}

const skillJijiangShaHandler: EventReplayHandler = {
  types: ['skill_jijiang_sha'],
  match: (e) => e.type === 'skill_jijiang_sha' && !!e.card,
  async replay(ctx) {
    ctx.tableActionHint.value = ctx.event.message ?? ''
    await ctx.sleep(320)
    ctx.tableActionHint.value = ''
  },
}

const weaponSkillHandler: EventReplayHandler = {
  types: ['weapon_skill'],
  match: (e) => e.type === 'weapon_skill',
  async replay(ctx) {
    ctx.tableActionHint.value = ctx.event.message ?? ''
    ctx.centerMessage.value = ctx.event.message ?? ctx.centerMessage.value
    await ctx.sleep(420)
    ctx.tableActionHint.value = ''
  },
}

const qilinDiscardHandler: EventReplayHandler = {
  types: ['qilin_discard'],
  match: (e) => e.type === 'qilin_discard' && e.target_index != null && !!e.card,
  async replay(ctx) {
    const { event, state, sleep } = ctx
    if (!state.value) return
    state.value = {
      ...state.value,
      players: state.value.players.map((p) => {
        if (p.index !== event.target_index) return p
        const next = { ...p }
        if (next.plus_horse?.id === event.card!.id) next.plus_horse = undefined
        if (next.minus_horse?.id === event.card!.id) next.minus_horse = undefined
        return next
      }),
      discard_count: state.value.discard_count + 1,
    }
    await sleep(360)
  },
}

const baguaJudgeHandler: EventReplayHandler = {
  types: ['bagua_judge'],
  match: (e) => e.type === 'bagua_judge' && !!e.card,
  async replay(ctx) {
    const { event, state, drawAreaRef, playAreaRef, tableActionHint } = ctx
    tableActionHint.value = event.message ?? ''
    await animateYzsBaguaJudge(
      drawAreaRef.value,
      playAreaRef.value,
      event.card!,
      (event.amount ?? 0) === 1,
    )
    if (state.value) {
      state.value = {
        ...state.value,
        draw_count: Math.max(0, state.value.draw_count - 1),
        discard_count: state.value.discard_count + 1,
      }
    }
    tableActionHint.value = ''
  },
}

const hitHandler: EventReplayHandler = {
  types: ['sha_hit', 'trick_hit'],
  match: (e) => typeIs(e.type, 'sha_hit', 'trick_hit'),
  async replay(ctx) {
    const { event, sleep, flashSeatHit, flashSeatBlocked } = ctx
    if (event.target_index != null) {
      const dmg = event.damage ?? 0
      if (dmg > 0) {
        await flashSeatHit(event.target_index)
      } else {
        await flashSeatBlocked(event.target_index)
      }
    }
    await sleep(280)
  },
}

const turnEndHandler: EventReplayHandler = {
  types: ['turn_end'],
  match: (e) => e.type === 'turn_end',
  async replay(ctx) {
    await ctx.sleep(280)
  },
}

const identityRevealedHandler: EventReplayHandler = {
  types: ['identity_revealed'],
  match: (e) => e.type === 'identity_revealed',
  async replay(ctx) {
    const { event, centerMessage, sleep } = ctx
    if (event.message) centerMessage.value = event.message
    await sleep(700)
  },
}

const gameOverHandler: EventReplayHandler = {
  types: ['game_over'],
  match: (e) => e.type === 'game_over',
  async replay(ctx) {
    await ctx.sleep(600)
  },
}

/** Registered event replay handlers; order matters for overlapping match rules. */
export const eventReplayerHandlers: EventReplayHandler[] = [
  drawNoopHandler,
  discardPhaseHandler,
  playShaHandler,
  playTrickHandler,
  playJiuEquipHandler,
  respondCardHandler,
  trickEffectHandler,
  trickHealHandler,
  lebuSkipHandler,
  bingliangSkipHandler,
  wuguPickHandler,
  shandianJudgeHandler,
  trickResponseHandler,
  wuxiekCancelHandler,
  phaseMessageHandler,
  peekDeckHandler,
  guanxingTriggerHandler,
  skillAwakenHandler,
  skillJiangHandler,
  judgeFlipHandler,
  luoshenGainHandler,
  luoshenStopHandler,
  skillTriggerHealHandler,
  gainCardSkillHandler('jianxiong_gain'),
  gainCardSkillHandler('fankui_take'),
  gainCardSkillHandler('tuxi_take'),
  guicaiReplaceHandler,
  skillGiveCardHandler,
  skillJijiangShaHandler,
  weaponSkillHandler,
  qilinDiscardHandler,
  baguaJudgeHandler,
  hitHandler,
  turnEndHandler,
  identityRevealedHandler,
  gameOverHandler,
]
