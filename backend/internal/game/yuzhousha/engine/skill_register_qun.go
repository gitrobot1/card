package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillChongzhen = "skill_chongzhen"
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
		CanActivate:    guidaoCanActivate,
		Activate:       guidaoActivate,
		AIPriority:     guidaoAIPriority,
		AIActivate:     guidaoAIActivate,
		CanModifyJudge: guidaoCanModifyJudge,
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

	// SP赵云-冲阵（声明式：攻击端 OnUseCardToTarget → TakeWindow，响应端 OnCardResolved → 拿牌）
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID: skill.IDChongzhen, Name: "冲阵", Kind: skill.KindPassive,
			Desc: "当你发动【龙胆】时，你可以获得对方的一张手牌。",
		},
		OnUseCardToTarget: chongzhenOnUseCardToTarget,
		OnCardResolved:    chongzhenOnCardResolved,
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

// guidaoCanModifyJudge 鬼道交互式改判能力声明（替代硬编码 hasSkill(SkillGuidao)）。
// 具体条件检查（黑色手牌等）由 offerNextModifyJudge 负责。
func guidaoCanModifyJudge(r skill.Runtime, seat int) (bool, string) {
	if !r.HasSkill(seat, skill.IDGuidao) {
		return false, ""
	}
	return true, skill.IDGuidao
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

// chongzhenOnCardResolved 冲阵的 OnCardResolved 回调（响应端：杀当闪打出后触发）。
// 通过 "longdan_activated" counter 检测龙胆是否被激活（而不是推断 OriginalKind）。
func chongzhenOnCardResolved(r skill.Runtime, ctx skill.CardResolvedCtx) error {
	gr := r.(*gameSkillRuntime)
	g := gr.g
	// 检查龙胆激活信号（攻击端由 OnUseCardToTarget 处理，这里只处理响应端）
	if g.getSkillCounter(ctx.Seat, "longdan_activated") <= 0 {
		return nil
	}
	g.setSkillCounter(ctx.Seat, "longdan_activated", 0) // 消耗信号
	// 只检查是否有正在进行的交互窗口（Take/Discard/Choice），普通 Respond 不跳过
	if g.Pending != nil && (g.Pending.WindowKind == WindowKindTake ||
		g.Pending.WindowKind == WindowKindDiscard || g.Pending.WindowKind == WindowKindChoice) {
		return nil
	}
	opponent := g.opponentOf(ctx.Seat)
	if opponent < 0 || len(g.Players[opponent].Hand) == 0 {
		return nil
	}
	taken := g.Players[opponent].Hand[0]
	g.Players[opponent].Hand = g.Players[opponent].Hand[1:]
	g.Players[ctx.Seat].Hand = append(g.Players[ctx.Seat].Hand, taken)
	g.SyncCounts()
	msg := fmt.Sprintf("%s 发动【冲阵】，获得 %s 的一张手牌", g.Players[ctx.Seat].Name, g.Players[opponent].Name)
	g.Message = msg
	*gr.events = append(*gr.events, GameEvent{
		Type: "chongzhen_take", PlayerIndex: ctx.Seat, TargetIndex: opponent,
		Card: &taken, SkillID: skill.IDChongzhen, Message: msg,
	})
	return nil
}

// chongzhenOnUseCardToTarget 冲阵的 OnUseCardToTarget 回调（电梯式：Hook 直接打开 TakeWindow）。
// 通过 "longdan_activated" counter 检测龙胆是否被激活（而不是推断 OriginalKind）。
func chongzhenOnUseCardToTarget(r skill.Runtime, ctx skill.UseCardCtx) error {
	gr := r.(*gameSkillRuntime)
	g := gr.g

	// 检查并消耗龙胆激活信号（必须在所有 return 之前，防止残留）
	if g.getSkillCounter(ctx.Seat, "longdan_activated") <= 0 {
		return nil
	}
	g.setSkillCounter(ctx.Seat, "longdan_activated", 0)

	pending := g.Pending
	if pending == nil || pending.Card.Kind != CardSha {
		return nil
	}
	if ctx.Seat != pending.SourceIndex {
		return nil
	}
	// 已执行过或已有窗口活跃，跳过
	if pending.ChongzhenDone || pending.WindowKind != "" {
		return nil
	}
	// 对手没有手牌则不触发
	if !g.hasTakeableCard(ctx.Target) {
		return nil
	}

	pending.ChongzhenDone = true
	return enterChongzhenTake(g, gr.events)
}

// enterChongzhenTake 打开冲阵选牌窗口（复用手顺手牵羊的 TakeWindow 模式）。
func enterChongzhenTake(g *Game, events *[]GameEvent) error {
	p := g.Pending
	if p == nil {
		return ErrWrongPhase
	}
	source := p.SourceIndex
	target := p.TargetIndex

	msg := fmt.Sprintf("%s 发动【冲阵】，获得 %s 的一张手牌", g.Players[source].Name, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDChongzhen, source, target, msg)
	return g.OpenTakeWindowOnPending(TakeWindowConfig{
		SkillID:          skill.IDChongzhen,
		ResponseMode:     ResponseModeSkillChongzhen,
		ActorSeat:        source,
		SubjectSeat:      target,
		OriginSeat:       source,
		MaxTake:          1,
		Destination:      TakeDestination{Zone: ZoneHand, Seat: source},
		Message:          msg,
		EventType:        "chongzhen_take",
		PassClosesWindow: true,
		PickTarget:       aiPickChongzhenTake,
		OnEachTake:       chongzhenOnEachTake,
		OnComplete:       chongzhenTakeComplete,
	}, events)
}

func chongzhenOnEachTake(g *Game, card Card, label string, events *[]GameEvent) error {
	p := g.Pending
	if p == nil {
		return ErrWrongPhase
	}
	source := p.ActorSeat
	target := p.SubjectSeat
	msg := fmt.Sprintf("%s 发动【冲阵】，获得 %s 的%s", g.Players[source].Name, g.Players[target].Name, label)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "chongzhen_take",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &card,
		SkillID:     skill.IDChongzhen,
		Message:     msg,
	})
	return nil
}

// chongzhenTakeComplete 冲阵选牌完成 → 清除窗口 → 回到杀流程。
func chongzhenTakeComplete(g *Game, events *[]GameEvent) error {
	p := g.Pending
	if p == nil {
		return ErrWrongPhase
	}
	p.ResponseMode = ""
	p.SkillID = ""
	FillPendingRoles(p)
	g.resetTimer()
	return g.advanceShaBeforeTargetResponse(events)
}

func aiPickChongzhenTake(g *Game, source, victim int) (zone, cardID string, ok bool) {
	if len(g.Players[victim].Hand) > 0 {
		return "hand", g.Players[victim].Hand[0].ID, true
	}
	_ = source
	return "", "", false
}
