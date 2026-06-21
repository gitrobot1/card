package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

func init() {
	registerWeiSkills()
}

func registerWeiSkills() {
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDFankui, Name: "反馈", Kind: skill.KindPassive,
			Desc: "当你受到1点伤害后，你可以获得伤害来源的一张牌。",
		},
		CanActivate: fankuiCanActivate,
		Activate:    fankuiActivate,
		AIPriority:  fankuiAIPriority,
		AIActivate:  fankuiAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDGuicai, Name: "鬼才", Kind: skill.KindActive,
			Desc: "在任意判定牌生效前，你可以打出一张手牌代替之。",
		},
		CanActivate: guicaiCanActivate,
		Activate:    guicaiActivate,
		AIPriority:  guicaiAIPriority,
		AIActivate:  guicaiAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLuoshen, Name: "洛神", Kind: skill.KindActive,
			Desc: "准备阶段，你可以进行判定：黑色则获得该牌并可以再次判定，红色则结束。",
		},
		PreparePhase: skill.PreparePhaseDecl{
			Offer: func(r skill.Runtime, seat int) bool {
				return r.HasSkill(seat, skill.IDLuoshen) && r.DrawPileCount() > 0
			},
		},
		CanActivate: luoshenCanActivate,
		Activate:    luoshenActivate,
		AIPriority:  luoshenAIPriority,
		AIActivate:  luoshenAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDJianxiong, Name: "奸雄", Kind: skill.KindPassive,
			Desc: "当你受到伤害后，你可以获得对你造成伤害的牌。",
		},
		CanActivate: jianxiongCanActivate,
		Activate:    jianxiongActivate,
		AIPriority:  jianxiongAIPriority,
		AIActivate:  jianxiongAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDGanglie, Name: "刚烈", Kind: skill.KindPassive,
			Desc: "当你受到1点伤害后，你可以进行判定：若结果不为红桃，伤害来源弃2张手牌或受到1点伤害。",
		},
		CanActivate: ganglieCanActivate,
		Activate:    ganglieActivate,
		AIPriority:  ganglieAIPriority,
		AIActivate:  ganglieAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDHujia, Name: "护驾", Kind: skill.KindLord,
			Desc: "主公技，当你需要使用或打出【闪】时，你可以令其他魏势力角色选择是否打出一张【闪】。",
		},
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDLuoyi, Name: "裸衣", Kind: skill.KindActive,
			Desc: "摸牌阶段，你可以放弃摸牌，若如此做，你于此回合内使用【杀】和【决斗】造成的伤害+1。",
		},
		CanActivate: luoyiCanActivate,
		Activate:    luoyiActivate,
		AIPriority:  luoyiAIPriority,
		AIActivate:  luoyiAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDTuxi, Name: "突袭", Kind: skill.KindActive,
			Desc: "摸牌阶段，你可以放弃摸牌，然后从至多2名对手中各获得一张牌。",
		},
		CanActivate: tuxiCanActivate,
		Activate:    tuxiActivate,
		AIPriority:  tuxiAIPriority,
		AIActivate:  tuxiAIActivate,
	})

	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDYiji, Name: "遗计", Kind: skill.KindPassive,
			Desc: "当你受到伤害后，你可以摸2张牌，然后将至多2张手牌交给其他角色。",
		},
		CanActivate: yijiCanActivate,
		Activate:    yijiActivate,
		AIPriority:  yijiAIPriority,
		AIActivate:  yijiAIActivate,
	})
}

func fankuiCanActivate(r skill.Runtime, seat int) bool {
	return r.PendingFankuiFor(seat)
}

func fankuiActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	_ = req.TargetIndex
	return r.FankuiTakeFrom(seat, req.TargetZone, req.TargetCardID)
}

func fankuiAIPriority(r skill.Runtime, seat int) int {
	if fankuiCanActivate(r, seat) {
		return 88
	}
	return 0
}

func fankuiAIActivate(r skill.Runtime, seat int) error {
	source := r.FankuiSourceSeat(seat)
	if source < 0 {
		return r.PassFankui(seat)
	}
	gr, ok := r.(*gameSkillRuntime)
	if !ok {
		if id := r.FirstTakeableCardID(source); id != "" {
			return r.FankuiTakeFrom(seat, "hand", id)
		}
		return r.PassFankui(seat)
	}
	p := &gr.g.Players[source]
	if len(p.Hand) > 0 {
		return r.FankuiTakeFrom(seat, "hand", p.Hand[0].ID)
	}
	zone, id := aiPickTakeTarget(gr.g, source)
	if id != "" {
		return r.FankuiTakeFrom(seat, zone, id)
	}
	return r.PassFankui(seat)
}

func guicaiCanActivate(r skill.Runtime, seat int) bool {
	return r.PendingGuicaiFor(seat)
}

func guicaiActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	return r.ApplyGuicaiReplace(seat, req.CardIDs[0])
}

func guicaiAIPriority(r skill.Runtime, seat int) int {
	if guicaiCanActivate(r, seat) {
		return 75
	}
	return 0
}

func guicaiAIActivate(r skill.Runtime, seat int) error {
	// 刚烈判定时，司马懿不应发动鬼才（刚烈是队友技能，改判红桃会导致刚烈失败，损人不利己）
	if r.PendingJudgeReason() == string(skill.JudgeGanglie) {
		return r.PassGuicai(seat)
	}
	ids := r.PlayerHandCardIDs(seat)
	if len(ids) == 0 {
		return r.PassGuicai(seat)
	}
	return r.ApplyGuicaiReplace(seat, ids[0])
}

func luoshenCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLuoshen) {
		return false
	}
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPrepare && r.CurrentTurn() == seat {
		return r.DrawPileCount() > 0
	}
	return false
}

func luoshenActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.StartLuoshen(seat)
}

func luoshenAIPriority(r skill.Runtime, seat int) int {
	if !luoshenCanActivate(r, seat) {
		return 0
	}
	if r.PlayerHandCount(seat) >= 8 {
		return 0
	}
	return 85
}

func luoshenAIActivate(r skill.Runtime, seat int) error {
	return r.StartLuoshen(seat)
}

func jianxiongCanActivate(r skill.Runtime, seat int) bool {
	return r.PendingJianxiongFor(seat)
}

func jianxiongActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.ApplyJianxiong(seat)
}

func jianxiongAIPriority(r skill.Runtime, seat int) int {
	if jianxiongCanActivate(r, seat) {
		return 90
	}
	return 0
}

func jianxiongAIActivate(r skill.Runtime, seat int) error {
	return r.ApplyJianxiong(seat)
}

func ganglieCanActivate(r skill.Runtime, seat int) bool {
	if r.PendingGanglieOfferFor(seat) {
		return true
	}
	return r.PendingGanglieChoiceFor(seat)
}

func ganglieActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if r.PendingGanglieOfferFor(seat) {
		return r.StartGanglieJudge(seat)
	}
	if r.PendingGanglieChoiceFor(seat) {
		if len(req.CardIDs) >= 2 {
			return r.GanglieDiscard(seat, req.CardIDs[:2])
		}
		if req.TargetZone == "take_damage" {
			return r.GanglieTakeDamage(seat)
		}
		if r.PlayerHandCount(seat) >= 2 {
			return r.GanglieDiscard(seat, r.PlayerHandCardIDs(seat)[:2])
		}
		return r.GanglieTakeDamage(seat)
	}
	return ErrWrongPhase
}

func ganglieAIPriority(r skill.Runtime, seat int) int {
	if ganglieCanActivate(r, seat) {
		return 87
	}
	return 0
}

func ganglieAIActivate(r skill.Runtime, seat int) error {
	return ganglieActivate(r, seat, skill.ActivateReq{})
}

func luoyiCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLuoyi) {
		return false
	}
	return r.PendingDrawPhaseChoiceFor(seat)
}

func luoyiActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.ActivateLuoyi(seat)
}

func luoyiAIPriority(r skill.Runtime, seat int) int {
	if !luoyiCanActivate(r, seat) {
		return 0
	}
	if r.HandPlaysAs(seat, CardSha) {
		return 75
	}
	return 0
}

func luoyiAIActivate(r skill.Runtime, seat int) error {
	return r.ActivateLuoyi(seat)
}

func tuxiCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDTuxi) {
		return false
	}
	if r.PendingTuxiTakeFor(seat) {
		return true
	}
	return r.PendingDrawPhaseChoiceFor(seat) && r.OpponentHasTakeableCard(seat)
}

func tuxiActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if r.PendingTuxiTakeFor(seat) {
		return r.TuxiTakeFrom(seat, req.TargetZone, req.TargetCardID)
	}
	// 新突袭：放弃摸牌，从至多2名对手中各获得一张牌
	return r.StartTuxi(seat)
}

func tuxiAIPriority(r skill.Runtime, seat int) int {
	if r.PendingTuxiTakeFor(seat) {
		return 90
	}
	if !r.PendingDrawPhaseChoiceFor(seat) || !r.OpponentHasTakeableCard(seat) {
		return 0
	}
	opp := r.OpponentOf(seat)
	if r.PlayerHandCount(opp) >= 4 {
		return 68
	}
	if r.PlayerHandCount(opp) >= 2 {
		return 58
	}
	return 0
}

func tuxiAIActivate(r skill.Runtime, seat int) error {
	if r.PendingTuxiTakeFor(seat) {
		source := r.TuxiSourceSeat(seat)
		if source < 0 {
			return r.PassTuxi(seat)
		}
		zone, cardID := r.BestTakeTarget(source)
		if zone == "" {
			return r.PassTuxi(seat)
		}
		return r.TuxiTakeFrom(seat, zone, cardID)
	}
	return r.StartTuxi(seat)
}

func yijiCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDYiji) {
		return false
	}
	return r.PendingYijiOfferFor(seat) || r.PendingYijiGiveFor(seat)
}

func yijiActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if r.PendingYijiOfferFor(seat) {
		return r.ApplyYiji(seat)
	}
	if r.PendingYijiGiveFor(seat) {
		target := req.TargetIndex
		if target < 0 {
			target = r.OpponentOf(seat)
		}
		return r.YijiGiveCards(seat, target, req.CardIDs)
	}
	return ErrWrongPhase
}

func yijiAIPriority(r skill.Runtime, seat int) int {
	if yijiCanActivate(r, seat) {
		return 88
	}
	return 0
}

func yijiAIActivate(r skill.Runtime, seat int) error {
	if r.PendingYijiOfferFor(seat) {
		return r.ApplyYiji(seat)
	}
	if r.PendingYijiGiveFor(seat) {
		target, ids := aiPickYijiGiveRuntime(r, seat)
		return r.YijiGiveCards(seat, target, ids)
	}
	return nil
}

func aiPickYijiGiveRuntime(r skill.Runtime, seat int) (target int, ids []string) {
	target = r.OpponentOf(seat)
	hand := r.PlayerHandCardIDs(seat)
	if len(hand) == 0 || len(hand) <= 4 {
		return target, nil
	}
	give := 2
	if len(hand)-2 < give {
		give = len(hand) - 2
	}
	if give <= 0 {
		return target, nil
	}
	return target, hand[len(hand)-give:]
}
