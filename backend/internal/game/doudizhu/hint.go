package doudizhu

import "github.com/time/card/backend/internal/game/card"

type HintAction string

const (
	HintPlay HintAction = "play"
	HintPass HintAction = "pass"
	HintNone HintAction = "none"
)

type HintResult struct {
	Action  HintAction `json:"action"`
	CardIDs []string   `json:"card_ids"`
	Message string     `json:"message"`
}

func (g *Game) Hint(playerIndex int) (*HintResult, error) {
	if g.Phase != PhasePlaying {
		return &HintResult{Action: HintNone, Message: "当前不是出牌阶段"}, ErrWrongPhase
	}
	if playerIndex != g.CurrentTurn {
		return &HintResult{Action: HintNone, Message: "还没轮到你出牌"}, ErrNotYourTurn
	}
	if g.WinnerIndex != nil {
		return &HintResult{Action: HintNone, Message: "游戏已结束"}, ErrAlreadyFinished
	}

	hand := g.Players[playerIndex].Hand
	if len(hand) == 0 {
		return &HintResult{Action: HintNone, Message: "没有手牌"}, nil
	}

	if g.LastPlay == nil || g.LastPlay.PlayerIndex == playerIndex && g.PassCount == 0 {
		cards := pickSmallestPattern(hand)
		return &HintResult{
			Action:  HintPlay,
			CardIDs: cardIDs(cards),
			Message: formatPlayHint(cards),
		}, nil
	}

	if g.LastPlay != nil && g.LastPlay.PlayerIndex != playerIndex {
		if beat := findBeatingCards(hand, g.LastPlay.Cards); len(beat) > 0 {
			return &HintResult{
				Action:  HintPlay,
				CardIDs: cardIDs(beat),
				Message: formatPlayHint(beat),
			}, nil
		}
		return &HintResult{
			Action:  HintPass,
			CardIDs: nil,
			Message: "建议不出，等待下一轮",
		}, nil
	}

	cards := pickSmallestPattern(hand)
	return &HintResult{
		Action:  HintPlay,
		CardIDs: cardIDs(cards),
		Message: formatPlayHint(cards),
	}, nil
}

func formatPlayHint(cards []card.Card) string {
	pattern, err := AnalyzePattern(cards)
	if err != nil {
		return "建议出牌"
	}
	switch pattern.Type {
	case PlaySingle:
		return "建议出单张"
	case PlayPair:
		return "建议出对子"
	case PlayTriple:
		return "建议出三张"
	case PlayBomb:
		return "建议出炸弹"
	case PlayRocket:
		return "建议出王炸"
	default:
		return "建议出牌"
	}
}
