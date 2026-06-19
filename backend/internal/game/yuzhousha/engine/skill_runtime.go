package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterRendeGiven     = "rende_given_play"
	counterRendeHealed    = "rende_healed_play"
	counterJijiangUseFailed = "jijiang_use_failed"
	counterWushengActive       = "wusheng_active"
	counterLuoyiActive         = "luoyi_active"
	counterDrawChoicePending   = "draw_choice_pending"
	counterTuxiDrawSkip        = "tuxi_draw_skip"
	counterQuhuUsed            = "quhu_used"

	ResponseModeSkillJijiang   = "skill_jijiang"
	ResponseModeSkillRende     = "skill_rende"
	ResponseModeSkillGuicai    = "skill_guicai"
	ResponseModeSkillFankui    = "skill_fankui"
	ResponseModeSkillPojunDiscard = "skill_pojun_discard"
)

type (
	SkillKind       = skill.Kind
	SkillMeta       = skill.Meta
	UseSkillRequest = skill.ActivateReq
)

const (
	SkillKindPassive = skill.KindPassive
	SkillKindActive  = skill.KindActive
	SkillKindLord    = skill.KindLord

	SkillRende   = skill.IDRende
	SkillJijiang = skill.IDJijiang
	SkillWusheng = skill.IDWusheng
	SkillLongdan = skill.IDLongdan
	SkillPaoxiao = skill.IDPaoxiao
	SkillGuanxing = skill.IDGuanxing
	SkillKongcheng = skill.IDKongcheng
	SkillTieqi   = skill.IDTieqi
	SkillMashi   = skill.IDMashi
	SkillJizhi   = skill.IDJizhi
	SkillQicai   = skill.IDQicai
	SkillFankui  = skill.IDFankui
	SkillGuicai  = skill.IDGuicai
	SkillLuoshen = skill.IDLuoshen
	SkillQingguo = skill.IDQingguo
	SkillJianxiong = skill.IDJianxiong
	SkillGanglie   = skill.IDGanglie
	SkillLuoyi     = skill.IDLuoyi
	SkillTuxi      = skill.IDTuxi
	SkillYiji      = skill.IDYiji
	SkillZhiheng   = skill.IDZhiheng
	SkillJiuyuan   = skill.IDJiuyuan
	SkillJieyin    = skill.IDJieyin
	SkillXiaoji    = skill.IDXiaoji
	SkillYingzi    = skill.IDYingzi
	SkillFanjian   = skill.IDFanjian
	SkillTianxiang = skill.IDTianxiang
	SkillHongyan   = skill.IDHongyan
	SkillQixi      = skill.IDQixi
	SkillYinghun   = skill.IDYinghun
	SkillLianying  = skill.IDLianying
	SkillGuose     = skill.IDGuose
	SkillLiuli     = skill.IDLiuli
	SkillKurou     = skill.IDKurou
	SkillKeji      = skill.IDKeji
	SkillJiang     = skill.IDJiang
	SkillHunzi     = skill.IDHunzi
	SkillJiji      = skill.IDJiji
	SkillWushuang  = skill.IDWushuang
	SkillBiyue     = skill.IDBiyue
	SkillShuangxiong = skill.IDShuangxiong
	SkillWansha      = skill.IDWansha
	SkillLuanwu      = skill.IDLuanwu
	SkillWeimu       = skill.IDWeimu
	SkillLeiji       = skill.IDLeiji
	SkillGuidao      = skill.IDGuidao
	SkillHuangtian   = skill.IDHuangtian
	SkillJueqing     = skill.IDJueqing
	SkillShangshi    = skill.IDShangshi
	SkillPojun       = skill.IDPojun
	SkillTushe       = skill.IDTushe
	SkillLimu         = skill.IDLimu

	CharLiuBei       = skill.CharLiuBei
	CharGuanYu       = skill.CharGuanYu
	CharZhangFei     = skill.CharZhangFei
	CharZhaoYun      = skill.CharZhaoYun
	CharZhugeLiang   = skill.CharZhugeLiang
	CharMaChao       = skill.CharMaChao
	CharHuangYueying = skill.CharHuangYueying
	CharSimaYi       = skill.CharSimaYi
	CharZhenJi       = skill.CharZhenJi
	CharCaoCao       = skill.CharCaoCao
	CharXiahouDun    = skill.CharXiahouDun
	CharXuChu        = skill.CharXuChu
	CharZhangLiao    = skill.CharZhangLiao
	CharGuoJia       = skill.CharGuoJia
	CharSunQuan      = skill.CharSunQuan
	CharSunShangxiang = skill.CharSunShangxiang
	CharZhouYu        = skill.CharZhouYu
	CharXiaoQiao      = skill.CharXiaoQiao
	CharGanNing       = skill.CharGanNing
	CharSunJian       = skill.CharSunJian
	CharLuXun         = skill.CharLuXun
	CharDaQiao        = skill.CharDaQiao
	CharHuangGai      = skill.CharHuangGai
	CharLvMeng        = skill.CharLvMeng
	CharSunCe         = skill.CharSunCe
	CharHuaTuo        = skill.CharHuaTuo
	CharLvBu          = skill.CharLvBu
	CharDiaoChan      = skill.CharDiaoChan
	CharYanLiangWenChou = skill.CharYanLiangWenChou
	CharJiaXu           = skill.CharJiaXu
	CharZhangJiao       = skill.CharZhangJiao
	CharZhangChunhua    = skill.CharZhangChunhua
	CharJieXuSheng      = skill.CharJieXuSheng
	KingdomShu       = skill.KingdomShu
	KingdomWei       = skill.KingdomWei
	KingdomWu        = skill.KingdomWu
	KingdomQun       = skill.KingdomQun
)

