# 04 - 伤害结算系统

> **用途**: 了解无名杀伤害事件的完整结算流程、属性系统、濒死检测。
>
> **源码文件**: `noname/library/element/content.js` 行8773-8924 (damage content), `noname/library/element/player.js` 行6384-6444 (damage方法)

---

## 1. 伤害事件创建

```javascript
// player.js 行6384-6444
player.damage(num, nature, source)
// 示例:
player.damage()                  // 默认1点伤害
player.damage(2)                 // 2点伤害
player.damage("fire")            // 1点火属性伤害
player.damage(3, "thunder")      // 3点雷属性伤害
player.damage(source)            // 来源为source的1点伤害
player.damage("nosource")        // 无来源伤害（如闪电）
player.damage("notrigger")       // 不触发技能
player.damage("unreal")          // 视为伤害（不扣血）

// 伤害值计算:
event.num = (event.baseDamage || 1) + (event.extraDamage || 0);
// baseDamage: 基础伤害（如杀=1，酒杀=2）
// extraDamage: 额外伤害（如古锭刀+1）
```

---

## 2. 伤害结算流程

```
damage 事件生命周期:
│
├── step 0: checkDamage1 → trigger("damageBegin1")
├── step 1: checkDamage2 → trigger("damageBegin2")
├── step 2: checkDamage3 → trigger("damageBegin3")  ← 藤甲等可在此取消伤害
├── step 3: checkDamage4 → trigger("damageBegin4")
│
├── step 4: 核心结算
│   ├── 播放伤害音效
│   ├── 记录日志和统计
│   ├── player.changeHp(-num)  ← 实际扣血
│   ├── 播放动画 ($damage, $fire, $thunder)
│   ├── trigger("damage")  或 trigger("damageZero")
│   └── [属性: fire/thunder/poison/ice]
│
├── step 5: 濒死检测
│   └── if (hp <= 0 && isAlive()) → player.dying(event)
│
└── step 6: trigger("damageSource")
```

---

## 3. 核心代码

```javascript
// content.js 行8773-8873
damage: function () {
    "step 0"; event.forceDie = true;
    if (event.unreal) { event.goto(4); return; }  // 视为伤害跳过前置
    game.callHook("checkDamage1", [event, player]);
    event.trigger("damageBegin1");
    "step 1"; game.callHook("checkDamage2", [event, player]);
    event.trigger("damageBegin2");
    "step 2"; game.callHook("checkDamage3", [event, player]);
    event.trigger("damageBegin3");
    "step 3"; game.callHook("checkDamage4", [event, player]);
    event.trigger("damageBegin4");
    "step 4";
        // 伤害音效
        // 日志记录: "X受到了来自Y的N点伤害"
        // 统计更新: stat.damaged / stat.damage
        player.changeHp(-num, false);  // 扣血
        // 动画: $damage, $fire, $thunder
        if (num == 0) event.trigger("damageZero");
        else event.trigger("damage");  // ← 卖血技触发点！
    "step 5";
        if (player.hp <= 0 && player.isAlive() && !event.nodying) {
            player.dying(event);  // 濒死求桃
        }
    "step 6"; event.trigger("damageSource");
}
```

---

## 4. 伤害事件的关键属性

```javascript
event.num           // 伤害值
event.original_num  // 原始伤害值（修改前）
event.source        // 伤害来源玩家
event.player        // 受伤者
event.nature        // 伤害属性: "fire"/"thunder"/"poison"/"ice"/null(普通)
event.card          // 造成伤害的牌（如杀）
event.unreal        // 是否视为伤害（不扣血，不触发卖血技）
event.notrigger     // 是否跳过触发
event.nodying       // 是否跳过濒死检测
event.change_history // 伤害值变更历史 [每次修改的差值]

// 属性检查方法
event.hasNature("fire")    // 是否包含火属性
event.hasNature("thunder") // 是否包含雷属性
```

---

## 5. 伤害触发链（完整时序）

```
damageBegin1 → damageBegin2 → damageBegin3 → damageBegin4
→ damage / damageZero        ← 卖血技能在此触发（如奸雄、刚烈）
→ [濒死检测] dying            ← 求桃流程
→ damageSource                ← 伤害来源相关技能（如狂暴）

事件结束后自动追加后缀:
→ damageEnd    ← 反馈、刚烈等在此触发
```

---

## 6. 卖血技示例

**刚烈** (受伤后判定，失败则来源受伤或弃牌):
```javascript
// character/standard/skill.js 行357-396
ganglie: {
    trigger: { player: "damageEnd" },  // 注意：是 damageEnd，不是 damage
    filter(event, player) {
        return event.source != undefined;  // 必须有伤害来源
    },
    async content(event, trigger, player) {
        // 判定：红桃 → 失败，其他 → 来源弃2牌或受1点伤害
        const judgeEvent = player.judge(card => {
            if (get.suit(card) == "heart") return -2;
            return 2;
        });
        // ...
    }
}
```

**反馈** (受伤后从来源获得一张牌):
```javascript
// character/standard/skill.js 行267-289
fankui: {
    trigger: { player: "damageEnd" },
    filter(event, player) {
        return event.source && 
               event.source.countGainableCards(player, "he") > 0;
    },
    async content(event, trigger, player) {
        player.gainPlayerCard(true, trigger.source, "he");
    }
}
```

---

## 7. 在你的项目中如何参考

设计你自己的伤害系统时，核心要点：

1. **分步结算**：伤害前置检查 (damageBegin1~4) → 实际伤害 (damage) → 濒死检测 → 来源结算
2. **属性系统**：fire/thunder/poison/ice，可组合
3. **伤害值可变**：通过 original_num 和 change_history 追踪修改
4. **卖血技在 damage 或 damageEnd 触发**：区别在于 damageEnd 在所有触发完成后执行
5. **濒死独立处理**：ddying 事件嵌套，支持多人求桃
