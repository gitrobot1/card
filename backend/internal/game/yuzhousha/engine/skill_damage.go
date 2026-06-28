package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

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
	// AoeResume AOE 恢复信息：伤害技能链处理完毕后继续 AOE 下一个人
	AoeResume struct {
		Source   int
		Amount   int
		Card     Card
		Rest     []int
		Active   bool
		Tiesuo   bool // true=铁索传导，false=南蛮/万箭
	}
}

// DamageSkillEntry 通用伤害技能队列条目。
// 卖血技通过 OnDamageEnd Hook 回调将自己加入 DamageAftermath.SkillQueue，
// 不再需要引擎层硬编码技能名。
type DamageSkillEntry struct {
	SkillID string                                 // 技能ID
	Left    int                                    // 剩余可执行次数（刚烈/反馈按伤害点数）
	OnOffer func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool
}

// DamageAftermath 一次伤害事件触发的技能队列。
// 技能通过 OnDamageEnd Hook 声明式入队，advanceDamageAftermath 通用排队执行。
// 执行顺序由入队顺序决定（参考 noname: arrangeTrigger 按 priority 排序）。
//
// OfferJianxiong/OfferYiji/GanglieLeft/FankuiLeft 保留为向后兼容字段，
// apply/pass 函数内部仍通过它们检查状态，后续可逐步迁移到 SkillQueue 驱动。
type DamageAftermath struct {
	Source, Target int
	Card           Card
	Resume         DamageResume
	SkillQueue     []DamageSkillEntry // 通用技能队列
	// 向后兼容字段（apply/pass 函数引用）
	OfferJianxiong bool
	OfferYiji      bool
	GanglieLeft    int
	FankuiLeft     int
}

const (
	damageResumeShaHit    = fankuiResumeShaHit
	damageResumeLightning = fankuiResumeLightning
)

// initDamageAftermath 初始化伤害技能链。
// 技能队列通过 HookDamageEnd → OnDamageEnd → enqueue 声明式填充，
// 引擎层零硬编码技能名。新增卖血技只需在 Decl 加 OnDamageEnd 回调。
func (g *Game) initDamageAftermath(source, target, damage int, card Card, resume DamageResume) {
	if damage <= 0 {
		return // 不清空旧的 DamageAftermath
	}
	a := &DamageAftermath{
		Source: source, Target: target, Card: card, Resume: resume,
	}

	// 先设置 damageAftermath（enqueue 函数需要访问 a.Card 等字段）
	oldAftermath := g.damageAftermath
	g.damageAftermath = a

	// 清空残留的 pendingDamageSkills（防止上次异常退出遗留数据）
	g.pendingDamageSkills = nil

	// 声明式收集卖血技能：通过 HookDamageEnd 广播
	// 技能在 OnDamageEnd 回调中调用 enqueueXxxSkill → 填充 g.pendingDamageSkills
	// 引擎层不知道有什么技能，全部由技能自行声明
	g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookDamageEnd, Seat: target, Role: skill.RolePlayer,
		Damage: &skill.DamageCtx{Source: source, Target: target, Amount: damage, Card: cardView(card)},
	})

	// 从 Hook 回调中收集技能队列
	if len(g.pendingDamageSkills) > 0 {
		a.SkillQueue = g.pendingDamageSkills
		g.pendingDamageSkills = nil
	}

	if len(a.SkillQueue) == 0 {
		// 无技能链的伤害（如刚烈扣血），恢复旧 DamageAftermath
		Logf("initDamageAftermath: no skills, keeping old")
		g.damageAftermath = oldAftermath
		return
	}
	Logf("initDamageAftermath: SET new Source=%d(%s) Target=%d(%s), skills=%d",
		source, g.Players[source].Name, target, g.Players[target].Name, len(a.SkillQueue))
}