type gameSkillRuntime struct {
	g      *Game
	events *[]GameEvent
}

func (g *Game) skillRuntime(events *[]GameEvent) *gameSkillRuntime {
	return &gameSkillRuntime{g: g, events: events}
}

func (r *gameSkillRuntime) ModeID() string                         { return r.g.ModeID() }
func (r *gameSkillRuntime) HasSkill(seat int, skillID string) bool { return r.g.hasSkill(seat, skillID) }
func (r *gameSkillRuntime) Phase() string                          { return r.g.Phase }
func (r *gameSkillRuntime) TurnStep() string                       { return r.g.TurnStep }
func (r *gameSkillRuntime) CurrentTurn() int                       { return r.g.CurrentTurn }
func (r *gameSkillRuntime) PlayerHandCount(seat int) int           { return len(r.g.Players[seat].Hand) }
func (r *gameSkillRuntime) PlayerHP(seat int) (int, int) {
	p := r.g.Players[seat]
	return p.HP, p.MaxHP
}
func (r *gameSkillRuntime) SkillCounter(seat int, key string) int { return r.g.getSkillCounter(seat, key) }
func (r *gameSkillRuntime) OpponentOf(seat int) int                 { return r.g.opponentOf(seat) }
func (r *gameSkillRuntime) PlayerCount() int                        { return r.g.PlayerCount() }
func (r *gameSkillRuntime) TeamOf(seat int) int                   { return r.g.teamOf(seat) }
func (r *gameSkillRuntime) EnemiesOf(seat int) []int                { return r.g.enemiesOf(seat) }
func (r *gameSkillRuntime) AlliesOf(seat int) []int                 { return r.g.alliesOf(seat) }
func (r *gameSkillRuntime) CanUseSha(seat int) bool                 { return r.g.canUseSha(seat) }
func (r *gameSkillRuntime) CanAttack(from, to int) bool             { return r.g.canAttack(from, to) }
func (r *gameSkillRuntime) ShuAllies(lord int) []int                { return r.g.shuAlliesOf(lord) }
func (r *gameSkillRuntime) PlayerHandCardIDs(seat int) []string {
	ids := make([]string, 0, len(r.g.Players[seat].Hand))
	for _, c := range r.g.Players[seat].Hand {
		ids = append(ids, c.ID)
	}
	return ids
}
func (r *gameSkillRuntime) PendingRequiredKind() string {
	if r.g.Pending == nil {
		return ""
	}
	return r.g.Pending.RequiredKind
}
func (r *gameSkillRuntime) PendingResponseMode() string {
	if r.g.Pending == nil {
		return ""
	}
	return r.g.Pending.ResponseMode
}
func (r *gameSkillRuntime) PendingTargetSeat() int {
	if r.g.Pending == nil {
		return -1
	}
	return r.g.Pending.TargetIndex
}
func (r *gameSkillRuntime) PendingWindowKind() string {
	return r.g.PendingWindowKind()
}
func (r *gameSkillRuntime) PendingActorSeat() int {
	return r.g.PendingActorSeat()
}
func (r *gameSkillRuntime) PendingSubjectSeat() int {
	return r.g.PendingSubjectSeat()
}
func (r *gameSkillRuntime) PendingOriginSeat() int {
	return r.g.PendingOriginSeat()
}
func (r *gameSkillRuntime) TakeOne(seat int, zone, cardID string) error {
	return r.g.TakeOne(seat, ZoneID(zone), cardID, r.events)
}
func (r *gameSkillRuntime) PassTake(seat int) error {
	return r.g.PassTake(seat, r.events)
}
func (r *gameSkillRuntime) DiscardWindowOne(seat int, cardID string) error {
	return r.g.DiscardOne(seat, cardID, r.events)
}
func (r *gameSkillRuntime) CardPlaysAs(seat int, cardKind, asKind, suit string) bool {
	return r.g.cardPlaysAs(seat, Card{Kind: cardKind, Suit: suit}, asKind)
}
func (r *gameSkillRuntime) HandPlaysAs(seat int, asKind string) bool {
	for _, c := range r.g.Players[seat].Hand {
		if r.g.cardPlaysAs(seat, c, asKind) {
			return true
		}
	}
	return false
}

