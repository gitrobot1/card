# 宇宙杀 · 交互窗口与 Pending 语义规范（API 草案 v0.1）

> **状态**：P0–P3 已实现（TakeWindow + DiscardWindow + 反馈/突袭/奇袭/破军）。P4 起待做。
> **读者**：后续 AI / 人类开发者。  
> **目的**：统一「从目标处取/弃牌」与「谁该操作 pending」两类重复逻辑，降低界徐盛类技能接入成本。  
> **关联文档**：`[dev-guide.md](./dev-guide.md)`、`[skill/doc.go](../../skill/doc.go)`

---

## 0. 文档约定（AI 必读）

### 0.1 优先级

1. 本文档与 `dev-guide.md` 冲突时，**交互窗口 / Pending 语义**以本文档为准。
2. 未覆盖场景仍走 `dev-guide.md` 决策树。
3. 实现时必须**先迁移再删旧代码**，禁止双轨长期并存。

### 0.2 术语


| 术语                | 含义                                                                               |
| ----------------- | -------------------------------------------------------------------------------- |
| **Actor**         | 当前必须做出 UI/API 操作的座位（谁点按钮）                                                        |
| **Subject**       | 被操作的角色座位（谁的牌被拿/弃/选）                                                              |
| **Origin**        | 事件来源座位（如杀的使用者、伤害来源），用于 resume / 伤害链                                              |
| **TakeWindow**    | 通用「从 Subject 的若干 Zone 取牌 → 放入 Destination」的 pending 状态机                          |
| **DiscardWindow** | 通用「从 Actor 自身某 Zone 弃牌」的 pending 状态机（破军弃「营」等）                                    |
| **Zone**          | 牌区：`hand` / `weapon` / `armor` / `plus_horse` / `minus_horse` / `judge` / `camp` |


### 0.3 实现阶段（禁止跳步）


| Phase       | ID                  | 交付物                                       | 验收                                    |
| ----------- | ------------------- | ----------------------------------------- | ------------------------------------- |
| P0-第一个ai已实现 | `pending-semantics` | Actor/Subject 字段 + 推导函数                   | 联机 1v1 超时、前端 `isMyResponse` 无 mode 特例 |
| P1-done          | `take-window-core`  | TakeWindow 引擎 + 迁移反馈/突袭/奇袭                | 原 scenario 全绿                         |
| P2-done          | `take-window-pojun` | 破军迁入 TakeWindow（dest=camp）                | 破军 + 古锭刀/藤甲 scenario                  |
| P3-done          | `discard-window`    | DiscardWindow + 破军回合末弃营                   | 同上                                    |
| P4          | `frontend-template` | `makeTakeWindowHandler` + 删 useYzsGame 特例 | `npm run build`                       |
| P5          | `cleanup`           | 删除旧 ResponseMode 重复逻辑                     | grep 无 `SourceIndex == seat` 技能特例     |


---

## 1. 问题陈述

### 1.1 现状痛点

1. `**PendingCombat.SourceIndex` / `TargetIndex` 语义过载**
  - 突袭：`SourceIndex` = 被拿牌的人，`TargetIndex` = 发动者  
  - 破军：`SourceIndex` = 杀使用者，`TargetIndex` = 被杀者  
  - 默认出闪：`TargetIndex` = 出闪者
2. **取牌逻辑复制**
  - `skill_fankui.go`、`skill_tuxi.go`、`skill_qixi.go`、`skill_pojun.go` 各自维护 pending、Pass、AI 循环、`takeTargetCard` 调用。
3. **前端重复**
  - 每个 mode 在 `handlers.ts` 写一套；`useYzsGame.ts` 的 `isMyResponse`、`showSeatSkillPanels` 需 per-mode 补丁。
4. **临时补丁**
  - `engine/pending_actor.go` 的 `PendingActorSeat()` 用 switch 补语义，但未写入 JSON，前端不可见。

### 1.2 目标

