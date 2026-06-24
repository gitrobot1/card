import {
  animateYzsBaguaJudge,
  animateYzsPlayEvent,
  animateYzsRevealCard,
  animateYzsTakeCard,
} from '../useYzsPlayAnimation'
import {
  equipSlotOf,
  judgeAreaCards,
  removeKnownCardFromPlayer,
  trickStaysInJudge,
} from '../playerCardHelpers'
import type { YzsPlayer } from '../../../types/yuzhousha'
import type { EventReplayHandler } from './types'
import type { EventReplayContext } from './context'

function typeIs(type: string, ...types: string[]) {
  return types.includes(type)
}

/** 辅助：根据事件类型决定飞线逻辑 */
async function flyBoltIfTargeted(ctx: EventReplayContext) {
  const { event, runShaFlyBolt, nextTick: tick } = ctx
  const source = event.player_index
  const target = event.target_index
  if (source == null) return

  await tick()

  // 群体锦囊（万箭齐发、南蛮入侵、桃园结义、五谷丰登、方天画戟）：一次性并发飞线到所有目标
  if (event.type === 'play_trick') {
    const aoeKinds = ['wanjian', 'nanman', 'taoyuan', 'wugu']
    if (event.card?.kind && aoeKinds.includes(event.card.kind)) {
      const players = ctx.state.value?.players ?? []
      await Promise.all(
        players
          .filter((p) => p.hp > 0 && p.index !== source)
          .map((p) => runShaFlyBolt(source, p.index)),
      )
      return
    }
  }
  // 方天画戟：一次性飞线到所有目标（类似南蛮入侵）
  if (event.type === 'play_sha' && (event as any).fangtian_targets) {
    const targets = (event as any).fangtian_targets as number[]
    if (targets.length > 0) {
      await Promise.all(targets.map((t) => runShaFlyBolt(source, t)))
      return
    }
  }

  // 对自己用的锦囊（无中生有）：无飞线
  if (event.type === 'play_trick' && event.card?.kind === 'wuzhong') {
    return
  }

  // 借刀杀人：双飞线（使用者 → 被借刀者 → 出杀目标）
  if (event.type === 'play_trick' && event.card?.kind === 'jiedao') {
    if (target != null && target !== source) {
      await runShaFlyBolt(source, target)
      await tick()
      const secondTarget = (event as any).second_target_index
      if (secondTarget != null && secondTarget !== target && secondTarget >= 0) {
        await runShaFlyBolt(target, secondTarget)
      }
    }
    return
  }

  // 无懈可击：飞线从打出者指向被抵消的锦囊来源
  if (event.type === 'play_wuxiek' && target != null) {
    await runShaFlyBolt(source, target)
    return
  }

  // 其他有目标的情况：杀、过河拆桥、顺手牵羊、决斗等
  // 跳过自指目标（酒自救、桃自救等），飞线无意义
  if (target != null && target !== source) {
    await runShaFlyBolt(source, target)
  }
}

/** 通用牌动画：飞入牌桌 + 显示在中间 */
async function playCardToTable(ctx: EventReplayContext, updateState: () => void) {
  const { event, mySeat, playAreaRef } = ctx
  await animateYzsPlayEvent(event, playAreaRef.value, mySeat, updateState)
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
    const { event, state, mySeat, displayedHand, setTableCard } = ctx
    await flyBoltIfTargeted(ctx)
    await playCardToTable(ctx, () => {
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
    const { event, state, mySeat, displayedHand, setTableCard } = ctx
    await flyBoltIfTargeted(ctx)
    await playCardToTable(ctx, () => {
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
            // 延时锦囊的判定区状态由 applyState 最终覆盖，不在这里手动修改
            return p
          }),
        }
      }
    })
  },
}

