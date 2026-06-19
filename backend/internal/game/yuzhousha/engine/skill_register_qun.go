package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func registerQunSkills() {
	// 刘焉-图射
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDTushe, Name: "图射", Kind: skill.KindPassive,
			Desc: "当你使用非装备牌指定目标后，若你没有基本牌，则你可以摸X张牌（X为此牌指定的目标数）。",
		},
		CanActivate: tusheCanActivate,
		Activate:    tusheActivate,
	})
	
	// 刘焉-立牧
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLimu, Name: "立牧", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以将一张方块牌当【乐不思蜀】对自己使用，然后回复1点体力；你的判定区有牌时，你对攻击范围内的其他角色使用牌没有次数和距离限制。",
		},
		CanActivate: limuCanActivate,
		Activate:    limuActivate,
		CardPlaysAs: limuCardPlaysAs,
		UnlimitedSha: limuUnlimitedSha,
		TrickIgnoresDistance: limuTrickIgnoresDistance,
	})
	
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDShuangxiong, Name: "双雄", Kind: skill.KindActive,
			Desc: "摸牌阶段，你可以改为亮出牌堆顶的一张牌并获得之；若如此做，本回合你可以将一张与之颜色不同的手牌当【决斗】使用。",
		},
		CanActivate: shuangxiongCanActivate,
		Activate:    shuangxiongActivate,
		AIPriority:  shuangxiongAIPriority,
		AIActivate:  shuangxiongAIActivate,
	})
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLuanwu, Name: "乱武", Kind: skill.KindLimited,
			Desc: "限定技，出牌阶段，令所有其他角色依次对除你外的一名角色使用一张【杀】，否则受到你造成的1点伤害。",
		},
		CanActivate: luanwuCanActivate,
		Activate:    luanwuActivate,
		AIPriority:  luanwuAIPriority,
		AIActivate:  luanwuAIActivate,
	})
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLeiji, Name: "雷击", Kind: skill.KindActive,
			Desc: "当你使用或打出一张【闪】时，你可以进行判定，若结果为黑色，你对一名其他角色造成2点雷电伤害。",
		},
		CanActivate: leijiCanActivate,
		Activate:    leijiActivate,
		AIPriority:  leijiAIPriority,
		AIActivate:  leijiAIActivate,
	})
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDGuidao, Name: "鬼道", Kind: skill.KindActive,
			Desc: "在任意判定牌生效前，你可以用一张黑色手牌替换之。",
		},
		CanActivate: guidaoCanActivate,
		Activate:    guidaoActivate,
		AIPriority:  guidaoAIPriority,
		AIActivate:  guidaoAIActivate,
	})
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDHuangtian, Name: "黄天", Kind: skill.KindLord,
			Desc: "主公技，其他群雄角色可以在你需要时给你一张【闪】。",
			InactiveIn1v1: true,
		},
	})
	
	// 华佗-青囊
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDQingnang, Name: "青囊", Kind: skill.KindActive,
			Desc: "出牌阶段限一次，你可以弃置一张手牌，令一名角色回复1点体力。",
		},
		CanActivate: qingnangCanActivate,
		Activate:    qingnangActivate,
	})
}

func shuangxiongCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDShuangxiong) {
		return false
	}
	if r.PendingDrawPhaseChoiceFor(seat) {
		return r.DrawPileCount() > 0
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	if r.SkillCounter(seat, counterShuangxiongActive) == 0 {
		return false
	}
	return r.HasShuangxiongJuedouCard(seat)
}

func shuangxiongActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if r.PendingDrawPhaseChoiceFor(seat) {
		return r.ActivateShuangxiongDraw(seat)
	}
	if len(req.CardIDs) != 1 {
		return ErrInvalidCard
	}
	return r.ActivateShuangxiongJuedou(seat, req.CardIDs[0])
}