// continueAfterDamage 扣血后的统一入口：濒死判定 → 铁索传导 → 伤害技能链 → 武器 follow-up → 恢复出牌。
func (g *Game) continueAfterDamage(source, target, damage int, card Card, resume DamageResume, events *[]GameEvent) bool {
	Logf("continueAfterDamage: source=%d target=%d damage=%d card.Kind=%s card.DamageType=%s target_chained=%v target_HP=%d",
		source, target, damage, card.Kind, card.DamageType, g.isChained(target), g.Players[target].HP)
	if damage <= 0 {
		return g.resumeAfterDamageNoSkill(resume, target, source, events)
	}
	g.tryJiangDraw(source, card, events)

	// 如果目标连环+属性伤害，先设置 Pending（濒死时自动保存/恢复）
	if g.isChained(target) && (card.DamageType == DamageTypeFire || card.DamageType == DamageTypeThunder) {
		chainSeats := make([]int, 0)
		for seat := range g.Players {
			if seat == target || !g.isChained(seat) || g.Players[seat].HP <= 0 {
				continue
			}
			chainSeats = append(chainSeats, seat)
		}
		g.setChained(target, false)
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  target,
			EffectTarget: target,
			Card:         card,
			Damage:       damage,
			AoeQueue:     chainSeats,
			ReturnIndex:  source,
			RequiredKind: "tiesuo",
		}
		Logf("continueAfterDamage: tiesuo setup, chainSeats=%v damage=%d", chainSeats, damage)
	}

	// 先濒死
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(source, target, damage, card, resume, events) {
			return true
		}
	}
	// 铁索传导：把 AOE 队列信息存入 resume，技能链处理后由 resumeAfterDamageNoSkill 恢复
	hasTiesuo := g.Pending != nil && g.Pending.RequiredKind == "tiesuo"
	if hasTiesuo {
		chainSeats := g.Pending.AoeQueue
		g.clearPending()
		if len(chainSeats) > 0 {
			g.setAoeResume(&resume, source, damage, card, chainSeats, true)
		}
	}

	// 伤害技能链（刚烈、反馈等）
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
	Logf("advanceDamageAftermath: Source=%d(%s) Target=%d(%s) queueLen=%d",
		a.Source, g.Players[a.Source].Name, a.Target, g.Players[a.Target].Name, len(a.SkillQueue))
	if g.Players[a.Target].HP <= 0 {
		if g.afterDamageApplied(a.Source, a.Target, 1, a.Card, a.Resume, events) {
			g.damageAftermath = nil
			return true
		}
	}
	// 通用技能队列：依次执行（由入队顺序决定优先级）
	for len(a.SkillQueue) > 0 {
		entry := &a.SkillQueue[0]
		if entry.Left <= 0 {
			a.SkillQueue = a.SkillQueue[1:]
			continue
		}
		if entry.OnOffer(g, a, entry, events) {
			return true // 等待玩家响应
		}
		a.SkillQueue = a.SkillQueue[1:]
	}
	g.damageAftermath = nil
	return g.resumeAfterDamageNoSkill(a.Resume, a.Target, a.Source, events)
}

