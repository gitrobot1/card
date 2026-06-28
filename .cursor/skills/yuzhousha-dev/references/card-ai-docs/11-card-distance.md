# 11 - 变牌系统与距离计算

> **用途**: 了解无名杀牌的获得/失去/弃置流程，以及距离计算规则。
>
> **源码文件**: `noname/library/element/content.js` 行6624-8328 (gain/gainPlayerCard/lose/discard), `noname/library/element/player.js` 行6140-6350 (gain/lose方法)

---

## 1. 获得牌 (gain)

```javascript
// player.js 行6140-6170
player.gain(cards, source, animate)
// animate: "draw" | "gain" | "gain2" | "draw2" | "give" | "giveAuto"

// 流程:
"step 0": 构建 losing_map（按原拥有者分组），调用 owner.lose() 移出牌
"step 1": 过滤被销毁的牌
"step 2": 更新统计 player.getStat().gain
"step 3": 根据 animate 类型播放动画
          "draw" → player.$draw(cards.length)    // 摸牌动画
          "gain" → player.$gain(cards)           // 获得动画
          "give" → 从源玩家动画转移
          → 插入手牌区DOM，按 sort_card 排序
```

---

## 2. 失去牌 (lose)

```javascript
// player.js 行6320-6350
player.lose(cards, position, visible)
// 从手牌区/装备区/判定区移除牌
// 更新UI
// 触发相关事件
```

---

## 3. 弃置 (discard)

```javascript
// content.js 行7943-7951
discard: function () {
    // 本质 = player.lose(cards, event.position, "visible")
    // 设置 type = "discard"
    // 完成后触发 event.trigger("discard")
}
```

---

## 4. 摸牌 (draw)

```javascript
player.draw(num)
// 从牌堆顶获取 num 张牌
// → player.gain(cards)
// 支持 game.modPhaseDraw 修改摸牌数
```

---

## 5. 获得他人牌 (gainPlayerCard)

```javascript
// content.js 行6624-6819
gainPlayerCard: function () {
    // 弹出选择界面
    // 可选区域: "h" (手牌), "e" (装备), "j" (判定区)
    // 支持 directresult 直接指定（跳过UI）
    // → player.gain(event.cards, target)
}
```

---

## 6. 牌区概念

| 区域 | 标识 | 说明 |
|------|------|------|
| 手牌区 | `h` | 手中的牌 |
| 装备区 | `e` | 已装备的牌 |
| 判定区 | `j` | 延时锦囊 |
| 处理区 | `ordering` | 正在处理的牌 |
| 弃牌堆 | `discardPile` | 已弃置的牌 |
| 牌堆 | `cardPile` | 未摸的牌 |

---

## 7. 距离计算

### 7.1 座位距离
```javascript
get.distance(playerA, playerB)
// 返回顺时针和逆时针座位差的最小值
```

### 7.2 攻击范围
```javascript
// 武器提供攻击距离
zhuge: { distance: { attackFrom: 1 } }    // 范围=1
qilin: { distance: { attackFrom: -5 } }   // 范围=5 (默认+5)

// 马匹
chitu: { distance: { globalFrom: -1 } }   // -1马（进攻）
dilu: { distance: { globalTo: +1 } }      // +1马（防御）

// 计算攻击范围
player.inRange(target)
// = 座位距离 + globalTo(target) + globalFrom(player) + attackFrom(player)
```

### 7.3 距离修正 (mod)
```javascript
// 技能可修改距离判定
qicai: {
    mod: {
        targetInRange(card, player, target, now) {
            if (["trick", "delay"].includes(get.type(card))) 
                return true;  // 锦囊无视距离
        }
    }
}
```

---

## 8. 在你的项目中如何参考

### 变牌系统：
1. **gain = lose + 动画**：先失去再获得，保证原子性
2. **牌区标识**：h/e/j 三元组覆盖所有位置
3. **动画类型**：draw/gain/give 区分不同视觉表现

### 距离系统：
1. **座位距离** = min(顺时针, 逆时针)
2. **攻击范围** = 座位距离 + 装备修正 + 技能修正
3. **-1马** (globalFrom) = 减少自己到别人的距离
4. **+1马** (globalTo) = 增加别人到自己的距离
