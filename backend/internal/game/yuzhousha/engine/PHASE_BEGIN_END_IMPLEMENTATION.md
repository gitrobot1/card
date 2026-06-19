# 回合阶段完善实现文档

## 概述

本文档描述了为宇宙杀游戏引擎完善回合阶段的工作。按照标准三国杀规则，一个完整的回合应该包含以下阶段：

1. **回合开始阶段** (StepStart)
2. **准备阶段** (StepPrepare)
3. **判定阶段** (StepJudge)
4. **摸牌阶段** (StepDraw)
5. **出牌阶段** (StepPlay)
6. **弃牌阶段** (StepDiscard)
7. **回合结束阶段** (StepFinish)

## 实现内容

### 1. 创建 `phase_begin.go` - 回合开始阶段

**文件位置**: `/Users/time/Project/card/backend/internal/game/yuzhousha/engine/phase_begin.go`

**主要功能**:

- `beginStartPhase(seat int, events *[]GameEvent) error`: 进入回合开始阶段
  - 设置阶段为 `StepStart`
  - 触发 `start_phase` 事件
  - 调用技能钩子（目前简化实现，未触发具体技能）
  - 自动继续到准备阶段

- `continueAfterStart(seat int, events *[]GameEvent) error`: 回合开始阶段结束后，进入准备阶段
  - 检查是否需要进入准备阶段
  - 如果需要，调用 `enterPreparePhase`
  - 如果不需要，直接进入判定阶段

- `PassStart(seat int, events *[]GameEvent) error`: 跳过回合开始阶段
  - 允许玩家主动跳过回合开始阶段
  - 目前简化实现，直接调用 `continueAfterStart`

**设计考虑**:

- 目前简化实现：回合开始阶段不触发具体技能，直接继续到准备阶段
- 未来扩展：可以在 `runTurnStartHooks` 中添加具体技能的触发逻辑（如洛神、励战等）

### 2. 完善 `phase_prepare.go` - 添加回合结束阶段

**文件位置**: `/Users/time/Project/card/backend/internal/game/yuzhousha/engine/phase_prepare.go`

**新增功能**:

- `enterFinishPhase(seat int, events *[]GameEvent) error`: 进入回合结束阶段
  - 设置阶段为 `StepFinish`
  - 触发 `finish_phase` 事件
  - 调用技能钩子（目前简化实现，未触发具体技能）
  - 完成回合结束阶段，进入下一个回合

- `finishTurn(seat int, events *[]GameEvent) error`: 完成回合结束阶段，进入下一个回合
  - 触发回合结束后的清理工作
  - 重置玩家状态（如醉酒状态）
  - 发送回合结束事件
  - 切换到下一个玩家
  - 开始下一个玩家的回合

- `runTurnEndCleanup(seat int, events *[]GameEvent)`: 运行回合结束时的清理工作
  - 清理回合相关的状态
  - 清理技能计数器、重置临时状态等

**修改内容**:

- 修改 `advanceToDiscardPhase` 函数，让它进入回合结束阶段而不是直接结束回合
- 删除了重复的 `runTurnEndHooks` 函数定义（已在 `skill_decl_hooks.go` 中定义）

### 3. 修改 `turn.go` - 完善回合流程

**文件位置**: `/Users/time/Project/card/backend/internal/game/yuzhousha/engine/turn.go`

**修改内容**:

- 修改 `beginTurn` 函数：
  - 原来：直接调用 `enterPreparePhase` 进入准备阶段
  - 现在：调用 `beginStartPhase` 进入回合开始阶段
  - 这样确保了回合按照正确的顺序执行：开始阶段 → 准备阶段 → 判定阶段 → ...

- 修改 `endTurn` 函数：
  - 原来：直接结束回合，切换到下一个玩家
  - 现在：调用 `enterFinishPhase` 进入回合结束阶段
  - 这样确保了回合结束阶段有机会触发技能（如遗计、固政等）

### 4. 修改 `constants.go` - 添加阶段常量

**文件位置**: `/Users/time/Project/card/backend/internal/game/yuzhousha/engine/constants.go`

**修改内容**:

- 添加 `StepFinish = "finish"` 常量，表示回合结束阶段

