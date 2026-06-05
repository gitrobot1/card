---
name: card-test
description: >-
  Card 项目后端测试 Agent。执行 backend/scripts/test.sh、解读 test/yuzhousha/sim_logs
  失败日志、补写 cardtest 测试。用户说「帮我测」「跑测试」「冒烟」「sim」「simrandom」
  「cardtest」「2v2」「2V二」「3v3」「identity_5」「identity_8」「五模式随机」「六模式随机」「七模式随机」或改宇宙杀/技能/引擎时使用。
disable-model-invocation: true
---

# Card 测试 Agent

> 用户 @card-test 或说「帮我测」时，**你必须自己跑命令**，读日志，汇报结果。不要只告诉用户命令。
>
> 宇宙杀架构与扩展规范见 `[backend/internal/game/yuzhousha/dev-guide.md](../../backend/internal/game/yuzhousha/dev-guide.md)`。
>
> **前端 UI / 浏览器操作**（登录、pending 窗口、布局）用 **[card-ui-test](../card-ui-test/SKILL.md)**，不要与本 skill 混用。

## 一键决策（用户没细说时）

```
改了什么？
├─ 宇宙杀技能/结算/AI     → smoke，再 yzs 或 scenario，大改后 sim
├─ 宇宙杀 2v2 / 敌友/选目标 → 2v2，再 smoke（1v1 回归）
├─ 宇宙杀 3p 杀上保下      → 3p_chain，再 smoke
├─ 宇宙杀 3p 斗地主        → 3p_ddz，再 smoke
├─ 宇宙杀 3v3 竞技         → 3v3，再 smoke
├─ 宇宙杀 identity_5 身份局 → identity_5，再 smoke
├─ 宇宙杀 identity_8 八人身份局 → identity_8，再 smoke
├─ 宇宙杀某个具体 bug     → yzs -run TestXxx 或 TestScenario
├─ 其他游戏               → ./scripts/test.sh <game> -v
├─ 合码前全量             → ./scripts/test.sh
└─ 「探索奇怪 bug / 大量随机」→ CARD_SIM=1 ./scripts/test.sh simrandom -v，读 sim_logs
└─ 「前端结算对不对 / 随机测 UI」→ **[card-ui-test](../card-ui-test/SKILL.md)** `frontend/scripts/ui-sim.sh`（或 `./scripts/test.sh uifixture` 只 harvest）
```

**工作目录始终是 `backend/`。**

---

## 唯一入口脚本：`backend/scripts/test.sh`

所有测试带 `-tags cardtest`，脚本内部已处理，**禁止**手写不带 tag 的 `go test`。

### 子命令一览


| 命令                               | 跑什么                                  | 耗时      | 何时用                           |
| -------------------------------- | ------------------------------------ | ------- | ----------------------------- |
| `./scripts/test.sh`              | 全部游戏 `./test/...`                    | 中       | 合码前                           |
| `./scripts/test.sh smoke`        | `TestSmoke_`* 全武将对开局 + 武将×牌型矩阵       | 快 (~1s) | **每次改 yzs 必跑**                |
| `./scripts/test.sh yzs`          | 宇宙杀全部测试                              | 中       | 改 yzs 后                       |
| `./scripts/test.sh sim`          | **全部** `TestSim_`*（含 1v1 全武将对 + 七模式） | 很慢      | 深度回归；自动 `CARD_SIM=1`          |
| `./scripts/test.sh simrandom`    | **七模式随机** AI 自对弈（见下表）                | 慢       | **推荐**：改引擎/AI 后大量随机           |
| `./scripts/test.sh sim2v2`       | `TestSim_2v2_`* 四人 AI 自对弈            | 慢       | 需 `CARD_SIM=1`；含全武将 + 随机种子    |
| `./scripts/test.sh sim3p_chain`  | `TestSim_3pChain_*` 三人链式 AI 自对弈      | 慢       | 需 `CARD_SIM=1`                |
| `./scripts/test.sh sim3p_ddz`    | `TestSim_3pDdz_*` 三人斗地主 AI 自对弈       | 慢       | 需 `CARD_SIM=1`                |
| `./scripts/test.sh sim3v3`       | `TestSim_3v3_*` 六人 3v3 AI 自对弈        | 慢       | 需 `CARD_SIM=1`                |
| `./scripts/test.sh simidentity`  | `TestSim_Identity5_*` 五人身份局 AI 自对弈   | 慢       | 需 `CARD_SIM=1`                |
| `./scripts/test.sh simidentity8` | `TestSim_Identity8_*` 八人身份局 AI 自对弈   | 慢       | 需 `CARD_SIM=1`                |
| `./scripts/test.sh 2v2`          | 2v2 冒烟 + 模式单测                        | 快       | 改 mode/targeting/dying/2v2 开局 |
| `./scripts/test.sh 3p_chain`     | 杀上保下 冒烟 + 链式模式单测                     | 快       | 改 3p 链式敌友/选目标/胜负              |
| `./scripts/test.sh 3p_ddz`       | 斗地主 冒烟 + 地主模式单测                      | 快       | 改 3p 斗地主/地主特权/团队胜负            |
| `./scripts/test.sh 3v3`          | 3v3 冒烟 + 模式单测                        | 快       | 改 3v3 团队/主帅胜负/选目标             |
| `./scripts/test.sh identity_5`   | 五人身份局 冒烟 + 模式单测                      | 快       | 改 identity_5 身份/胜负/选目标        |
| `./scripts/test.sh identity_8`   | 八人身份局 冒烟 + 模式单测                      | 快       | 改 identity_8 身份/胜负/选目标        |
| `./scripts/test.sh uno`          | UNO                                  | 短       | 改 uno                         |
| `./scripts/test.sh doudizhu`     | 斗地主                                  | 短       | 改 doudizhu                    |
| `./scripts/test.sh douniu`       | 斗牛                                   | 短       | 改 douniu                      |
| `./scripts/test.sh zhajinhua`    | 炸金花                                  | 短       | 改 zhajinhua                   |


