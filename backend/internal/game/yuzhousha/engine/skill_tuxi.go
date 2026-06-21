package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeSkillTuxi = "skill_tuxi"

func (g *Game) shouldOfferTuxiDrawChoice(seat int) bool {
	if !g.hasSkill(seat, SkillTuxi) {
		return false
	}
	for _, e := range g.enemiesOf(seat) {
		if g.hasTakeableCard(e) {
			return true
		}
	}
	return false
}

func (g *Game) shouldOfferDrawPhaseChoice(seat int) bool {
	if g.hasSkill(seat, SkillLuoyi) {
		return true
	}
	if g.hasSkill(seat, SkillShuangxiong) {
		return true
	}
	return g.shouldOfferTuxiDrawChoice(seat)
}

func (g *Game) offerDrawPhaseChoice(seat int, events *[]GameEvent) {
	g.setSkillCounter(seat, counterDrawChoicePending, 1)
	g.TurnStep = StepDraw
	g.Message = fmt.Sprintf("%s 摸牌阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "draw_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
}

func (g *Game) isDrawPhaseChoicePending(seat int) bool {
	return g.Phase == PhasePlaying && g.TurnStep == StepDraw && g.CurrentTurn == seat &&
		g.getSkillCounter(seat, counterDrawChoicePending) > 0
}

// StartTuxi 发动突袭：放弃摸牌，从至多2名对手中各获得一张牌
func (g *Game) StartTuxi(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isDrawPhaseChoicePending(seat) {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillTuxi) {
		return ErrInvalidTarget
	}

	// 检查是否有至少1名有牌的对手
	hasValidTarget := false
	for _, enemy := range g.enemiesOf(seat) {
		if g.hasTakeableCard(enemy) {
			hasValidTarget = true
			break
		}
	}
	if !hasValidTarget {
		return ErrInvalidTarget
	}

	g.setSkillCounter(seat, counterDrawChoicePending, 0)

	// 初始化突袭状态：已选择0名对手，最多可选2名
	g.setSkillCounter(seat, "tuxi_selected", 0)
	g.setSkillCounter(seat, "tuxi_max", 2)

	// 开始第一次选择
	return g.startTuxiTake(seat, events)
}

// startTuxiTake 开始一次突袭拿牌
func (g *Game) startTuxiTake(seat int, events *[]GameEvent) error {
	// 查找第一个有牌的对手
	opp := -1
	for _, enemy := range g.enemiesOf(seat) {
		if g.hasTakeableCard(enemy) {
			opp = enemy
			break
		}
	}

	if opp < 0 {
		// 没有可选择的对手，结束突袭
		return g.finishTuxi(seat, events)
	}

	selected := g.getSkillCounter(seat, "tuxi_selected")
	msg := fmt.Sprintf("%s 发动【突袭】（%d/2），请选择获得 %s 的牌",
		g.Players[seat].Name, selected+1, g.Players[opp].Name)

	actor := seat
	return g.OpenTakeWindow(TakeWindowConfig{
		SkillID:          skill.IDTuxi,
		ResponseMode:     ResponseModeSkillTuxi,
		ActorSeat:        seat,
		SubjectSeat:      opp,
		OriginSeat:       seat,
		MaxTake:          1, // 每次只能拿1张
		Destination:      TakeDestination{Zone: ZoneHand, Seat: seat},
		Message:          msg,
		EventType:        "tuxi_take",
		SkillEventLabel:  "突袭",
		PassClosesWindow: false, // 不关闭窗口，可以继续选择第二名对手
		OnComplete: func(g *Game, events *[]GameEvent) error {
			// 拿牌完成后，增加计数
			g.addSkillCounter(seat, "tuxi_selected", 1)
			selected := g.getSkillCounter(seat, "tuxi_selected")

			// 如果已经选择了2名对手，结束突袭
			if selected >= 2 {
				return g.finishTuxi(actor, events)
			}

			// 否则，检查是否还有第二名对手可选
			hasMore := false
			for _, enemy := range g.enemiesOf(seat) {
				if g.hasTakeableCard(enemy) {
					hasMore = true
					break
				}
			}

			if !hasMore {
				// 没有更多对手可选，结束突袭
				return g.finishTuxi(actor, events)
			}

			// 继续选择第二名对手
			return g.continueTuxi(actor, events)
		},
	}, events)
}

// continueTuxi 继续选择第二名对手
func (g *Game) continueTuxi(seat int, events *[]GameEvent) error {
	// 查找另一个有牌的对手（排除已经选择过的）
	opp := -1
	for _, enemy := range g.enemiesOf(seat) {
		if g.hasTakeableCard(enemy) {
			opp = enemy
			break
		}
	}

	if opp < 0 {
		return g.finishTuxi(seat, events)
	}

	msg := fmt.Sprintf("%s 可继续发动【突袭】（2/2），请选择获得 %s 的牌，或点击跳过",
		g.Players[seat].Name, g.Players[opp].Name)

	actor := seat
	return g.OpenTakeWindow(TakeWindowConfig{
		SkillID:          skill.IDTuxi,
		ResponseMode:     ResponseModeSkillTuxi,
		ActorSeat:        seat,
		SubjectSeat:      opp,
		OriginSeat:       seat,
		MaxTake:          1,
		Destination:      TakeDestination{Zone: ZoneHand, Seat: seat},
		Message:          msg,
		EventType:        "tuxi_take",
		SkillEventLabel:  "突袭",
		PassClosesWindow: true, // 第二次可以选择跳过
		OnComplete: func(g *Game, events *[]GameEvent) error {
			g.addSkillCounter(seat, "tuxi_selected", 1)
			return g.finishTuxi(actor, events)
		},
	}, events)
}

// TuxiTakeFrom 突袭拿牌（TakeWindow 薄封装）。
func (g *Game) TuxiTakeFrom(seat int, zone, cardID string, events *[]GameEvent) error {
	if zone == "" {
		zone = "hand"
	}
	return g.TakeOne(seat, ZoneID(zone), cardID, events)
}

// PassTuxi 结束突袭窗口（TakeWindow 薄封装）。
func (g *Game) PassTuxi(seat int, events *[]GameEvent) error {
	return g.PassTake(seat, events)
}

func (g *Game) finishTuxi(seat int, events *[]GameEvent) error {
	// 清理突袭状态
	g.setSkillCounter(seat, "tuxi_selected", 0)
	g.setSkillCounter(seat, "tuxi_max", 0)

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay

	// 突袭放弃摸牌，所以不需要再摸牌
	// 直接进入出牌阶段
	return g.advanceTurnAfterDraw(seat, events)
}

func aiPickTakeTarget(g *Game, target int) (zone, cardID string) {
	p := &g.Players[target]

	// 优先级：手牌区（随机） → 装备区（随机） → 判定区（随机）
	// 手牌区
	if len(p.Hand) > 0 {
		return "hand", ""
	}
	// 装备区：随机选一个非空槽位
	equips := make([]struct {
		zone string
		card *Card
	}, 0, 4)
	if p.Weapon != nil {
		equips = append(equips, struct {
			zone string
			card *Card
		}{EquipWeapon, p.Weapon})
	}
	if p.Armor != nil {
		equips = append(equips, struct {
			zone string
			card *Card
		}{EquipArmor, p.Armor})
	}
	if p.MinusHorse != nil {
		equips = append(equips, struct {
			zone string
			card *Card
		}{EquipMinusHorse, p.MinusHorse})
	}
	if p.PlusHorse != nil {
		equips = append(equips, struct {
			zone string
			card *Card
		}{EquipPlusHorse, p.PlusHorse})
	}
	if len(equips) > 0 {
		idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(equips))
		return equips[idx].zone, equips[idx].card.ID
	}
	// 判定区
	if len(p.JudgeArea) > 0 {
		idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(p.JudgeArea))
		return "judge", p.JudgeArea[idx].ID
	}
	return "", ""
}
