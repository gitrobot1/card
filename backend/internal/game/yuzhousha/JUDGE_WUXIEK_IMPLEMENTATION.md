# 判定前无懈可击窗口实现文档

## 概述

本文档描述了三国杀游戏引擎中判定前无懈可击窗口的实现。该功能允许玩家在判定牌生效前使用无懈可击来抵消判定牌的效果。

## 功能特性

### 1. 判定前无懈可击窗口
- 在判定阶段，当判定区有牌时，系统会先启动一个无懈可击响应窗口
- **响应顺序**：按照三国杀规则，从当前回合玩家开始，逆时针顺序依次响应
- 所有玩家（包括AI）都可以响应这个窗口，使用无懈可击来抵消判定牌的效果
- 如果没有人有无懈可击，系统会直接执行判定

### 2. 反无懈可击窗口
- 当第一张无懈可击被打出后，系统会启动一个反无懈可击窗口
- **响应顺序**：从打出无懈可击的玩家下家开始，逆时针顺序依次响应
- 所有玩家都可以再次打出无懈可击来抵消第一张无懈可击
- 这实现了无懈可击的递归响应机制

### 3. 响应队列管理
- 使用`ResponseQueue`字段管理响应顺序
- 使用`ResponseIndex`字段跟踪当前响应者
- 当当前响应者跳过时，自动移动到队列中的下一个响应者
- 支持AI自动响应

### 4. 支持所有判定牌类型
- **乐不思蜀**：判定生效后跳过出牌阶段
- **兵粮寸断**：判定生效后跳过摸牌阶段
- **闪电**：判定生效后受到3点雷电伤害，否则转移到下家

## 实现细节

### 核心函数

#### 1. `startJudgeWuxiekWindow(seat int, judgeCard Card, events *[]GameEvent) error`
- **位置**：`phase_prepare.go`
- **功能**：在判定前启动无懈可击响应窗口
- **逻辑**：
  1. 根据判定牌类型设置响应模式（ResponseModeWuxiekLebu/Bingliang/Shandian）
  2. 创建响应队列：从当前回合玩家开始，逆时针顺序
  3. 检查队列中是否有玩家可以使用无懈可击
  4. 如果没有人有无懈可击，直接执行判定
  5. 否则，启动无懈可击响应窗口，设置响应队列

#### 2. `createResponseQueue(startSeat int) []int`
- **位置**：`phase_prepare.go`
- **功能**：创建响应队列，按照三国杀规则：从指定座位开始，逆时针顺序
- **逻辑**：
  1. 从起始座位开始
  2. 按逆时针顺序（下家是 (currentSeat + 1) % len(g.Players)）添加所有存活玩家
  3. 返回响应队列

#### 2. `handleJudgeWuxiekResponse(seat int, pending PendingCombat, events *[]GameEvent) error`
- **位置**：`response.go`
- **功能**：处理判定前无懈可击的响应
- **逻辑**：
  1. 当玩家打出无懈可击后，启动反无懈可击窗口
  2. 创建反无懈可击的响应队列：从打出无懈可击的玩家下家开始，逆时针顺序
  3. 保存原始判定信息到SavedPending字段
  4. 允许所有玩家响应反无懈可击

#### 3. `advanceJudgeWuxiekQueue(seat int, events *[]GameEvent) error`
- **位置**：`response.go`
- **功能**：推进判定前无懈可击的响应队列
- **逻辑**：
  1. 移动到下一个响应者
  2. 如果所有玩家都响应过了，执行判定
  3. 如果下一个响应者是AI，自动处理

#### 4. `advanceWuxiekResponseQueue(events *[]GameEvent) error`
- **位置**：`response.go`
- **功能**：推进反无懈可击的响应队列
- **逻辑**：
  1. 移动到下一个响应者
  2. 如果所有玩家都响应过了，第一张无懈可击生效
  3. 如果下一个响应者是AI，自动处理

#### 5. `resumeJudgeAfterWuxiek(seat int, events *[]GameEvent) error`
- **位置**：`phase_prepare.go`
- **功能**：无懈可击响应后的恢复逻辑
- **逻辑**：
  1. 恢复保存的判定信息
  2. 清除Pending状态
  3. 执行判定

#### 6. `handleWuxiekCounterPass(events *[]GameEvent) error`
- **位置**：`phase_prepare.go`
- **功能**：处理反无懈可击窗口的跳过
- **逻辑**：
  1. 当反无懈可击窗口关闭时，第一张无懈可击生效
  2. 将判定牌从判定区移除，放到弃牌堆
  3. 继续处理下一张判定牌

### 响应模式

新增了以下响应模式（在`constants.go`中定义）：
- `ResponseModeWuxiekLebu`：乐不思蜀的无懈可击响应
- `ResponseModeWuxiekBingliang`：兵粮寸断的无懈可击响应
- `ResponseModeWuxiekShandian`：闪电的无懈可击响应

### 数据结构修改

在`PendingCombat`结构体中新增了以下字段（在`model.go`中定义）：
- `ResponseQueue []int`：响应队列，按顺序排列
- `ResponseIndex int`：当前响应者在队列中的索引

