# 05 - 判定与改判系统

> **用途**: 了解无名杀判定流程、改判技能如何介入、判定结果的计算。
>
> **源码文件**: `noname/library/element/content.js` 行9474-9562 (judge content), `noname/library/element/player.js` 行7033-7067 (judge方法)

---

## 1. 判定流程

```
牌堆顶取牌 → 亮出判定牌 → trigger("judge")  ← 【改判技介入点】
→ 构建 result 对象 → judge函数计算 → mod.judge 修改
→ trigger("judgeFixing") → callback 回调
```

---

## 2. 核心代码

```javascript
// content.js 行9474-9562
judge: function () {
    "step 0";
    // 1. 从牌堆顶取一张牌
    var cardj = player.getTopCards()[0] || get.cards()[0];
    // 2. 移入处理区
    owner.lose(cardj, "visible", ui.ordering);
    // 3. 放入判定栈
    player.judging.unshift(cardj);
    // 4. 显示判定牌动画
    // 5. 触发 "judge" 事件 ← 改判技能在这里介入！
    event.trigger("judge");
    
    "step 1";
    // 构建判定结果对象
    event.result = {
        card: player.judging[0],     // 判定牌
        name: player.judging[0].name, // 牌名
        number: get.number(player.judging[0]),  // 点数
        suit: get.suit(player.judging[0]),      // 花色: spade/heart/club/diamond
        color: get.color(player.judging[0]),    // 颜色: red/black
    };
    
    // 应用预设结果（跳过随机判定）
    if (event.fixedResult) {
        Object.assign(event.result, event.fixedResult);
    }
    
    // 执行判定函数
    event.result.judge = event.judge(event.result);
    if (event.result.judge > 0) event.result.bool = true;      // 判定成功
    else if (event.result.judge < 0) event.result.bool = false; // 判定失败
    else event.result.bool = null;                               // 无结果
    
    // 移出判定栈
    player.judging.shift();
    
    // 执行 mod.judge（技能修改判定结果）
    game.checkMod(player, event.result, "judge", player);
    
    // 二次判定（控制动画）
    if (event.judge2) {
        var judge2 = event.judge2(event.result);
        if (typeof judge2 == "boolean") player.tryJudgeAnimate(judge2);
    }
    
    // 触发 "judgeFixing"
    event.trigger("judgeFixing");
    
    // 执行回调
    if (event.callback) {
        event.callback(event.result);
    }
}
```

---

## 3. 判定函数返回值含义

```javascript
// judge(card) 返回值:
//   > 0 (正数)  → result.bool = true   → 判定成功
//   < 0 (负数)  → result.bool = false  → 判定失败
//   = 0        → result.bool = null   → 无结果

// 典型判定函数示例:
// 乐不思蜀: 红桃 → 1 (成功/失效), 其他 → -2 (失败/生效)
lebu: { judge(card) {
    if (get.suit(card) == "heart") return 1;
    return -2;
}}

// 闪电: 黑桃2-9 → -5 (失败/生效), 其他 → 1 (成功/失效)
shandian: { judge(card) {
    if (get.suit(card) == "spade" && get.number(card) > 1 && get.number(card) < 10) 
        return -5;
    return 1;
}}

// 刚烈: 红桃 → -2 (失败), 其他 → 2 (成功)
ganglie: { judge(card) {
    if (get.suit(card) == "heart") return -2;
    return 2;
}}
```

---

## 4. 改判的两种方式

### 方式一：替换判定牌（鬼才模式）

```javascript
// character/standard/skill.js 行290-356
guicai: {
    trigger: { global: "judge" },  // 监听全局判定事件
    filter(event, player) {
        return player.countCards("hs") > 0;  // 有手牌才能改判
    },
    async content(event, trigger, player) {
        // 玩家选择一张手牌
        // 替换判定牌: trigger.player.judging[0] = 选中的牌
        // 直接替换判定栈顶的牌即可改变判定结果
    }
}
```

### 方式二：mod.judge 修改结果

```javascript
// 技能声明 mod.judge
skill: {
    mod: {
        judge(player, result) {
            // 直接修改 result 的属性
            result.number = 10;       // 修改点数
            result.suit = "spade";    // 修改花色
            result.color = "black";   // 修改颜色
            result.bool = true;       // 修改判定结果
        }
    }
}

// 系统调用 (game.checkMod 遍历所有技能的 mod.judge)
game.checkMod(player, event.result, "judge", player);
```

---

## 5. 改判的关键时机

```
判定流程时间线:
  取牌 → 亮出 → 【"judge" 触发】→ 构建result → judge计算
  → 【mod.judge 修改】→ 【"judgeFixing" 触发】→ callback

改判可介入点:
  1. "judge" 事件触发时 → 替换 judging[0]（鬼才类）
  2. mod.judge → 直接修改 result 属性（技能被动修改）
  3. "judgeFixing" → 判定修正后的最终确认
```

---

## 6. 判定相关属性

```javascript
event.judgestr       // 判定描述（如"闪电"、"乐不思蜀"）
event.fixedResult    // 预设判定结果（跳过随机判定）
event.judge          // 判定函数 fn(result) → number
event.judge2         // 二次判定函数 fn(result) → boolean（控制动画方向）
event.noJudgeTrigger // 跳过 "judge" 事件触发（不触发改判技）
event.directresult   // 直接指定判定牌
player.judging[]     // 当前判定牌栈（栈顶 = 正在判定的牌）
```

---

## 7. 在你的项目中如何参考

设计你自己的判定系统时，核心要点：

1. **判定是事件**：有自己的 Before/Begin/End 钩子
2. **改判通过监听 "judge" 事件实现**：替换 judging 栈顶的牌
3. **判定结果结构化**：card/name/number/suit/color/bool
4. **mod.judge 被动修改**：不需要玩家操作，自动生效
5. **judgeFixing 最后确认**：所有修改完成后触发
