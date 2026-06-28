# 13 - 功能完整性检查清单

> **用途**: AI 实现任何功能后，必须对照本清单逐项检查。禁止在清单未完成时声称"已完成"。
> **原则**: 每个功能类型都对应一组**必须修改的文件**，缺一不可。

---

## 一、牌当牌技能（武圣/龙胆/倾国/奇袭）

### 必须涉及的文件（全部 6 项）

| # | 文件 | 必须做的事 | 检查命令 |
|---|------|-----------|---------|
| 1 | `skill/ids.go` | 定义技能 ID 常量 | `grep "IDXxx" skill/ids.go` |
| 2 | `skill/catalog_skills.go` 或 `engine/skill_register_*.go` | 注册 `CardPlaysAs` Hook | `grep "CardPlaysAs" skill/catalog_skills.go` |
| 3 | `engine/skill_register_*.go` | 注册 `CanActivate` + `Activate`（激活/取消逻辑） | `grep "xxxCanActivate\|xxxActivate" engine/skill_register_*.go` |
| 4 | `engine/skill_actions.go` | 实现 toggle 函数（设置 SkillCounter 来激活） | `grep "toggleXxx" engine/skill_actions.go` |
| 5 | `engine/skill_runtime.go` | Runtime 接口方法 + gameSkillRuntime 实现 | `grep "ToggleXxx" engine/skill_runtime.go` |
| 6 | `engine/skill_hooks.go` 的 `cardPlaysAsViaHooks` | **已存在，不需要改**（引擎已接入 HookCardPlaysAs） | - |

### 检查清单

```bash
# 1. ID 已定义
grep "IDWusheng" skill/ids.go

# 2. CardPlaysAs Hook 已注册（带 SkillCounter 检查）
grep -A5 "IDWusheng" skill/catalog_skills.go | grep "CardPlaysAs"

# 3. CanActivate + Activate 已注册
grep -A5 "IDWusheng" engine/skill_register.go | grep "CanActivate\|Activate"

# 4. toggle 函数已实现
grep "toggleWusheng" engine/skill_actions.go

# 5. Runtime 方法已实现
grep "ToggleWusheng" engine/skill_runtime.go

# 6. 测试验证变牌生效
grep "TestWusheng" backend/test/yuzhousha/skill_test.go
```

---

## 二、卖血技（刚烈/反馈/奸雄/遗计）

### 必须涉及的文件（全部 5 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `skill/ids.go` | 定义技能 ID |
| 2 | `engine/skill_register_*.go` | 注册 CanActivate/Activate/AIPriority/AIActivate |
| 3 | `engine/skill_xxx.go`（独立文件） | 实现 offer/apply/pass 三步 PendingCombat 流程 |
| 4 | `engine/game.go` 或伤害结算处 | 在伤害结算后调用 offer 函数 |
| 5 | `skill/types.go` Runtime 接口 | 添加 `PendingXxxFor`/`ApplyXxx`/`PassXxx` 方法签名 |

### 检查清单

```bash
# 1. ID
grep "IDGanglie" skill/ids.go

# 2. 注册
grep -A5 "IDGanglie" engine/skill_register_wei.go

# 3. offer/apply/pass 三步实现
grep "offerGanglieWindow\|StartGanglieJudge\|PassGanglieOffer\|applyGanglieJudgeResult\|GanglieTakeDamage\|GanglieDiscard" engine/skill_ganglie.go

# 4. 伤害结算后调用
grep "offerGanglieWindow\|initDamageAftermath" engine/game.go engine/phase_hp_change.go

# 5. Runtime 方法
grep "PendingGanglieOfferFor\|StartGanglieJudge\|PassGanglieOffer\|GanglieTakeDamage\|GanglieDiscard" skill/types.go
```

---

## 三、主动技（青囊/仁德/结姻/制衡）

### 必须涉及的文件（全部 4 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `skill/ids.go` | 定义技能 ID |
| 2 | `engine/skill_register_*.go` | 注册 CanActivate/Activate/AIPriority/AIActivate |
| 3 | `engine/skill_actions.go` | 实现引擎方法（如 recover/give/discard） |
| 4 | `skill/types.go` Runtime 接口 | 添加 `ActivateXxx` 方法签名 |

### 检查清单

```bash
# 1. ID
grep "IDQingnang" skill/ids.go

# 2. 注册（含 CanActivate 阶段检查）
grep -A5 "IDQingnang" engine/skill_register_shu.go | grep "PhasePlaying\|StepPlay\|CurrentTurn"

# 3. 引擎实现
grep "ActivateQingnang\|qingnangActivate" engine/skill_actions.go engine/skill_register_shu.go

# 4. Runtime 方法
grep "ActivateQingnang\|Qingnang" skill/types.go
```