### 脚本参数


| 参数              | 含义                                                  |
| --------------- | --------------------------------------------------- |
| `-v`            | 详细输出（失败时必加）                                         |
| `-run TestName` | 只跑匹配的测试，支持 `|` 正则，如 `-run TestScenario|TestGanglie` |


示例：

```bash
cd backend
./scripts/test.sh smoke -v
./scripts/test.sh yzs -run 'TestSmoke_HeroCardKindMatrix|TestSmoke_CardMatrix' -v
./scripts/test.sh yzs -run TestScenario -v
./scripts/test.sh yzs -run TestGanglie -v
./scripts/test.sh 2v2 -v
./scripts/test.sh 3p_chain -v
./scripts/test.sh 3p_ddz -v
./scripts/test.sh 3v3 -v
./scripts/test.sh identity_5 -v
./scripts/test.sh identity_8 -v
./scripts/test.sh sim2v2 -v
./scripts/test.sh sim3p_chain -v
./scripts/test.sh sim3p_ddz -v
./scripts/test.sh sim3v3 -v
./scripts/test.sh simidentity -v
./scripts/test.sh simidentity8 -v
./scripts/test.sh simrandom -v
./scripts/test.sh sim -v
./scripts/test.sh simrandom -run TestSim_3pDdz_RandomTriosSeeded/19 -v
```

### 环境变量（仅 sim 相关）


| 变量                  | 默认          | 作用                                             |
| ------------------- | ----------- | ---------------------------------------------- |
| `CARD_SIM=1`        | sim 子命令自动设置 | 启用 `TestSim_All*` / `TestSim_Random*`（否则 Skip） |
| `CARD_SIM_ROUNDS=N` | `80`        | **每个**随机测试的种子数 1..N（七模式合计约 7×N 局）              |
| `CARD_SIM_TRACE=1`  | 关           | 模拟时记录最近 25 条事件写入失败日志                           |
| `CARD_SIM_STRICT=1` | 关           | 牌数不守恒也判 FAIL（默认只警告+写日志）                        |


组合示例：

```bash
cd backend
# 七模式各 100 种子（约 700 局）
CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v
CARD_SIM_TRACE=1 CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v
CARD_SIM_STRICT=1 ./scripts/test.sh simrandom -v
# 深度回归（含 1v1 全武将对 32×32，极慢）
CARD_SIM=1 ./scripts/test.sh sim -v
```

---

## 七模式大量随机测试（核心）

宇宙杀现有 **7 个单机模式**，各有独立的随机种子自对弈。改引擎 / AI / 模式规则后，优先跑 `**simrandom`**。

### 一键七模式随机

```bash
cd backend
./scripts/test.sh simrandom -v
# 默认 CARD_SIM_ROUNDS=40 → 七模式各 40 种子 ≈ 280 局
CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v
```

失败日志：`test/yuzhousha/sim_logs/`（`simrandom` 会清空 `failures-summary.log` 再跑）。

### 七模式对照表


| 模式 ID             | 人数  | 子命令（冒烟）      | 子命令（sim 全量）    | 随机测试                                   | 开局 API                       |
| ----------------- | --- | ------------ | -------------- | -------------------------------------- | ---------------------------- |
| `1v1`             | 2   | `smoke`      | `sim`          | `TestSim_RandomHeroMixSeeded`          | `NewSolo1v1`                 |
| `2v2`             | 4   | `2v2`        | `sim2v2`       | `TestSim_2v2_RandomQuadsSeeded`        | `NewSolo2v2WithHeroes`       |
| `3p_chain` 杀上保下   | 3   | `3p_chain`   | `sim3p_chain`  | `TestSim_3pChain_RandomTriosSeeded`    | `NewSolo3pChainWithHeroes`   |
| `3p_ddz` 斗地主      | 3   | `3p_ddz`     | `sim3p_ddz`    | `TestSim_3pDdz_RandomTriosSeeded`      | `NewSolo3pDdzWithHeroes`     |
| `3v3` 竞技          | 6   | `3v3`        | `sim3v3`       | `TestSim_3v3_RandomHexesSeeded`        | `NewSolo3v3WithHeroes`       |
| `identity_5` 五人身份 | 5   | `identity_5` | `simidentity`  | `TestSim_Identity5_RandomPentasSeeded` | `NewSoloIdentity5WithHeroes` |
| `identity_8` 八人身份 | 8   | `identity_8` | `simidentity8` | `TestSim_Identity8_RandomOctasSeeded`  | `NewSoloIdentity8WithHeroes` |