const playJiuEquipHandler: EventReplayHandler = {
  types: ['play_jiu', 'equip'],
  match: (e) => (e.type === 'play_jiu' || e.type === 'equip') && !!e.card && !e.heal,
  async replay(ctx) {
    const { event, state, mySeat, displayedHand, setTableCard } = ctx
    await flyBoltIfTargeted(ctx)
    await playCardToTable(ctx, () => {
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
  types: ['play_tao', 'play_jiu', 'skill_jiji', 'respond_shan', 'respond_sha'],
  match: (e) =>
    typeIs(e.type, 'play_tao', 'play_jiu', 'skill_jiji', 'respond_shan', 'respond_sha') && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, displayedHand, setTableCard } = ctx
    await flyBoltIfTargeted(ctx)
    await playCardToTable(ctx, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value && event.player_index != null) {
        const healSeat =
          (event.type === 'play_tao' || event.type === 'play_jiu' || event.type === 'skill_jiji') && event.heal
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
    const { event, state, mySeat, centerMessage, tableActionHint, displayedHand, appendDrawnCards, setTableCard, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }

    const isDelayTrick = event.card ? trickStaysInJudge(event.card.kind) : false
    const isTake = event.amount === 1 // 顺手牵羊（拿牌）
    const isDiscard = !isTake && !isDelayTrick && !!event.card && event.target_index != null && event.player_index != null // 过河拆桥（拆牌）

    // 过河拆桥：牌从被拆者飞到牌桌中央（复用正常打牌的动画）
    if (isDiscard) {
      // 先显示飞线：从拆牌者飞到被拆牌者
      await flyBoltIfTargeted(ctx)
      // 临时修改 event.player_index 为目标玩家，让动画从被拆者位置飞出
      const originalPlayerIndex = event.player_index
      event.player_index = event.target_index!
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
      })
      // 恢复原始的 player_index
      event.player_index = originalPlayerIndex
    }

    // 顺手牵羊：牌从被拿者飞到拿牌者手上
    if (isTake && event.player_index != null) {
      await animateYzsTakeCard(event.target_index!, event.player_index, event.card!, () => {
        if (event.player_index === mySeat) {
          appendDrawnCards([event.card!])
        }
      })
    }

    if (state.value && event.player_index != null && event.target_index != null && !isDelayTrick) {
      if (isTake) {
        if (event.player_index === mySeat && event.card) {
          appendDrawnCards([event.card])
        }
      }
      if (event.target_index === mySeat && event.card) {
        displayedHand.value = displayedHand.value.filter((card) => card.id !== event.card!.id)
      }
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (isTake) {
            if (p.index === event.player_index) return { ...p, hand_count: p.hand_count + 1 }
            if (p.index === event.target_index) return removeKnownCardFromPlayer(p, event.card)
          } else if (isDiscard && p.index === event.target_index) {
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

function createFlyCardEl(card: { suit?: string; label?: string; name?: string; kind?: string }): HTMLElement {
  const el = document.createElement('div')
  const suitSymbol = card.suit === 'H' ? '♥' : card.suit === 'D' ? '♦' : card.suit === 'S' ? '♠' : card.suit === 'C' ? '♣' : ''
  const suitColor = card.suit === 'H' || card.suit === 'D' ? '#dc2626' : '#1e293b'
  el.className = 'yzs-fly-card'
  el.innerHTML = `
    <span style="position:absolute;top:4px;left:5px;font-size:14px;color:${suitColor};font-weight:700">${suitSymbol}</span>
    <span style="position:absolute;top:4px;right:5px;font-size:11px;font-weight:700;color:${suitColor}">${card.label ?? card.name ?? ''}</span>
    <span style="position:absolute;bottom:6px;left:50%;translate:-50% 0;font-size:11px;font-weight:700;color:#1e293b">${card.name ?? card.label ?? ''}</span>
  `
  el.style.cssText = `
    background: linear-gradient(180deg, #fefce8 0%, #f5f0d0 100%);
    border: 2px solid #d4c8a0;
    border-radius: 8px;
    box-shadow: 0 4px 16px rgba(0,0,0,0.25);
    pointer-events: none;
    position: fixed;
    z-index: 9999;
  `
  return el
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

/** 摸牌阶段被跳过（兵粮寸断生效后的提示） */
const drawPhaseSkipHandler: EventReplayHandler = {
  types: ['draw_phase_skip'],
  match: (e) => e.type === 'draw_phase_skip',
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      centerMessage.value = event.message
      tableActionHint.value = event.message
    }
    if (state.value && event.player_index != null) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) =>
          p.index === event.player_index ? { ...p, skip_draw: false } : p,
        ),
      }
    }
    await sleep(600)
    tableActionHint.value = ''
  },
}

/** 出牌阶段被跳过（乐不思蜀生效） */
const playPhaseSkipHandler: EventReplayHandler = {
  types: ['play_phase_skip'],
  match: (e) => e.type === 'play_phase_skip',
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      centerMessage.value = event.message
      tableActionHint.value = event.message
    }
    if (state.value && event.player_index != null) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) =>
          p.index === event.player_index ? { ...p, skip_play: false } : p,
        ),
      }
    }
    await sleep(800)
    tableActionHint.value = ''
  },
}