- 新「拿牌/弃牌窗口」类技能：**≤ 3 个后端文件触点**（技能触发 + Register + 测试），不再改 `play.go` / `response.go` / `ai.go` 多处。  
- 前端：**1 个 TakeWindow 模板 handler**，仅配置 `pickFromSeat` / `skillId` / `hint`。  
- 联机 / 单机 / AI 共用同一套 Actor 推导。

### 1.3 非目标（v0.1 不做）

- Sha Pipeline 全量 hook 表（另开 `dev-sha-pipeline.md` 预留）。  
- 观星 / 五谷 / 刚烈选弃 等非「从指定 Zone 取/弃」的复杂 UI。  
- 将 `UseSkill` 大 switch 一次性改为表驱动（可 P5 后迭代）。

---

## 2. Pending 语义模型

### 2.1 新字段（`PendingCombat` 扩展）

在 `engine/model.go` 的 `PendingCombat` 增加**显式语义字段**（JSON 同步给前端）：

```go
// PendingCombat 扩展字段（v0.1）
type PendingCombat struct {
    // ... 现有字段保留 ...

    // 语义字段（v0.1 新增，优先于 SourceIndex/TargetIndex 推导）
    ActorSeat   int    `json:"actor_seat"`             // 当前应操作的座位；-1 表示无
    SubjectSeat int    `json:"subject_seat,omitempty"` // 被操作座位（拿牌/选目标区）
    OriginSeat  int    `json:"origin_seat,omitempty"`  // 事件来源（resume/伤害链）

    // 窗口类型（与 ResponseMode 并存，便于前端模板路由）
    WindowKind  string `json:"window_kind,omitempty"`  // "" | "take" | "discard" | "choice" | "respond"
}
```

**兼容策略（P0）**：

- 旧 `response_mode` **保留**，不 breaking。  
- 新开窗口时**必须**填 `ActorSeat` / `SubjectSeat` / `WindowKind`。  
- `PendingActorSeat()` 改为：`if p.ActorSeat >= 0 { return p.ActorSeat }` 再 fallback 旧 switch（迁移完成后删除 fallback）。

### 2.2 推导规则表（WindowKind → 默认 Actor/Subject）


| WindowKind | ActorSeat | SubjectSeat | 典型 response_mode                                          |
| ---------- | --------- | ----------- | --------------------------------------------------------- |
| `respond`  | 响应者       | 响应者         | 出闪/出桃/无懈（RequiredKind 有值）                                 |
| `take`     | 拿牌发动者     | 被拿牌角色       | `skill_fankui`, `skill_tuxi`, `skill_qixi`, `skill_pojun` |
| `discard`  | 弃牌者       | 弃牌者         | `skill_pojun_discard`, `skill_yinghun_discard`            |
| `choice`   | 选择者       | 选择者         | `skill_ganglie_choice`, `skill_yinghun`                   |
| `peek`     | 看牌者       | 看牌者         | `peek_deck`                                               |
| `pick`     | 选牌者       | 选牌者         | `wugu_pick`（Subject 可为共享牌堆，seat=-1 特例）                    |


**特殊：TieqiPending**  

- `ActorSeat = SourceIndex`（杀使用者决定是否发动铁骑）  
- `WindowKind = "choice"` 或保留 `TieqiPending` 布尔直到迁移完成

**特殊：Dying**  

- `ActorSeat = SourceIndex`（当前 askSeat，询问谁出桃）  
- `SubjectSeat = TargetIndex`（濒死角色）

### 2.3 引擎 API（P0 必做）

```go
// engine/pending_semantics.go

// FillPendingRoles 在创建 Pending 后调用，按 WindowKind 与 mode 填充 Actor/Subject/Origin。
func FillPendingRoles(p *PendingCombat)

// PendingActorSeat 返回当前应操作的人类/AI 座位；-1 无 pending。
func (g *Game) PendingActorSeat() int

// PendingSubjectSeat 返回被操作座位；-1 不适用。
func (g *Game) PendingSubjectSeat() int

// IsActorSeat 判断 seat 是否为当前 Actor（用于 ValidateAction / 前端 human_player 对齐）
func (g *Game) IsActorSeat(seat int) bool
```

**超时 / AI**：

