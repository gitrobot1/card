# 09 - 武器与装备系统

> **用途**: 了解无名杀武器、防具、马匹的定义和技能触发机制。
>
> **源码文件**: `card/standard.js` (武器卡牌定义 行721-850+), `noname/library/element/content.js` 行700-994 (equip内容)

---

## 1. 武器列表

| 武器 | 范围 | 技能 | 核心效果 |
|------|------|------|----------|
| 诸葛连弩 (`zhuge`) | 1 | `zhuge_skill` | 杀无次数限制 |
| 雌雄双股剑 (`cixiong`) | 2 | `cixiong_skill` | 对异性出杀→目标弃牌或摸牌 |
| 青釭剑 (`qinggang`) | 2 | `qinggang_skill` | 无视目标防具 |
| 青龙刀 (`qinglong`) | 3 | `qinglong_skill` | 杀被闪后追加杀 |
| 丈八蛇矛 (`zhangba`) | 3 | `zhangba_skill` | 2手牌当杀 |
| 贯石斧 (`guanshi`) | 3 | `guanshi_skill` | 弃2牌强制命中 |
| 方天画戟 (`fangtian`) | 4 | `fangtian_skill` | 最后手牌可多目标 |
| 麒麟弓 (`qilin`) | 5 | `qilin_skill` | 造成伤害可弃目标马 |

---

## 2. 武器定义结构

```javascript
zhuge: {
    type: "equip",
    subtype: "equip1",     // equip1=武器, equip2=防具, equip3=马, equip4=宝物
    distance: { attackFrom: 1 },  // 攻击范围
    skills: ["zhuge_skill"],       // 装备时获得的技能
    ai: { basic: { equipValue: 5 } }
}
```

---

## 3. 典型武器技能

**诸葛连弩** (无限出杀)：
```javascript
zhuge_skill: {
    equipSkill: true,
    mod: {
        cardUsable: function(card, player, num) {
            if (card.name == "sha") return Infinity;  // 无限次
        }
    }
}
```

**丈八蛇矛** (牌当杀)：
```javascript
zhangba_skill: {
    equipSkill: true,
    enable: ["chooseToUse", "chooseToRespond"],  // 使用和响应时可用
    filterCard: function(card) { return true; },
    selectCard: 2,
    position: "hs",
    viewAs: { name: "sha" },  // 视为杀
    prompt: "将两张手牌当杀使用或打出",
}
```

**贯石斧** (强制命中)：
```javascript
guanshi_skill: {
    equipSkill: true,
    trigger: { player: ["shaMiss", "eventNeutralized"] },  // 杀被闪或被抵消
    filter(event, player) {
        return event.type == "card" && event.card.name == "sha";
    },
    content: function() {
        // 弃置两张牌 → untrigger → trigger("shaHit")
    }
}
```

**青釭剑** (无视防具)：
```javascript
qinggang_skill: {
    equipSkill: true,
    trigger: { player: "useCardToPlayered" },
    filter(event, player) {
        return event.card.name == "sha";
    },
    content: function() {
        // 给目标添加临时技能 qinggang2（使防具技能失效）
        // 在 damage/damageCancelled/shaMiss/useCardEnd 时自动移除
    }
}
```

---

## 4. 装备流程

```javascript
// content.js 行700-887
equip: function () {
    // 1. 牌从原拥有者处 lose
    // 2. player.addVirtualEquip(card, cards)  // 加入虚拟装备区
    // 3. 触发 "replaceEquip" 事件（顶替旧装备）
    // 4. 执行 onEquip 回调
    // 5. 创建 "equip_{cardName}" 事件
    // 6. 注册装备技能: player.addSkillTrigger(info.skills[i])
}
```

---

## 5. 卸载流程

```javascript
// player.js 行8550-8579
removeEquipTrigger(card, hasMove) {
    // 检查技能是否还被其他同名装备使用
    // 若无 → removeSkillTrigger(skill)
}
```

---

## 6. 防具示例

**八卦阵**：
```javascript
bagua_skill: {
    equipSkill: true,
    trigger: { player: "chooseToRespondBegin" },
    filter(event, player) {
        return event.respondTo && event.respondTo[1] && 
               event.respondTo[1].name == "sha";  // 需要出闪时
    },
    content: function() {
        // 判定：红色 → 视为打出了闪
    }
}
```

---

## 7. 在你的项目中如何参考

设计你自己的装备系统时，核心要点：

1. **装备=卡牌+技能**：装备牌自带 skills 数组，装备时注册技能
2. **equipSkill 标记**：装备技能自动在卸下时移除
3. **武器用 mod.cardUsable 控制次数**
4. **武器用 viewAs 实现"牌当牌"效果**
5. **装备流程**：lose → addVirtualEquip → replaceEquip → addSkillTrigger
6. **卸载流程**：removeEquipTrigger → removeSkillTrigger（检查同名装备）
