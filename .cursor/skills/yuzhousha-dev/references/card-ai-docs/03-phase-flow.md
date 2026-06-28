# 03 - 阶段流转系统（回合制）

> **用途**: 了解无名杀的六大阶段系统，包括完整回合时序、阶段跳过、轮数管理。
>
> **源码文件**: `noname/library/element/content.js`
> **关键行号**: 3028-3051 (phaseLoop), 3638-3852 (phase/阶段流转), 3886-4046 (各阶段实现)

---

## 1. 六大阶段

```javascript
// content.js 行3644
event.phaseList = [
    "phaseZhunbei",   // 准备阶段
    "phaseJudge",     // 判定阶段
    "phaseDraw",      // 摸牌阶段
    "phaseUse",       // 出牌阶段
    "phaseDiscard",   // 弃牌阶段
    "phaseJieshu"     // 结束阶段
];
```

---

## 2. 完整回合时序

```
phaseLoop (全局循环，所有玩家轮流转)
│
├── phaseBefore          ← 回合开始前
├── phaseBeforeStart     ← 回合开始后①
├── phaseBeforeEnd       ← 回合开始后④
├── [翻面检测]            ← 翻面则跳过整个回合
├── phaseBeginStart      ← 回合开始后⑦
├── phaseBegin           ← 回合开始后⑨
│
├── phaseChange          ← 阶段切换时
│   ├── phaseZhunbei     ← 准备阶段
│   │   └── trigger("phaseZhunbei")
│   │
│   ├── phaseJudge       ← 判定阶段
│   │   ├── 遍历判定区牌 (后进先出)
│   │   ├── trigger("phaseJudge") ← 无懈可击介入点
│   │   ├── 判定 → effect() 或 cancel()
│   │   └── 循环处理下一张
│   │
│   ├── phaseDraw        ← 摸牌阶段
│   │   ├── phaseDrawBegin1 / phaseDrawBegin2
│   │   ├── draw(num) 摸牌
│   │   └── game.modPhaseDraw 可修改摸牌数
│   │
│   ├── phaseUse         ← 出牌阶段
│   │   ├── phaseUseBefore / phaseUseBegin
│   │   ├── chooseToUse() 循环
│   │   ├── phaseUseEnd / phaseUseAfter
│   │   └── 清理 ui.tempnowuxie
│   │
│   ├── phaseDiscard     ← 弃牌阶段
│   │   ├── needsToDiscard() 计算需弃牌数
│   │   ├── trigger("phaseDiscard")
│   │   └── chooseToDiscard()
│   │
│   └── phaseJieshu      ← 结束阶段
│       └── trigger("phaseJieshu")
│
├── phaseEnd             ← 回合结束
├── phaseAfter           ← 回合结束后
└── [下一名玩家]          ← phaseLoop 循环
```

---

## 3. phaseLoop：全局回合循环

```javascript
// content.js 行3028-3051
phaseLoop: function () {
    "step 0";
    // 分配座位号
    var num = 1, current = player;
    while (current.getSeatNum() == 0) {
        current.setSeatNum(num);
        current = current.next;
        num++;
    }
    "step 1";
    // 执行 onphase 回调 → 当前玩家执行 phase()
    for (var i = 0; i < lib.onphase.length; i++) {
        lib.onphase[i]();
    }
    player.phase();
    "step 2";
    event.trigger("phaseOver");
    "step 3";
    // 找下一位玩家 → 循环
    if (!game.players.includes(event.player.next)) {
        event.player = game.findNext(event.player.next);
    } else {
        event.player = event.player.next;
    }
    event.goto(1);  // 回到 step 1，下一位玩家开始回合
}
```

---

## 4. phase()：单个玩家回合入口