func (g *Game) resumeAfterDamageNoSkill(resume DamageResume, target, source int, events *[]GameEvent) bool {
	// AOE 恢复：伤害技能链（刚烈等）处理完毕后，继续 AOE 下一个人
	if resume.AoeResume.Active {
		Logf("resumeAfterDamageNoSkill: AOE resume active, Tiesuo=%v Card=%s Rest=%v", resume.AoeResume.Tiesuo, resume.AoeResume.Card.Kind, resume.AoeResume.Rest)
		resume.AoeResume.Active = false
		ar := resume.AoeResume
		if ar.Tiesuo {
			// 铁索传导：继续逐人扣血
			if len(ar.Rest) > 0 {
				g.startTiesuoAoe(ar.Source, ar.Amount, ar.Card, ar.Rest, events)
			} else {
				g.finishTiesuoAoe(ar.Source, events)
			}
		} else {
			// 南蛮/万箭：继续逐人无懈窗口
			if ar.Card.Kind == CardNanMan {
				g.continueNanManAfterTarget(ar.Source, ar.Rest, events)
			} else {
				g.continueWanJianAfterTarget(ar.Source, ar.Rest, events)
			}
		}
		return true
	}
	if resume.ResumeLuanwu {
		_ = g.finishLuanwu(resume.LuanwuOwner, events)
		return true
	}
	if resume.LeijiResumeShan {
		_ = g.finishShanDodgeSuccess(resume.LeijiShanSeat, resume.LeijiSaved, events, "")
		return true
	}
	// 麒麟弓已迁移到 TagEquipSkill → OnShaHit(RoleSource) Decl Hook
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

// ============================================================================
// DamageSkillEntry.OnOffer 工厂函数（通用技能队列的 offer 适配器）
// 每个卖血技通过 OnDamageEnd Hook → 调用对应的 enqueue 函数 → 入队
// ============================================================================

// enqueueJianxiongSkill 奸雄入队：获得造成伤害的牌。
func (g *Game) enqueueJianxiongSkill(target int) {
	if !g.hasSkill(target, SkillJianxiong) {
		return
	}
	a := g.damageAftermath
	if a == nil || !g.damageCardObtainable(a.Card) {
		return
	}
	entry := DamageSkillEntry{
		SkillID: skill.IDJianxiong,
		Left:    1,
		OnOffer: func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool {
			// 设置兼容旧字段（apply/pass 函数内部检查用）
			a.OfferJianxiong = true
			return g.offerJianxiongWindow(a, events)
		},
	}
	g.pendingDamageSkills = append(g.pendingDamageSkills, entry)
}

// enqueueYijiSkill 遗计入队：摸2张牌后可将至多2张手牌交给其他角色。
func (g *Game) enqueueYijiSkill(target int) {
	if !g.hasSkill(target, SkillYiji) {
		return
	}
	// 安全检查：只在 damageAftermath 已初始化时入队
	if g.damageAftermath == nil {
		return
	}
	entry := DamageSkillEntry{
		SkillID: skill.IDYiji,
		Left:    1,
		OnOffer: func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool {
			a.OfferYiji = true
			return g.offerYijiWindow(a, events)
		},
	}
	g.pendingDamageSkills = append(g.pendingDamageSkills, entry)
}

// enqueueGanglieSkill 刚烈入队：判定→非红桃则来源弃2牌或受伤。
func (g *Game) enqueueGanglieSkill(target, damage int) {
	if !g.hasSkill(target, SkillGanglie) {
		return
	}
	a := g.damageAftermath
	if a == nil {
		return
	}
	entry := DamageSkillEntry{
		SkillID: skill.IDGanglie,
		Left:    damage,
		OnOffer: func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool {
			if entry.Left <= 0 {
				return false
			}
			// 设置兼容字段：刚烈判定时需要 GanglieLeft 追踪剩余次数
			a.GanglieLeft = entry.Left
			return g.offerGanglieWindow(a, events)
		},
	}
	g.pendingDamageSkills = append(g.pendingDamageSkills, entry)
}

// enqueueFankuiSkill 反馈入队：获得伤害来源一张牌。
func (g *Game) enqueueFankuiSkill(target, source, damage int) {
	if !g.hasSkill(target, SkillFankui) || !g.hasTakeableCard(source) {
		return
	}
	// 安全检查：只在 damageAftermath 已初始化时入队
	if g.damageAftermath == nil {
		return
	}
	entry := DamageSkillEntry{
		SkillID: skill.IDFankui,
		Left:    damage,
		OnOffer: func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool {
			if entry.Left <= 0 || !g.hasTakeableCard(a.Source) {
				return false
			}
			a.FankuiLeft = entry.Left
			return g.offerFankuiFromAftermath(a, events)
		},
	}
	g.pendingDamageSkills = append(g.pendingDamageSkills, entry)
}