### 5. 创建测试文件 `phase_begin_end_test.go`

**文件位置**: `/Users/time/Project/card/backend/internal/game/yuzhousha/engine/phase_begin_end_test.go`

**测试内容**:

- `TestStartPhase`: 测试回合开始阶段
  - 验证 `start_phase` 事件是否正确触发
  - 验证阶段转换是否正确

- `TestFinishPhase`: 测试回合结束阶段
  - 验证 `finish_phase` 事件是否正确触发
  - 验证回合结束后是否正确切换到下一个玩家

- `TestFullTurn`: 测试完整的回合流程
  - 验证回合是否按照正确的顺序执行各个阶段
  - 验证从开始到结束的整个流程是否正确

- `TestTurnSteps`: 测试所有阶段的常量定义
  - 验证所有阶段常量是否都已正确定义

## 技术细节

### 阶段转换流程

```
beginTurn
  ↓
beginStartPhase (StepStart)
  ↓
continueAfterStart
  ↓
enterPreparePhase (StepPrepare)
  ↓
continueAfterPrepare
  ↓
enterJudgePhase (StepJudge)
  ↓
advanceToDrawPhase (StepDraw)
  ↓
advanceToPlayPhase (StepPlay)
  ↓
EndPlay / finishPlayWithKejiOrDiscard
  ↓
advanceToDiscardPhase (StepDiscard)
  ↓
enterFinishPhase (StepFinish)
  ↓
finishTurn
  ↓
beginTurn (下一个玩家的回合)
```

### 技能钩子

- **回合开始阶段**: `runTurnStartHooks(seat, events)`
  - 目前简化实现，未触发具体技能
  - 未来可以添加：洛神（甄宓）、励战（SP关羽）等

- **回合结束阶段**: `runTurnEndHooks(seat, events)`
  - 已在 `skill_decl_hooks.go` 中实现
  - 可以触发：遗计（郭嘉）、固政（张昭张纮）等

- **回合结束清理**: `runTurnEndCleanup(seat, events)`
  - 目前简化实现，未触发具体技能
  - 未来可以添加清理逻辑

## 测试验证

所有测试均已通过：

```bash
cd /Users/time/Project/card/backend && go test ./internal/game/yuzhousha/engine/...
```

**测试结果**:
- `TestStartPhase`: PASS
- `TestFinishPhase`: PASS
- `TestFullTurn`: PASS
- `TestTurnSteps`: PASS
- 所有现有测试: PASS

## 未来扩展

### 1. 添加回合开始阶段的技能触发

在 `runTurnStartHooks` 中添加具体技能的触发逻辑：

```go
func (g *Game) runTurnStartHooks(seat int, events *[]GameEvent) {
	// 洛神（甄宓）
	// 励战（SP关羽）
	// 等等...
}
```

### 2. 添加回合结束阶段的技能触发

在 `runTurnEndHooks` 中完善技能触发逻辑（已在 `skill_decl_hooks.go` 中实现）：

```go
func (g *Game) runTurnEndHooks(seat int, events *[]GameEvent) {
	rt := g.skillRuntime(events)
	for _, h := range g.playerSkillHandlers(seat) {
		if err := h.OnTurnEnd(rt, seat); err != nil {
			return
		}
	}
}
```

### 3. 完善玩家交互

目前回合开始阶段和回合结束阶段都是自动完成的，未来可以添加玩家交互：

- 回合开始阶段：允许玩家触发技能（如洛神）
- 回合结束阶段：允许玩家触发技能（如遗计）

需要实现相应的响应模式和Pending管理。

## 总结

本次工作成功完善了宇宙杀游戏引擎的回合阶段，实现了标准的回合流程：

1. ✅ 创建独立的回合开始阶段 (`phase_begin.go`)
2. ✅ 创建独立的回合结束阶段 (`phase_prepare.go`)
3. ✅ 修改回合流程，确保按照正确顺序执行各个阶段 (`turn.go`)
4. ✅ 添加阶段常量定义 (`constants.go`)
5. ✅ 创建测试文件，验证实现的正确性
6. ✅ 所有测试通过，确保修改没有破坏现有功能

未来可以根据需要添加更多技能触发逻辑和玩家交互功能。