func shuangxiongAIPriority(r skill.Runtime, seat int) int {
	if !shuangxiongCanActivate(r, seat) {
		return 0
	}
	if r.PendingDrawPhaseChoiceFor(seat) {
		return 52
	}
	opp := r.OpponentOf(seat)
	ohp, _ := r.PlayerHP(opp)
	if ohp <= 2 {
		return 70
	}
	return 58
}

func shuangxiongAIActivate(r skill.Runtime, seat int) error {
	if r.PendingDrawPhaseChoiceFor(seat) {
		return r.ActivateShuangxiongDraw(seat)
	}
	for _, id := range r.PlayerHandCardIDs(seat) {
		if err := r.ActivateShuangxiongJuedou(seat, id); err == nil {
			return nil
		}
	}
	return nil
}

func luanwuCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLuanwu) {
		return false
	}
	if r.SkillCounter(seat, counterLuanwuUsed) > 0 {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	return r.PendingResponseMode() == ""
}

func luanwuActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	_ = req
	return r.ActivateLuanwu(seat)
}

func luanwuAIPriority(r skill.Runtime, seat int) int {
	if !luanwuCanActivate(r, seat) {
		return 0
	}
	opp := r.OpponentOf(seat)
	ohp, _ := r.PlayerHP(opp)
	if ohp <= 2 {
		return 75
	}
	return 45
}

func luanwuAIActivate(r skill.Runtime, seat int) error {
	return r.ActivateLuanwu(seat)
}

func leijiCanActivate(r skill.Runtime, seat int) bool {
	return r.PendingLeijiOfferFor(seat)
}

func leijiActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	_ = req
	return r.StartLeijiJudge(seat)
}

func leijiAIPriority(r skill.Runtime, seat int) int {
	if leijiCanActivate(r, seat) {
		return 62
	}
	return 0
}

func leijiAIActivate(r skill.Runtime, seat int) error {
	return r.StartLeijiJudge(seat)
}

func guidaoCanActivate(r skill.Runtime, seat int) bool {
	return r.PendingGuidaoFor(seat)
}

func guidaoActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	return r.ApplyGuidaoReplace(seat, req.CardIDs[0])
}

func guidaoAIPriority(r skill.Runtime, seat int) int {
	if guidaoCanActivate(r, seat) {
		return 74
	}
	return 0
}

func guidaoAIActivate(r skill.Runtime, seat int) error {
	for _, id := range r.PlayerHandCardIDs(seat) {
		// Runtime has no per-card suit; try each card until black succeeds.
		if err := r.ApplyGuidaoReplace(seat, id); err == nil {
			return nil
		}
	}
	return r.PassGuidao(seat)
}

func qingnangCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDQingnang) {
		return false
	}
	// 出牌阶段才能激活
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	// 本回合已使用过青囊，不能再次激活
	if r.SkillCounter(seat, "qingnang_used") > 0 {
		return false
	}
	// 检查是否有手牌
	return r.PlayerHandCount(seat) > 0
}

func qingnangActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	// 弃置一张手牌
	g := r.(*gameSkillRuntime).g
	events := r.(*gameSkillRuntime).events
	idx, _, ok := g.findCard(seat, req.CardIDs[0])
	if !ok {
		return ErrInvalidCard
	}
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)
	
	// 令目标回复1点体力
	target := req.TargetIndex
	if target >= 0 && target < len(g.Players) {
		p := &g.Players[target]
		if p.HP < p.MaxHP {
			p.HP++
			*events = append(*events, GameEvent{
				Type:        "skill_heal",
				PlayerIndex: target,
				TargetIndex: seat,
				SkillID:     skill.IDQingnang,
				Message:     fmt.Sprintf("%s 对 %s 使用【青囊】，回复1点体力", g.Players[seat].Name, g.Players[target].Name),
			})
		}
	}
	
	// 标记本回合已使用过青囊
	g.setSkillCounter(seat, "qingnang_used", 1)
	g.appendSkillEvent(events, skill.IDQingnang, seat, target, g.Message)
	g.resetTimer()
	return nil
}
