---
name: yuzhousha-dev
description: >
  This skill should be used whenever the user works on the Yuzhousha (宇宙杀) card game
  module within the Card Hub project (Go + Vue 3). It provides authoritative reference to
  the Sanguosha/Noname rule system and precise implementation templates. Trigger when the
  user: adds a new hero/skill/card, fixes a rule bug, refactors an existing skill to match
  Noname behavior, asks "参考无名杀" or "参考noname", mentions specific Noname skills
  (刚烈/反馈/鬼才/武圣/咆哮/马术/奇才/青囊/仁德/结姻/反间/制衡/奸雄/遗计/突袭/裸衣/洛神/铁骑/
  龙胆/集智/空城/观星/英姿/克己/无双/离间/乱武/帷幕/完杀/雷击/鬼道/断粮/激昂/魂姿/破军等),
  discusses game mechanics (判定/改判/伤害结算/阶段流转/无懈可击/延迟锦囊/武器/装备/距离),
  or works on any file under backend/internal/game/yuzhousha/ or frontend/src/**/yuzhousha/.
---

# 宇宙杀开发 Skill

## 概述

此 Skill 用于在 Card Hub 项目 (`/Users/time/Project/card`) 中开发宇宙杀模块。
项目技术栈：后端 Go + Gin，前端 Vue 3 + TypeScript + PixiJS。

**核心原则**：
- 所有游戏规则参考无名杀 (Noname)，但用本项目框架实现
- 不要自行发明新的架构模式，遵循现有框架
- 新增技能必须在 `skill/ids.go` 定义 ID，在 `engine/skill_register_*.go` 注册

---

## 项目结构速查

```
backend/internal/game/yuzhousha/
├── skill/
│   ├── ids.go              # ★ 所有技能 ID 常量
│   ├── types.go            # Decl 结构体、Runtime 接口（不要改）
│   ├── hooks.go            # HookKind 枚举、上下文结构体（不要改）
│   └── catalog_skills.go   # 声明式被动技注册
├── engine/
│   ├── skill_register_wei.go  # ★ 魏国技能注册
│   ├── skill_register_shu.go  # ★ 蜀国技能注册
│   ├── skill_register_wu.go   # ★ 吴国技能注册
│   ├── skill_register_qun.go  # ★ 群雄技能注册
│   ├── skill_ganglie.go       # 刚烈引擎实现
│   ├── skill_fankui.go        # 反馈引擎实现
│   ├── skill_judge.go         # 判定/改判系统
│   ├── skill_hooks.go         # runSkillHooks 统一分发
│   ├── weapons.go             # 武器技能实现
│   └── game.go                # Game 结构体
└── dev-guide.md               # 开发指南

frontend/src/
├── composables/yuzhousha/
│   ├── useYzsGame.ts          # 主游戏状态
│   ├── pending/
│   │   ├── registry.ts        # response_mode → handler 映射
│   │   └── handlers.ts        # ★ 前端技能 UI 处理器
│   └── animation/
│       └── handlers.ts        # 动画处理器
├── types/yuzhousha.ts         # 类型定义
└── components/yuzhousha/      # UI 组件
```

---

## 工作模式

此 Skill 支持两种工作模式：

### 模式 A：新增技能/武将
用户要求添加全新内容时，遵循下面的"实现新技能的流程"。

### 模式 B：参考无名杀重构
用户说"参考无名杀的 XX 技能重构"时：
1. 先读取 `references/card-ai-docs/` 中对应的无名杀文档
2. 定位无名杀源码中的具体实现（文档中有关键行号）
3. 对照本项目现有框架，找出差异
4. 用本项目的 API 重写，不引入无名杀的特有模式
5. 修改完成后，**必须编写行为验证测试**（见下方"行为验证测试模板"），然后跑 `go build ./...` 验证

---

## 行为验证测试（关键！编译通过 ≠ 功能正确）

**⚠️ 这是解决"AI 说完成了但实际不能用"的核心机制。**

冒烟测试只验证"不崩溃"，AI 模拟只验证"不卡死"。两者都不验证"效果是否正确"。
因此每完成一个功能，**必须写行为验证测试**——精确控制输入、逐步推进、断言输出。

### 测试模板

