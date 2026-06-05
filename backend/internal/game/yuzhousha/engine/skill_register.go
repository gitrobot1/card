package engine

import (
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func init() {
	registerComplexSkills()
	registerQunSkills()
}

func registerComplexSkills() {
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDRende, Name: "仁德", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以将任意张手牌交给其他角色；若本阶段累计给出至少 2 张，你回复 1 点体力（每阶段一次）。",
		},
		CanActivate: rendeCanActivate,
		Activate:    rendeActivate,
		AIPriority:  rendeAIPriority,
		AIActivate:  rendeAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDJijiang, Name: "激将", Kind: skill.KindLord,
			Desc: "当你需要使用或打出【杀】时，你可以令其他蜀角色代替你使用或打出【杀】。",
		},
		CanActivate: jijiangCanActivate,
		Activate:    jijiangActivate,
		AIPriority:  jijiangAIPriority,
		AIActivate:  jijiangAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDWusheng, Name: "武圣", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以将一张红色牌当【杀】使用或打出（含红色装备牌）；未发动时红色牌按原牌型使用。",
		},
		CanActivate: wushengCanActivate,
		Activate:    wushengActivate,
		CardPlaysAs: wushengCardPlaysAs,
		AIPriority:  wushengAIPriority,
		AIActivate:  wushengAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDTieqi, Name: "铁骑", Kind: skill.KindActive,
			Desc: "当你使用【杀】指定目标后，你可以进行判定；若不为红色，目标不能出【闪】。",
		},
		CanActivate: tieqiCanActivate,
		Activate:    tieqiActivate,
		AIPriority:  tieqiAIPriority,
		AIActivate:  tieqiAIActivate,
	})
}

func rendeCanActivate(r skill.Runtime, seat int) bool {
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	if !r.HasSkill(seat, skill.IDRende) {
		return false
	}
	return r.PlayerHandCount(seat) > 0
}

func rendeActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	target := req.TargetIndex
	if target == seat || target < 0 {
		return ErrInvalidTarget
	}
	return r.GiveRende(seat, target, req.CardIDs)
}

func rendeAIPriority(r skill.Runtime, seat int) int {
	if !rendeCanActivate(r, seat) {
		return 0
	}
	hp, maxHP := r.PlayerHP(seat)
	// 1v1 AI 仅在仁德回血时发动，避免互赠手牌导致模拟死循环
	if hp < maxHP && r.PlayerHandCount(seat) >= 2 && r.SkillCounter(seat, counterRendeHealed) == 0 {
		return 80
	}
	return 0
}

func rendeAIActivate(r skill.Runtime, seat int) error {
	ids := r.PlayerHandCardIDs(seat)
	give := 2
	if len(ids) < give {
		give = len(ids)
	}
	if give == 0 {
		return nil
	}
	return r.GiveRende(seat, r.OpponentOf(seat), ids[:give])
}

func jijiangCanActivate(r skill.Runtime, seat int) bool {
	if !lordSkillsActive(r.ModeID()) {
		return false
	}
	if len(r.ShuAllies(seat)) == 0 {
		return false
	}
	if r.SkillCounter(seat, counterJijiangUseFailed) > 0 {
		return false
	}
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPlay && r.CurrentTurn() == seat && r.CanUseSha(seat) {
		return true
	}
	if r.Phase() == PhaseResponse && r.PendingRequiredKind() == CardSha && r.PendingTargetSeat() == seat {
		return true
	}
	return false
}

func jijiangActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	if len(r.ShuAllies(seat)) == 0 {
		return ErrInvalidTarget
	}
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPlay {
		return r.StartJijiangForUse(seat, r.OpponentOf(seat))
	}
	if r.Phase() == PhaseResponse && r.PendingRequiredKind() == CardSha && r.PendingTargetSeat() == seat {
		return r.StartJijiangForResponse(seat)
	}
	return ErrWrongPhase
}

func jijiangAIPriority(r skill.Runtime, seat int) int {
	if !jijiangCanActivate(r, seat) {
		return 0
	}
	if r.HandPlaysAs(seat, CardSha) {
		return 0
	}
	if r.Phase() == PhaseResponse {
		return 85
	}
	return 75
}

func jijiangAIActivate(r skill.Runtime, seat int) error {
	if r.Phase() == PhaseResponse {
		return r.StartJijiangForResponse(seat)
	}
	return r.StartJijiangForUse(seat, r.OpponentOf(seat))
}

func wushengCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDWusheng) {
		return false
	}
	if r.SkillCounter(seat, counterWushengActive) > 0 {
		// 出牌阶段取消武圣由出牌区「取消武圣」按钮；响应阶段仍可在技能栏切换
		return r.Phase() == PhaseResponse && r.PendingRequiredKind() == CardSha && r.PendingTargetSeat() == seat
	}
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPlay && r.CurrentTurn() == seat && r.CanUseSha(seat) {
		return true
	}
	if r.Phase() == PhaseResponse && r.PendingRequiredKind() == CardSha && r.PendingTargetSeat() == seat {
		return true
	}
	return false
}

func wushengActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.ToggleWusheng(seat)
}

func wushengCardPlaysAs(r skill.Runtime, seat int, _, asKind, suit string) bool {
	if !r.HasSkill(seat, skill.IDWusheng) || asKind != CardSha || !skill.IsRedSuit(suit) {
		return false
	}
	return r.SkillCounter(seat, counterWushengActive) > 0
}

func wushengAIPriority(r skill.Runtime, seat int) int {
	if !wushengCanActivate(r, seat) || r.SkillCounter(seat, counterWushengActive) > 0 {
		return 0
	}
	if r.HandPlaysAs(seat, CardSha) {
		return 0
	}
	if r.Phase() == PhaseResponse && r.PendingRequiredKind() == CardSha {
		return 90
	}
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPlay && r.CanUseSha(seat) {
		return 70
	}
	return 0
}

func wushengAIActivate(r skill.Runtime, seat int) error {
	return r.ToggleWusheng(seat)
}

func tieqiCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDTieqi) {
		return false
	}
	return r.PendingTieqiForSource(seat)
}

func tieqiActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.ApplyTieqi(seat)
}

func tieqiAIPriority(r skill.Runtime, seat int) int {
	if tieqiCanActivate(r, seat) {
		return 95
	}
	return 0
}

func tieqiAIActivate(r skill.Runtime, seat int) error {
	return r.ApplyTieqi(seat)
}
