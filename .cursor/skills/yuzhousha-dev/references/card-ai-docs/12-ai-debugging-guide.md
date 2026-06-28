# AI 排查指南

> **用途**: 当宇宙杀测试失败时，AI 按本文档流程排查，不要盲目猜测。

---

## 一、测试体系速览

| 测试类型 | 命令 | 用途 | 耗时 |
|---------|------|------|------|
| 编译检查 | `go build ./...` | 语法/类型/导入正确 | 5秒 |
| 冒烟测试 | `./scripts/test.sh smoke -v` | 全武将开局不崩溃 | 30秒 |
| 场景测试 | `./scripts/test.sh yzs -run TestScenario -v` | 特定流程正确性 | 30秒 |
| AI 模拟 1v1 | `CARD_SIM=1 ./scripts/test.sh sim -v` | 全武将两两自对弈 | 2-5分钟 |
| AI 模拟全部 | `CARD_SIM=1 ./scripts/test.sh simall -v` | 全部模式 | 10-30分钟 |

---

## 二、失败类型分类与排查

### 2.1 `go build` 编译失败

**排查步骤**：
1. 读错误信息 → 定位文件和行号
2. 检查是否引用了不存在的函数/常量（新技能 ID 未在 `skill/ids.go` 定义？）
3. 检查是否使用了错误的类型（`skill.Runtime` vs `*gameSkillRuntime`？）
4. 修复 → `go build ./...` 验证

### 2.2 冒烟测试 panic / 开局失败

**常见原因**：
- 新武将数据文件缺失或格式错误
- 新技能 Decl 注册但 `CanActivate` 返回逻辑有问题
- Hook 函数 panic（nil pointer 等）

**排查步骤**：
1. 跑 `./scripts/test.sh smoke -v` 看具体哪个武将在哪个阶段 panic
2. 读 panic 堆栈，定位代码行
3. 最常见：`g.Players[seat]` 越界 → 加边界检查
4. 次常见：`g.Pending` 为 nil 但后续代码假设非 nil → 加 nil 检查

### 2.3 AI 模拟失败：`stuck`（卡住）

**这是最常见的失败类型。**

**根因分类**：
| 子类型 | 典型表现 | 排查入口 |
|--------|---------|---------|
| AI 无法出牌 | 指纹停在 Play 步 | `engine/ai.go` → `runAIPlayPhase` |
| AI 无法响应 | 指纹停在 Response 步 | Pending 窗口未正确关闭 |
| 无限循环 | 步数暴增 | `forceProgress` 未触发 |
| 牌不守恒 | 报告显示牌数不对 | 牌的 gain/lose/discard 有 bug |

**排查步骤**：
1. 读失败报告中的"局面"信息
2. 看 `phase=` 和 `step=` 卡在哪个阶段
3. 看 `Pending` 字段：如果有值但 AI 没响应 → `handlers.ts` 中缺 response_mode 处理
4. 看 `牌数` 是否守恒：不守恒 → 牌的移入/移出有 bug
5. 复现：`CARD_SIM=1 CARD_SIM_TRACE=1 ./scripts/test.sh sim -run TestXxx -v`

### 2.4 AI 模拟失败：`force_error`（AI 动作异常）

**常见原因**：
- AI 选了不合法的牌（filterCard 与 AI 逻辑不一致）
- AI 选了不存在的目标（filterTarget 边界问题）
- Runtime 方法 panic

**排查步骤**：
1. 读错误信息中的具体 error
2. 定位是哪个 Runtime 方法出错
3. 检查该方法的 AI 调用路径

### 2.5 AI 模拟失败：`card_loss`（卡牌不守恒）

**常见原因**：
- 牌被重复移除
- 牌进入弃牌堆但未追加到 `g.DiscardPile`
- 牌从手牌移除但未追加到任何地方

**排查步骤**：
1. `CARD_SIM_STRICT=1 CARD_SIM_TRACE=1` 复现
2. 看最近 25 条事件中涉及牌操作的事件
3. 追踪出问题的牌：从 gain → lose → discard 的完整链路

### 2.6 AI 模拟失败：`timeout`（超时）

**常见原因**：
- 两个 AI 互相不进攻，一直在弃牌/摸牌循环
- AI 选择逻辑导致无限拖延

**排查步骤**：
1. 检查是否两个武将都是防御型（如甄姬 vs 司马懿）
2. 看 AI 的攻击倾向配置是否合理
3. 可临时提高超时步数测试，看最终能否结束

---

## 三、新技能实现后的验证清单

每次实现一个新技能后，按顺序跑：

```bash
# 1. 编译
cd /Users/time/Project/card/backend && go build ./...

# 2. 冒烟（快速验证不崩溃）
./scripts/test.sh smoke -v

# 3. 场景测试（如果写了）
./scripts/test.sh yzs -run TestScenario -v

# 4. AI 模拟（验证不会卡死）
CARD_SIM=1 ./scripts/test.sh sim -run TestSim_AllHeroPairsAIVsAI -v

# 5. 前端构建
cd frontend && npm run build
```

---

## 四、常见 Bug 模式速查

### 4.1 PendingCombat 未清理

```go
// ❌ 错误
g.Pending = &PendingCombat{...}
if condition {
    return ErrXxx  // 忘了 g.Pending = nil！
}

// ✅ 正确
g.Pending = &PendingCombat{...}
if condition {
    g.Pending = nil
    return ErrXxx
}
```

### 4.2 改判后未更新判定牌

改判技（鬼才/鬼道）替换判定牌后，后续 judge 函数必须使用**修改后的牌**：

```go
// ❌ 错误：用原始牌
func completeJudgeResume(..., card Card, ...) error {
    result := buildJudgeResult(judgeSeat, reason, card, judgeFunc)
    // card 可能已被鬼才替换，但这里用的是旧 card
}

// ✅ 正确：用栈顶的牌
func completeJudgeResume(..., card Card, ...) error {
    currentCard := g.getCurrentJudgeCard()  // 获取判定栈顶（可能已改）
    result := buildJudgeResult(judgeSeat, reason, currentCard, judgeFunc)
}
```

### 4.3 装备技能未在卸下时移除

```go
// 装备技能必须标记 TagEquipSkill
skill.Register(skill.Decl{
    Meta: skill.Meta{ID: skill.IDXxx, ...},
    Tags: []skill.SkillTag{skill.TagEquipSkill},  // ← 关键
    CardPlaysAs: func(...) bool { ... },
})
```

### 4.4 伤害来源死亡后仍触发技能

```go
// 卖血技必须检查来源是否存活
if a.Source >= 0 && g.Players[a.Source].HP <= 0 {
    // 来源已死亡，跳过此技能
    g.advanceDamageAftermath(events)
    return nil
}
```

### 4.5 牌操作越界

```go
// 任何涉及 seat 的操作前必须检查
if seat < 0 || seat >= len(g.Players) {
    return ErrInvalidTarget
}
// 取牌前检查
idx, card, ok := g.findCard(seat, cardID)
if !ok {
    return ErrInvalidCard
}
```

### 4.6 多技能同时触发时顺序混乱

```go
// 多个卖血技同时触发时，按优先级排列
// FirstDo: true 的最先（如无懈可击）
// Priority 高的先
// 同优先级按注册顺序
// LastDo: true 的最后
```