const wuguPickHandler: EventReplayHandler = {
  types: ['wugu_pick'],
  match: (e) => e.type === 'wugu_pick' && e.player_index != null && !!e.card,
  async replay(ctx) {
    const { event, state, sleep } = ctx
    if (!state.value) return
    // 五谷选牌不飞动画，只在框内变灰。仅更新手牌数。
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

const huogongRevealHandler: EventReplayHandler = {
  types: ['huogong_reveal'],
  match: (e) => e.type === 'huogong_reveal' && !!e.card,
  async replay(ctx) {
    const { event, centerMessage, tableActionHint, playAreaRef, setTableCard, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    // 目标的手牌飞到牌桌中央展示（不从手牌移除）
    if (event.player_index != null && event.card) {
      await animateYzsRevealCard(
        event.player_index,
        event.card,
        playAreaRef.value,
        () => setTableCard(event.card!),
      )
      // 确保牌一定被设置（动画跳过时 fallback）
      setTableCard(event.card!)
      await sleep(600)
    }
    tableActionHint.value = ''
  },
}

const trickResponseHandler: EventReplayHandler = {
  types: ['trick_response', 'wuxiek_offer'],
  match: (e) => typeIs(e.type, 'trick_response', 'wuxiek_offer'),
  async replay(ctx) {
    if (ctx.event.message) {
      ctx.tableActionHint.value = ctx.event.message
    }
    await ctx.sleep(800)
    ctx.tableActionHint.value = ''
  },
}

const tiesuoChainHandler: EventReplayHandler = {
  types: ['tiesuo_chain'],
  match: (e) => e.type === 'tiesuo_chain' && e.target_index != null,
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    // 更新连环状态：根据消息中的"横置"/"重置"设置，而不是 toggle
    if (state.value && event.target_index != null) {
      const seat = event.target_index
      const isChaining = event.message?.includes('横置') ?? false
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (p.index !== seat) return p
          const counters = { ...(p.skill_counters ?? {}) }
          if (isChaining) {
            counters.chained = 1
          } else {
            delete counters.chained
          }
          return { ...p, skill_counters: counters }
        }),
      }
    }
    await sleep(500)
    tableActionHint.value = ''
  },
}

const tiesuoSpreadHandler: EventReplayHandler = {
  types: ['tiesuo_spread'],
  match: (e) => e.type === 'tiesuo_spread',
  async replay(ctx) {
    const { event, state, centerMessage, tableActionHint, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    // 传导完毕：重置所有连环角色
    if (state.value) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (!p.skill_counters?.chained) return p
          const counters = { ...p.skill_counters }
          delete counters.chained
          return { ...p, skill_counters: counters }
        }),
      }
    }
    await sleep(600)
    tableActionHint.value = ''
  },
}

const wuxiekCancelHandler: EventReplayHandler = {
  types: ['play_wuxiek', 'trick_cancelled', 'wuxiek_recursive'],
  match: (e) => typeIs(e.type, 'play_wuxiek', 'trick_cancelled', 'wuxiek_recursive'),
  async replay(ctx) {
    const { event, mySeat, displayedHand, setTableCard, sleep } = ctx
    // 无懈可击打出时：牌飞行动画 + 飞线
    if (event.card && event.type === 'play_wuxiek') {
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
        if (event.player_index === mySeat) {
          displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
        }
      })
    }
    await sleep(320)
  },
}

const phaseMessageHandler: EventReplayHandler = {
  types: ['prepare_phase', 'draw_phase', 'judge_phase', 'play_phase', 'discard_phase'],
  match: (e) => typeIs(e.type, 'prepare_phase', 'draw_phase', 'judge_phase', 'play_phase', 'discard_phase'),
  async replay(ctx) {
    if (ctx.event.message) {
      ctx.centerMessage.value = ctx.event.message
      ctx.tableActionHint.value = ctx.event.message
    }
    await ctx.sleep(800)
    ctx.tableActionHint.value = ''
  },
}

const turnEndHandler: EventReplayHandler = {
  types: ['turn_end'],
  match: (e) => e.type === 'turn_end',
  async replay(ctx) {
    if (ctx.event.message) {
      ctx.centerMessage.value = ctx.event.message
      ctx.tableActionHint.value = ctx.event.message
    }
    await ctx.sleep(800)
    ctx.tableActionHint.value = ''
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
    const { event, setTableCard } = ctx
    // 判定牌翻到牌桌上展示
    if (event.card) {
      setTableCard(event.card)
    }
    if (event.message) {
      ctx.centerMessage.value = event.message
      ctx.tableActionHint.value = event.message
    }
    await ctx.sleep(2500)
    ctx.tableActionHint.value = ''
  },
}