```go
func TestXxx(t *testing.T) {
    // 1. 创建指定武将的对局
    g, err := engine.NewSolo1v1("test-id", "玩家", engine.CharXxx, engine.CharYyy)
    if err != nil { t.Fatal(err) }

    // 2. 精确设置手牌、装备、牌堆
    g.Players[0].Hand = []engine.Card{
        {ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
    }
    g.DrawPile = []engine.Card{
        {ID: "j-1", Suit: "H", Kind: engine.CardSha, Name: "杀", Label: "红桃A"},
    }

    // 3. 设置游戏阶段
    g.Phase = engine.PhasePlaying
    g.TurnStep = engine.StepPlay
    g.CurrentTurn = 0
    g.SyncCounts()

    // 4. 逐步执行操作
    var events []engine.GameEvent
    if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
        t.Fatal(err)
    }

    // 5. 断言中间态（Pending 模式、HP、手牌数等）
    if g.Pending == nil || g.Pending.ResponseMode != "expected_mode" {
        t.Fatalf("expected pending mode expected_mode, got %+v", g.Pending)
    }

    // 6. 继续推进
    if err := g.PassResponse(1, &events); err != nil {
        t.Fatal(err)
    }

    // 7. 断言最终结果
    if g.Players[1].HP != expectedHP {
        t.Fatalf("expected hp=%d, got %d", expectedHP, g.Players[1].HP)
    }
}
```

### 测试必须覆盖的场景

| 场景类型 | 必须验证 | 示例 |
|---------|---------|------|
| **正常流程** | 技能在正确时机触发，效果正确 | 刚烈判定红桃→无效，其他→来源选择 |
| **边界条件** | 技能在条件不满足时不触发 | 来源已死亡→刚烈不触发 |
| **多技能联动** | 多个技能同时触发时顺序正确 | 反馈+刚烈+奸雄同时触发 |
| **入口连通** | 新卡牌的 useCard/playTrick 入口存在 | `grep "CardXxx" engine/play.go` |
| **前端连通** | 新 response_mode 的 handler 存在 | `grep "mode_name" handlers.ts` |
| **文件变更完整性** | 每种功能类型必须修改的文件全部到位 | 参考下方"功能完整性检查清单" |

### 功能完整性检查清单

**这是解决"AI 说做完了但实际缺文件"的核心机制。**

每种功能类型都对应一组必须修改的文件。完成功能后，**必须跑对应 grep 命令**逐项验证。
详见 `references/card-ai-docs/13-checklist.md`。

快速对照：

| 功能类型 | 必须涉及的文件数 | 最易遗漏 |
|---------|----------------|---------|
| 牌当牌技能 | 6 个文件 | `skill_actions.go` 的 toggle 函数 |
| 卖血技 | 5 个文件 | `game.go` 伤害结算后调用 offer |
| 主动技 | 4 个文件 | `skill/types.go` Runtime 方法签名 |
| 锁定技/mod | 2-3 个文件 | 几乎不会遗漏 |
| 锦囊牌 | 6 个文件 | `play.go` useCard switch case |
| 装备牌 | 5 个文件 | `weapons.go` 特效实现 |
| 改判技 | 3 个文件 | `skill_judge.go` collectModifyJudgeSeats |

### 现有测试参考

参考 `backend/test/yuzhousha/skill_test.go` 和 `backend/test/yuzhousha/scenario_test.go` 中的测试写法。

---

## 实现新技能的流程

### 1. 确定技能模式

参考 `references/card-ai-docs/10-skill-structure.md`，技能分为五种模式：

| 模式 | 触发方式 | 典型示例 | 实现难度 |
|------|---------|---------|---------|
| 卖血技 | 受伤后触发 (damageEnd) | 刚烈/反馈/奸雄/遗计 | 中（需要引擎层 PendingCombat） |
| 改判技 | 全局判定介入 (judge) | 鬼才/鬼道 | 高（需要修改 skill_judge.go） |
| 主动技 | 出牌阶段主动使用 (phaseUse) | 青囊/仁德/结姻 | 中（需要 Runtime 方法） |
| 锁定技/牌当牌 | 被动生效 (mod/viewAs) | 武圣/咆哮/马术/奇才 | 低（纯 Decl Hook） |
| 装备技 | 装备时获得 (equipSkill) | 诸葛连弩/丈八蛇矛 | 低（纯 Decl Hook + TagEquipSkill） |

### 2. 添加技能 ID

在 `backend/internal/game/yuzhousha/skill/ids.go` 添加常量：
```go
const IDXxx = "xxx"
```

### 3. 注册技能

在对应王国的 `engine/skill_register_*.go` 中注册：

