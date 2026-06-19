/** 标准版武器 / 装备展示信息（与后端 kind 对应） */
export interface YzsWeaponMeta {
  name: string
  range: number
  effect: string
}

export interface YzsEquipMeta {
  name: string
  effect: string
}

export const YZS_WEAPON_META: Record<string, YzsWeaponMeta> = {
  weapon_1: {
    name: '诸葛连弩',
    range: 1,
    effect: '出牌阶段使用【杀】无次数限制',
  },
  weapon_2: {
    name: '青釭剑',
    range: 2,
    effect: '使用【杀】无视目标防具',
  },
  weapon_3: {
    name: '青龙偃月刀',
    range: 3,
    effect: '【杀】被【闪】抵消时可再对其使用【杀】',
  },
  weapon_4: {
    name: '方天画戟',
    range: 4,
    effect: '【杀】为最后手牌时可额外指定目标',
  },
  weapon_5: {
    name: '麒麟弓',
    range: 5,
    effect: '【杀】造成伤害时可弃置目标坐骑',
  },
  weapon_6: {
    name: '古锭刀',
    range: 2,
    effect: '锁定技，若目标没有手牌，【杀】伤害+1',
  },
  weapon_7: {
    name: '朱雀羽扇',
    range: 4,
    effect: '使用普通【杀】时，可将其转为【火杀】',
  },
  weapon_8: {
    name: '雌雄双股剑',
    range: 2,
    effect: '使用【杀】指定异性角色后，其需弃一张手牌，否则你摸一张牌',
  },
  weapon_9: {
    name: '贯石斧',
    range: 3,
    effect: '【杀】被【闪】抵消时，可弃置两张牌令此【杀】依然命中',
  },
}

export const YZS_EQUIP_META: Record<string, YzsEquipMeta> = {
  armor: {
    name: '八卦阵',
    effect: '需出【闪】时可判定，红色视为出【闪】',
  },
  armor_vine: {
    name: '藤甲',
    effect: '锁定技，南蛮入侵、万箭齐发、普通杀对你无效；受到的火焰伤害+1',
  },
  plus_horse: {
    name: '+1马',
    effect: '其他角色计算与你距离+1',
  },
  minus_horse: {
    name: '-1马',
    effect: '你计算与其他角色距离-1',
  },
}

export function weaponMetaForKind(kind: string | undefined): YzsWeaponMeta | undefined {
  if (!kind) return undefined
  return YZS_WEAPON_META[kind]
}

export function equipMetaForKind(kind: string | undefined): YzsEquipMeta | undefined {
  if (!kind) return undefined
  return YZS_EQUIP_META[kind]
}

export function weaponRangeForKind(kind: string | undefined): number {
  return weaponMetaForKind(kind)?.range ?? 1
}

export function equipDisplayName(card: { kind: string; name: string }): string {
  const w = weaponMetaForKind(card.kind)
  if (w) return w.name
  const e = equipMetaForKind(card.kind)
  if (e) return e.name
  return card.name
}

export function equipDisplaySummary(card: { kind: string; name: string }): string {
  const w = weaponMetaForKind(card.kind)
  if (w) return w.name
  const e = equipMetaForKind(card.kind)
  if (e) return e.name
  return card.name
}
