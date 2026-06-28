# 无名杀 AI 参考文档索引

> 本目录包含 12 份拆分后的核心系统文档，每份文档聚焦一个最小场景，可直接让 AI 阅读参考。

---

## 文档列表

| 编号 | 文件 | 内容 | 适用场景 |
|------|------|------|----------|
| 01 | [01-event-lifecycle.md](./01-event-lifecycle.md) | GameEvent 事件生命周期（核心状态机） | 设计事件系统 |
| 02 | [02-trigger-system.md](./02-trigger-system.md) | 技能触发-响应系统 | 设计技能触发框架 |
| 03 | [03-phase-flow.md](./03-phase-flow.md) | 阶段流转（回合制） | 设计回合/阶段系统 |
| 04 | [04-damage-system.md](./04-damage-system.md) | 伤害结算 | 设计伤害/濒死系统 |
| 05 | [05-judge-system.md](./05-judge-system.md) | 判定与改判 | 设计判定/改判系统 |
| 06 | [06-trick-card.md](./06-trick-card.md) | 锦囊牌系统 | 设计锦囊/卡牌使用 |
| 07 | [07-wuxie-system.md](./07-wuxie-system.md) | 无懈可击 | 设计抵消/响应链 |
| 08 | [08-delay-trick.md](./08-delay-trick.md) | 延迟锦囊 | 设计延迟锦囊 |
| 09 | [09-weapon-equip.md](./09-weapon-equip.md) | 武器与装备 | 设计装备系统 |
| 10 | [10-skill-structure.md](./10-skill-structure.md) | **技能实现完全指南** | 实现新技能（必读） |
| 11 | [11-card-distance.md](./11-card-distance.md) | 变牌系统与距离计算 | 设计牌操作/距离 |
| 12 | [12-ai-debugging-guide.md](./12-ai-debugging-guide.md) | **AI 排查指南** | 测试失败时排查（必读） |

---

## 使用方法

### 按场景选择文档

**要移植某个功能时，选择对应的 1-2 份文档**：

| 要移植的功能 | 需要阅读的文档 |
|-------------|---------------|
| 武将的卖血技（刚烈/反馈） | 02 + 04 + 10 |
| 武将的改判技（鬼才） | 02 + 05 + 10 |
| 武将的主动技（青囊/仁德） | 02 + 10 |
| 武将的牌当杀技（武圣/龙胆） | 09 + 10 |
| 新锦囊（过河拆桥等） | 06 + 07 + 11 |
| 延迟锦囊（乐不思蜀等） | 05 + 08 + 03 |
| 武器（丈八/贯石斧等） | 09 + 10 |
| 防具（八卦阵/藤甲等） | 09 + 10 |
| 回合/阶段系统重构 | 01 + 03 |
| 伤害结算重构 | 01 + 04 |
| 判定系统重构 | 01 + 05 |

### 对话模板

```
请阅读 card-ai-docs/04-damage-system.md 和 card-ai-docs/10-skill-structure.md，
参考无名杀的伤害结算和技能框架，在我的项目中实现"刚烈"技能。
```

---

## 源码快速定位

| 源码 | 路径 |
|------|------|
| 事件系统 | `noname/library/element/gameEvent.js` |
| Content函数 | `noname/library/element/content.js` |
| Player类 | `noname/library/element/player.js` |
| 标准卡牌 | `card/standard.js` |
| 标准技能 | `character/standard/skill.js` |
