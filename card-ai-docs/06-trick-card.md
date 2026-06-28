# 06 - 锦囊牌系统

> **用途**: 了解无名杀锦囊牌的定义结构、使用流程、AOE结算。
>
> **源码文件**: `card/standard.js` (锦囊定义), `noname/library/element/content.js` 行7138-7520 (useCard)

---

## 1. 卡牌类型体系

| type | 含义 | 示例 |
|------|------|------|
| `"basic"` | 基础牌 | 杀、闪、桃、酒 |
| `"trick"` | 非延时锦囊 | 过河拆桥、无懈可击、南蛮入侵 |
| `"delay"` | 延时锦囊 | 乐不思蜀、闪电、兵粮寸断 |
| `"equip"` | 装备牌 | 武器、防具、马匹 |

---

## 2. 锦囊牌定义结构（以"过河拆桥"为例）

```javascript
// card/standard.js
guohe: {
    type: "trick",                    // 类型：锦囊
    enable: true,                     // 是否可用
    selectTarget: [1, 1],            // 可选目标数 [min, max]（-1=全体）
    reverseOrder: false,             // 是否逆序结算（AOE用）
    
    // 目标合法性检查
    filterTarget: function(card, player, target) {
        return target != player && target.countCards("he") > 0;
    },
    
    // 生效内容（step写法）
    content: function() {
        "step 0";
        if (target.countCards("he")) {
            player.gainPlayerCard(target, "he", true);
        }
    },
    
    // AI 相关
    ai: {
        wuxie: function(target, card, player, viewer) {
            // AI判断是否使用无懈可击
        },
        basic: { order: 9, value: [4, 1], useful: [4, 1] },
        result: { target: function(player, target) { ... } }
    }
}
```

---

## 3. 锦囊使用流程

```
chooseToUse → 选择锦囊和合法目标
    ↓
createEvent("useCard") → 创建使用事件
    ↓
useCard0 → useCard1 → useCard2 → useCard  (逐步触发)
    ↓
遍历 targets:
    对每个目标:
    ├── useCardToPlayer → useCardToTarget
    ├── useCardToBegin  ← 【无懈可击触发点】
    ├── 执行 content()   ← 锦囊效果
    └── useCardToEnd → useCardToAfter
    ↓
useCardEnd → useCardAfter
```

---

## 4. 群体锦囊（AOE）特殊处理

```javascript
// 南蛮入侵
nanman: {
    type: "trick",
    selectTarget: -1,     // -1 = 全体其他角色
    reverseOrder: true,   // 从当前回合玩家逆序结算
    filterTarget: function(card, player, target) {
        return target != player;
    },
    content: function() {
        "step 0";
        target.chooseToRespond({ name: "sha" }, "请打出一张杀");
        "step 1";
        if (!result.bool) {
            target.damage(source);
        }
    }
}
```

---

## 5. 关键字段说明

```javascript
{
    type: "trick",           // 卡牌类型
    enable: true,            // 是否可用
    selectTarget: [1, 1],    // 目标数量范围 [min, max]，-1=全体
    reverseOrder: false,     // 是否逆序结算
    multitarget: false,      // 是否多目标（可分别无懈）
    wuxieable: true,         // 是否可被无懈（默认true）
    filterTarget(fn),        // 目标合法性检查
    content(fn),             // 生效内容
    precontent(fn),          // 使用前额外操作
    ai: {                    // AI相关
        wuxie: fn,           // AI无懈判断
        basic: { order, value, useful },
        result: { player, target }
    }
}
```

---

## 6. 在你的项目中如何参考

设计你自己的锦囊系统时，核心要点：

1. **卡牌类型区分**：basic/trick/delay/equip
2. **使用流程分层**：useCard → useCardToPlayer → useCardToBegin → content
3. **目标选择灵活**：selectTarget [min, max]，-1=全体
4. **AOE逆序结算**：从当前回合玩家逆时针依次结算
5. **无懈可击在 useCardToBegin 介入**