**身份局共用规则**（`identity_5` / `identity_8`，见 `engine/mode/identity.go`）：主公 +1 体力公开；其余身份隐藏至阵亡；可攻击除自己外任意角色；禁【闪电】；主公技可用（`LordSkillsActive`）。胜负：主公阵亡 → 反贼胜（**主内单挑时主公阵亡** → 内奸胜）；反贼与内奸全灭 → 主公阵营胜；仅剩内奸 → 内奸胜。身份字段 `lord` / `loyalist` / `spy` / `rebel`。


| 模式           | 配比                    |
| ------------ | --------------------- |
| `identity_5` | 1 主 + 1 忠 + 1 内 + 2 反 |
| `identity_8` | 1 主 + 2 忠 + 1 内 + 4 反 |


### 每个模式的 sim 三层


| 层      | 测试后缀               | 需 CARD_SIM | 规模                              |
| ------ | ------------------ | ---------- | ------------------------------- |
| 快      | `SingleQuick`      | 否          | 1 局固定阵容，跟 `yzs` 套件跑             |
| 全武将    | `AllHeroesAsSeat0` | 是          | 32 武将各作 0 号位，余座随机不重复            |
| **随机** | `Random*Seeded`    | 是          | 种子 1..`CARD_SIM_ROUNDS`，每种子随机选将 |


### 单模式随机（只跑一种）

```bash
cd backend
# 1v1 随机
./scripts/test.sh sim -run TestSim_RandomHeroMixSeeded -v
./scripts/test.sh sim -run TestSim_RandomHeroMixSeeded/19 -v    # 复现种子 19

# 2v2 随机
./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded -v
./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded/19 -v

# 杀上保下 随机
./scripts/test.sh sim3p_chain -run TestSim_3pChain_RandomTriosSeeded -v
./scripts/test.sh sim3p_chain -run TestSim_3pChain_RandomTriosSeeded/19 -v

# 斗地主 随机
./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_RandomTriosSeeded -v
./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_RandomTriosSeeded/19 -v

# 3v3 随机
./scripts/test.sh sim3v3 -run TestSim_3v3_RandomHexesSeeded -v
./scripts/test.sh sim3v3 -run TestSim_3v3_RandomHexesSeeded/19 -v

# 五人身份局 随机
./scripts/test.sh simidentity -run TestSim_Identity5_RandomPentasSeeded -v
./scripts/test.sh simidentity -run TestSim_Identity5_RandomPentasSeeded/19 -v

# 八人身份局 随机
./scripts/test.sh simidentity8 -run TestSim_Identity8_RandomOctasSeeded -v
./scripts/test.sh simidentity8 -run TestSim_Identity8_RandomOctasSeeded/19 -v
```

### 推荐合码前 sim 套餐

```bash
cd backend
./scripts/test.sh smoke -v                    # 1v1 矩阵 + 全模式冒烟 hook
./scripts/test.sh 2v2 -v && ./scripts/test.sh 3p_chain -v && ./scripts/test.sh 3p_ddz -v && ./scripts/test.sh 3v3 -v && ./scripts/test.sh identity_5 -v && ./scripts/test.sh identity_8 -v
CARD_SIM=1 ./scripts/test.sh simrandom -v     # 七模式随机（默认 280 局）
# 或加量：
CARD_SIM_ROUNDS=100 CARD_SIM=1 ./scripts/test.sh simrandom -v
```

### 用户说「五模式随机 / 大量随机 / sim 全模式」→ 你跑


| 用户说                          | 你跑                                                             |
| ---------------------------- | -------------------------------------------------------------- |
| 五模式随机 / 六模式随机 / 七模式随机 / 大量随机 | `CARD_SIM=1 ./scripts/test.sh simrandom -v`                    |
| 加量随机（100 种子）                 | `CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v`           |
| 只测某模式随机                      | 见上「单模式随机」                                                      |
| 单种子复现                        | `./scripts/test.sh simrandom -run TestSim_*_Random*/<seed> -v` |
| 深度回归（含 1v1 全武将对）             | `CARD_SIM=1 ./scripts/test.sh sim -v`（很慢）                      |
| 合码前 sim                      | `smoke -v` + 各模式冒烟 + `simrandom -v`                            |


---

## 测试分层（宇宙杀 yzs）


| 层        | 文件                                                                                                                  | 测试前缀                         | 测什么                            |
| -------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------- | ------------------------------ |
| 冒烟       | `smoke_test.go` `card_matrix_test.go`                                                                               | `TestSmoke_`                 | 全武将对开局 + **武将×牌型矩阵**（基本/锦囊/武器） |
| 单元/技能    | `skill_test.go` `engine_test.go`                                                                                    | `TestXxx`                    | 手工摆盘的精确断言                      |
| 场景       | `scenario_test.go`                                                                                                  | `TestScenario_`              | 复杂结算链（伤害/装备/无懈）                |
| 快 sim    | `sim_test.go` `sim_2v2_test.go` `sim_3p_*_test.go` `sim_3v3_test.go` `sim_identity_test.go` `sim_identity8_test.go` | `TestSim_*SingleQuick`       | 1 局固定阵容（无需 CARD_SIM）           |
| 随机 sim   | 同上                                                                                                                  | `TestSim_*Random*Seeded`     | 种子 1..N 随机选将（需 CARD_SIM=1）     |
| 全武将对 sim | `sim_test.go`                                                                                                       | `TestSim_AllHeroPairsAIVsAI` | 仅 1v1，32×32（需 CARD_SIM=1，极慢）   |


