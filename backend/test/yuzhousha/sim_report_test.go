package engine_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

var simLogsAbsDir = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("test", "yuzhousha", "sim_logs")
	}
	return filepath.Join(filepath.Dir(file), "sim_logs")
}()

const simLogDir = "test/yuzhousha/sim_logs"

type simContext struct {
	Label  string
	Mode   string // "1v1" | "2v2" | "3p_chain" | "3p_ddz"
	Heroes []string
	Hero0  string
	Hero1  string
	Hero2  string
	Hero3  string
	Seed   int64
	Reason string // stuck | timeout | card_loss | force_error | no_winner
}

func (ctx simContext) is2v2() bool {
	return ctx.Mode == "2v2" || len(ctx.Heroes) == 4
}

func (ctx simContext) is3p() bool {
	return ctx.Mode == "3p_chain" || ctx.Mode == "3p_ddz" || len(ctx.Heroes) == 3
}

type simRun struct {
	result     simResult
	lastEvents []engine.GameEvent
	stuckAtFP  string
	forceErr   string
}

func (ctx simContext) matchup() string {
	if ctx.Label != "" {
		return ctx.Label
	}
	if ctx.is2v2() && len(ctx.Heroes) == 4 {
		return fmt.Sprintf("%s vs %s vs %s vs %s", ctx.Heroes[0], ctx.Heroes[1], ctx.Heroes[2], ctx.Heroes[3])
	}
	if ctx.is3p() && len(ctx.Heroes) == 3 {
		return fmt.Sprintf("%s vs %s vs %s", ctx.Heroes[0], ctx.Heroes[1], ctx.Heroes[2])
	}
	if ctx.Seed > 0 {
		return fmt.Sprintf("seed %d (%s vs %s)", ctx.Seed, ctx.Hero0, ctx.Hero1)
	}
	return fmt.Sprintf("%s vs %s", ctx.Hero0, ctx.Hero1)
}

func simProblemHint(g *engine.Game) (category, hint string) {
	if g.Phase == engine.PhaseFinished {
		return "牌堆守恒", "engine/play.go、tricks.go、skill_*.go — 局结束时牌总数不对，某处多弃/少回收"
	}
	if g.Pending == nil {
		switch g.Phase {
		case engine.PhaseResponse:
			return "响应阶段", "engine/response.go — phase=response 但 Pending 为空，状态机未正确收尾"
		case engine.PhasePlaying:
			if g.TurnStep == engine.StepPlay {
				return "出牌阶段", "engine/ai.go runAIPlayPhase — AI 可能无法出牌也未 EndPlay（空城/距离/技能冲突）"
			}
			return "回合流程", fmt.Sprintf("engine/turn.go — step=%s 可能未推进", g.TurnStep)
		default:
			return "未知", fmt.Sprintf("phase=%s step=%s", g.Phase, g.TurnStep)
		}
	}
	p := g.Pending
	mode := p.ResponseMode
	switch mode {
	case engine.ResponseModeSkillJianxiong:
		return "伤害链", "engine/skill_jianxiong.go + skill_damage.go — 奸雄窗口"
	case engine.ResponseModeSkillGanglieOffer, engine.ResponseModeSkillGanglieChoice:
		return "伤害链", "engine/skill_ganglie.go + ai.go — 刚烈判定/弃牌选择"
	case engine.ResponseModeSkillFankui:
		return "伤害链", "engine/skill_fankui.go — 反馈拿牌"
	case engine.ResponseModeSkillGuicai:
		return "判定", "engine/skill_judge.go — 鬼才改判"
	case engine.ResponseModeQilinBow:
		return "装备", "engine/weapons.go — 麒麟弓弃马"
	case engine.ResponseModeGuanYuFollow:
		return "装备", "engine/weapons.go — 青龙偃月跟进"
	case engine.ResponseModeWuguPick:
		return "锦囊", "engine/tricks.go + ai.go — 五谷丰登选牌（检查 WuguPickSeat）"
	case engine.ResponseModePeekDeck:
		return "阶段技", "engine/phase_prepare.go — 观星/洛神看牌堆"
	case engine.ResponseModeWuxiekTrick, engine.ResponseModeWuxiekLebu,
		engine.ResponseModeWuxiekBingliang, engine.ResponseModeWuxiekShandian:
		return "无懈", "engine/response.go + tricks.go — 无懈可击窗口"
	case engine.ResponseModeDdzJudgeCancel:
		return "斗地主判定", "engine/skill_ddz.go + ai.go — 地主弃2张取消判定"
	case engine.ResponseModeSkillJijiang:
		return "主公技", "engine/skill_register.go — 激将（1v1 需检查 ShuAllies）"
	case engine.ResponseModeDying:
		return "濒死救援", "engine/skill_dying.go + ai.go — 2v2 队友出桃；检查 nextDyingAskSeat"
	default:
		if p.RequiredKind == engine.CardSha {
			return "AOE/杀响应", "engine/response.go + play.go — 需出【杀】（南蛮/激将/决斗链）"
		}
		if p.RequiredKind == engine.CardShan || p.RequiredKind == "" {
			return "杀/闪响应", "engine/response.go — 需出【闪】或八卦判定"
		}
		return "响应窗", fmt.Sprintf("engine/response.go — mode=%q required=%q", mode, p.RequiredKind)
	}
}

