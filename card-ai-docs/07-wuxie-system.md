# 07 - 无懈可击系统

> **用途**: 了解无名杀无懈可击的触发时机、抵消链机制。
>
> **源码文件**: `card/standard.js` 行3350-3661 (_wuxie 技能), `card/standard.js` 行2461+ (wuxie卡牌定义)

---

## 1. 核心技能：`_wuxie`

```javascript
// card/standard.js 行3350-3368
_wuxie: {
    trigger: { player: ["useCardToBegin", "phaseJudge"] },
    priority: 5,
    forced: true,      // 强制触发（系统自动询问，不询问是否发动）
    silent: true,      // 无描述，同优先级自动优先执行
    popup: false,      // 不弹出技能名
    
    filter: function(event, player) {
        // 不可无懈的情况：
        // 1. card.storage.nowuxie 标记
        // 2. info.wuxieable === false
        // 3. event.getParent().nowuxie 标记
        // 4. player.hasSkillTag("playernowuxie")  ← 如帷幕
        // 5. 不是 trick 类型且未标记 wuxieable
        if (event.card.storage && event.card.storage.nowuxie) return false;
        var info = get.info(card);
        if (info.wuxieable === false) return false;
        if (event.name != "phaseJudge") {
            if (event.getParent().nowuxie) return false;
            if (event.player.hasSkillTag("playernowuxie", false, event.card)) return false;
            if (get.type(event.card) != "trick" && !info.wuxieable) return false;
        }
        return true;
    },
}
```

---

## 2. 无懈可击的触发时机

| 时机 | 事件名 | 说明 |
|------|--------|------|
| 锦囊对目标生效前 | `useCardToBegin` | 锦囊即将对某个目标生效 |
| 判定阶段 | `phaseJudge` | 延时锦囊即将判定 |

---

## 3. 无懈可击抵消链机制

这是无名杀最精妙的设计之一——**对无懈可击的无懈可击**。

```javascript
// card/standard.js 行3371-3661 (content简化)
_wuxie.content = function() {
    "step 0";
    // 创建状态映射
    var map = {
        card: trigger.card,
        player: trigger.player,
        target: trigger.target,
        state: 1,  // 1=锦囊将生效, -1=锦囊将失效
        isJudge: trigger.name == "phaseJudge",
    };
    
    // 如果是对无懈的无懈，翻转 state
    if (card.name == "wuxie") {
        // 向上查找父级_wuxie事件
        var evt = event;
        while (true) {
            evt = evt.getParent(5);
            if (evt && evt.name == "_wuxie") {
                state = !state;  // 翻转状态
            } else break;
        }
    }
    
    "step 1";
    // 收集所有拥有无懈可击的玩家
    var list = game.filterPlayer(function(current) {
        return current.hasWuxie(map);
    });
    list.sortBySeat(_status.currentPhase);
    
    "step 2";
    // 按座位依次询问
    if (event.list.length == 0) {
        event.finish();  // 无人可无懈
    } else {
        event.current = event.list.shift();
        event.send(event.current, event._info_map);  // 询问此玩家
    }
    
    "step 3";
    if (result.bool) {
        // 有人使用了无懈可击 → 创建新的 _wuxie 事件（抵消链）
        event.wuxieresult = event.current;
        event.goto(8);
    } else {
        event.goto(2);  // 继续询问下一个玩家
    }
}
```

**无懈链示意**：
```
锦囊A即将生效 (state=1)
  → 玩家X使用无懈可击 → 锦囊A即将失效 (state=-1)
    → 玩家Y使用无懈可击 → 锦囊A即将生效 (state=1)
      → 无人再出 → 最终state=1 → 锦囊A生效
```

---

## 4. 关键设计要点

1. **state 翻转**：每次无懈翻转生效/失效状态
2. **座位顺序询问**：从当前回合玩家开始逆时针
3. **临时"不无懈"按钮**：群体锦囊时显示快捷跳过按钮
4. **firstDo/forced/silent**：确保 _wuxie 在所有其他技能之前处理

---

## 5. 在你的项目中如何参考

设计你自己的无懈可击系统时，核心要点：

1. **两个触发点**：useCardToBegin（锦囊）+ phaseJudge（延时锦囊）
2. **抵消链**：对无懈的无懈通过状态翻转实现
3. **座位顺序询问**：从当前回合玩家开始
4. **不可无懈的检查**：卡牌标记/技能标签/类型限制
5. **状态追踪**：用 boolean state 追踪当前是"生效"还是"失效"状态
