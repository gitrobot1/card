#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

GO_VERSION="$(tr -d '[:space:]' < .go-version)"
GVM_GO="$HOME/.gvm/gos/${GO_VERSION}/bin/go"

if [[ -x "$GVM_GO" ]]; then
  GO="$GVM_GO"
elif command -v go >/dev/null 2>&1; then
  GO="go"
else
  echo "Go ${GO_VERSION} not found. Install with: gvm install ${GO_VERSION}" >&2
  exit 1
fi

usage() {
  cat <<'EOF'
Usage: ./scripts/test.sh [game] [-v] [-run TestName]

Run backend game tests (requires -tags cardtest).

Games:
  all          all games (default)
  yzs          宇宙杀 (yuzhousha)
  smoke        宇宙杀冒烟（全武将开局 + hook，快）
  sim          宇宙杀 AI 自对弈（需 CARD_SIM=1，见下）
  2v2          宇宙杀 2v2 冒烟 + 模式单测（敌友/选目标/全武将开局）
  3p_chain     宇宙杀 杀上保下 冒烟 + 链式模式单测
  3p_ddz       宇宙杀 斗地主 冒烟 + 地主模式单测
  sim3p_chain  宇宙杀 3 人链式 AI 自对弈（需 CARD_SIM=1）
  sim3p_ddz    宇宙杀 3 人斗地主 AI 自对弈（需 CARD_SIM=1）
  sim2v2       宇宙杀 2v2 AI 自对弈（需 CARD_SIM=1，见下）
  simrandom    四模式随机 AI 自对弈（1v1/2v2/3p_chain/3p_ddz，需 CARD_SIM=1）
  uno          UNO
  doudizhu     斗地主
  douniu       斗牛
  zhajinhua    炸金花

Examples:
  ./scripts/test.sh
  ./scripts/test.sh smoke -v
  ./scripts/test.sh sim -v          # CARD_SIM=1，失败日志 → test/yuzhousha/sim_logs/
  CARD_SIM_TRACE=1 ./scripts/test.sh sim -v   # 附带最近 25 条事件
  CARD_SIM_ROUNDS=100 ./scripts/test.sh sim -v
  ./scripts/test.sh yzs -v
  ./scripts/test.sh yzs -run TestGanglie -v
  ./scripts/test.sh 2v2 -v
  ./scripts/test.sh 3p_chain -v
  ./scripts/test.sh 3p_ddz -v
  ./scripts/test.sh sim3p_chain -v
  ./scripts/test.sh sim3p_ddz -v
  ./scripts/test.sh sim2v2 -v
  CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v   # 四模式各 100 种子
EOF
}

SIM_ENV=()

GAME="all"
VERBOSE=""
RUN=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -v)
      VERBOSE="-v"
      shift
      ;;
    -run)
      RUN="-run"
      RUN_PATTERN="${2:?missing pattern after -run}"
      shift 2
      ;;
    all|yzs|yuzhousha|smoke|sim|simrandom|sim2v2|2v2|3p_chain|3p_ddz|sim3p_chain|sim3p_ddz|uno|doudizhu|douniu|zhajinhua)
      GAME="$1"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

case "$GAME" in
  all) PKG=(./test/...) ;;
  yzs|yuzhousha) PKG=(./test/yuzhousha/...) ;;
  smoke)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSmoke"
    ;;
  sim)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSim"
    CARD_SIM=1
    SIM_ENV=(env CARD_SIM=1)
    rm -f test/yuzhousha/sim_logs/failures-summary.log
    mkdir -p test/yuzhousha/sim_logs
    ;;
  2v2)
    PKG=(./internal/game/yuzhousha/engine/... ./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSmoke_2v2|TestDefaultEnemy|TestEnemiesOf|TestAlliesOf|TestValidPlayTargets2v2|TestValidateHeroForMode|TestLookup"
    ;;
  sim2v2)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSim_2v2"
    CARD_SIM=1
    SIM_ENV=(env CARD_SIM=1)
    rm -f test/yuzhousha/sim_logs/failures-summary.log
    mkdir -p test/yuzhousha/sim_logs
    ;;
  3p_chain)
    PKG=(./internal/game/yuzhousha/engine/... ./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSmoke_3pChain|Test3pChain|TestEvaluateHumanChain|TestValidPlayTargets3pChain|TestLookup|TestNormalizeID"
    ;;
  3p_ddz)
    PKG=(./internal/game/yuzhousha/engine/... ./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSmoke_3pDdz|TestTeamOf_3pDdz|TestIs3pDdz|TestLookup|TestNormalizeID"
    ;;
  sim3p_chain)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSim_3pChain"
    CARD_SIM=1
    SIM_ENV=(env CARD_SIM=1)
    rm -f test/yuzhousha/sim_logs/failures-summary.log
    mkdir -p test/yuzhousha/sim_logs
    ;;
  sim3p_ddz)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSim_3pDdz"
    CARD_SIM=1
    SIM_ENV=(env CARD_SIM=1)
    rm -f test/yuzhousha/sim_logs/failures-summary.log
    mkdir -p test/yuzhousha/sim_logs
    ;;
  simrandom)
    PKG=(./test/yuzhousha/...)
    RUN="-run"
    RUN_PATTERN="TestSim_RandomHeroMixSeeded|TestSim_2v2_RandomQuadsSeeded|TestSim_3pChain_RandomTriosSeeded|TestSim_3pDdz_RandomTriosSeeded"
    CARD_SIM=1
    SIM_ENV=(env CARD_SIM=1)
    rm -f test/yuzhousha/sim_logs/failures-summary.log
    mkdir -p test/yuzhousha/sim_logs
    ;;
  uno) PKG=(./test/uno/...) ;;
  doudizhu) PKG=(./test/doudizhu/...) ;;
  douniu) PKG=(./test/douniu/...) ;;
  zhajinhua) PKG=(./test/zhajinhua/...) ;;
esac

ARGS=(-tags cardtest "${PKG[@]}" -count=1)
[[ -n "$VERBOSE" ]] && ARGS+=("$VERBOSE")
[[ -n "$RUN" ]] && ARGS+=("$RUN" "$RUN_PATTERN")

if [[ ${#SIM_ENV[@]} -gt 0 ]]; then
  echo "→ ${SIM_ENV[*]} $GO test ${ARGS[*]}"
  exec "${SIM_ENV[@]}" "$GO" test "${ARGS[@]}"
else
  echo "→ $GO test ${ARGS[*]}"
  exec "$GO" test "${ARGS[@]}"
fi