func formatPlayerState(g *engine.Game, i int) string {
	p := g.Players[i]
	skills := strings.Join(p.Character.SkillIDs, ",")
	eq := []string{}
	if p.Weapon != nil {
		eq = append(eq, "武器:"+p.Weapon.Name)
	}
	if p.Armor != nil {
		eq = append(eq, "防具:"+p.Armor.Name)
	}
	if p.PlusHorse != nil {
		eq = append(eq, "+1:"+p.PlusHorse.Name)
	}
	if p.MinusHorse != nil {
		eq = append(eq, "-1:"+p.MinusHorse.Name)
	}
	judge := ""
	if len(p.JudgeArea) > 0 {
		parts := make([]string, len(p.JudgeArea))
		for j, c := range p.JudgeArea {
			parts[j] = c.Name
		}
		judge = " 判定区:" + strings.Join(parts, ",")
	}
	return fmt.Sprintf("  [%d] %s (%s) HP=%d/%d 手牌=%d 技能=[%s] %s%s",
		i, p.Name, p.Character.ID, p.HP, p.MaxHP, len(p.Hand), skills,
		strings.Join(eq, " "), judge)
}

func formatPending(p *engine.PendingCombat) string {
	if p == nil {
		return "  (nil)"
	}
	return fmt.Sprintf(`  mode=%q required=%q src=%d tgt=%d return=%d effect=%d
  card=%s damage=%d wuxiek=%v bagua=%v ignoreArmor=%v tieqi=%v unblock=%v
  wuguPickSeat=%d revealed=%d msg_card=%+v`,
		p.ResponseMode, p.RequiredKind, p.SourceIndex, p.TargetIndex, p.ReturnIndex, p.EffectTarget,
		p.Card.Name, p.Damage, p.AllowWuxiek, p.BaguaUsed, p.IgnoreArmor, p.TieqiPending, p.ShaUnblockable,
		p.WuguPickSeat, len(p.RevealedCards), p.Card)
}

func formatEvents(events []engine.GameEvent) string {
	if len(events) == 0 {
		return "  (无记录 — 用 CARD_SIM_TRACE=1 开启逐步事件)\n"
	}
	var b strings.Builder
	for _, e := range events {
		card := ""
		if e.Card != nil {
			card = " card=" + e.Card.Name
		}
		fmt.Fprintf(&b, "  [%s] p%d→t%d dmg=%d heal=%d skill=%s%s %s\n",
			e.Type, e.PlayerIndex, e.TargetIndex, e.Damage, e.Heal, e.SkillID, card, e.Message)
	}
	return b.String()
}

