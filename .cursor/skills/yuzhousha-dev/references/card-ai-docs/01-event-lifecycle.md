# 01 - GameEvent 事件生命周期（核心状态机）

> **用途**: 了解无名杀事件驱动的核心状态机。所有游戏行为（出牌、受伤、阶段切换）都是 GameEvent，遵循统一的生命周期。
>
> **源码文件**: `noname/library/element/gameEvent.js`
> **关键行号**: 1-200 (类定义), 1082-1152 (loop方法), 767-907 (trigger方法), 1066-1081 (start方法)

---

## 1. 事件生命周期状态图

```
                    ┌──────────────────┐
                    │   事件被创建       │
                    │ _triggered = 0    │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  checkSkipped()   │  ← 检查 skipList，如乐不思蜀跳过出牌阶段
                    └────────┬─────────┘
                             │ (未被跳过)
                    ┌────────▼─────────┐
                    │ trigger("Before") │  ← 触发 {事件名}Before 技能
                    │ _triggered → 1   │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │ trigger("Begin")  │  ← 触发 {事件名}Begin 技能
                    │ _triggered → 2   │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  content(this)    │  ← 执行事件实际内容（核心逻辑）
                    │  #inContent=true  │
                    └────────┬─────────┘
                             │ (finished=true)
                    ┌────────▼─────────┐
                    │ trigger("End")    │  ← 触发 {事件名}End 技能
                    │ _triggered → 3   │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │ trigger("After")  │  ← 触发 {事件名}After 技能
                    │ _triggered → 4   │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  事件结束/销毁     │
                    └──────────────────┘
```

**如果事件在 Begin 前被 cancel**:
```
Before → trigger("Omitted") → 事件结束 (跳过 End 和 After)
```

---

## 2. 核心代码：事件循环 `loop()`

```javascript
// gameEvent.js 行1082-1115
async loop() {
    const trigger = async (trigger, to) => {
        this._triggered = to;
        if (this.type == "card") await this.trigger("useCardTo" + trigger);
        await this.trigger(this.name + trigger);
    };
    if (await this.checkSkipped()) return;
    while (true) {
        await this.waitNext();
        if (!this.finished) {
            if (this._triggered === 0) await trigger("Before", 1);
            else if (this._triggered === 1) await trigger("Begin", 2);
            else {
                // 执行 content（实际游戏逻辑）
                this.#inContent = true;
                let next = this.content(this);
                await next.finally(() => (this.#inContent = false));
            }
        } else {
            if (this._triggered === 1) await trigger("Omitted", 4);
            else if (this._triggered === 2) await trigger("End", 3);
            else if (this._triggered === 3) await trigger("After", 4);
            else return;
        }
    }
}
```

---

## 3. 事件创建与启动

```javascript
// gameEvent.js 行16-28 (构造函数)
constructor(name = "", trigger = true, manager = _status.eventManager) {
    this.name = name;
    this.manager = manager;
    if (trigger && !game.online) this._triggered = 0;  // 0 = 可以触发 Before
}

// gameEvent.js 行1066-1081 (启动)
start() {
    if (this.#start) return this.#start;
    this.#start = (async () => {
        if (this.parent) this.parent.childEvents.push(this);
        this.manager.eventStack.push(this);        // 入栈
        await this.loop().finally(() => {
            this.manager.eventStack.pop();         // 出栈
        });
    })();
    return this.#start;
}
```

---

## 4. 事件跳过检查

```javascript
// gameEvent.js 行1117-1124
async checkSkipped() {
    if (!this.player || !this.player.skipList.includes(this.name)) return false;
    this.player.skipList.remove(this.name);
    if (lib.phaseName.includes(this.name)) 
        this.player.getHistory("skipped").add(this.name);
    this.finish();                                    // 标记完成
    await this.trigger(this.name + "Skipped");       // 触发跳过事件
    return true;
}
```

---

## 5. 事件属性

```javascript
event.name          // 事件名（如 "phaseUse", "damage", "useCard"）
event.player        // 事件主体玩家
event.source        // 事件来源（如伤害来源）
event.target        // 目标
event.targets       // 多目标数组
event.card          // 关联卡牌
event.cards         // 关联卡牌数组
event.num           // 数值（伤害值、摸牌数等）
event.result        // 事件结果
event.finished      // 是否已完成
event.cancelled     // 是否被取消
event._triggered    // 当前阶段: 0=Before, 1=Begin, 2=Content, 3=End, 4=After, 5=已取消
event.type          // 事件类型: "card" / "player" / "phase"
event.parent        // 父事件
event.childEvents   // 子事件列表
event.skill         // 关联技能
event.forced        // 是否强制
event.notrigger     // 是否跳过触发
event.nature        // 属性 (fire/thunder/poison等)
```

---

## 6. 关键控制方法

```javascript
event.goto(N)     // 跳转到 "step N"（配合 Step 编译器使用）
event.redo()      // 重复当前 step
event.finish()    // 标记完成 → 进入 End → After 流程
event.cancel()    // 取消事件 → 走 Omitted 或直接跳过 End/After
event.trigger(name)  // 触发技能响应（详见 02-trigger-system.md）
event.untrigger(all, player)  // 阻止后续触发
```

---

## 7. 事件嵌套与栈

```javascript
// gameEvent.js 行933 (manager属性)
event.manager           // GameEventManager 实例
event.manager.eventStack  // 事件栈数组，栈顶=当前正在处理的事件
event.manager.rootEvent   // 根事件
event.manager.tempEvent   // 临时事件（高优先级插入）

// 当新事件创建时自动入栈，完成时自动出栈
// parent → child 关系自动维护
```

---

## 8. 在你的项目中如何参考

设计你自己的事件系统时，核心要点：

1. **每个事件有固定生命周期钩子**：Before → Begin → Content → End → After
2. **钩子名自动拼接**：`event.name + "Before"` → 如 `"damageBegin"`
3. **事件栈追踪嵌套**：子事件自动入栈/出栈
4. **跳过机制**：通过 `skipList` 实现阶段跳过（如乐不思蜀）
5. **取消机制**：`cancel()` 在 Begin 前走 Omitted 路径，之后跳过 End/After