- `IsHumanPending()` → `PendingActorSeat()` 对应玩家 `!IsAI`  
- `ApplyHumanTimeout()` → 对 `PendingActorSeat()` 执行 Pass（窗口 Pass 见 §3.4）  
- **禁止**再读 `g.HumanPlayer` 决定 pending（`HumanPlayer` 仅保留为 solo 视角默认值，逐步废弃）

### 2.4 PublicState JSON（前端）

`PublicViewForSeat` 输出：

```json
{
  "pending": {
    "response_mode": "skill_pojun",
    "window_kind": "take",
    "actor_seat": 0,
    "subject_seat": 1,
    "origin_seat": 0,
    "source_index": 0,
    "target_index": 1
  }
}
```

**前端规则（P4）**：

```typescript
// 统一替代 isMyResponse 中的 mode 特例
function isMyPendingActor(state: YuzhoushaState, mySeat: number): boolean {
  const p = state.pending
  if (!p || state.phase !== 'response') return false
  if (p.actor_seat != null && p.actor_seat >= 0) return p.actor_seat === mySeat
  // fallback 旧逻辑（迁移期）
  return p.target_index === mySeat || p.source_index === mySeat
}

function pickFromSeat(state: YuzhoushaState): number {
  const p = state.pending!
  if (p.subject_seat != null && p.subject_seat >= 0) return p.subject_seat
  // fallback: tuxi 反转语义
  if (p.response_mode === 'skill_tuxi') return p.source_index ?? 0
  if (p.response_mode === 'skill_pojun') return p.target_index ?? 0
  return p.source_index ?? p.target_index ?? 0
}
```

---

## 3. TakeWindow API（后端）

### 3.1 配置结构

```go
// engine/take_window.go

type ZoneID string

const (
    ZoneHand       ZoneID = "hand"
    ZoneWeapon     ZoneID = "weapon"
    ZoneArmor      ZoneID = "armor"
    ZonePlusHorse  ZoneID = "plus_horse"
    ZoneMinusHorse ZoneID = "minus_horse"
    ZoneJudge      ZoneID = "judge"
    ZoneCamp       ZoneID = "camp"
)

type TakeDestination struct {
    Zone   ZoneID // hand | camp | discard | void
    Seat   int    // 目标座位（通常 ActorSeat）
}

type TakeWindowConfig struct {
    // 身份
    SkillID     string
    ResponseMode string // 现有 JSON 值，如 "skill_fankui"

    ActorSeat   int
    SubjectSeat int
    OriginSeat  int

    // 窗口行为
    MaxTake     int      // 最多取几张；0 = 不限直到 Pass
    MinTake     int      // 少于此不能 Pass（默认 0）
    AllowedZones []ZoneID // 空 = 全部可拿区

    Destination TakeDestination

    // 回调（可选）
    OnEachTake  func(g *Game, card Card, events *[]GameEvent) error
    OnComplete  func(g *Game, events *[]GameEvent) error // 全部完成或 Pass 后

    // UI / 事件
    Message     string
    EventType   string // 默认 "take_window"
}
```

### 3.2 生命周期

```text
OpenTakeWindow(cfg) → PhaseResponse + Pending(WindowKind=take)
  ↓
TakeOne(actor, zone, cardID)  // HTTP: UseSkill + target_zone/target_card_id
  ↓ 重复直到 MaxTake 或无牌可拿
PassTake(actor)               // HTTP: PassResponse 或 UseSkill 空提交
  ↓
OnComplete → 恢复 SavedResume（若有）或进入下一 pending
```

### 3.3 引擎方法

```go
// 开启窗口；写入 Pending + FillPendingRoles；返回 error 若当前不可开窗
func (g *Game) OpenTakeWindow(cfg TakeWindowConfig, resume *PendingResume, events *[]GameEvent) error

// 取一张；actor 必须 == Pending.ActorSeat
func (g *Game) TakeOne(actor int, zone ZoneID, cardID string, events *[]GameEvent) error

// 结束拿牌（可未满 MaxTake）
func (g *Game) PassTake(actor int, events *[]GameEvent) error

// AI：尽量拿满 MaxTake 或直到无牌
func (g *Game) AutoTakeWindow(actor int, events *[]GameEvent)

// 查询 Subject 在允许 Zone 内是否还有牌
func (g *Game) HasTakeableInWindow(subject int, zones []ZoneID) bool
```

