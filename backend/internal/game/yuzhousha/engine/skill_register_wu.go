package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

func init() {
	registerWuSkills()
}

func registerWuSkills() {
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDZhiheng, Name: "制衡", Kind: skill.KindActive,
			Desc: "出牌阶段限一次，你可以弃置任意张牌，然后摸等量的牌。",
		},
		CanActivate: zhihengCanActivate,
		Activate:    zhihengActivate,
		AIPriority:  zhihengAIPriority,
		AIActivate:  zhihengAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDJiuyuan, Name: "救援", Kind: skill.KindLord,
			Desc: "主公技，当需要使用或打出【桃】时，你可以令其他吴势力角色选择是否打出一张【桃】。",
		},
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDJieyin, Name: "结姻", Kind: skill.KindActive,
			Desc: "出牌阶段限一次，你可以弃置2张手牌，若对方体力比你少且已受伤，则你与其各回复1点体力。",
		},
		CanActivate: jieyinCanActivate,
		Activate:    jieyinActivate,
		AIPriority:  jieyinAIPriority,
		AIActivate:  jieyinAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDFanjian, Name: "反间", Kind: skill.KindActive,
			Desc: "出牌阶段限一次，你可以将一张手牌交给对手，其选择一种花色并展示该牌；若猜中则受到1点伤害。",
		},
		CanActivate: fanjianCanActivate,
		Activate:    fanjianActivate,
		AIPriority:  fanjianAIPriority,
		AIActivate:  fanjianAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDTianxiang, Name: "天香", Kind: skill.KindPassive,
			Desc: "当你受到伤害时，你可以弃置一张红桃手牌并选择一名其他角色。若如此做，你将此伤害转移给该角色，然后其摸X张牌（X为其已损失体力值）。",
		},
		CanActivate: tianxiangCanActivate,
		Activate:    tianxiangActivate,
		AIPriority:  tianxiangAIPriority,
		AIActivate:  tianxiangAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDQixi, Name: "奇袭", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以将一张黑色的牌当过河拆桥打出，包括装备区的牌。",
		},
		CanActivate: qixiCanActivate,
		Activate:    qixiActivate,
		CardPlaysAs: qixiCardPlaysAs,
		AIPriority:  qixiAIPriority,
		AIActivate:  qixiAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDYinghun, Name: "英魂", Kind: skill.KindActive,
			Desc: "准备阶段，若你已受伤，你可以选择一项：1.令对手摸X张牌，然后其弃置一张牌；2.令对手摸一张牌，然后其弃置X张牌（X为你已损失的体力值）。",
		},
		PreparePhase: skill.PreparePhaseDecl{
			Offer: func(r skill.Runtime, seat int) bool {
				return r.HasSkill(seat, skill.IDYinghun)
			},
		},
		CanActivate: yinghunCanActivate,
		Activate:    yinghunActivate,
		AIPriority:  yinghunAIPriority,
		AIActivate:  yinghunAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDGuose, Name: "国色", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以将一张方块牌当【乐不思蜀】使用。",
		},
		CanActivate: guoseCanActivate,
		Activate:    guoseActivate,
		AIPriority:  guoseAIPriority,
		AIActivate:  guoseAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLiuli, Name: "流离", Kind: skill.KindPassive,
			Desc: "当你成为【杀】的目标时，你可以弃置一张牌，将此【杀】转移给攻击范围内的另一名其他角色。",
		},
		CanActivate: liuliCanActivate,
		Activate:    liuliActivate,
		AIPriority:  liuliAIPriority,
		AIActivate:  liuliAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDKurou, Name: "苦肉", Kind: skill.KindActive,
			Desc: "出牌阶段，你可以失去 1 点体力，然后摸两张牌。",
		},
		CanActivate: kurouCanActivate,
		Activate:    kurouActivate,
		AIPriority:  kurouAIPriority,
		AIActivate:  kurouAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDPojun, Name: "破军", Kind: skill.KindActive,
			Desc: "当你使用【杀】指定一个目标后，你可以将其至多X张牌扣置于该角色的武将牌旁（X为其体力值），若如此做，当前回合结束后，该角色获得这些牌。当你使用【杀】对手牌数与装备数均不大于你的角色造成伤害时，此伤害+1。",
		},
		CanActivate: pojunCanActivate,
		Activate:    pojunActivate,
		AIPriority:  pojunAIPriority,
		AIActivate:  pojunAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDHunzi, Name: "魂姿", Kind: skill.KindAwakening,
			Desc: "觉醒技，准备阶段，若你的体力值不大于 1，你减 1 点体力上限，并获得技能「英姿」和「英魂」。",
		},
		PreparePhase: skill.PreparePhaseDecl{
			Offer: func(r skill.Runtime, seat int) bool {
				if !r.HasSkill(seat, skill.IDHunzi) {
					return false
				}
				hp, _ := r.PlayerHP(seat)
				return hp <= 1
			},
		},
		CanActivate: hunziCanActivate,
		Activate:    hunziActivate,
		AIPriority:  hunziAIPriority,
		AIActivate:  hunziAIActivate,
	})
}

func zhihengCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDZhiheng) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	return r.SkillCounter(seat, counterZhihengUsed) == 0 && r.PlayerHandCount(seat) > 0
}

func zhihengActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	return r.ActivateZhiheng(seat, req.CardIDs)
}

func zhihengAIPriority(r skill.Runtime, seat int) int {
	if !zhihengCanActivate(r, seat) {
		return 0
	}
	n := r.PlayerHandCount(seat)
	hp, _ := r.PlayerHP(seat)
	// 手牌明显超过体力上限时再制衡，避免每回合空换手牌拖长对局
	if n > hp+2 {
		return 55 + n
	}
	if n >= 8 {
		return 50
	}
	return 0
}

func zhihengAIActivate(r skill.Runtime, seat int) error {
	ids := r.PlayerHandCardIDs(seat)
	if len(ids) == 0 {
		return nil
	}
	hp, _ := r.PlayerHP(seat)
	discard := len(ids) - hp
	if discard < 1 {
		discard = 1
	}
	if discard > 2 {
		discard = 2
	}
	if discard > len(ids) {
		discard = len(ids)
	}
	return r.ActivateZhiheng(seat, ids[len(ids)-discard:])
}

func jieyinCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDJieyin) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	if r.SkillCounter(seat, counterJieyinUsed) > 0 || r.PlayerHandCount(seat) < 2 {
		return false
	}
	hp, maxHP := r.PlayerHP(seat)
	if hp >= maxHP {
		return false
	}
	return canJieyinTargetRuntime(r, seat, r.OpponentOf(seat))
}

func canJieyinTargetRuntime(r skill.Runtime, actor, target int) bool {
	if target < 0 {
		return false
	}
	hp, _ := r.PlayerHP(actor)
	thp, tmax := r.PlayerHP(target)
	return thp < hp && thp < tmax
}

func jieyinActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	target := req.TargetIndex
	if target < 0 {
		target = r.OpponentOf(seat)
	}
	if len(req.CardIDs) != 2 {
		return ErrInvalidCard
	}
	return r.ActivateJieyin(seat, target, req.CardIDs)
}

func jieyinAIPriority(r skill.Runtime, seat int) int {
	if jieyinCanActivate(r, seat) {
		return 68
	}
	return 0
}

func jieyinAIActivate(r skill.Runtime, seat int) error {
	target := r.OpponentOf(seat)
	ids := r.PlayerHandCardIDs(seat)
	if len(ids) < 2 {
		return nil
	}
	return r.ActivateJieyin(seat, target, ids[len(ids)-2:])
}

func fanjianCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDFanjian) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	return r.SkillCounter(seat, counterFanjianUsed) == 0 && r.PlayerHandCount(seat) > 0
}

func fanjianActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) != 1 {
		return ErrInvalidCard
	}
	return r.ActivateFanjian(seat, req.CardIDs[0])
}

func fanjianAIPriority(r skill.Runtime, seat int) int {
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return 0
	}
	if r.SkillCounter(seat, counterFanjianUsed) > 0 || r.PlayerHandCount(seat) < 2 {
		return 0
	}
	return 45
}

func fanjianAIActivate(r skill.Runtime, seat int) error {
	ids := r.PlayerHandCardIDs(seat)
	if len(ids) == 0 {
		return nil
	}
	return r.ActivateFanjian(seat, ids[len(ids)-1])
}

func tianxiangCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDTianxiang) {
		return false
	}
	return r.Phase() == PhaseResponse && r.PendingResponseMode() == ResponseModeSkillTianxiang &&
		r.PendingTargetSeat() == seat && r.HasRedHandCard(seat)
}

func tianxiangActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return r.PassTianxiang(seat)
	}
	return r.ApplyTianxiang(seat, req.CardIDs[0])
}

func tianxiangAIPriority(r skill.Runtime, seat int) int {
	if tianxiangCanActivate(r, seat) {
		return 75
	}
	return 0
}

func tianxiangAIActivate(r skill.Runtime, seat int) error {
	if !r.HasRedHandCard(seat) {
		return r.PassTianxiang(seat)
	}
	opp := r.OpponentOf(seat)
	hp, _ := r.PlayerHP(seat)
	ohp, _ := r.PlayerHP(opp)
	if ohp < hp {
		return r.PassTianxiang(seat)
	}
	for _, id := range r.PlayerHandCardIDs(seat) {
		if err := r.ApplyTianxiang(seat, id); err == nil {
			return nil
		}
	}
	return r.PassTianxiang(seat)
}

func yinghunCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDYinghun) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPrepare || r.CurrentTurn() != seat {
		return false
	}
	if r.SkillCounter(seat, counterYinghunUsed) > 0 {
		return false
	}
	// 必须已受伤才能发动
	hp, maxHP := r.PlayerHP(seat)
	return hp < maxHP
}

func yinghunActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	target := req.TargetIndex
	if target < 0 {
		target = r.OpponentOf(seat)
	}
	return r.ActivateYinghun(seat, target)
}

func yinghunAIPriority(r skill.Runtime, seat int) int {
	if yinghunCanActivate(r, seat) {
		return 72
	}
	return 0
}

func yinghunAIActivate(r skill.Runtime, seat int) error {
	return r.ActivateYinghun(seat, r.OpponentOf(seat))
}

func guoseCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDGuose) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	return r.HasDiamondHandCard(seat)
}

func guoseActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) != 1 {
		return ErrInvalidCard
	}
	target := req.TargetIndex
	if target < 0 {
		target = r.OpponentOf(seat)
	}
	return r.ActivateGuose(seat, target, req.CardIDs[0])
}

func guoseAIPriority(r skill.Runtime, seat int) int {
	if guoseCanActivate(r, seat) {
		return 58
	}
	return 0
}

func guoseAIActivate(r skill.Runtime, seat int) error {
	for _, id := range r.PlayerHandCardIDs(seat) {
		if err := r.ActivateGuose(seat, r.OpponentOf(seat), id); err == nil {
			return nil
		}
	}
	return nil
}

func liuliCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLiuli) {
		return false
	}
	return r.Phase() == PhaseResponse && r.PendingResponseMode() == ResponseModeSkillLiuli &&
		r.PendingTargetSeat() == seat && r.PlayerHandCount(seat) > 0
}

func liuliActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return r.PassLiuli(seat)
	}
	target := req.TargetIndex
	if target < 0 {
		target = r.OpponentOf(seat)
	}
	return r.ApplyLiuli(seat, req.CardIDs[0], target)
}

func liuliAIPriority(r skill.Runtime, seat int) int {
	if liuliCanActivate(r, seat) {
		return 70
	}
	return 0
}

func liuliAIActivate(r skill.Runtime, seat int) error {
	ids := r.PlayerHandCardIDs(seat)
	if len(ids) == 0 {
		return r.PassLiuli(seat)
	}
	return r.ApplyLiuli(seat, ids[len(ids)-1], r.OpponentOf(seat))
}

func kurouCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDKurou) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	hp, _ := r.PlayerHP(seat)
	return hp > 1
}

func kurouActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	_ = req
	return r.ActivateKurou(seat)
}

func kurouAIPriority(r skill.Runtime, seat int) int {
	if !kurouCanActivate(r, seat) {
		return 0
	}
	hp, _ := r.PlayerHP(seat)
	hand := r.PlayerHandCount(seat)
	if hand <= 2 && hp >= 3 {
		return 64
	}
	if hand <= 4 && hp >= 4 {
		return 52
	}
	return 0
}

func kurouAIActivate(r skill.Runtime, seat int) error {
	return r.ActivateKurou(seat)
}

func hunziCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDHunzi) {
		return false
	}
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPrepare || r.CurrentTurn() != seat {
		return false
	}
	hp, maxHP := r.PlayerHP(seat)
	return hp <= 1 && maxHP > 1
}

func hunziActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	_ = req
	return r.AwakenHunzi(seat)
}

func hunziAIPriority(r skill.Runtime, seat int) int {
	if hunziCanActivate(r, seat) {
		return 90
	}
	return 0
}

func hunziAIActivate(r skill.Runtime, seat int) error {
	return r.AwakenHunzi(seat)
}

func pojunCanActivate(r skill.Runtime, seat int) bool {
	return r.Phase() == PhaseResponse && r.PendingResponseMode() == ResponseModeSkillPojun &&
		r.PendingPojunForSource(seat) && r.HasSkill(seat, skill.IDPojun)
}

func pojunActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if req.TargetZone == "" && len(req.CardIDs) == 0 {
		return r.PassPojun(seat)
	}
	zone := req.TargetZone
	cardID := req.TargetCardID
	if len(req.CardIDs) > 0 {
		cardID = req.CardIDs[0]
	}
	return r.PojunPlace(seat, zone, cardID)
}

func pojunAIPriority(r skill.Runtime, seat int) int {
	if pojunCanActivate(r, seat) {
		return 85
	}
	return 0
}

func pojunAIActivate(r skill.Runtime, seat int) error {
	return r.AutoPojunPlacing(seat)
}