```go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID: skill.IDXxx, Name: "技能名", Kind: skill.KindPassive,
        Desc: "技能描述",
    },
    CanActivate: xxxCanActivate,
    Activate:    xxxActivate,
    AIPriority:  xxxAIPriority,
    AIActivate:  xxxAIActivate,
})
```

### 4. 如果需要引擎支持

- **卖血技**：需要在引擎层创建 PendingCombat 窗口，实现 offer/apply/pass 三步
- **主动技**：需要在 Runtime 接口添加方法，在 gameSkillRuntime 中实现
- **锁定技/牌当牌**：只需要 Decl Hook 字段，不需要引擎改动

### 5. 如果需要前端 UI

在 `frontend/src/composables/yuzhousha/pending/handlers.ts` 中添加对应 `response_mode` 的处理器。

---

## 强制交付验证流程

**⚠️ 每完成一个功能，必须按以下顺序执行验证，不得跳过任何步骤。**
**禁止在测试通过前声称"已完成"。**

### Step 1: 编译验证
```bash
cd /Users/time/Project/card/backend && go build ./...
cd /Users/time/Project/card/frontend && npm run build
```
- 失败 → 修复后再继续
- 通过 → 进入 Step 2

### Step 2: 新功能的入口连通性验证（最容易被忽略）

**这是最关键的步骤。AI 经常写完了逻辑代码但忘记注册入口。**

| 检查项 | 检查方法 | 常见遗漏 |
|--------|---------|---------|
| 新卡牌/技能 ID 已定义 | `grep "CardXxx\|IDXxx" skill/ids.go` | 常量未定义 |
| 新卡牌的 `useCard` 入口已注册 | `grep "CardXxx" engine/play.go` | `useCard` switch 漏了 case |
| 新技能的 `CanActivate` 能被触发 | 确认引擎在对应时机调用了 `r.PendingXxxFor` | Pending 窗口未创建 |
| 新卡牌在 `playTrickWithCard` 有路由 | `grep "CardXxx" engine/play.go` | 卡牌使用后无处理 |
| 前端 `pendingHandlers` 已注册 | `grep "response_mode_name" frontend/.../handlers.ts` | 后端发了 Pending 但前端无 UI |
| 前端动画已注册 | `grep "event_type" frontend/.../animation/handlers.ts` | 事件无动画 |

**强制命令**（验证入口连通性）：
```bash
# 验证新卡牌从使用到结算的完整链路
grep -rn "CardXxx" backend/internal/game/yuzhousha/engine/ --include="*.go"

# 验证新技能注册
grep -rn "IDXxx" backend/internal/game/yuzhousha/ --include="*.go"

# 验证前端 handler 注册
grep -rn "response_mode_name" frontend/src/composables/yuzhousha/pending/ --include="*.ts"
```

### Step 3: 冒烟测试（验证不崩溃）
```bash
cd /Users/time/Project/card/backend && ./scripts/test.sh smoke -v
```
- 失败 → 读 panic 信息，定位代码行修复
- 通过 → 进入 Step 4

### Step 4: AI 模拟测试（验证不会卡死）
```bash
cd /Users/time/Project/card/backend && CARD_SIM=1 ./scripts/test.sh sim -v
```
- 失败 → 读失败报告，按 `references/card-ai-docs/12-ai-debugging-guide.md` 排查
- 通过 → 进入 Step 5

### Step 5: 全模式 AI 模拟（如果有时间）
```bash
cd /Users/time/Project/card/backend && CARD_SIM=1 ./scripts/test.sh simall -v
```

### 交付声明

完成所有验证后，报告格式：
```
✅ 编译通过
✅ 入口连通性验证通过
✅ 冒烟测试通过
✅ AI 模拟通过（1v1）
⏭️ 全模式模拟（可选）
```

**如果任何一步失败，必须修复后重跑，不得跳过。**
**禁止说"逻辑上应该没问题"，必须跑测试证明。**

### ⚠️ 严禁在 `finishSoloSetup` 或任何初始化函数中留下测试代码

以下行为会导致游戏直接崩坏：
- 在初始化时扣玩家血量（如 `g.Players[i].HP--`）
- 在初始化时跳过回合阶段
- 在初始化时自动执行 AI 操作
- 任何"临时测试用"的代码残留在生产路径中

**修改 `solo_setup.go` 后必须跑 `go test -tags cardtest ./test/yuzhousha/... -run "TestSim_Human|TestFlow_" -count=1 -v` 验证。**

---

## 参考文档

当需要了解无名杀的具体规则实现时，读取 `references/card-ai-docs/` 目录下的对应文档：

