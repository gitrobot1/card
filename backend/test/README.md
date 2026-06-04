# 游戏逻辑测试

各游戏的单元/集成测试集中在此目录，与 `internal/game/` 业务代码分离。

## 运行

```bash
cd backend
go test -tags cardtest ./test/... -count=1
```

`-tags cardtest` 会编译各游戏包里的 `testhook.go`，供本目录下的外部测试访问必要的内部钩子（不影响正常 `go build`）。

## 目录

| 路径 | 说明 |
|------|------|
| `doudizhu/` | 斗地主 |
| `douniu/` | 斗牛 |
| `zhajinhua/` | 炸金花 |
| `uno/` | UNO |
| `yuzhousha/` | 宇宙杀 |
| `yuzhousha/scenario_test.go` | 复杂结算场景示例（伤害链、装备叠加、Pending 中间态） |

业务代码内不应再保留 `*_test.go`（`testhook.go` 除外）。

复杂场景测试见 `yuzhousha/scenario_test.go`；冒烟/AI 自对弈见 `smoke_test.go`、`sim_test.go`。

```bash
cd backend
./scripts/test.sh smoke    # 全武将开局冒烟（快）
./scripts/test.sh sim -v   # AI 自对弈；失败见 test/yuzhousha/sim_logs/
CARD_SIM_TRACE=1 ./scripts/test.sh sim -v   # 失败日志含最近事件
```