func buildSimReport(g *engine.Game, ctx simContext, run simRun) string {
	cat, hint := simProblemHint(g)
	cards := countCardsInPlay(g)
	var b strings.Builder
	fmt.Fprintf(&b, "=== 宇宙杀 AI 模拟失败报告 ===\n")
	fmt.Fprintf(&b, "时间: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&b, "对局: %s\n", ctx.matchup())
	if ctx.is2v2() {
		fmt.Fprintf(&b, "模式: 2v2\n")
	} else if ctx.is3p() {
		fmt.Fprintf(&b, "模式: %s\n", ctx.Mode)
	}
	fmt.Fprintf(&b, "失败类型: %s\n", ctx.Reason)
	fmt.Fprintf(&b, "步数: %d / %d\n", run.result.steps, defaultSimMaxSteps)
	if run.stuckAtFP != "" {
		fmt.Fprintf(&b, "卡住指纹: %s\n", run.stuckAtFP)
	}
	if run.forceErr != "" {
		fmt.Fprintf(&b, "forceProgress 错误: %s\n", run.forceErr)
	}
	fmt.Fprintf(&b, "\n--- 可能问题区域 ---\n")
	fmt.Fprintf(&b, "分类: %s\n", cat)
	fmt.Fprintf(&b, "建议查: %s\n", hint)
	fmt.Fprintf(&b, "\n--- 局面 ---\n")
	fmt.Fprintf(&b, "phase=%s step=%s turn=%d message=%q\n", g.Phase, g.TurnStep, g.CurrentTurn, g.Message)
	fmt.Fprintf(&b, "牌堆=%d 弃牌=%d 牌总数=%d (期望 %d", len(g.DrawPile), len(g.DiscardPile), cards, expectedDeckSize)
	if cards != expectedDeckSize {
		b.WriteString(" ⚠ 牌数不守恒")
	}
	b.WriteString(")\n")
	for i := range g.Players {
		fmt.Fprintf(&b, "%s\n", formatPlayerState(g, i))
	}
	fmt.Fprintf(&b, "\n--- Pending ---\n")
	fmt.Fprintf(&b, "%s\n", formatPending(g.Pending))
	fmt.Fprintf(&b, "\n--- 最近事件 (最多 25 条) ---\n")
	fmt.Fprintf(&b, "%s", formatEvents(run.lastEvents))
	fmt.Fprintf(&b, "\n--- 复现 ---\n")
	switch {
	case ctx.Mode == "3p_chain" && ctx.Seed > 0:
		fmt.Fprintf(&b, "  CARD_SIM=1 CARD_SIM_ROUNDS=%d ./scripts/test.sh sim3p_chain -run TestSim_3pChain_RandomTriosSeeded/%d -v\n", ctx.Seed, ctx.Seed)
	case ctx.Mode == "3p_chain" && ctx.Hero0 != "":
		fmt.Fprintf(&b, "  CARD_SIM=1 ./scripts/test.sh sim3p_chain -run TestSim_3pChain_AllHeroesAsSeat0/%s -v\n", sanitizeLogName(ctx.Hero0))
	case ctx.Mode == "3p_ddz" && ctx.Seed > 0:
		fmt.Fprintf(&b, "  CARD_SIM=1 CARD_SIM_ROUNDS=%d ./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_RandomTriosSeeded/%d -v\n", ctx.Seed, ctx.Seed)
	case ctx.Mode == "3p_ddz" && ctx.Hero0 != "":
		fmt.Fprintf(&b, "  CARD_SIM=1 ./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_AllHeroesAsSeat0/%s -v\n", sanitizeLogName(ctx.Hero0))
	case ctx.is2v2() && ctx.Seed > 0:
		fmt.Fprintf(&b, "  CARD_SIM=1 CARD_SIM_ROUNDS=%d ./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded/%d -v\n", ctx.Seed, ctx.Seed)
	case ctx.is2v2() && ctx.Hero0 != "":
		fmt.Fprintf(&b, "  CARD_SIM=1 ./scripts/test.sh sim2v2 -run TestSim_2v2_AllHeroesAsSeat0/%s -v\n", sanitizeLogName(ctx.Hero0))
	case ctx.Seed > 0:
		fmt.Fprintf(&b, "  CARD_SIM=1 CARD_SIM_ROUNDS=%d ./scripts/test.sh sim -run TestSim_RandomHeroMixSeeded/%d -v\n", ctx.Seed, ctx.Seed)
	default:
		fmt.Fprintf(&b, "  CARD_SIM=1 ./scripts/test.sh sim -run TestSim_AllHeroPairsAIVsAI/%s -v\n", sanitizeLogName(ctx.matchup()))
	}
	return b.String()
}

func sanitizeLogName(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(s, "_")
	return s
}

func writeSimReport(report string, baseName string) (string, error) {
	dir := simLogsAbsDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := sanitizeLogName(baseName) + ".log"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		return "", err
	}
	// 追加到汇总文件，方便一次 sim 跑完扫一眼
	summary := filepath.Join(dir, "failures-summary.log")
	f, err := os.OpenFile(summary, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = fmt.Fprintf(f, "\n%s\n%s\n", strings.Repeat("-", 60), report)
		_ = f.Close()
	}
	return path, nil
}

func emitSimReport(t *testing.T, g *engine.Game, ctx simContext, run simRun, level string) {
	t.Helper()
	report := buildSimReport(g, ctx, run)
	path, err := writeSimReport(report, ctx.matchup())
	if err != nil {
		t.Logf("write sim log: %v", err)
	} else {
		t.Logf("sim 日志已写入: %s", path)
		t.Logf("汇总: test/yuzhousha/sim_logs/failures-summary.log")
	}
	cat, hint := simProblemHint(g)
	msg := fmt.Sprintf("sim [%s] %s | %s | %s | 步数=%d | 详见 test/yuzhousha/sim_logs/",
		ctx.Reason, ctx.matchup(), cat, hint, run.result.steps)
	if level == "error" {
		t.Error(msg)
	} else {
		t.Log("警告: " + msg)
	}
}

func reportSimFailure(t *testing.T, g *engine.Game, ctx simContext, run simRun) {
	emitSimReport(t, g, ctx, run, "error")
}

func assertSimFinished(t *testing.T, g *engine.Game, ctx simContext, run simRun) {
	t.Helper()
	if run.forceErr != "" {
		ctx.Reason = "force_error"
		reportSimFailure(t, g, ctx, run)
		t.FailNow()
	}
	if run.result.stuck {
		ctx.Reason = "stuck"
		reportSimFailure(t, g, ctx, run)
		t.FailNow()
	}
	if !run.result.finished {
		ctx.Reason = "timeout"
		reportSimFailure(t, g, ctx, run)
		t.FailNow()
	}
	if g.WinnerIndex == nil {
		ctx.Reason = "no_winner"
		reportSimFailure(t, g, ctx, run)
		t.FailNow()
	}
	if cards := countCardsInPlay(g); cards != expectedDeckSize {
		ctx.Reason = "card_loss"
		if os.Getenv("CARD_SIM_STRICT") == "1" {
			reportSimFailure(t, g, ctx, run)
			t.FailNow()
		}
		emitSimReport(t, g, ctx, run, "warn")
		return
	}
}