**cardtest 约定**：测试只在 `backend/test/`；`internal/game/**/testhook.go`（`//go:build cardtest`）供外部测试调内部方法。

### 武将×牌型矩阵（`card_matrix_test.go`）

- `TestSmoke_HeroCardKindMatrix`：每位可选武将 × 牌堆全部 kind（基本/锦囊/装备），出牌后消化响应窗，断言 57 张守恒
- `TestSmoke_CardMatrixCatalogCoversDeckKinds`：矩阵目录与 `NewBasicDeck()` kind 一致
- 仅跑矩阵：`./scripts/test.sh yzs -run 'TestSmoke_HeroCardKindMatrix|TestSmoke_CardMatrix' -v`

---

## 2v2 全量测试

> 详细清单见 `backend/internal/game/yuzhousha/dev-guide.md` §3。

### 自动化

```bash
cd backend
./scripts/test.sh 2v2 -v                              # 冒烟：TestSmoke_2v2 + mode 单测（无需 CARD_SIM）
CARD_SIM=1 ./scripts/test.sh sim2v2 -v                # 全量 sim：SingleQuick + 全武将 + 随机种子
CARD_SIM=1 ./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded -v   # 仅随机
CARD_SIM=1 ./scripts/test.sh sim2v2 -run TestSim_2v2_AllHeroesAsSeat0 -v     # 仅全武将矩阵
./scripts/test.sh smoke -v                            # 1v1 回归
```


| 测试                                 | 测什么                                 |
| ---------------------------------- | ----------------------------------- |
| `TestSmoke_2v2_AllHeroesBootstrap` | 32 武将各作 0 号位 2v2 开局                 |
| `TestSim_2v2_AllHeroesAsSeat0`     | 32 武将 × AI 四人自对弈（`CARD_SIM=1`）      |
| `TestSim_2v2_RandomQuadsSeeded`    | 种子 1..`CARD_SIM_ROUNDS` 随机四人（默认 40） |
| mode 单测                            | 敌友、杀队友、registry                     |


失败日志：`test/yuzhousha/sim_logs/`；2v2 濒死卡住查 `skill_dying.go` + `ai.go`（队友出桃）。

### 用户说「测 2v2 / 2V二 / 2v2 全量 / 2v2 随机」→ 你跑


| 用户说              | 你跑                                                                                             |
| ---------------- | ---------------------------------------------------------------------------------------------- |
| 2v2 冒烟 / 2v2 单测  | `./scripts/test.sh 2v2 -v`                                                                     |
| 2v2 全量 / 2v2 sim | `CARD_SIM=1 ./scripts/test.sh sim2v2 -v`                                                       |
| 2v2 随机           | `CARD_SIM=1 CARD_SIM_ROUNDS=40 ./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded -v` |
| 合码前 2v2          | `2v2 -v` + `smoke -v` + `CARD_SIM=1 sim2v2 -run TestSim_2v2_SingleQuick -v`                    |


---

## 3p 模式全量测试（杀上保下 / 斗地主）

> 详细清单见 `backend/internal/game/yuzhousha/dev-guide.md` §4。

### 自动化

```bash
cd backend
./scripts/test.sh 3p_chain -v                           # 杀上保下冒烟 + 链式 mode 单测
./scripts/test.sh 3p_ddz -v                             # 斗地主冒烟 + 地主 mode 单测
CARD_SIM=1 ./scripts/test.sh sim3p_chain -v             # 链式 AI 自对弈（全武将 + 随机种子）
CARD_SIM=1 ./scripts/test.sh sim3p_ddz -v               # 斗地主 AI 自对弈
CARD_SIM=1 ./scripts/test.sh sim3p_chain -run TestSim_3pChain_RandomTriosSeeded/1 -v  # 单种子复现
CARD_SIM=1 ./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_RandomTriosSeeded/1 -v
./scripts/test.sh smoke -v                              # 1v1 回归
```


| 测试                                     | 测什么                                                |
| -------------------------------------- | -------------------------------------------------- |
| `TestSmoke_3pChain_AllHeroesBootstrap` | 32 武将各作 0 号位链式开局                                   |
| `TestSmoke_3pDdz_AllHeroesBootstrap`   | 32 武将各作地主开局，团队/摸牌加成                                |
| `TestSmoke_3pDdz_LandlordPerks`        | 地主多摸牌、双杀                                           |
| `TestSim_3pChain_SingleQuick`          | 固定三人链式快速 AI 局                                      |
| `TestSim_3pChain_AllHeroesAsSeat0`     | 全武将 × 随机对手 AI 自对弈（`CARD_SIM=1`）                    |
| `TestSim_3pChain_RandomTriosSeeded`    | 种子 1..N 随机三人阵容                                     |
| `TestSim_3pDdz_SingleQuick`            | 固定三人斗地主快速 AI 局                                     |
| `TestSim_3pDdz_AllHeroesAsSeat0`       | 全武将 × 随机农民 AI 自对弈（`CARD_SIM=1`）                    |
| `TestSim_3pDdz_RandomTriosSeeded`      | 种子 1..N 随机三人阵容                                     |
| mode 单测                                | `chain_test.go` / `ddz_test.go` — 敌友、AOE 跳过下家、链式胜负 |


