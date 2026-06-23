package engine

import (
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

type gameTargetCtx struct{ g *Game }

func (g *Game) targetCtx() gameTargetCtx { return gameTargetCtx{g: g} }

func (t gameTargetCtx) ModeID() string        { return t.g.ModeID() }
func (t gameTargetCtx) PlayerCount() int      { return t.g.PlayerCount() }
func (t gameTargetCtx) AliveHP(seat int) int  { return t.g.AliveHP(seat) }
func (t gameTargetCtx) CanAttack(from, to int) bool {
	return t.g.canAttack(from, to)
}
func (t gameTargetCtx) HasTakeableCard(target int) bool {
	return t.g.hasTakeableCard(target)
}
func (t gameTargetCtx) CanBingliangTarget(from, to int) bool {
	return t.g.canBingliangTarget(from, to)
}
func (t gameTargetCtx) HandCount(seat int) int {
	if seat < 0 || seat >= len(t.g.Players) {
		return 0
	}
	p := &t.g.Players[seat]
	if len(p.Hand) > 0 {
		return len(p.Hand)
	}
	return p.HandCount
}
func (t gameTargetCtx) HasJudgeKind(target int, kind string) bool {
	return t.g.Players[target].hasJudgeKind(kind)
}

func (t gameTargetCtx) TargetBlocked(target int, cardKind string) bool {
	if t.g.vineBlocksTrick(target, cardKind) {
		return true
	}
	return t.g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTargetBlocked, Target: target, CardKind: cardKind,
	}).Bool
}
func (t gameTargetCtx) PlayerHP(seat int) (hp, maxHP int) {
	if seat < 0 || seat >= len(t.g.Players) {
		return 0, 0
	}
	p := t.g.Players[seat]
	return p.HP, p.MaxHP
}
func (t gameTargetCtx) LimuActive(source int) bool {
	return t.g.hasSkill(source, SkillLimu) && len(t.g.Players[source].JudgeArea) > 0
}

func (t gameTargetCtx) TrickIgnoresDistance(source int, trickKind string) bool {
	return t.g.trickIgnoresDistance(source, trickKind)
}

func (g *Game) validPlayTargets(source int, cardKind string) []int {
	return mode.ValidPlayTargets(g.targetCtx(), source, cardKind)
}

func (g *Game) isValidPlayTarget(source, target int, cardKind string) bool {
	return mode.IsValidPlayTarget(g.targetCtx(), source, target, cardKind)
}