```javascript
// content.js 行3638-3852 (简化)
phase: function() {
    "step 0"; event.trigger("phaseBefore");
    "step 1"; game.phaseNumber++; 
              event.phaseList = ["phaseZhunbei","phaseJudge","phaseDraw","phaseUse","phaseDiscard","phaseJieshu"];
              // 轮数检测，触发 roundStart
    "step 2"; event.trigger("phaseBeforeStart");
    "step 3"; event.trigger("phaseBeforeEnd");
    "step 4"; // 翻面检测 → 翻面则 event.cancel() 跳过回合
    "step 5"; // 更新当前回合角色显示
    "step 6"; event.trigger("phaseBeginStart");
    "step 7"; event.trigger("phaseBegin");
    
    // ===== 阶段循环 =====
    "step 8"; if (num < phaseList.length) trigger("phaseChange");
    "step 9"; // 执行当前阶段 player[phaseList[num]]()
    "step 10"; // phaseUse 后清理 ui.tempnowuxie
    "step 11"; if (num < phaseList.length) goto(8);  // 下一阶段
               else trigger("phaseEnd");
    "step 12"; event.trigger("phaseAfter");
    "step 13"; // 清理当前回合角色
}
```

---

## 5. 各阶段实现

```javascript
// 准备阶段 (行3886)
phaseZhunbei: function () {
    event.trigger(event.name);  // 触发 "phaseZhunbei"
}

// 判定阶段 (行3890-3944)
phaseJudge: function () {
    "step 0"; event.cards = player.getVCards("j");  // 获取判定区牌
    "step 1"; // 取出一张 → trigger("phaseJudge") → 无懈检查
    "step 2"; player.judge(event.card);  // 执行判定
    "step 3"; // cancelled → cancel(); 否则 → effect()
    event.goto(1);  // 循环处理下一张
}

// 摸牌阶段 (行3948-3975)
phaseDraw: function () {
    "step 0"; trigger("phaseDrawBegin1");
    "step 1"; trigger("phaseDrawBegin2");
    "step 2"; player.draw(num);  // 摸牌
    // game.modPhaseDraw 可修改摸牌数
}

// 出牌阶段 (行3977-4023)
phaseUse: function () {
    "step 0"; // 重置技能和卡牌使用次数
    "step 1"; trigger("phaseUseBefore");
    "step 2"; trigger("phaseUseBegin");
    "step 3"; player.chooseToUse();  // 玩家操作循环
    "step 4"; if (result.bool) goto(3);  // 继续出牌
    "step 5"; trigger("phaseUseEnd");
    "step 6"; trigger("phaseUseAfter");
}

// 弃牌阶段 (行4025-4042)
phaseDiscard: function () {
    "step 0"; event.num = player.needsToDiscard();  // 计算需弃牌数
              trigger("phaseDiscard");
    "step 1"; player.chooseToDiscard(num, true);  // 弃牌
}

// 结束阶段 (行4043)
phaseJieshu: function () {
    event.trigger(event.name);  // 触发 "phaseJieshu"
}
```

---

## 6. 阶段跳过机制

```javascript
// 跳过某个阶段
player.skip("phaseUse");    // 跳过出牌阶段（如乐不思蜀）
player.skip("phaseDraw");   // 跳过摸牌阶段（如兵粮寸断）

// 系统检查 (gameEvent.js 行1117-1124)
async checkSkipped() {
    if (!this.player || !this.player.skipList.includes(this.name)) 
        return false;
    this.player.skipList.remove(this.name);
    this.finish();  // 标记事件完成，跳过 content
    await this.trigger(this.name + "Skipped");
    return true;
}
```

---

## 7. 轮数管理

```javascript
// content.js 行3650-3676
// 检测是否新的一轮（所有玩家都行动过一次）
if (isRound) {
    game.roundNumber++;              // 轮数+1
    event._roundStart = true;
    event.trigger("roundStart");     // 触发 "roundStart" 事件
    
    // 处理出局玩家的出局计数
    for (var i = 0; i < game.players.length; i++) {
        if (game.players[i].isOut() && game.players[i].outCount > 0) {
            game.players[i].outCount--;
            if (game.players[i].outCount == 0) {
                game.players[i].in();  // 重新加入游戏
            }
        }
    }
}
```

---

## 8. 在你的项目中如何参考

设计你自己的回合系统时，核心要点：

1. **阶段列表是数据**：`["准备","判定","摸牌","出牌","弃牌","结束"]` — 可配置
2. **每个阶段都是独立事件**：有自己的 Before/Begin/End/After 钩子
3. **阶段可被跳过**：通过 skipList 实现，系统自动检查
4. **回合循环**：phaseLoop → phase → 遍历阶段 → 下一玩家
5. **轮数追踪**：通过座位号判断是否完成一轮，触发 roundStart