/** 判定结果：展示 ✅ 或 ❌，并移除判定区对应的延时锦囊 */
const judgeResultHandler: EventReplayHandler = {
  types: ['judge_result'],
  match: (e) => e.type === 'judge_result',
  async replay(ctx) {
    const { event, setTableCard, state, sleep } = ctx
    // 再次确保判定牌在桌上
    if (event.card) {
      setTableCard(event.card)
    }
    // 显示结果
    const resultIcon = event.success ? '✅' : '❌'
    if (event.message) {
      ctx.centerMessage.value = `${resultIcon} ${event.message}`
      ctx.tableActionHint.value = `${resultIcon} ${event.message}`
    }
    // 从判定区移除第一张延时锦囊（乐/兵/电）
    if (state.value && event.player_index != null) {
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (p.index !== event.player_index || !p.judge_area?.length) return p
          const area = p.judge_area.slice(1)
          return { ...p, judge_area: area.length ? area : undefined }
        }),
      }
    }
    await sleep(3000)
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
    const { event, state, centerMessage, tableActionHint, mySeat, displayedHand, setTableCard, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    // 技能触发带牌（如丈八蛇矛出杀）：牌飞到牌桌中央
    if (event.card && event.type === 'skill_trigger') {
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
        if (event.player_index === mySeat) {
          displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
        }
        if (state.value && event.player_index != null) {
          state.value = {
            ...state.value,
            discard_count: state.value.discard_count + 1,
            players: state.value.players.map((p) =>
              p.index === event.player_index
                ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
                : p,
            ),
          }
        }
      })
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

/** 牌从被拿者飞向拿牌者的通用 handler（反馈、突袭、冲阵等） */
const takeCardHandler: EventReplayHandler = {
  types: ['fankui_take', 'tuxi_take', 'chongzhen_take'],
  match: (e) =>
    typeIs(e.type, 'fankui_take', 'tuxi_take', 'chongzhen_take') && !!e.card,
  async replay(ctx) {
    const { event, state, mySeat, centerMessage, tableActionHint, displayedHand, appendDrawnCards, sleep } = ctx
    if (event.message) {
      tableActionHint.value = event.message
      centerMessage.value = event.message
    }
    const taker = event.player_index // 拿牌者
    const victim = event.target_index // 被拿者
    // 牌从被拿者座位飞向拿牌者座位
    if (event.card && taker != null && victim != null) {
      await animateYzsTakeCard(victim, taker, event.card, () => {
        if (taker === mySeat) {
          appendDrawnCards([event.card])
        }
      })
    }
    // 更新手牌状态
    if (state.value && taker != null && victim != null) {
      if (victim === mySeat && event.card) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      state.value = {
        ...state.value,
        players: state.value.players.map((p) => {
          if (p.index === taker) return { ...p, hand_count: p.hand_count + 1 }
          if (p.index === victim) return removeKnownCardFromPlayer(p, event.card)
          return p
        }),
      }
    }
    await sleep(360)
    tableActionHint.value = ''
  },
}

const guicaiReplaceHandler: EventReplayHandler = {
  types: ['guicai_replace'],
  match: (e) => e.type === 'guicai_replace' && !!e.message,
  async replay(ctx) {
    const { event, mySeat, displayedHand, setTableCard, state, sleep } = ctx
    ctx.tableActionHint.value = event.message!
    ctx.centerMessage.value = event.message!
    // 鬼才改判：手牌飞到牌桌中央展示
    if (event.card) {
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
        if (event.player_index === mySeat) {
          displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
        }
        if (state.value && event.player_index != null) {
          state.value = {
            ...state.value,
            discard_count: state.value.discard_count + 1,
            players: state.value.players.map((p) =>
              p.index === event.player_index
                ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
                : p,
            ),
          }
        }
      })
    }
    await sleep(1000)
    ctx.tableActionHint.value = ''
  },
}

/** 鬼道改判：黑色手牌代替判定牌，展示在牌桌 */
const guidaoReplaceHandler: EventReplayHandler = {
  types: ['guidao_replace'],
  match: (e) => e.type === 'guidao_replace' && !!e.message,
  async replay(ctx) {
    const { event, mySeat, displayedHand, setTableCard, state, sleep } = ctx
    ctx.tableActionHint.value = event.message!
    ctx.centerMessage.value = event.message!
    if (event.card) {
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
        if (event.player_index === mySeat) {
          displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
        }
        if (state.value && event.player_index != null) {
          state.value = {
            ...state.value,
            discard_count: state.value.discard_count + 1,
            players: state.value.players.map((p) =>
              p.index === event.player_index
                ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
                : p,
            ),
          }
        }
      })
    }
    await sleep(1000)
    ctx.tableActionHint.value = ''
  },
}