| 需要了解的系统 | 读取的文档 |
|---------------|-----------|
| 事件生命周期、状态机 | `references/card-ai-docs/01-event-lifecycle.md` |
| 技能触发-响应 | `references/card-ai-docs/02-trigger-system.md` |
| 阶段流转 | `references/card-ai-docs/03-phase-flow.md` |
| 伤害结算 | `references/card-ai-docs/04-damage-system.md` |
| 判定与改判 | `references/card-ai-docs/05-judge-system.md` |
| 锦囊牌 | `references/card-ai-docs/06-trick-card.md` |
| 无懈可击 | `references/card-ai-docs/07-wuxie-system.md` |
| 延迟锦囊 | `references/card-ai-docs/08-delay-trick.md` |
| 武器与装备 | `references/card-ai-docs/09-weapon-equip.md` |
| **技能实现完全指南** | `references/card-ai-docs/10-skill-structure.md` |
| 变牌与距离 | `references/card-ai-docs/11-card-distance.md` |
| **AI 排查指南** | `references/card-ai-docs/12-ai-debugging-guide.md` |

**注意**：每次只读取当前任务相关的 1-2 份文档，不要全部加载。

---

## 测试失败排查流程

当测试失败时，按以下流程排查，**不要盲目猜测**：

1. **读失败报告**：`./scripts/test.sh sim -v` 会输出详细报告（stuck/force_error/timeout/card_loss/no_winner）
2. **定位失败类型**：参考 `references/card-ai-docs/12-ai-debugging-guide.md` 中"失败类型分类与排查"
3. **复现**：报告中会给出精确复现命令，如 `CARD_SIM=1 CARD_SIM_TRACE=1 ./scripts/test.sh sim -run TestXxx -v`
4. **修复后验证**：`go build ./...` → `./scripts/test.sh smoke -v` → `./scripts/test.sh sim -run TestXxx -v`

### 快速验证命令

```bash
# 编译检查（5秒）
cd /Users/time/Project/card/backend && go build ./...

# ★ 流程冒烟测试（0.5秒）—— 验证核心流程节点通畅
cd /Users/time/Project/card/backend && go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_" -count=1 -v -timeout 30s

# 冒烟测试（30秒）
cd /Users/time/Project/card/backend && ./scripts/test.sh smoke -v

# AI 模拟（2-5分钟）
CARD_SIM=1 ./scripts/test.sh sim -v
```

### 流程冒烟测试覆盖的节点

`backend/test/yuzhousha/smoke_flow_test.go` 覆盖了 11 个核心流程节点：

| 测试 | 验证的流程 |
|------|-----------|
| `TestFlow_PlayShaAndShan` | 杀→闪 基础响应 |
| `TestFlow_PlayGuoHe` | 过河拆桥 锦囊使用 |
| `TestFlow_PlayNanman` | 南蛮入侵 AOE |
| `TestFlow_PlayLebu` | 乐不思蜀 延时锦囊 |
| `TestFlow_EquipWeapon` | 装备武器 |
| `TestFlow_GanglieTrigger` | 刚烈 卖血技触发 |
| `TestFlow_TieqiThenGuicai` | 铁骑→改判 联动 |
| `TestFlow_DyingRescue` | 濒死求桃 |
| `TestFlow_RendeActivate` | 仁德 主动技 |
| `TestFlow_WushengRedAsSha` | 武圣 牌当牌 |
| `TestFlow_FullTurn` | 出牌→结束 回合流转 |

**任何重构后，先跑 `TestFlow_*`，0.5 秒内就能知道核心流程有没有断。**

---

## 关键约束

1. **不要修改框架文件**：`skill/types.go`、`skill/hooks.go`、`engine/skill_hooks.go` 是框架代码
2. **新增技能必须在 ids.go 定义常量**：不要使用字符串字面量作为技能 ID
3. **所有 PendingCombat 窗口必须正确清理**：每个分支都要设置 `g.Pending = nil`
4. **座位号使用前检查范围**：`if seat < 0 || seat >= len(g.Players)` 
5. **伤害来源死亡后不应触发技能**：检查 `g.Players[source].HP <= 0`
6. **牌操作前检查牌是否存在**：使用 `g.findCard(seat, cardID)` 检查返回值

---

## 验收命令

```bash
# 每次修改后必跑
cd /Users/time/Project/card/backend && go build ./...

# 跑测试
cd /Users/time/Project/card/backend && go test ./internal/game/yuzhousha/... -count=1

# 前端构建
cd /Users/time/Project/card/frontend && npm run build
```
