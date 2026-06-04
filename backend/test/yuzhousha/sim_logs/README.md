# AI 模拟失败日志

运行 `./scripts/test.sh sim` 后，失败对局会写入此目录：

- `<对局>.log` — 单次失败详情（局面、Pending、可能问题文件、复现命令）
- `failures-summary.log` — 一次 sim 跑完的所有失败汇总

可选：`CARD_SIM_TRACE=1` 在日志中附带最近 25 条游戏事件。