创建固定三人盘：

- 链式：`engine.NewSolo3pChainWithHeroes(id, [3]string{seat0, seat1, seat2})`
- 斗地主：`engine.NewSolo3pDdzWithHeroes(id, [3]string{landlord, farmer, farmer})`

### 用户说「测 3p / 杀上保下 / 斗地主 sim」→ 你跑


| 用户说       | 你跑                                                                                                                                                    |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 3p 链式冒烟   | `./scripts/test.sh 3p_chain -v`                                                                                                                       |
| 斗地主冒烟     | `./scripts/test.sh 3p_ddz -v`                                                                                                                         |
| 3p 链式 sim | `CARD_SIM=1 ./scripts/test.sh sim3p_chain -v`                                                                                                         |
| 斗地主 sim   | `CARD_SIM=1 ./scripts/test.sh sim3p_ddz -v`                                                                                                           |
| 合码前 3p    | `3p_chain -v` + `3p_ddz -v` + `CARD_SIM=1 sim3p_chain -run TestSim_3pChain_SingleQuick -v` + `CARD_SIM=1 sim3p_ddz -run TestSim_3pDdz_SingleQuick -v` |


---

## 3v3 模式全量测试

> 详细清单见 `backend/internal/game/yuzhousha/dev-guide.md` §3.6。

### 自动化

```bash
cd backend
./scripts/test.sh 3v3 -v                              # 冒烟 + mode 单测（无需 CARD_SIM）
CARD_SIM=1 ./scripts/test.sh sim3v3 -v                # 全量 sim：SingleQuick + 全武将 + 随机种子
CARD_SIM=1 ./scripts/test.sh sim3v3 -run TestSim_3v3_RandomHexesSeeded -v   # 仅随机
CARD_SIM=1 ./scripts/test.sh sim3v3 -run TestSim_3v3_AllHeroesAsSeat0 -v     # 仅全武将矩阵
./scripts/test.sh smoke -v                            # 1v1 回归
```


| 测试                                 | 测什么                             |
| ---------------------------------- | ------------------------------- |
| `TestSmoke_3v3_AllHeroesBootstrap` | 32 武将各作 0 号位（暖主帅）6 人开局          |
| `TestSmoke_3v3_SingleQuick`        | 团队/主帅座位不变量                      |
| `TestSim_3v3_SingleQuick`          | 固定六人快速 AI 局                     |
| `TestSim_3v3_AllHeroesAsSeat0`     | 全武将 × 随机五人 AI 自对弈（`CARD_SIM=1`） |
| `TestSim_3v3_RandomHexesSeeded`    | 种子 1..`CARD_SIM_ROUNDS` 随机六人阵容  |
| mode 单测                            | 敌友、主帅胜负、registry、禁闪电            |


创建固定六人盘：`engine.NewSolo3v3WithHeroes(id, [6]string{seat0..seat5})`（0 暖主帅、2 冷主帅）。

### 用户说「测 3v3 / 3v3 sim / 3v3 随机」→ 你跑


| 用户说              | 你跑                                                                                             |
| ---------------- | ---------------------------------------------------------------------------------------------- |
| 3v3 冒烟 / 3v3 单测  | `./scripts/test.sh 3v3 -v`                                                                     |
| 3v3 全量 / 3v3 sim | `CARD_SIM=1 ./scripts/test.sh sim3v3 -v`                                                       |
| 3v3 随机           | `CARD_SIM=1 CARD_SIM_ROUNDS=40 ./scripts/test.sh sim3v3 -run TestSim_3v3_RandomHexesSeeded -v` |
| 合码前 3v3          | `3v3 -v` + `smoke -v` + `CARD_SIM=1 sim3v3 -run TestSim_3v3_SingleQuick -v`                    |


---

## identity_5 模式全量测试

> 标准五人身份场；模式规则见 `engine/mode/identity.go`、`engine/mode/registry.go`。

### 自动化

```bash
cd backend
./scripts/test.sh identity_5 -v                              # 冒烟 + mode 单测（无需 CARD_SIM）
CARD_SIM=1 ./scripts/test.sh simidentity -v                  # 全量 sim：SingleQuick + 全武将 + 随机种子
CARD_SIM=1 ./scripts/test.sh simidentity -run TestSim_Identity5_RandomPentasSeeded -v   # 仅随机
CARD_SIM=1 ./scripts/test.sh simidentity -run TestSim_Identity5_AllHeroesAsSeat0 -v     # 仅全武将矩阵
./scripts/test.sh smoke -v                                   # 1v1 回归
```


