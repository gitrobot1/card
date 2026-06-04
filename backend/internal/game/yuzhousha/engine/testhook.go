//go:build cardtest

package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// 以下导出仅供 backend/test 下的外部测试使用（需 -tags cardtest）。

func (g *Game) SyncCounts() { g.syncCounts() }

func (g *Game) CanUseSha(seat int) bool { return g.canUseSha(seat) }

func (g *Game) CardPlaysAsForTest(seat int, card Card, asKind string) bool {
	return g.cardPlaysAs(seat, card, asKind)
}

func (g *Game) TargetBlockedBySkillForTest(target int, cardKind string) bool {
	return g.targetBlockedBySkill(target, cardKind)
}

func (g *Game) PlaySha(seat int, cardID string, targetIndex int, events *[]GameEvent) error {
	return g.playSha(seat, cardID, targetIndex, events)
}

func (g *Game) RunSkillHooks(events *[]GameEvent, call skill.HookCall) skill.HookResult {
	return g.runSkillHooks(events, call)
}

func (g *Game) ApplyDamageForTest(source, target, amount int, cardKind, cardName string, events *[]GameEvent) int {
	return g.applyDamage(source, target, amount, Card{Kind: cardKind, Name: cardName}, events)
}

func (g *Game) NotifyInstantTrickUsedForTest(seat int, trickKind string, events *[]GameEvent) {
	g.notifyInstantTrickUsed(seat, trickKind, events)
}

func (g *Game) BeginTurnForTest(events *[]GameEvent) { g.beginTurn(events) }

func (g *Game) CanBingliangTargetForTest(from, to int) bool { return g.canBingliangTarget(from, to) }

func (g *Game) DistanceBetween(from, to int) int { return g.distanceBetween(from, to) }

func (g *Game) HasJudgeKindForTest(seat int, kind string) bool {
	return g.Players[seat].hasJudgeKind(kind)
}

func (g *Game) QilinDiscardHorseForTest(seat int, zone string, events *[]GameEvent) error {
	return g.qilinDiscardHorse(seat, zone, events)
}

func (g *Game) PickWuguCardForTest(seat int, cardID string, events *[]GameEvent) error {
	return g.pickWuguCard(seat, cardID, events)
}

func (g *Game) AwakenHunziForTest(seat int, events *[]GameEvent) error {
	return g.AwakenHunzi(seat, events)
}

func (g *Game) TryJiangDrawForTest(seat int, card Card, events *[]GameEvent) {
	g.tryJiangDraw(seat, card, events)
}

func (g *Game) HasSkillForTest(seat int, skillID string) bool {
	return g.hasSkill(seat, skillID)
}

func (g *Game) CanUseJijiHealForTest(seat int, card Card) bool {
	return g.canUseJijiHeal(seat, card)
}

func (g *Game) PlayJijiHealForTest(seat int, cardID string, events *[]GameEvent) error {
	return g.playJijiHeal(seat, cardID, events)
}

func (g *Game) SetSkillCounterForTest(seat int, key string, value int) {
	g.setSkillCounter(seat, key, value)
}

func (g *Game) GetSkillCounterForTest(seat int, key string) int {
	return g.getSkillCounter(seat, key)
}

func (g *Game) RespondWuxiekForTest(seat int, cardID string, events *[]GameEvent) error {
	return g.RespondWuxiek(seat, cardID, events)
}

func (g *Game) WeimuBlocksTrickForTest(target int, card Card) bool {
	return g.weimuBlocksTrick(target, card)
}

func (g *Game) ActivateLuanwuForTest(seat int, events *[]GameEvent) error {
	return g.ActivateLuanwu(seat, events)
}

func (g *Game) PassLuanwuForTest(seat int, events *[]GameEvent) error {
	return g.passLuanwu(seat, events)
}

func (g *Game) PlayLuanwuShaForTest(seat int, cardID string, target int, events *[]GameEvent) error {
	return g.playLuanwuSha(seat, cardID, target, events)
}

func (g *Game) StartLeijiJudgeForTest(seat int, events *[]GameEvent) error {
	return g.StartLeijiJudge(seat, events)
}

func (g *Game) ApplyGuidaoReplaceForTest(seat int, cardID string, events *[]GameEvent) error {
	return g.ApplyGuidaoReplace(seat, cardID, events)
}

func (g *Game) PassGuidaoForTest(seat int, events *[]GameEvent) error {
	return g.PassGuidao(seat, events)
}

func (g *Game) RespondCardForTest(seat int, cardID string, events *[]GameEvent) error {
	return g.RespondCard(seat, cardID, events)
}

func (g *Game) AfterJudgeFlipForTest(judgeSeat int, card Card, events *[]GameEvent) error {
	return g.afterJudgeFlip(judgeSeat, skill.JudgeLeiji, guicaiResumeLeiji, card, events)
}

func (g *Game) SetLeijiContextForTest(shanSeat int) {
	g.leijiShanSeat = shanSeat
}

func (g *Game) PendingResponseModeForTest() string {
	if g.Pending == nil {
		return ""
	}
	return g.Pending.ResponseMode
}

func (g *Game) ForceSkipDeadTurnForTest(events *[]GameEvent) error {
	if g.AliveHP(g.CurrentTurn) > 0 {
		return fmt.Errorf("current seat alive")
	}
	return g.endTurn(events)
}

func (g *Game) DrawCountForTest(seat int) int {
	return g.drawCountFor(seat)
}

func (g *Game) CanUseShaForTest(seat int) bool {
	return g.canUseSha(seat)
}