// HasBlackCard 检查玩家是否有黑色牌（手牌或装备区）
func (r *gameSkillRuntime) HasBlackCard(seat int) bool {
	// 检查手牌
	for _, c := range r.g.Players[seat].Hand {
		if skill.IsBlackSuit(c.Suit) {
			return true
		}
	}
	// 检查装备区
	for _, card := range []*Card{
		r.g.Players[seat].Weapon,
		r.g.Players[seat].Armor,
		r.g.Players[seat].PlusHorse,
		r.g.Players[seat].MinusHorse,
	} {
		if card != nil && skill.IsBlackSuit(card.Suit) {
			return true
		}
	}
	return false
}

func (r *gameSkillRuntime) AlivePlayerCount() int { return r.g.alivePlayerCount() }
func (r *gameSkillRuntime) DrawPileCount() int  { return len(r.g.DrawPile) }
func (r *gameSkillRuntime) StartPeekDeck(seat int, skillID string) error {
	return r.g.StartPeekDeck(seat, skillID, r.events)
}

func (r *gameSkillRuntime) GiveRende(source, target int, cardIDs []string) error {
	return r.g.executeRendeGive(source, target, cardIDs, r.events)
}
func (r *gameSkillRuntime) StartJijiangForUse(lord, target int) error {
	return r.g.startJijiangForUse(lord, target, r.events)
}
func (r *gameSkillRuntime) StartJijiangForResponse(lord int) error {
	return r.g.startJijiangForResponse(lord, r.events)
}

func (r *gameSkillRuntime) ToggleWusheng(seat int) error {
	return r.g.toggleWusheng(seat, r.events)
}

func (r *gameSkillRuntime) ToggleQixi(seat int) error {
	return r.g.toggleQixi(seat, r.events)
}

func (g *Game) playerSkillIDs(seat int) []string {
	if seat < 0 || seat >= len(g.Players) {
		return nil
	}
	return append([]string(nil), g.Players[seat].Character.SkillIDs...)
}

func (g *Game) hasSkill(seat int, skillID string) bool {
	for _, id := range g.playerSkillIDs(seat) {
		if id == skillID {
			return true
		}
	}
	return false
}

func (g *Game) playerSkillHandlers(seat int) []skill.Handler {
	ids := g.playerSkillIDs(seat)
	out := make([]skill.Handler, 0, len(ids))
	for _, id := range ids {
		if h, ok := skill.Lookup(id); ok {
			out = append(out, h)
		}
	}
	return out
}

func (g *Game) cardPlaysAs(seat int, card Card, asKind string) bool {
	return g.cardPlaysAsViaHooks(seat, card, asKind)
}