| 测试                                       | 测什么                                       |
| ---------------------------------------- | ----------------------------------------- |
| `TestSmoke_Identity5_AllHeroesBootstrap` | 32 武将各作 0 号位（主公）5 人开局                     |
| `TestSmoke_Identity5_SingleQuick`        | 身份分配（1 忠 1 内 2 反）、主公先行动、可选任意目标            |
| `TestSmoke_Identity5_SetupShuffle`       | 随机发身份（4 非主公位洗牌）                           |
| `TestSim_Identity5_SingleQuick`          | 固定五人快速 AI 局                               |
| `TestSim_Identity5_AllHeroesAsSeat0`     | 全武将 × 主公位 + 随机四人 AI（`CARD_SIM=1`）         |
| `TestSim_Identity5_RandomPentasSeeded`   | 种子 1..`CARD_SIM_ROUNDS` 随机选将 + 随机身份分配     |
| mode 单测                                  | `ValidateIdentity5Roles`、胜负（含内奸独活）、敌友、选目标 |


创建固定五人盘：`engine.NewSoloIdentity5WithHeroes(id, [5]string{heroes}, [5]string{roles})`；人机默认 `NewSoloIdentity5(id, name, heroID)`（0 号位固定主公，余座随机 1 忠 1 内 2 反）。

### 用户说「测身份局 / identity sim / 内奸」→ 你跑


| 用户说          | 你跑                                                                                                         |
| ------------ | ---------------------------------------------------------------------------------------------------------- |
| 身份局冒烟 / 单测   | `./scripts/test.sh identity_5 -v`                                                                          |
| 身份局全量 sim    | `CARD_SIM=1 ./scripts/test.sh simidentity -v`                                                              |
| 身份局随机        | `CARD_SIM=1 CARD_SIM_ROUNDS=40 ./scripts/test.sh simidentity -run TestSim_Identity5_RandomPentasSeeded -v` |
| 合码前 identity | `identity_5 -v` + `smoke -v` + `CARD_SIM=1 simidentity -run TestSim_Identity5_SingleQuick -v`              |


### 身份局结算机制核查清单（改 `mode/identity.go` / `skill_dying.go` / `skill_zhangchunhua.go` 必跑）


| 条件           | 期望                                         | 测试                                                                                                |
| ------------ | ------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| 主公阵亡         | 反贼阵营胜（即使反贼已全灭）                             | `TestEvaluateIdentityWin_LordDeathNoLivingRebels` / `TestScenario_Identity_LordDeath_*`           |
| 主公+内奸单挑，主公阵亡 | 内奸胜（含内奸击杀、闪电、未知来源）                         | `TestEvaluateIdentityWin_LordDeathSpyDuel*` / `TestScenario_Identity_SpyDuel*`                    |
| 反贼 + 内奸全灭    | 主公阵营胜                                      | `TestEvaluateIdentityWin_RebelsEliminated` / `TestScenario_Identity_LordFactionWin_*`             |
| 仅剩内奸存活       | 内奸胜                                        | `TestEvaluateIdentityWin_SpySolo` / `TestScenario_Identity_SpySoloWin`                            |
| 反贼全灭但内奸存活    | **继续**                                     | `TestEvaluateIdentityWin_RebelsDeadSpyAlive` / `TestScenario_Identity_ContinuesWhenSpyStillAlive` |
| 主公 + 内奸双存活   | **继续**                                     | `TestScenario_Identity_ContinuesLordAndSpyRemaining`                                              |
| 阵亡           | 身份揭示                                       | `TestScenario_Identity_RevealOnDeath`                                                             |
| 绝情致死         | 走 `checkIdentityDeath`（非 1v1 `finishGame`） | `TestScenario_Identity_JueqingLordDeath_`* / `TestScenario_Identity8_JueqingLordDeath_*`          |


---

## identity_8 模式全量测试

> 标准八人身份场；胜负/敌友与 `identity_5` 共用 `engine/mode/identity.go`（`IsIdentityMode`）；布局 `octagon_8`。

### 自动化

```bash
cd backend
./scripts/test.sh identity_8 -v                              # 冒烟 + mode 单测 + 主公技（无需 CARD_SIM）
CARD_SIM=1 ./scripts/test.sh simidentity8 -v                  # 全量 sim：SingleQuick + 全武将 + 随机种子
CARD_SIM=1 ./scripts/test.sh simidentity8 -run TestSim_Identity8_RandomOctasSeeded -v   # 仅随机
CARD_SIM=1 ./scripts/test.sh simidentity8 -run TestSim_Identity8_AllHeroesAsSeat0 -v     # 仅全武将矩阵
./scripts/test.sh smoke -v                                   # 1v1 回归
```