const skillGiveCardHandler: EventReplayHandler = {
  types: ['skill_give_card'],
  match: (e) => e.type === 'skill_give_card' && e.target_index != null && !!e.card,
  async replay(ctx) {
    const { event, mySeat, appendDrawnCards, displayedHand, sleep } = ctx
    // 给牌直接到手上，不在牌桌展示
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
    const { event, mySeat, displayedHand, setTableCard, state } = ctx
    await flyBoltIfTargeted(ctx)
    await playCardToTable(ctx, () => {
      setTableCard(event.card!)
      if (event.player_index === mySeat) {
        displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
      }
      if (state.value && event.player_index != null) {
        state.value = {
          ...state.value,
          discard_count: state.value.discard_count + 1,
          players: state.value.players.map((p) =>
            p.index === event.player_index
              ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
              : p,
          ),
        }
      }
    })
    ctx.tableActionHint.value = ''
  },
}

const weaponSkillHandler: EventReplayHandler = {
  types: ['weapon_skill'],
  match: (e) => e.type === 'weapon_skill',
  async replay(ctx) {
    const { event, mySeat, displayedHand, setTableCard, state } = ctx
    if (event.card) {
      await flyBoltIfTargeted(ctx)
      await playCardToTable(ctx, () => {
        setTableCard(event.card!)
        if (event.player_index === mySeat) {
          displayedHand.value = displayedHand.value.filter((c) => c.id !== event.card!.id)
        }
        if (state.value && event.player_index != null) {
          state.value = {
            ...state.value,
            discard_count: state.value.discard_count + 1,
            players: state.value.players.map((p) =>
              p.index === event.player_index
                ? { ...p, hand_count: Math.max(0, p.hand_count - 1) }
                : p,
            ),
          }
        }
      })
    }
    ctx.tableActionHint.value = event.message ?? ''
    ctx.centerMessage.value = event.message ?? ctx.centerMessage.value
    await ctx.sleep(420)
    ctx.tableActionHint.value = ''
  },
}

const qilinDiscardHandler: EventReplayHandler = {
  types: ['qilin_discard'],
  match: (e) => e.type === 'qilin_discard' && e.target_index != null && !!e.card,
  async replay(ctx) {
    const { event, state, setTableCard, sleep } = ctx
    if (!state.value) return
    // 弃置的装备飞到牌桌中央展示
    await playCardToTable(ctx, () => {
      setTableCard(event.card!)
      if (state.value && event.target_index != null) {
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
      }
    })
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
    const { event, sleep, flashSeatHit, flashSeatBlocked, state, runHitSlash } = ctx
    if (event.target_index != null) {
      const dmg = event.damage ?? 0
      if (dmg > 0) {
        // 受伤穿梭线 → 震动 → 掉血
        await runHitSlash(event.target_index)
        await flashSeatHit(event.target_index)
      } else {
        await flashSeatBlocked(event.target_index)
      }
    }
    // 震动后再更新 HP（applyState 中保留了旧 HP，这里才更新）
    if (state.value && event.target_index != null && (event.damage ?? 0) > 0) {
      const seat = event.target_index
      const dmg = event.damage ?? 0
      state.value = {
        ...state.value,
        players: state.value.players.map((p) =>
          p.index === seat
            ? { ...p, hp: Math.max(0, p.hp - dmg) }
            : p,
        ),
      }
    }
    await sleep(280)
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
  takeCardHandler,
  trickHealHandler,
  lebuSkipHandler,
  bingliangSkipHandler,
  drawPhaseSkipHandler,
  playPhaseSkipHandler,
  wuguPickHandler,
  shandianJudgeHandler,
  huogongRevealHandler,
  trickResponseHandler,
  tiesuoChainHandler,
  tiesuoSpreadHandler,
  wuxiekCancelHandler,
  phaseMessageHandler,
  peekDeckHandler,
  guanxingTriggerHandler,
  skillAwakenHandler,
  skillJiangHandler,
  judgeFlipHandler,
  judgeResultHandler,
  luoshenGainHandler,
  luoshenStopHandler,
  skillTriggerHealHandler,
  gainCardSkillHandler('jianxiong_gain'),
  gainCardSkillHandler('fankui_take'),
  gainCardSkillHandler('tuxi_take'),
  guicaiReplaceHandler,
  guidaoReplaceHandler,
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