func (g *Game) skillUnlimitedSha(seat int) bool {
	if g.hasWeaponKind(seat, CardWeapon1) {
		return true
	}
	return g.skillUnlimitedShaViaHooks(seat)
}

func (g *Game) getSkillCounter(seat int, key string) int {
	p := &g.Players[seat]
	if p.SkillCounters == nil {
		return 0
	}
	return p.SkillCounters[key]
}

func (g *Game) addSkillCounter(seat int, key string, delta int) {
	p := &g.Players[seat]
	if p.SkillCounters == nil {
		p.SkillCounters = map[string]int{}
	}
	p.SkillCounters[key] += delta
}

func (g *Game) resetPlayPhaseSkillCounters(seat int) {
	p := &g.Players[seat]
	if p.SkillCounters == nil {
		return
	}
	delete(p.SkillCounters, counterRendeGiven)
	delete(p.SkillCounters, counterRendeHealed)
	delete(p.SkillCounters, counterJijiangUseFailed)
	delete(p.SkillCounters, counterWushengActive)
	delete(p.SkillCounters, counterLuoyiActive)
	delete(p.SkillCounters, counterDrawChoicePending)
	delete(p.SkillCounters, counterZhihengUsed)
	delete(p.SkillCounters, counterJieyinUsed)
	delete(p.SkillCounters, counterFanjianUsed)
	delete(p.SkillCounters, counterQixiActive)
	delete(p.SkillCounters, counterYinghunUsed)
	delete(p.SkillCounters, counterShuangxiongActive)
	delete(p.SkillCounters, counterShuangxiongRefRed)
	// 突袭相关计数器
	delete(p.SkillCounters, "tuxi_selected")
	delete(p.SkillCounters, "tuxi_max")
}

func (g *Game) setSkillCounter(seat int, key string, value int) {
	p := &g.Players[seat]
	if p.SkillCounters == nil {
		p.SkillCounters = map[string]int{}
	}
	p.SkillCounters[key] = value
}

func (g *Game) isShuCharacter(seat int) bool {
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	return g.Players[seat].Character.Kingdom == KingdomShu
}

func (g *Game) shuAlliesOf(lordSeat int) []int {
	out := make([]int, 0, len(g.Players)-1)
	for i := range g.Players {
		if i == lordSeat {
			continue
		}
		if g.isShuCharacter(i) {
			out = append(out, i)
		}
	}
	return out
}

func (g *Game) appendSkillEvent(events *[]GameEvent, skillID string, source, target int, message string) {
	*events = append(*events, GameEvent{
		Type:        "skill_trigger",
		PlayerIndex: source,
		TargetIndex: target,
		SkillID:     skillID,
		Message:     message,
	})
}

func (g *Game) ListActivatableSkills(seat int) []SkillMeta {
	if g.IsFinished() {
		return nil
	}
	rt := g.skillRuntime(nil)
	out := make([]SkillMeta, 0)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.CanActivate(rt, seat) {
			out = append(out, h.Meta())
		}
	}
	return out
}