**内部复用**：

- 底层仍调用现有 `takeTargetCard(subject, PlayTarget{...})` + 新 `placeCard(dest)`。  
- `placeCard` P1 可简化为：`dest.Zone == hand` → append Hand；`camp` → CampCards；`discard` → DiscardPile。

### 3.4 Pass 路由（`response.go` 收敛）

```go
// PassResponse 内部（伪代码）
switch g.Pending.WindowKind {
case "take":
    if seat != g.Pending.ActorSeat { return ErrNotYourTurn }
    return g.PassTake(seat, events)
case "discard":
    return g.PassDiscardWindow(seat, events) // §4
default:
    // 现有逻辑
}
```

### 3.5 UseSkill 路由（`skill_runtime.go` 收敛）

```go
case "take": // WindowKind
    return g.TakeOne(seat, req.TargetZone, req.TargetCardID, events)
```

技能 Register 的 `CanActivate`：

```go
func takeWindowCanActivate(r skill.Runtime, seat int, skillID string) bool {
    return r.PendingWindowKind() == "take" &&
        r.PendingActorSeat() == seat &&
        r.HasSkill(seat, skillID)
}
```

### 3.6 PendingResume（嵌套窗口）

杀链中插入 TakeWindow 时必须保存 resume：

```go
type PendingResume struct {
    ReturnIndex int
    Saved       *PendingCombat // 或 SavedSha *ShaPendingSnapshot（后续 Sha Pipeline 统一）
    AfterTake   func(g *Game, events *[]GameEvent) error
}
```

**破军示例**：

```go
OpenTakeWindow(TakeWindowConfig{
    SkillID: "pojun", ResponseMode: ResponseModeSkillPojun,
    ActorSeat: source, SubjectSeat: target, OriginSeat: source,
    MaxTake: victimHP, AllowedZones: allTakeable,
    Destination: {Zone: ZoneCamp, Seat: target},
    OnComplete: func(g, ev) { return g.advanceShaBeforeTargetResponse(ev) },
}, shaResume, events)
```

---

## 4. DiscardWindow API（后端）

### 4.1 配置

```go
type DiscardWindowConfig struct {
    SkillID      string
    ResponseMode string
    ActorSeat    int
    SourceZone   ZoneID   // camp | hand
    MinDiscard   int
    MaxDiscard   int      // 通常 == MinDiscard
    Message      string
    OnComplete   func(g *Game, events *[]GameEvent) error
}
```

### 4.2 方法

```go
func (g *Game) OpenDiscardWindow(cfg DiscardWindowConfig, events *[]GameEvent) error
func (g *Game) DiscardOne(actor int, cardID string, events *[]GameEvent) error
func (g *Game) AutoDiscardWindow(actor int, events *[]GameEvent)
```

**破军回合末**：`OpenDiscardWindow` + `OnComplete → endTurn hooks`。

---

## 5. 现有技能迁移映射


| 技能   | 现文件             | Window  | Actor | Subject   | Dest         | Max             | Phase      |
| ---- | --------------- | ------- | ----- | --------- | ------------ | --------------- | ---------- |
| 反馈   | skill_fankui.go | take    | 受伤者   | 伤害来源      | actor/hand   | FankuiRemaining | 伤害后        |
| 突袭   | skill_tuxi.go   | take    | 发动者   | 对手        | actor/hand   | TuxiRemaining   | 摸牌阶段       |
| 奇袭   | skill_qixi.go   | take    | 发动者   | 对手        | actor/hand   | 1               | 出牌阶段       |
| 破军   | skill_pojun.go  | take    | 杀来源   | 杀目标       | subject/camp | 目标体力            | 杀链         |
| 破军弃营 | skill_pojun.go  | discard | 受害者   | self/camp | discard      | counter         | 回合末        |
| 麒麟弓  | weapons.go      | take    | 杀来源   | 目标        | discard      | 1               | 杀后（可选 P2+） |