| 测试                                          | 测什么                                                                                                                                                             |
| ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TestSmoke_Identity8_AllHeroesBootstrap`    | 32 武将各作 0 号位（主公）8 人开局                                                                                                                                           |
| `TestSmoke_Identity8_SingleQuick`           | 身份分配（2 忠 1 内 4 反）、主公先行动、可对任意目标出杀（含忠臣）                                                                                                                           |
| `TestSmoke_Identity8_SetupShuffle`          | 随机发身份（7 非主公位洗牌）                                                                                                                                                 |
| `TestIdentity8_LordSkillsNotInactiveIn1v1`  | 主公技在 identity_8 非 `inactive_in_1v1`                                                                                                                             |
| `TestListHeroes_Identity8_LordSkillsActive` | 武将目录 `mode=identity_8` 主公技可用                                                                                                                                    |
| `TestSim_Identity8_SingleQuick`             | 固定八人快速 AI 局                                                                                                                                                     |
| `TestSim_Identity8_AllHeroesAsSeat0`        | 全武将 × 主公位 + 随机七人 AI（`CARD_SIM=1`）                                                                                                                               |
| `TestSim_Identity8_RandomOctasSeeded`       | 种子 1..`CARD_SIM_ROUNDS` 随机选将 + 随机身份分配                                                                                                                           |
| mode 单测                                     | `ValidateIdentity8Roles`、`TestTeamOf_Identity8`、`TestValidPlayTargets_Identity8AnyOther`、`TestEvaluateIdentityWin`、`IsIdentityMode`、`LordSkillsActive`、registry |
| 场景单测                                        | `TestScenario_Identity8_*`（与 identity_5 场景一一对应）                                                                                                                 |


创建固定八人盘：`engine.NewSoloIdentity8WithHeroes(id, [8]string{heroes}, [8]string{roles})`；人机默认 `NewSoloIdentity8(id, name, heroID)`（0 号位固定主公，余座随机 2 忠 1 内 4 反）。`TestSim_Identity8_*` 在开局后调用 `SetDeckSeedForTest` 固定洗牌，与 `RandomOctasSeeded` 一致。

### 用户说「测八人身份 / identity_8 sim / 八人内奸」→ 你跑


| 用户说            | 你跑                                                                                                         |
| -------------- | ---------------------------------------------------------------------------------------------------------- |
| 八人身份冒烟 / 单测    | `./scripts/test.sh identity_8 -v`                                                                          |
| 八人身份全量 sim     | `CARD_SIM=1 ./scripts/test.sh simidentity8 -v`                                                             |
| 八人身份随机         | `CARD_SIM=1 CARD_SIM_ROUNDS=40 ./scripts/test.sh simidentity8 -run TestSim_Identity8_RandomOctasSeeded -v` |
| 合码前 identity_8 | `identity_8 -v` + `smoke -v` + `CARD_SIM=1 simidentity8 -run TestSim_Identity8_SingleQuick -v`             |


---

## Sim 失败日志

### 位置（相对 repo）

```text
backend/test/yuzhousha/sim_logs/<对局名>.log     # 单次失败
backend/test/yuzhousha/sim_logs/failures-summary.log  # 一次 sim 全部失败汇总
backend/test/yuzhousha/sim_logs/README.md
```

`./scripts/test.sh sim` 会清空 `failures-summary.log` 再跑。`simrandom` / `sim2v2` / `sim3p_*` / `sim3v3` / `simidentity` / `simidentity8` 同理。

### Agent 读日志流程

1. sim 失败后读 `failures-summary.log` 看有哪些对局
2. 打开对应 `.log` 看 **「可能问题区域」** 和 **「复现」**
3. 按「建议查」路径读引擎代码，最小修复
4. 用日志里的复现命令再跑，直到绿

### 日志结构（固定章节）

```text
=== 宇宙杀 AI 模拟失败报告 ===
时间 / 对局 / 失败类型          ← 见下方「失败类型」表
步数: N / 8000
卡住指纹: ...                   ← 可选，state 不变时

--- 可能问题区域 ---
分类: ...                       ← 如「伤害链」「AOE/杀响应」
建议查: engine/xxx.go — ...     ← **优先从这里下手**

--- 局面 ---
phase / step / turn / message
牌堆 / 弃牌 / 牌总数 (期望 57)  ← 57 = len(NewBasicDeck())
[0] 玩家0 HP 手牌 技能 装备 判定区
[1] 玩家1 ...

--- Pending ---                 ← 当前响应窗，卡死时最关键
  mode / required / src / tgt / card ...

--- 最近事件 ---                ← CARD_SIM_TRACE=1 才有内容

