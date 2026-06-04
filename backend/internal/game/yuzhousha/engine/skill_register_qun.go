package engine

import (
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func registerQunSkills() {
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