---

## 四、锁定技 / mod 被动技（马术/奇才/空城/克己/咆哮）

### 必须涉及的文件（全部 2-3 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `skill/ids.go` | 定义技能 ID |
| 2 | `skill/catalog_skills.go` | 注册 Decl Hook（DistanceDelta/TrickIgnoresDistance/BlocksTarget 等） |
| 3 | `engine/skill_hooks.go`（如新增 HookKind） | 如现有 Hook 不覆盖，需新增 HookKind + runSkillHooks 分支 |

### 检查清单

```bash
# 1. ID
grep "IDMashi" skill/ids.go

# 2. Decl Hook（锁定技标记 TagForced）
grep -A10 "IDMashi" skill/catalog_skills.go | grep "DistanceDelta\|TagForced"

# 3. 引擎已接入（不需要改）
grep "HookDistanceDelta" engine/skill_hooks.go
```

---

## 五、锦囊牌

### 必须涉及的文件（全部 6 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `engine/constants.go` | 定义 `CardXxx` 常量 + 标签映射 |
| 2 | `engine/deck.go` | 加入牌堆（TrickScope 定义） |
| 3 | `engine/play.go` 的 `useCard` switch | **加入 case！** |
| 4 | `engine/play.go` 的 `playTrickWithCard` | 路由到处理函数 |
| 5 | `engine/play.go` 或独立文件 | 实现效果函数（resolveXxx） |
| 6 | `frontend/.../pending/handlers.ts` | 如需要玩家交互，注册 handler |

### 检查清单

```bash
# 1. 常量
grep "CardJieDao" engine/constants.go

# 2. 牌堆
grep "jiedao" engine/deck.go

# 3. useCard switch（最易遗漏！）
grep "CardJieDao" engine/play.go | grep "case"

# 4. playTrickWithCard 路由
grep "CardJieDao" engine/play.go | grep "playTrickWithCard\|resolveJieDao"

# 5. 效果函数
grep "resolveJieDao\|ApplyJieDaoSha\|ApplyJieDaoGiveWeapon" engine/card_jiedao.go

# 6. 前端 handler
grep "jiedao" frontend/src/composables/yuzhousha/pending/handlers.ts
```

---

## 六、装备牌（武器/防具/马）

### 必须涉及的文件（全部 5 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `engine/constants.go` | 定义 `CardXxx` 常量 + `EquipZone` |
| 2 | `engine/deck.go` | 加入牌堆 |
| 3 | `skill/catalog_skills.go` | 注册 Decl Hook（如 TagEquipSkill） |
| 4 | `engine/weapons.go` 或 `engine/armors.go` | 实现武器/防具特效 |
| 5 | `engine/game.go` | 装备时/卸下时调用技能注册/注销 |

### 检查清单

```bash
# 1. 常量 + 装备区
grep "CardWeapon1\|EquipWeapon" engine/constants.go

# 2. 牌堆
grep "zhuge\|诸葛连弩" engine/deck.go

# 3. Decl Hook（TagEquipSkill）
grep "TagEquipSkill" skill/catalog_skills.go

# 4. 武器特效
grep "zhuge\|Zhuge\|连弩" engine/weapons.go

# 5. 装备流程
grep "equipCard\|EquipWeapon" engine/game.go
```

---

## 七、改判技（鬼才/鬼道）

### 必须涉及的文件（全部 3 项）

| # | 文件 | 必须做的事 |
|---|------|-----------|
| 1 | `skill/ids.go` | 定义技能 ID |
| 2 | `engine/skill_register_*.go` | 注册 CanActivate/Activate |
| 3 | `engine/skill_judge.go` 的 `collectModifyJudgeSeats` | **加入新改判技的判断条件！** |

### 检查清单

```bash
# 1. ID
grep "IDGuicai" skill/ids.go

# 2. 注册
grep "IDGuicai" engine/skill_register_wei.go

# 3. collectModifyJudgeSeats（最易遗漏！）
grep "SkillGuicai\|SkillGuidao" engine/skill_judge.go
```

---

## 使用方式

AI 完成任何功能后，**必须跑对应类型的检查命令**，输出结果。如果有任何一个 `grep` 返回空，说明遗漏了，立即修复。

**禁止在检查清单未全部通过时声称"已完成"。**