### 事件类型

新增了以下事件类型：
- `wuxiek_offer`：发起无懈可击响应窗口
- `wuxiek_counter_offer`：发起反无懈可击响应窗口
- `wuxiek_cancel`：无懈可击成功抵消判定牌

## 测试用例

### 1. `TestJudgeWuxiekWindow`
- **功能**：测试判定前的无懈可击窗口是否正常启动
- **验证点**：
  - 游戏阶段是否正确切换到PhaseResponse
  - Pending是否正确设置
  - 响应模式是否正确

### 2. `TestJudgeWuxiekCancel`
- **功能**：测试无懈可击成功抵消判定牌
- **验证点**：
  - 无懈可击是否正确打出
  - 反无懈可击窗口是否正确启动
  - 判定牌是否被正确抵消（从判定区移除）

## 使用流程

### 正常流程（无人使用无懈可击）
1. 进入判定阶段
2. 检查判定区是否有牌
3. 创建响应队列：从当前回合玩家开始，逆时针顺序
4. 检查队列中是否有玩家可以使用无懈可击
5. 如果没有，直接执行判定
6. 根据判定结果处理相应效果

### 无懈可击响应流程（按队列顺序）
1. 进入判定阶段
2. 检查判定区是否有牌
3. 创建响应队列（从当前回合玩家开始，逆时针顺序）
4. 启动无懈可击响应窗口，当前响应者为队列第一个玩家
5. 当前响应者打出无懈可击 或 跳过
6. 如果跳过，移动到队列下一个玩家
7. 如果打出无懈可击，启动反无懈可击窗口
8. 反无懈可击窗口也按队列顺序（从打出无懈可击的玩家下家开始）
9. 如果没有人再打出无懈可击，第一张无懈可击生效
10. 判定牌被抵消，从判定区移除
11. 继续处理下一张判定牌（如果有）

### 响应顺序规则
1. **判定前无懈可击**：从当前回合玩家开始，逆时针顺序
   - 例如：4人局，当前回合是0号，顺序为 0 → 1 → 2 → 3
2. **反无懈可击**：从打出无懈可击的玩家下家开始，逆时针顺序
   - 例如：1号打出无懈可击，顺序为 2 → 3 → 0
3. **AI自动响应**：当轮到AI响应时，系统会自动判断是否使用无懈可击

## 注意事项

1. **响应顺序**：严格按照三国杀规则，判定前无懈可击从当前回合玩家开始逆时针响应，反无懈可击从打出无懈可击的玩家下家开始逆时针响应
2. **递归响应**：无懈可击支持递归响应，可以实现多张无懈可击的连锁
3. **AI支持**：AI玩家也可以参与无懈可击的响应，系统会自动判断是否需要使用无懈可击
4. **性能考虑**：递归深度理论上没有限制，但实际应用中很少超过2-3层
5. **状态恢复**：系统会正确保存和恢复判定状态，确保流程的正确性
6. **响应队列管理**：使用ResponseQueue和ResponseIndex字段精确管理响应顺序，避免混乱

## 文件修改清单

### 新增文件
- `judge_wuxiek_test.go`：测试用例

### 修改文件
- `model.go`：
  - 在`PendingCombat`结构体中新增`ResponseQueue`和`ResponseIndex`字段
- `phase_prepare.go`：
  - 完善`startJudgeWuxiekWindow`函数，添加响应队列管理
  - 新增`createResponseQueue`函数，创建响应队列
  - 新增`resumeJudgeAfterWuxiek`函数
  - 新增`handleWuxiekCounterPass`函数
- `response.go`：
  - 修改`RespondWuxiek`函数，添加判定前无懈可击响应处理
  - 新增`handleJudgeWuxiekResponse`函数，添加反无懈可击响应队列
  - 新增`advanceJudgeWuxiekQueue`函数，推进判定前无懈可击响应队列
  - 新增`advanceWuxiekResponseQueue`函数，推进反无懈可击响应队列
  - 新增`handleAIWuxiekResponse`函数，处理AI的无懈可击响应
  - 修改`PassResponse`函数，使用响应队列管理

## 测试验证

所有测试均已通过：
```
=== RUN   TestJudgeWuxiekWindow
    judge_wuxiek_test.go:49: 判定前无懈可击窗口测试通过！
--- PASS: TestJudgeWuxiekWindow (0.00s)

=== RUN   TestJudgeWuxiekCancel
    judge_wuxiek_test.go:104: 当前阶段: play（可能已进入摸牌阶段或继续处理判定）
    judge_wuxiek_test.go:107: 无懈可击抵消判定牌测试通过！
--- PASS: TestJudgeWuxiekCancel (0.00s)
```

## 总结

判定前无懈可击窗口的实现完善了游戏引擎的判定阶段逻辑，使得游戏流程更加符合三国杀的规则。该实现支持递归响应，能够处理复杂的无懈可击连锁情况，同时保持了代码的清晰和可维护性。