**迁移顺序**：反馈 → 突袭 → 奇袭 → 破军 → 破军弃营。

每个迁移 **必须**：

1. 原 `response_mode` 字符串不变。
2. 原 scenario 测试不改断言，只改 setup 若必要。
3. 删除旧函数前先 `@deprecated` 注释一个 commit。

---

## 6. skill.Runtime 接口变更（P1）

在 `skill/types.go` 增加（替代 per-skill PassPojun / TuxiTakeFrom 膨胀）：

```go
type Runtime interface {
    // ... 现有 ...

    PendingWindowKind() string
    PendingActorSeat() int
    PendingSubjectSeat() int
    PendingOriginSeat() int

    TakeOne(actor int, zone, cardID string) error
    PassTake(actor int) error
    DiscardWindowOne(actor int, cardID string) error
}
```

**废弃路径**（P5 删除）：

- `FankuiTakeFrom`, `TuxiTakeFrom`, `PojunPlace`, `PassPojun`, `PojunDiscardCamp` 等改为 thin wrapper 调 `TakeOne` / `PassTake` / `DiscardOne`。

---

## 7. 前端 Pending 注册表（P4）

### 7.1 扩展 `PendingHandler`

```typescript
// pending/types.ts 扩展
export interface PendingHandler {
  // ... 现有 ...
  windowKind?: 'take' | 'discard' | 'choice' | 'respond' | 'peek' | 'pick'
  actorSeat?: (state: YuzhoushaState) => number
  subjectSeat?: (state: YuzhoushaState) => number
}
```

### 7.2 工厂函数

```typescript
// pending/templates/takeWindow.ts

export function makeTakeWindowHandler(opts: {
  modes: string[]
  skillId: string
  hint?: string | ((ctx: PendingContext) => string)
  zones?: string[] // 限制可选区；默认全部
}): PendingHandler
```

**注册示例**：

```typescript
const fankuiHandler = makeTakeWindowHandler({
  modes: ['skill_fankui'],
  skillId: 'fankui',
  hint: '【反馈】：选择来源的一张牌',
})
```

### 7.3 待删除的前端特例（P4 完成后 grep）

- `useYzsGame.ts`：`skill_pojun` 的 `isMyResponse` 分支  
- `useYzsGame.ts`：`fankuiSourceSeat` / `tuxiSourceSeat` / `pojunVictimSeat` 若可由 `subject_seat` 替代  
- `pending/registry.ts`：`TARGET_PICK_MODES` 硬编码 set → 改为 `window_kind === 'take'`

---

## 8. 文件落点（实现 checklist）

### P0 — Pending 语义


| 操作                                    | 文件                                                      |
| ------------------------------------- | ------------------------------------------------------- |
| 加字段                                   | `engine/model.go`                                       |
| FillPendingRoles / PendingSubjectSeat | `engine/pending_semantics.go`（新）                        |
| 改 PendingActorSeat                    | `engine/pending_actor.go` → 合并进 pending_semantics.go    |
| 改 IsHumanPending / ApplyHumanTimeout  | `engine/game.go`                                        |
| PublicView 输出 actor/subject           | `engine/game.go`                                        |
| 前端类型                                  | `frontend/src/types/yuzhousha.ts`                       |
| 前端 isMyPendingActor                   | `frontend/src/composables/yuzhousha/pending/helpers.ts` |


### P1 — TakeWindow 核心


| 操作               | 文件                                                  |
| ---------------- | --------------------------------------------------- |
| TakeWindow 实现    | `engine/take_window.go`（新）                          |
| placeCard 辅助     | `engine/card_zones.go`（新，或扩 model.go）               |
| Pass/UseSkill 路由 | `engine/response.go`, `engine/skill_runtime.go`     |
| AI               | `engine/ai.go` → `autoTakeWindowIfNeeded`           |
| 迁移反馈             | `engine/skill_fankui.go`                            |
| 迁移突袭/奇袭          | `engine/skill_tuxi.go`, `engine/skill_qixi.go`      |
| Runtime          | `skill/types.go`, `engine/skill_tieqi.go` 等 adapter |


