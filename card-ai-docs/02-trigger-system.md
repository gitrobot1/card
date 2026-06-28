# 02 - 技能触发-响应系统 (Trigger)

> **用途**: 了解无名杀技能如何声明触发时机、如何按优先级收集和排序执行。
>
> **源码文件**: `noname/library/element/gameEvent.js` 行767-907 (trigger方法), `noname/library/element/content.js` 行3163-3219 (arrangeTrigger)

---

## 1. 技能触发机制总览

当事件调用 `event.trigger(eventName)` 时：
1. 按**座位顺序**遍历所有玩家（从当前回合玩家开始）
2. 查找每个玩家是否拥有监听该事件名的技能
3. 将匹配的技能按**优先级**排序
4. 依次执行技能

---

## 2. 技能声明的四种角色维度

```javascript
// 技能声明中的 trigger 字段
trigger: {
    player: "damageEnd",      // 自己是事件主体时触发（如自己受伤结束）
    source: "damageSource",   // 自己是事件来源时触发（如自己造成伤害）
    target: "useCardToTarget",// 自己是事件目标时触发（如被指定为卡牌目标）
    global: "judge",          // 全局触发（任何人发生事件都触发，如改判技）
}
```

| 角色 | 含义 | 典型技能 |
|------|------|----------|
| `player` | 事件主体 | 刚烈 (`damageEnd`) — 自己受伤 |
| `source` | 事件来源 | 狂暴 (`damageSource`) — 自己造成伤害 |
| `target` | 事件目标 | 流离 (`useCardToTarget`) — 被指定为目标 |
| `global` | 全局监听 | 鬼才 (`judge`) — 任何人判定 |

---

## 3. trigger() 核心代码

```javascript
// gameEvent.js 行767-907 (关键部分)
trigger(name) {
    if (!lib.hookmap[name] && !lib.config.compatiblemode) return;
    
    // 确定起始玩家
    let start = [_status.currentPhase, event.source, event.player, game.me, game.players[0]]
        .find(i => get.itemtype(i) == "player");
    
    // 四个角色维度
    const roles = ["player", "source", "target", "global"];
    const playerMap = game.players.concat(game.dead).sortBySeat(start);
    
    // 按座位遍历玩家
    let player = start;
    do {
        const doing = {
            player: player,
            todoList: [],    // 待执行的技能列表
            doneList: [],    // 已执行的技能列表
            addList(skill) {
                // 收集匹配的技能到 todoList
                const info = lib.skill[skill];
                // firstDo → 最优先, lastDo → 最后执行
                const list = info.firstDo ? firstDo.todoList 
                           : info.lastDo ? lastDo.todoList 
                           : this.todoList;
                list.push({
                    skill: skill,
                    player: this.player,
                    priority: get.priority(skill),  // 优先级
                });
                list.sort((a, b) => b.priority - a.priority);
            }
        };
        
        // 按角色维度匹配技能
        roles.forEach(role => {
            doing.addList(lib.hook.globalskill[role + "_" + name]);
            doing.addList(lib.hook[player.playerid + "_" + role + "_" + name]);
        });
        
        doingList.push(doing);
        player = player.nextSeat;
    } while (player && player !== start);
    
    // 创建 arrangeTrigger 事件排序执行
    const next = game.createEvent("arrangeTrigger", false, event);
    next.setContent("arrangeTrigger");
    next.doingList = doingList;
    return next;
}
```

---

## 4. arrangeTrigger：技能排序与执行

```javascript
// content.js 行3163-3219
arrangeTrigger: async function (event, trigger, player) {
    const doingList = event.doingList.slice(0);
    
    while (doingList.length > 0) {
        event.doing = doingList.shift();
        while (true) {
            // 过滤出合法技能（通过 filter 检查）
            const usableSkills = event.doing.todoList.filter(info => {
                return lib.filter.filterTrigger(trigger, info.player, 
                    event.triggername, info.skill, info.indexedData);
            });
            
            if (usableSkills.length == 0) break;
            
            // 只保留最高优先级
            event.doing.todoList = event.doing.todoList.filter(
                i => i.priority <= usableSkills[0].priority
            );
            
            // 同优先级多技能 → 玩家选择顺序
            event.choice = usableSkills.filter(
                n => n.priority == usableSkills[0].priority
            );
            
            if (event.choice.length > 1) {
                // 弹出选择框让玩家决定顺序
                const next = currentPlayer.chooseControl(skillsToChoose);
                const { result } = await next;
                event.current = usableSkills.find(
                    info => info.skill == result.control
                );
            } else {
                event.current = event.choice[0];
            }
            
            // 执行技能
            const result = await game.createTrigger(
                event.triggername, event.current.skill, 
                event.current.player, trigger, event.current.indexedData
            ).forResult();
        }
    }
}
```

---

## 5. 优先级系统

```javascript
// 技能声明
{
    trigger: { player: "damageEnd" },
    priority: 5,       // 数字越大优先级越高（默认0）
    firstDo: true,     // 始终最先执行（如无懈可击）
    lastDo: true,      // 始终最后执行
    silent: true,      // 无描述，同优先级中自动优先执行
}

// 执行顺序:
// 1. firstDo 技能
// 2. 按 priority 降序排列
// 3. 同优先级由玩家选择顺序（或 silent 自动优先）
// 4. lastDo 技能
```

---

## 6. 技能收集的完整流程

```
event.trigger("damage") 被调用
    ↓
1. 从 _status.currentPhase 确定起始玩家
    ↓
2. 按座位顺序遍历每个玩家:
    ├── 检查 player.additionalSkills（附加技能）
    ├── 检查 player.tempSkills（临时技能，自动清理过期的）
    ├── 按 role: ["player","source","target","global"] 匹配
    └── 收集到 doingList
    ↓
3. firstDo / lastDo 分别收集到首尾
    ↓
4. 创建 arrangeTrigger 事件
    ↓
5. 依次执行每个玩家的技能（同优先级可选顺序）
```

---

## 7. 在你的项目中如何参考

设计你自己的技能触发系统时，核心要点：

1. **四角色维度**：`player/source/target/global` 覆盖所有触发场景
2. **座位顺序遍历**：从当前回合玩家开始，按座位顺序收集
3. **优先级排序**：数字越大越先执行，firstDo/lastDo 处理特殊情况
4. **同优先级选择**：让玩家决定发动顺序（如多个卖血技同时触发）
5. **filter 过滤**：每个技能有 filter 函数，运行时判断是否满足条件
