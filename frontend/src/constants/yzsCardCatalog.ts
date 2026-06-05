import { YZS_EQUIP_META, YZS_WEAPON_META } from './yzsWeapons'

export interface CodexCardEntry {
  kind: string
  name: string
  effect: string
  /** 武器 / 防具 / 坐骑 等子类 */
  subtype?: string
  range?: number
  /** 延时锦囊等标签 */
  tag?: string
}

export const YZS_BASIC_CARDS: CodexCardEntry[] = [
  {
    kind: 'sha',
    name: '杀',
    effect: '出牌阶段，对攻击范围内的一名其他角色使用；目标需打出【闪】，否则受到 1 点伤害（【酒】、武器或技能可增伤）。',
  },
  {
    kind: 'shan',
    name: '闪',
    effect: '当成为【杀】或【万箭齐发】的目标时，可打出以抵消该次对你生效的【杀】或【万箭】效果。',
  },
  {
    kind: 'tao',
    name: '桃',
    effect: '出牌阶段，若体力未满可对自己使用并回复 1 点体力；当处于濒死状态时，可对自己或该角色使用以回复 1 点体力。',
  },
  {
    kind: 'jiu',
    name: '酒',
    effect: '出牌阶段对自己使用：本回合下一张【杀】伤害 +1；当处于濒死状态时，可对自己使用并回复 1 点体力。',
  },
]

export const YZS_WEAPON_CARDS: CodexCardEntry[] = [
  ...Object.entries(YZS_WEAPON_META).map(([kind, meta]) => ({
    kind,
    name: meta.name,
    effect: meta.effect,
    subtype: '武器',
    range: meta.range,
  })),
  ...Object.entries(YZS_EQUIP_META).map(([kind, meta]) => ({
    kind,
    name: meta.name,
    effect: meta.effect,
    subtype: kind === 'plus_horse' || kind === 'minus_horse' ? '坐骑' : '防具',
  })),
]

export const YZS_TRICK_CARDS: CodexCardEntry[] = [
  {
    kind: 'guohe',
    name: '过河拆桥',
    effect: '出牌阶段，对一名有牌的其他角色使用，弃置其区域内的一张牌（【藤甲】持有者无效）。',
  },
  {
    kind: 'tannang',
    name: '顺手牵羊',
    effect: '出牌阶段，对距离 1 以内有牌的其他角色使用，获得其区域内的一张牌（【藤甲】持有者无效）。',
  },
  {
    kind: 'wuzhong',
    name: '无中生有',
    effect: '出牌阶段对自己使用，摸两张牌。',
  },
  {
    kind: 'wugu',
    name: '五谷丰登',
    effect: '出牌阶段对自己使用，亮出等于存活角色数的牌，每名角色按顺序选一张获得。',
  },
  {
    kind: 'taoyuan',
    name: '桃园结义',
    effect: '出牌阶段使用，所有角色各回复 1 点体力。',
  },
  {
    kind: 'nanman',
    name: '南蛮入侵',
    effect: '出牌阶段使用，所有其他角色依次需打出【杀】，否则受到 1 点伤害（【藤甲】持有者无效）。',
  },
  {
    kind: 'wanjian',
    name: '万箭齐发',
    effect: '出牌阶段使用，所有其他角色依次需打出【闪】，否则受到 1 点伤害（【藤甲】持有者无效）。',
  },
  {
    kind: 'juedou',
    name: '决斗',
    effect: '出牌阶段，对一名其他角色使用；由目标开始，双方轮流打出【杀】，先不出者受到 1 点伤害（【藤甲】持有者无效）。',
  },
  {
    kind: 'wuxiek',
    name: '无懈可击',
    effect: '当一张锦囊牌生效前，可打出以抵消该锦囊对一名角色的效果（可连锁抵消）。',
  },
  {
    kind: 'huogong',
    name: '火攻',
    effect: '出牌阶段，对一名有手牌的角色使用；展示其一张手牌，你需弃置一张同花色手牌，否则其受到 1 点火焰伤害。',
  },
  {
    kind: 'tiesuo',
    name: '铁索连环',
    effect: '出牌阶段，对一名角色使用令其横置/重置；或重铸（弃置此牌并摸一张）。横置角色受到火焰伤害时，其他横置角色受到相同传导伤害。',
  },
  {
    kind: 'lebu',
    name: '乐不思蜀',
    tag: '延时锦囊',
    effect: '出牌阶段，对一名其他角色使用并置入其判定区；其判定阶段判定，若不为红桃则跳过出牌阶段。',
  },
  {
    kind: 'bingliang',
    name: '兵粮寸断',
    tag: '延时锦囊',
    effect: '出牌阶段，对距离 1 以内的其他角色使用并置入其判定区；其判定阶段判定，若不为梅花则跳过摸牌阶段。',
  },
  {
    kind: 'shandian',
    name: '闪电',
    tag: '延时锦囊',
    effect: '出牌阶段对自己使用并置入判定区；判定阶段判定，若为黑桃 2–9 则受到 3 点雷电伤害，否则传给下家判定区。',
  },
]

export const YZS_PACK_LABELS: Record<string, string> = {
  standard: '标准包',
  sp: 'SP',
  shen: '神',
}
