# 宇宙杀 Browser 场景清单

> Agent 按 ID 执行；每步后 `browser_snapshot`，关键节点 `browser_take_screenshot`。

## 公共前置

1. `http://127.0.0.1:6677/` 登录 `ui_test_bot`
2. 进入 `/games/yuzhousha`

---

## yzs-mode-list

**目的**：模式 catalog API → 卡片渲染。

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 等待加载完成 | 无「加载模式中…」 |
| 2 | 快照 | 存在 `1v1`、`2v2` 卡片 |
| 3 | 快照 | 存在 `identity_5` 或「身份」类模式 |
| 4 | 快照 | 存在 `identity_8` 或「八人」类模式（若已上线） |
| 5 | 点任一「开始 xxx」 | 进入 `/games/yuzhousha/solo/pick` |

---

## yzs-pick-1v1

**目的**：选将页、技能标签。

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 模式页点「开始 1v1」 | URL `/solo/pick`（无 mode query 或 mode=1v1） |
| 2 | 快照 | 武将 grid `.yzs-pick__grid` 有卡片 |
| 3 | 点「刘备」或第一个蜀将 | 跳转 `/games/yuzhousha/play/:gameId` |

**identity_5/8 主公技**：在 pick 页看刘备「仁德/激将」**无**「1v1不可用」标签（`skillBlockedInMode`）。

---

## yzs-pick-identity8

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 模式页点「开始 identity_8」 | pick URL 带 `?mode=identity_8` |
| 2 | 快照 | subtitle/标题含「八人」或 8 人身份说明 |
| 3 | 选将开局 | 进入 play 页 |

---

## yzs-play-1v1-quick

**目的**：对局壳层、操作栏、不卡死。

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 1v1 选刘备开局 | phase 非 loading |
| 2 | 快照 | 己方手牌区有牌；有「← 返回」 |
| 3 | 若有「结束出牌」且可点 | 点击后 message/阶段变化 |
| 4 | 等待 3～5s 或再 snapshot | 回合推进（current_turn / message 变） |
| 5 | 无 pending 时 | 不应出现空白屏或无限 loading |

---

## yzs-layout-identity8

**目的**：八人场布局。

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | identity_8 选将开局 | — |
| 2 | 快照 | 存在 `.yzs__arena--identity8` 或 octagon 座位 |
| 3 | 数 AI 座位 | 7 个非 0 号位（1～7） |
| 4 | 快照 | 0 号位（主公）在下方或标记为主公 |

---

## yzs-layout-2v2

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 2v2 选将开局 | `.yzs__arena--2v2` 或 cross 布局 |
| 2 | 快照 | 4 座位；队友/敌区分 visible |

---

## yzs-pending-pass

**目的**：响应窗基础交互（1v1 最易触发）。

| 步 | 操作 | 期望 |
|----|------|------|
| 1 | 1v1 开局，多次「结束出牌」推进 | — |
| 2 | 直到 snapshot 出现响应 UI | 如「不出」「出闪」「无懈」等 |
| 3 | 点「不出」或等价 pass 按钮 | pending 关闭或进入下一阶段 |
| 4 | 不应 | 按钮可点但点击无反应；或 toast 报错后仍卡在 response |

**response_mode 对照**（查 `handlers.ts`）：

| mode | UI 期望 |
|------|---------|
| `wuxiek_trick` | 可出无懈或点过 |
| `wugu_pick` | 可选亮牌之一 |
| `peek_deck` | 观星分配 UI |
| `skill_ganglie_choice` | 刚烈选牌 |
| `skill_jijiang` | 激将出杀/过 |
| `discard`（出牌阶段末） | 选手牌弃置 |

---

## Pending 全量回归（发版前，耗时长）

按 `frontend/src/composables/yuzhousha/pending/handlers.ts` 每个 handler 至少人工/Agent 走一次。优先高频：

1. 杀/闪响应（默认 sha/shan）
2. `wuxiek_*`
3. `wugu_pick` / `peek_deck`
4. `skill_fankui` / `skill_ganglie_*`
5. `skill_jijiang`（2v2 / identity）
6. 弃牌阶段 `StepDiscard`

**技巧**：后端 `TestScenario_*` 摆盘到指定 pending 后，若已导出 JSON fixture，Browser 可跳过「玩到该状态」；否则 2v2/identity 人机较长，优先 1v1 或固定武将。

---

## 合码前最小集（宇宙杀 UI）

```
[ ] ui-smoke.sh
[ ] yzs-mode-list
[ ] yzs-pick-1v1
[ ] yzs-play-1v1-quick
[ ] （若改 identity_8）yzs-pick-identity8 + yzs-layout-identity8
[ ] （若改 pending）对应 response_mode 场景
```
