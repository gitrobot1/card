package engine

import "fmt"

// DamageResume 伤害结算后恢复对局流程的上下文（反馈、麒麟弓等挂起后需 resume）。
type DamageResume struct {
	Mode        string
	Card        Card
	ReturnIndex int
	ResumeLuanwu bool
	LuanwuOwner  int
	LeijiResumeShan bool
	LeijiSaved      *PendingCombat
	LeijiShanSeat   int
	// OfferQilin 是否在反馈结束后尝试麒麟弓（仅杀命中且未死亡时）。
	OfferQilin bool
	// SkipTianxiang 为 true 时不再向同一受害者提供【天香】窗口（已跳过或已处理）。
	SkipTianxiang bool
	// IgnoreArmor 青釭剑等无视防具（藤甲加伤与八卦均不生效）。
	IgnoreArmor bool
}

// DamageAftermath 一次伤害事件触发的可选技能链（奸雄 → 刚烈×N → 反馈×N）。
type DamageAftermath struct {
	Source, Target int
	Card           Card
	Resume         DamageResume
	OfferJianxiong bool
	OfferYiji      bool
	GanglieLeft    int
	FankuiLeft     int
}

const (
	damageResumeShaHit    = fankuiResumeShaHit
	damageResumeLightning = fankuiResumeLightning
)

func (g *Game) initDamageAftermath(source, target, damage int, card Card, resume DamageResume) {
	if damage <= 0 {
		g.damageAftermath = nil
		return
	}
	a := &DamageAftermath{
		Source: source, Target: target, Card: card, Resume: resume,
	}
	if g.hasSkill(target, SkillJianxiong) && g.damageCardObtainable(card) {
		a.OfferJianxiong = true
	}
	if g.hasSkill(target, SkillGanglie) {
		a.GanglieLeft = damage
	}
	if g.hasSkill(target, SkillYiji) {
		a.OfferYiji = true
	}
	if g.hasSkill(target, SkillFankui) && g.hasTakeableCard(source) {
		a.FankuiLeft = damage
	}
	if !a.OfferJianxiong && !a.OfferYiji && a.GanglieLeft == 0 && a.FankuiLeft == 0 {
		g.damageAftermath = nil
		return
	}
	g.damageAftermath = a
}

// continueAfterDamage 扣血后的统一入口：死亡判定 → 伤害技能链 → 武器 follow-up → 恢复出牌。
func (g *Game) continueAfterDamage(source, target, damage int, card Card, resume DamageResume, events *[]GameEvent) bool {
	if damage <= 0 {
		return g.resumeAfterDamageNoSkill(resume, target, source, events)
	}
	g.tryJiangDraw(source, card, events)
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(source, target, damage, card, resume, events) {
			return true
		}
	}
	if g.isJueqingHarm(source) {
		return g.resumeAfterDamageNoSkill(resume, target, source, events)
	}
	g.initDamageAftermath(source, target, damage, card, resume)
	if g.damageAftermath == nil {
		return g.resumeAfterDamageNoSkill(resume, target, source, events)
	}
	return g.advanceDamageAftermath(events)
}

func (g *Game) advanceDamageAftermath(events *[]GameEvent) bool {
	a := g.damageAftermath
	if a == nil {
		return false
	}
	if g.Players[a.Target].HP <= 0 {
		if g.afterDamageApplied(a.Source, a.Target, 1, a.Card, a.Resume, events) {
			g.damageAftermath = nil
			return true
		}
	}
	if a.OfferJianxiong {
		if g.offerJianxiongWindow(a, events) {
			return true
		}
		a.OfferJianxiong = false
	}
	if a.OfferYiji {
		if g.offerYijiWindow(a, events) {
			return true
		}
		a.OfferYiji = false
	}
	for a.GanglieLeft > 0 {
		if g.offerGanglieWindow(a, events) {
			return true
		}
		a.GanglieLeft--
	}
	if a.FankuiLeft > 0 && g.hasTakeableCard(a.Source) {
		return g.offerFankuiFromAftermath(a, events)
	}
	g.damageAftermath = nil
	return g.resumeAfterDamageNoSkill(a.Resume, a.Target, a.Source, events)
}

func (g *Game) resumeAfterDamageNoSkill(resume DamageResume, target, source int, events *[]GameEvent) bool {
	if resume.ResumeLuanwu {
		_ = g.finishLuanwu(resume.LuanwuOwner, events)
		return true
	}
	if resume.LeijiResumeShan {
		_ = g.finishShanDodgeSuccess(resume.LeijiShanSeat, resume.LeijiSaved, events, "")
		return true
	}
	if resume.OfferQilin && resume.Card.Kind == CardSha && g.offerQilinBow(source, target, resume.ReturnIndex, events) {
		return true
	}
	switch resume.Mode {
	case damageResumeShaHit:
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = resume.ReturnIndex
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[resume.ReturnIndex].Name)
		g.resetTimer()
		return true
	case damageResumeLightning:
		_ = g.continueTurnAfterJudge(target, events)
		return true
	default:
		return false
	}
}

func (g *Game) finishDamageAftermathChain(events *[]GameEvent) bool {
	a := g.damageAftermath
	if a == nil {
		return false
	}
	g.damageAftermath = nil
	return g.resumeAfterDamageNoSkill(a.Resume, a.Target, a.Source, events)
}