func (g *Game) UseSkill(seat int, req UseSkillRequest, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase == PhaseResponse && g.Pending != nil {
		g.ensurePendingRoles()
		if g.Pending.WindowKind == WindowKindTake && g.takeWindow != nil {
			if req.SkillID != "" && req.SkillID != g.Pending.SkillID {
				return ErrWrongPhase
			}
			// 破军批量选牌：一次性提交多张
			if g.Pending.ResponseMode == ResponseModeSkillPojun && len(req.CardIDs) > 0 {
				for _, cardID := range req.CardIDs {
					if g.takeWindow == nil {
						break
					}
					// 自动检测牌所在的 zone
					zone := g.findCardZone(g.Pending.SubjectSeat, cardID)
					if err := g.TakeOne(seat, zone, cardID, events); err != nil {
						return err
					}
				}
				return nil
			}
			zone := req.TargetZone
			if zone == "" {
				zone = "hand"
			}
			return g.TakeOne(seat, ZoneID(zone), req.TargetCardID, events)
		}
		if g.Pending.WindowKind == WindowKindDiscard && g.discardWindow != nil {
			if req.SkillID != "" && req.SkillID != g.Pending.SkillID {
				return ErrWrongPhase
			}
			cardID := req.TargetCardID
			if cardID == "" && len(req.CardIDs) > 0 {
				cardID = req.CardIDs[0]
			}
			if cardID == "" {
				return ErrInvalidCard
			}
			return g.DiscardOne(seat, cardID, events)
		}
		if g.Pending.TieqiPending && g.Pending.SourceIndex == seat && req.SkillID == SkillTieqi {
			if !g.hasSkill(seat, req.SkillID) {
				return ErrInvalidCard
			}
			h, ok := skill.Lookup(req.SkillID)
			if !ok {
				return ErrInvalidCard
			}
			rt := g.skillRuntime(events)
			if !h.CanActivate(rt, seat) {
				return ErrWrongPhase
			}
			return h.Activate(rt, seat, req)
		}
		switch g.Pending.ResponseMode {
		case ResponseModeSkillRende:
			return g.executeRendeGive(seat, req.TargetIndex, req.CardIDs, events)
		case ResponseModeSkillJijiang:
			return ErrWrongPhase
		case ResponseModeSkillFankui:
			if req.SkillID != SkillFankui {
				return ErrWrongPhase
			}
			return g.FankuiTakeFrom(seat, req.TargetZone, req.TargetCardID, events)
		case ResponseModeSkillGuicai:
			if req.SkillID != SkillGuicai {
				return ErrWrongPhase
			}
			if len(req.CardIDs) == 0 {
				return ErrInvalidCard
			}
			return g.ApplyGuicaiReplace(seat, req.CardIDs[0], events)
		case ResponseModeDdzJudgeCancel:
			if req.SkillID != SkillDdzJudgeCancel {
				return ErrWrongPhase
			}
			return g.ApplyDdzJudgeCancel(seat, req.CardIDs, events)
		case ResponseModeSkillGuidao:
			if req.SkillID != SkillGuidao {
				return ErrWrongPhase
			}
			if len(req.CardIDs) == 0 {
				return ErrInvalidCard
			}
			return g.ApplyGuidaoReplace(seat, req.CardIDs[0], events)
		case ResponseModeSkillLeijiOffer:
			if req.SkillID != SkillLeiji {
				return ErrWrongPhase
			}
			return g.StartLeijiJudge(seat, events)
		case ResponseModeSkillJianxiong:
			if req.SkillID != SkillJianxiong {
				return ErrWrongPhase
			}
			return g.ApplyJianxiong(seat, events)
		case ResponseModeSkillYijiOffer:
			if req.SkillID != SkillYiji {
				return ErrWrongPhase
			}
			return g.ApplyYiji(seat, events)
		case ResponseModeSkillYijiGive:
			if req.SkillID != SkillYiji {
				return ErrWrongPhase
			}
			return g.YijiGiveCards(seat, req.TargetIndex, req.CardIDs, events)
		case ResponseModeSkillGanglieOffer:
			if req.SkillID != SkillGanglie {
				return ErrWrongPhase
			}
			return g.StartGanglieJudge(seat, events)
		case ResponseModeSkillGanglieChoice:
			if req.SkillID != SkillGanglie {
				return ErrWrongPhase
			}
			if len(req.CardIDs) >= 2 {
				return g.GanglieDiscard(seat, req.CardIDs[:2], events)
			}
			if req.TargetZone == "take_damage" {
				return g.GanglieTakeDamage(seat, events)
			}
			return ErrInvalidCard
		case ResponseModeSkillTuxi:
			if req.SkillID != SkillTuxi {
				return ErrWrongPhase
			}
			return g.TuxiTakeFrom(seat, req.TargetZone, req.TargetCardID, events)
		case ResponseModeSkillPojun:
			if req.SkillID != SkillPojun {
				return ErrWrongPhase
			}
			return g.PojunPlace(seat, req.TargetZone, req.TargetCardID, events)
		}
	}
	if g.Phase == PhaseResponse {
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillFankui && g.Pending.TargetIndex == seat && req.SkillID == SkillFankui {
			return g.FankuiTakeFrom(seat, req.TargetZone, req.TargetCardID, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillGuicai && g.Pending.TargetIndex == seat && req.SkillID == SkillGuicai {
			if len(req.CardIDs) == 0 {
				return ErrInvalidCard
			}
			return g.ApplyGuicaiReplace(seat, req.CardIDs[0], events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillGuidao && g.Pending.TargetIndex == seat && req.SkillID == SkillGuidao {
			if len(req.CardIDs) == 0 {
				return ErrInvalidCard
			}
			return g.ApplyGuidaoReplace(seat, req.CardIDs[0], events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillLeijiOffer && g.Pending.TargetIndex == seat && req.SkillID == SkillLeiji {
			return g.StartLeijiJudge(seat, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillJianxiong && g.Pending.TargetIndex == seat && req.SkillID == SkillJianxiong {
			return g.ApplyJianxiong(seat, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillYijiOffer && g.Pending.TargetIndex == seat && req.SkillID == SkillYiji {
			return g.ApplyYiji(seat, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillYijiGive && g.Pending.TargetIndex == seat && req.SkillID == SkillYiji {
			return g.YijiGiveCards(seat, req.TargetIndex, req.CardIDs, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillGanglieOffer && g.Pending.TargetIndex == seat && req.SkillID == SkillGanglie {
			return g.StartGanglieJudge(seat, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillGanglieChoice && g.Pending.TargetIndex == seat && req.SkillID == SkillGanglie {
			if len(req.CardIDs) >= 2 {
				return g.GanglieDiscard(seat, req.CardIDs[:2], events)
			}
			if req.TargetZone == "take_damage" {
				return g.GanglieTakeDamage(seat, events)
			}
			return ErrInvalidCard
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillTuxi && g.Pending.TargetIndex == seat && req.SkillID == SkillTuxi {
			return g.TuxiTakeFrom(seat, req.TargetZone, req.TargetCardID, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillFanjianSuit && g.Pending.TargetIndex == seat {
			if req.TargetZone == "" {
				return ErrInvalidTarget
			}
			return g.ResolveFanjianSuit(seat, req.TargetZone, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillTianxiang && g.Pending.TargetIndex == seat {
			if req.SkillID != SkillTianxiang {
				return ErrWrongPhase
			}
			if len(req.CardIDs) == 0 {
				return g.PassTianxiang(seat, events)
			}
			return g.ApplyTianxiang(seat, req.CardIDs[0], events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillYinghun && g.Pending.TargetIndex == seat {
			discard := ""
			if len(req.CardIDs) > 0 {
				discard = req.CardIDs[0]
			}
			return g.resolveYinghunChoice(seat, req.TargetZone, discard, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillYinghunDiscard && g.Pending.TargetIndex == seat {
			if len(req.CardIDs) == 0 {
				return ErrInvalidCard
			}
			return g.YinghunDiscard(seat, req.CardIDs, events)
		}
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeSkillLiuli && g.Pending.TargetIndex == seat {
			if req.SkillID != SkillLiuli {
				return ErrWrongPhase
			}
			if len(req.CardIDs) == 0 {
				return g.PassLiuli(seat, events)
			}
			target := req.TargetIndex
			if target < 0 {
				target = g.Pending.EffectTarget
			}
			return g.ApplyLiuli(seat, req.CardIDs[0], target, events)
		}
		if g.Pending == nil || g.Pending.TargetIndex != seat {
			return ErrNotYourTurn
		}
		if !g.hasSkill(seat, req.SkillID) {
			return ErrInvalidCard
		}
		h, ok := skill.Lookup(req.SkillID)
		if !ok {
			return ErrInvalidCard
		}
		rt := g.skillRuntime(events)
		if !h.CanActivate(rt, seat) {
			return ErrWrongPhase
		}
		return h.Activate(rt, seat, req)
	}
	if g.Phase == PhasePlaying && g.TurnStep == StepPrepare && g.CurrentTurn == seat {
		if !g.hasSkill(seat, req.SkillID) {
			return ErrInvalidCard
		}
		h, ok := skill.Lookup(req.SkillID)
		if !ok {
			return ErrInvalidCard
		}
		rt := g.skillRuntime(events)
		if h.PeekDeckConfig() != nil {
			return g.StartPeekDeck(seat, req.SkillID, events)
		}
		if !h.CanActivate(rt, seat) {
			return ErrWrongPhase
		}
		return h.Activate(rt, seat, req)
	}
	if g.Phase == PhasePlaying && g.TurnStep == StepDraw && g.CurrentTurn == seat {
		if req.SkillID == skill.IDLuoyi {
			return g.ActivateLuoyi(seat, events)
		}
		if !g.hasSkill(seat, req.SkillID) {
			return ErrInvalidCard
		}
		h, ok := skill.Lookup(req.SkillID)
		if !ok {
			return ErrInvalidCard
		}
		rt := g.skillRuntime(events)
		if !h.CanActivate(rt, seat) {
			return ErrWrongPhase
		}
		return h.Activate(rt, seat, req)
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrNotYourTurn
	}
	if !g.hasSkill(seat, req.SkillID) {
		return ErrInvalidCard
	}
	// 出牌阶段：取消武圣（不要求 CanActivate，因技能钮仅用于发动）
	if req.SkillID == skill.IDWusheng && g.getSkillCounter(seat, counterWushengActive) > 0 {
		return g.toggleWusheng(seat, events)
	}
	h, ok := skill.Lookup(req.SkillID)
	if !ok {
		return ErrInvalidCard
	}
	rt := g.skillRuntime(events)
	if !h.CanActivate(rt, seat) {
		return ErrWrongPhase
	}
	return h.Activate(rt, seat, req)
}

func (g *Game) runAIActiveSkills(seat int, events *[]GameEvent) bool {
	rt := g.skillRuntime(events)
	type candidate struct {
		priority int
		handler  skill.Handler
	}
	var options []candidate
	for _, h := range g.playerSkillHandlers(seat) {
		p := h.AIPriority(rt, seat)
		if p > 0 && h.CanActivate(rt, seat) {
			options = append(options, candidate{priority: p, handler: h})
		}
	}
	if len(options) == 0 {
		return false
	}
	best := options[0]
	for _, o := range options[1:] {
		if o.priority > best.priority {
			best = o
		}
	}
	if err := best.handler.AIActivate(rt, seat); err != nil {
		return false
	}
	return true
}

func buildCharacter(charID string) Character {
	def, ok := skill.CharacterByID(charID)
	if !ok {
		return Character{ID: charID, Name: charID, MaxHP: DefaultMaxHP}
	}
	display := skill.ResolveHeroDisplay(charID, "")
	return Character{
		ID:            def.ID,
		Name:          def.Name,
		MaxHP:         def.MaxHP,
		Kingdom:       def.Kingdom,
		Gender:        def.Gender,
		SkillIDs:      append([]string(nil), def.SkillIDs...),
		Skills:        skill.MetasForCharacter(charID),
		DefaultSkinID: display.SkinID,
		SkinID:        display.SkinID,
		Display:       &display,
	}
}

func validateCharacterIDStatic(id string) error {
	if _, ok := skill.CharacterByID(id); !ok {
		return fmt.Errorf("unknown character: %s", id)
	}
	return nil
}

func RandomAICharacter(excludeID string) string {
	return skill.RandomPickableCharacter(excludeID)
}

// PlayerHandCards 获取玩家的手牌
func (r *gameSkillRuntime) PlayerHandCards(seat int) []skill.CardView {
	if seat < 0 || seat >= len(r.g.Players) {
		return nil
	}
	
	// 将 []Card 转换为 []skill.CardView
	cards := r.g.Players[seat].Hand
	result := make([]skill.CardView, len(cards))
	for i, card := range cards {
		result[i] = skill.CardView{
			ID:   card.ID,
			Kind: card.Kind,
			Suit: card.Suit,
			Rank: card.Rank,
			Name: card.Name,
			Label: card.Label,
		}
	}
	return result
}

// UseLonghunCards 使用龙魂转化的牌
func (r *gameSkillRuntime) UseLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool) error {
	return r.g.useLonghunCards(seat, cardIDs, asKind, useTwoCards, isRed, isBlack, r.events)
}

// ResponseLonghunCards 打出龙魂转化的牌
func (r *gameSkillRuntime) ResponseLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool) error {
	return r.g.responseLonghunCards(seat, cardIDs, asKind, useTwoCards, isRed, isBlack, r.events)
}