--- 复现 ---                    ← 复制这条命令再跑
```

### 失败类型 `Reason`


| 值             | 含义                   | Agent 动作                                 |
| ------------- | -------------------- | ---------------------------------------- |
| `stuck`       | 状态指纹不变，AI+force 都推不动 | 读 Pending + 建议查；修 ai.go 或 response       |
| `timeout`     | 8000 步未结束            | 可能死循环或局太长；先 repro 单局                     |
| `force_error` | forceProgress 报错     | 读 phase/step，修状态机                        |
| `card_loss`   | 结束牌总数 ≠ 57           | 查 play/tricks/skill 丢牌逻辑；STRICT=1 才 FAIL |
| `no_winner`   | 结束了无 winner          | 查 finishGame 路径                          |


### 分类 → 源码速查


| 分类      | 常见文件                                                   |
| ------- | ------------------------------------------------------ |
| 伤害链     | `skill_damage.go`, `skill_jianxiong/ganglie/fankui.go` |
| AOE/杀响应 | `response.go`, `play.go`（南蛮/激将/决斗）                     |
| 无懈      | `response.go`, `tricks.go`                             |
| 装备      | `weapons.go`（青龙、麒麟弓）                                   |
| 阶段技     | `phase_prepare.go`（观星、洛神）                              |
| 锦囊      | `tricks.go`（五谷看 `WuguPickSeat`）                        |
| 主公技     | `skill_register.go`（激将 1v1 看 `ShuAllies`）              |
| 牌堆守恒    | `play.go`, `tricks.go`, 各 `skill_*.go`                 |
| 出牌阶段    | `ai.go` `runAIPlayPhase`                               |


---

## Agent 标准流程（每次执行）

```
1. cd backend
2. 按「一键决策」选命令并执行（带 -v）
3. 失败 → 读输出；sim 失败 → 读 sim_logs/*.log
4. 修代码或补测试（最小 diff）
5. 重跑同一命令确认
6. 用下方「报告模板」回复用户
```

### 用户常见说法 → 命令


| 用户说                          | 你跑                                                             |
| ---------------------------- | -------------------------------------------------------------- |
| 帮我测 / 跑一下测试                  | `smoke` + 相关 `yzs`；改动大则 `sim`                                  |
| 冒烟                           | `./scripts/test.sh smoke -v`                                   |
| 自对弈 / sim / 随机组合             | `CARD_SIM=1 ./scripts/test.sh simrandom -v`                    |
| 五模式随机 / 六模式随机 / 七模式随机 / 大量随机 | `CARD_SIM_ROUNDS=40 ./scripts/test.sh simrandom -v`            |
| 测宇宙杀 / yzs                   | `./scripts/test.sh yzs -v`                                     |
| 测刚烈/某技能                      | `./scripts/test.sh yzs -run TestGanglie -v`                    |
| 测场景/结算                       | `./scripts/test.sh yzs -run TestScenario -v`                   |
| 测 2v2 / 2V二                  | `./scripts/test.sh 2v2 -v` + `smoke -v` + 见 dev-guide 手动清单     |
| 测 3p / 杀上保下                  | `./scripts/test.sh 3p_chain -v`                                |
| 测 斗地主 yzs 模式                 | `./scripts/test.sh 3p_ddz -v`                                  |
| 测 3v3 / 3v3 竞技               | `./scripts/test.sh 3v3 -v`                                     |
| 测 identity_5 / 五人身份 / 内奸     | `./scripts/test.sh identity_5 -v`                              |
| 测 identity_8 / 八人身份          | `./scripts/test.sh identity_8 -v`                              |
| 3p sim / 斗地主 sim             | `CARD_SIM=1 ./scripts/test.sh sim3p_chain -v` 或 `sim3p_ddz -v` |
| 3v3 sim                      | `CARD_SIM=1 ./scripts/test.sh sim3v3 -v`                       |
| identity_5 sim               | `CARD_SIM=1 ./scripts/test.sh simidentity -v`                  |
| identity_8 sim               | `CARD_SIM=1 ./scripts/test.sh simidentity8 -v`                 |
| 全量                           | `./scripts/test.sh`                                            |


---

## 写/补测试（简要）

- 精确场景：抄 `scenario_test.go` 四步法（选将 → 摆盘 → 逐步推进 → 断言 Pending）
- 2v2 场景：`engine.NewSolo2v2(id, name, heroID)` 开局；敌友用 `mode.EnemiesOf` 断言
- 3p 链式：`engine.NewSolo3pChainWithHeroes(id, [3]string{...})`；上家/下家用 `mode.MarkTarget` / `ProtectTarget`
- 3p 斗地主：`engine.NewSolo3pDdzWithHeroes(id, [3]string{landlord, f1, f2})`；团队用 `mode.TeamOf`
- 3v3：`engine.NewSolo3v3WithHeroes(id, [6]string{seat0..seat5})`；主帅胜负用 `mode.IsCommander3v3` / `mode.TeamOf3v3`
- identity_5：`engine.NewSoloIdentity5WithHeroes(id, [5]string{heroes}, [5]string{roles})`；身份/胜负用 `mode.ValidateIdentity5Roles` / `mode.EvaluateIdentityWin` / `mode.RoleSpy`
- identity_8：`engine.NewSoloIdentity8WithHeroes(id, [8]string{heroes}, [8]string{roles})`；身份校验 `mode.ValidateIdentity8Roles`；胜负仍用 `mode.EvaluateIdentityWin`（与 identity_5 共用）
- 新技能：读 `internal/game/yuzhousha/skill/doc.go` + `internal/game/yuzhousha/dev-guide.md`
- 需内部 API：在 `engine/testhook.go` 加 `XxxForTest`，`//go:build cardtest`
- 至少 1 正向 + 1 边界用例

---

## 安全

- 不起 MySQL / Redis / Docker / 不监听端口
- 不申请 `required_permissions: all`

---

## 报告模板（回复用户必用）

```markdown
## 测试报告
- **命令**: `cd backend && ./scripts/test.sh ...`
- **结果**: pass / N failed
- **Sim 日志**（如有）: `test/yuzhousha/sim_logs/xxx.log` → 分类 / 建议查
- **失败摘要**（如有）: 测试名 → 一行原因
- **改动**（如有）: 文件 + 原因
```

