# 08 - 延迟锦囊系统

> **用途**: 了解无名杀延迟锦囊（闪电、乐不思蜀、兵粮寸断）的定义和判定逻辑。
>
> **源码文件**: `card/standard.js` 行2531-2666 (lebu/shandian), `card/extra.js` 行392-438 (bingliang)

---

## 1. 延迟锦囊 vs 普通锦囊

| 特性 | 普通锦囊 (`trick`) | 延迟锦囊 (`delay`) |
|------|-------------------|-------------------|
| 结算时机 | 使用后立即结算 | 进入判定区，在判定阶段结算 |
| 无懈时机 | `useCardToBegin` | `phaseJudge` |
| 存放位置 | 手牌→处理区 | 手牌→判定区(`j`) |
| 可被顺/拆 | 不能 | 能（判定区牌可被获得/弃置） |

---

## 2. 乐不思蜀

```javascript
// card/standard.js 行2531-2586
lebu: {
    type: "delay",
    filterTarget: function(card, player, target) {
        return lib.filter.judge(card, player, target) && player != target;
    },
    judge: function(card) {
        if (get.suit(card) == "heart") return 1;   // 红桃 → 失效
        return -2;                                   // 其他 → 生效（跳过出牌阶段）
    },
    judge2: function(result) {
        if (result.bool == false) return true;      // 判定失败 → 播放生效动画
        return false;
    },
    effect: function() {
        if (result.bool == false) {
            player.skip("phaseUse");  // 跳过出牌阶段
        }
    },
}
```

---

## 3. 闪电

```javascript
// card/standard.js 行2587-2666
shandian: {
    type: "delay",
    cardnature: "thunder",             // 雷电属性
    filterTarget: function(card, player, target) {
        return lib.filter.judge(card, player, target) && player == target;
    },
    toself: true,                       // 只能对自己使用
    selectTarget: [-1, -1],            // 默认选择自己
    
    judge: function(card) {
        // 黑桃2-9 → 生效（造成伤害）
        if (get.suit(card) == "spade" && 
            get.number(card) > 1 && get.number(card) < 10) 
            return -5;
        return 1;  // 其他 → 失效
    },
    
    judge2: function(result) {
        if (result.bool == false) return true;
        return false;
    },
    
    effect: function() {
        if (result.bool == false) {
            player.damage(3, "thunder", "nosource");  // 3点雷电伤害
        } else {
            player.addJudgeNext(card);  // 不生效 → 传给下家
        }
    },
    
    cancel: function() {
        player.addJudgeNext(card);  // 被无懈后也传给下家
    },
}
```

---

## 4. 兵粮寸断

```javascript
// card/extra.js 行392-438
bingliang: {
    type: "delay",
    range: { global: 1 },              // 距离为1的角色
    filterTarget: function(card, player, target) {
        return lib.filter.judge(card, player, target) && player != target;
    },
    judge: function(card) {
        if (get.suit(card) == "club") return 1;  // 梅花 → 失效
        return -2;                                 // 其他 → 生效
    },
    effect: function() {
        if (result.bool == false) {
            player.skip("phaseDraw");  // 跳过摸牌阶段
        }
    },
}
```

---

## 5. 判定区管理

```javascript
// player.js
// 检查能否加入判定区
canAddJudge(card) { /* 检查禁用、重复、出局、mod限制 */ }

// 将牌置入判定区
addJudge(card, cards) { /* 创建 addJudge 事件 */ }

// 传给下家（闪电核心机制）
addJudgeNext(card) { /* 找到下家 → 移入其判定区 */ }
```

---

## 6. 判定阶段的处理流程

```javascript
// content.js 行3890-3944
phaseJudge: function () {
    "step 0";
    event.cards = player.getVCards("j");  // 获取判定区牌（后进先出）
    
    "step 1";
    event.card = cards.pop();  // 取最后放入的
    player.lose(event.card.cards);  // 移出判定区
    event.trigger("phaseJudge");  // ← 无懈可击介入点
    
    "step 2";
    player.judge(event.card);  // 执行判定
    
    "step 3";
    if (event.cancelled) {
        // 被无懈 → 执行 cancel()
        lib.card[name].cancel();
    } else {
        // 未被无懈 → 执行 effect()
        lib.card[name].effect();
    }
    event.goto(1);  // 循环处理下一张
}
```

---

## 7. 在你的项目中如何参考

设计你自己的延迟锦囊时，核心要点：

1. **type = "delay"** 区分延迟锦囊
2. **进入判定区**：使用后不立即结算，放入 j 区
3. **判定阶段结算**：后进先出（最后放的先判定）
4. **三段函数**：judge(判定条件) + effect(生效) + cancel(被无懈)
5. **judge2 控制动画**：boolean 返回控制动画方向
6. **闪电特殊机制**：不生效或被无懈后传给下家
