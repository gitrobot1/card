package skill

// catalogSkills 声明式注册的被动/简单技能；复杂主动技在 engine/skill_register*.go 挂载。
func catalogSkills() []Decl {
	return []Decl{
		{
			Meta: Meta{
				ID: IDPaoxiao, Name: "咆哮", Kind: KindPassive,
				Desc: "锁定技，你使用【杀】没有次数限制。",
			},
			UnlimitedSha: func(r Runtime, seat int) bool {
				return r.HasSkill(seat, IDPaoxiao)
			},
		},
		{
			Meta: Meta{
				ID: IDLongdan, Name: "龙胆", Kind: KindPassive,
				Desc: "你可以将一张【杀】当【闪】、【闪】当【杀】使用或打出。",
			},
			CardPlaysAs: func(r Runtime, seat int, cardKind, asKind, suit string) bool {
				if !r.HasSkill(seat, IDLongdan) {
					return false
				}
				_ = suit
				return (cardKind == "shan" && asKind == "sha") || (cardKind == "sha" && asKind == "shan")
			},
		},
		{
			Meta: Meta{
				ID: IDKongcheng, Name: "空城", Kind: KindPassive,
				Desc: "锁定技，若你没有手牌，你不能成为【杀】或【决斗】的目标。",
			},
			BlocksTarget: func(r Runtime, target int, cardKind string) bool {
				if !r.HasSkill(target, IDKongcheng) {
					return false
				}
				if r.PlayerHandCount(target) > 0 {
					return false
				}
				return cardKind == "sha" || cardKind == "juedou"
			},
		},
		{
			Meta: Meta{
				ID: IDMashi, Name: "马术", Kind: KindPassive,
				Desc: "锁定技，你计算与其他角色的距离时始终-1。",
			},
			DistanceDelta: func(r Runtime, from, to int) int {
				_ = to
				if r.HasSkill(from, IDMashi) {
					return -1
				}
				return 0
			},
		},
		{
			Meta: Meta{
				ID: IDJizhi, Name: "集智", Kind: KindPassive,
				Desc: "当你使用一张非延时类锦囊牌时，你可以摸一张牌。",
			},
			OnInstantTrickUsed: func(r Runtime, seat int, trickKind string) error {
				if !r.HasSkill(seat, IDJizhi) {
					return nil
				}
				return r.DrawCards(seat, 1)
			},
		},
		{
			Meta: Meta{
				ID: IDQicai, Name: "奇才", Kind: KindPassive,
				Desc: "锁定技，你使用的锦囊牌没有距离限制。",
			},
			TrickIgnoresDistance: func(r Runtime, seat int, trickKind string) bool {
				_ = trickKind
				return r.HasSkill(seat, IDQicai)
			},
		},
		{
			Meta: Meta{
				ID: IDQingguo, Name: "倾国", Kind: KindPassive,
				Desc: "你可以将一张黑色手牌当【闪】使用或打出。",
			},
			CardPlaysAs: func(r Runtime, seat int, cardKind, asKind, suit string) bool {
				if !r.HasSkill(seat, IDQingguo) || asKind != "shan" {
					return false
				}
				return IsBlackSuit(suit)
			},
		},
		{
			Meta: Meta{
				ID: IDJiji, Name: "急救", Kind: KindPassive,
				Desc: "锁定技，你的回合外，你可以将一张红色牌当【桃】使用。",
			},
			CardPlaysAs: func(r Runtime, seat int, cardKind, asKind, suit string) bool {
				if !r.HasSkill(seat, IDJiji) || asKind != "tao" {
					return false
				}
				if r.CurrentTurn() == seat {
					return false
				}
				_ = cardKind
				return IsRedSuit(suit)
			},
		},
		{
			Meta: Meta{
				ID: IDWushuang, Name: "无双", Kind: KindPassive,
				Desc: "锁定技，当你使用【杀】或【决斗】指定目标后，你令此【杀】或【决斗】需要依次使用两张【闪】或【杀】才能抵消。",
			},
			ExtraResponsesNeeded: func(r Runtime, source int, cardKind string) int {
				if !r.HasSkill(source, IDWushuang) {
					return 0
				}
				if cardKind == "sha" || cardKind == "juedou" {
					return 1
				}
				return 0
			},
		},
		{
			Meta: Meta{
				ID: IDBiyue, Name: "闭月", Kind: KindPassive,
				Desc: "锁定技，结束阶段，你可以摸一张牌。",
			},
			OnTurnEnd: func(r Runtime, seat int) error {
				if !r.HasSkill(seat, IDBiyue) {
					return nil
				}
				return r.DrawSkillCards(seat, IDBiyue, 1, "")
			},
		},
		{
			Meta: Meta{
				ID: IDWansha, Name: "完杀", Kind: KindPassive,
				Desc: "锁定技，你的回合内，除处于濒死状态的角色外，其他角色不能使用【桃】。",
			},
			BlocksPeachUse: func(r Runtime, userSeat int) bool {
				lord := r.CurrentTurn()
				if !r.HasSkill(lord, IDWansha) || userSeat == lord {
					return false
				}
				return !r.IsSeatInDyingRescue(userSeat)
			},
		},
		{
			Meta: Meta{
				ID: IDWeimu, Name: "帷幕", Kind: KindPassive,
				Desc: "锁定技，你不能成为黑色锦囊牌的目标。",
			},
			BlocksTrickTarget: func(r Runtime, target int, trickKind, suit string) bool {
				if !r.HasSkill(target, IDWeimu) || !IsJinnangKind(trickKind) {
					return false
				}
				return IsBlackSuit(suit)
			},
		},
		{
			Meta: Meta{
				ID: IDJueqing, Name: "绝情", Kind: KindPassive,
				Desc: "锁定技，你造成的伤害都视为失去体力。",
			},
			DamageAsHPLoss: func(r Runtime, source int) bool {
				return r.HasSkill(source, IDJueqing)
			},
		},
		{
			Meta: Meta{
				ID: IDShangshi, Name: "伤逝", Kind: KindPassive,
				Desc: "锁定技，除弃牌阶段外，若你的体力值不大于1，每当你失去手牌区里的牌时，你摸一张牌。",
			},
			OnCardsDiscarded: func(r Runtime, ctx CardsDiscardedCtx) error {
				if !r.HasSkill(ctx.Seat, IDShangshi) || ctx.Reason == "discard_phase" {
					return nil
				}
				hp, _ := r.PlayerHP(ctx.Seat)
				if hp > 1 || len(ctx.Cards) == 0 {
					return nil
				}
				return r.DrawSkillCards(ctx.Seat, IDShangshi, len(ctx.Cards), "")
			},
		},
		{
			Meta: Meta{
				ID: IDYingzi, Name: "英姿", Kind: KindPassive,
				Desc: "锁定技，摸牌阶段你多摸一张牌。",
			},
			DrawCountBonus: func(r Runtime, seat int) int {
				if r.HasSkill(seat, IDYingzi) {
					return 1
				}
				return 0
			},
		},
		{
			Meta: Meta{
				ID: IDLianying, Name: "连营", Kind: KindPassive,
				Desc: "当你失去最后的手牌时，你摸一张牌。",
			},
			OnHandEmpty: func(r Runtime, seat int) error {
				if !r.HasSkill(seat, IDLianying) {
					return nil
				}
				return r.DrawSkillCards(seat, IDLianying, 1, "")
			},
		},
		{
			Meta: Meta{
				ID: IDXiaoji, Name: "枭姬", Kind: KindPassive,
				Desc: "当你失去装备区里的牌时，你可以摸2张牌。",
			},
			OnEquipLost: func(r Runtime, ctx EquipLostCtx) error {
				if !r.HasSkill(ctx.Seat, IDXiaoji) {
					return nil
				}
				return r.DrawSkillCards(ctx.Seat, IDXiaoji, 2, "")
			},
		},
		{
			Meta: Meta{
				ID: IDHongyan, Name: "红颜", Kind: KindPassive,
				Desc: "锁定技，你的黑桃牌视为红桃。",
			},
			EffectiveSuit: func(r Runtime, seat int, suit string) string {
				if r.HasSkill(seat, IDHongyan) && suit == "S" {
					return "H"
				}
				return suit
			},
		},
		{
			Meta: Meta{
				ID: IDKeji, Name: "克己", Kind: KindPassive,
				Desc: "若你于出牌阶段内没有使用或打出过【杀】，你可以跳过弃牌阶段。",
			},
			SkipsDiscardPhase: func(r Runtime, seat int) bool {
				if !r.HasSkill(seat, IDKeji) {
					return false
				}
				return r.SkillCounter(seat, CounterShaInPlayPhase) == 0
			},
		},
		{
			Meta: Meta{
				ID: IDJiang, Name: "激昂", Kind: KindPassive,
				Desc: "每当你的【决斗】或【红色杀】被【无懈可击】抵消时，或你使用的【决斗】或【红色杀】生效后，你可以摸一张牌。",
			},
			OnCardResolved: func(r Runtime, ctx CardResolvedCtx) error {
				if !r.HasSkill(ctx.Seat, IDJiang) || !IsJiangCard(ctx.Card.Kind, ctx.Card.Suit) {
					return nil
				}
				return r.DrawSkillCards(ctx.Seat, IDJiang, 1, "")
			},
		},
	}
}