### P2–P3 — 破军


| 操作            | 文件                            |
| ------------- | ----------------------------- |
| 迁移拿牌          | `engine/skill_pojun.go`       |
| DiscardWindow | `engine/discard_window.go`（新） |
| 回合末           | `engine/turn.go`              |


### P4 — 前端模板


| 操作                    | 文件                                             |
| --------------------- | ---------------------------------------------- |
| makeTakeWindowHandler | `frontend/.../pending/templates/takeWindow.ts` |
| 迁移 handlers           | `frontend/.../pending/handlers.ts`             |


---

## 9. 测试要求

### 9.1 每个 Phase 必跑

```bash
cd backend && ./scripts/test.sh smoke -v
cd backend && ./scripts/test.sh yzs -run TestScenario -v
cd frontend && npm run build
```

### 9.2 新增测试（P1 起）


| 测试 ID                                | 内容                                      |
| ------------------------------------ | --------------------------------------- |
| `TestTakeWindow_FankuiTakeOne`       | 开窗 → TakeOne hand → 牌进 actor 手          |
| `TestTakeWindow_PassEarly`           | MaxTake=3 但 Pass 只拿 1 张                 |
| `TestDiscardWindow_PojunCampOne`     | 开窗 → DiscardOne → 牌进弃牌堆                  |
| `TestDiscardWindow_AISweeps`         | AI 弃满 required 张                           |
| `TestPendingSemantics_PojunActor`    | Actor=source, Subject=target            |
| `TestPendingSemantics_TuxiActor`     | 迁移后 Actor=seat0, Subject=opponent       |
| `TestPendingSemantics_OnlineTimeout` | 双真人 seat1 pending 时 seat0 tick 不代为 pass |


### 9.3 禁止回归

- 联机 WS `broadcastGame` 行为不变。  
- 现有 `response_mode` JSON 值不变。

---

## 10. 工作量估算（人日，供排期）


| Phase                  | 后端   | 前端   | 测试   | 合计         |
| ---------------------- | ---- | ---- | ---- | ---------- |
| P0 Pending 语义          | 1d   | 0.5d | 0.5d | **2d**     |
| P1 TakeWindow + 3 技能迁移 | 2d   | —    | 1d   | **3d**     |
| P2 破军 take             | 0.5d | —    | 0.5d | **1d**     |
| P3 DiscardWindow       | 1d   | —    | 0.5d | **1.5d**   |
| P4 前端模板                | —    | 1.5d | 0.5d | **2d**     |
| P5 清理废弃                | 1d   | 0.5d | 0.5d | **2d**     |
| **总计**                 |      |      |      | **~11.5d** |


> 若 P0+P1 完成，新「拿牌窗」类技能接入可从 **~3 天降至 ~0.5 天**。

---

## 11. AI 实施指令模板（复制即用）

```text
任务：实现宇宙杀 dev-interaction-window.md Phase P{n}

约束：
1. 严格按 Phase 顺序，不得跳过 P0
2. response_mode 字符串不得改
3. 迁移一个技能一组 commit：反馈 / 突袭 / 奇袭 / 破军
4. 每个 commit 跑：./scripts/test.sh yzs -run TestScenario -v
5. 前端改动仅在 P4 开始
6. 完成后更新本文档 Phase 表格状态为 done

验收：
- [ ] PendingCombat 含 actor_seat/subject_seat/window_kind
- [ ] PendingActorSeat 优先读 ActorSeat
- [ ] TakeWindow 单测通过
- [ ] 迁移技能 scenario 全绿
- [ ] npm run build 通过
```

---

## 12. 版本历史


| 版本   | 日期         | 说明                                    |
| ---- | ---------- | ------------------------------------- |
| v0.1 | 2026-06-04 | 初稿：TakeWindow + Pending Actor/Subject |


后续会有大量武将、装备、锦囊重新调整效果，要重新sim以及skill

联机模式3v3，五人身份、8人身份未做

所有模式都没有经过真人测试