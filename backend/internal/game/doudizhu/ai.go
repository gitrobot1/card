package doudizhu

import (
	"math/rand"

	"github.com/time/card/backend/internal/game/card"
)
func RunAITurns(game *Game, events *[]GameEvent) {
	for !game.IsFinished() && !game.IsHumanTurn() {
		if game.Phase == PhaseCalling {
			playerIndex := game.CallingIndex
			want := autoCall(game, playerIndex)
			name := game.Players[playerIndex].Name
			if err := game.CallLandlord(playerIndex, want); err != nil {
				return
			}
			appendCallEvent(events, playerIndex, name, want)
			continue
		}

		playerIndex := game.CurrentTurn
		name := game.Players[playerIndex].Name
		hand := game.Players[playerIndex].Hand
		if len(hand) == 0 {
			if !game.IsFinished() {
				game.finish(playerIndex)
			}
			return
		}

		if game.LastPlay == nil {
			cards := pickSmallestPattern(hand)
			record, err := game.Play(playerIndex, cardIDs(cards))
			if err != nil {
				return
			}
			appendPlayEvent(events, record)
			if game.IsFinished() {
				return
			}
			continue
		}

		if game.LastPlay.PlayerIndex != playerIndex {
			if beat := findBeatingCards(hand, game.LastPlay.Cards); len(beat) > 0 {
				record, err := game.Play(playerIndex, cardIDs(beat))
				if err != nil {
					return
				}
				appendPlayEvent(events, record)
				if game.IsFinished() {
					return
				}
				continue
			}
			if err := game.Pass(playerIndex); err != nil {
				return
			}
			appendPassEvent(events, playerIndex, name)
			continue
		}

		cards := pickSmallestPattern(hand)
		record, err := game.Play(playerIndex, cardIDs(cards))
		if err != nil {
			return
		}
		appendPlayEvent(events, record)
		if game.IsFinished() {
			return
		}
	}
}

func autoCall(game *Game, playerIndex int) bool {
	if playerIndex == game.HumanPlayer {
		return false
	}
	return rand.Intn(100) < 35
}

func pickSmallestPattern(hand []card.Card) []card.Card {
	sorted := append([]card.Card(nil), hand...)
	card.SortByRank(sorted)
	if len(sorted) > 0 {
		return []card.Card{sorted[0]}
	}
	return nil
}

func findBeatingCards(hand []card.Card, previous []card.Card) []card.Card {
	prevPattern, err := AnalyzePattern(previous)
	if err != nil {
		return nil
	}

	candidates := generateCandidates(hand, prevPattern)
	for _, candidate := range candidates {
		pattern, err := AnalyzePattern(candidate)
		if err != nil {
			continue
		}
		if CanBeat(pattern, prevPattern) {
			return candidate
		}
	}
	return nil
}

func generateCandidates(hand []card.Card, previous *HandPattern) [][]card.Card {
	sorted := append([]card.Card(nil), hand...)
	card.SortByRank(sorted)
	counts := make(map[card.Rank][]card.Card)
	for _, c := range sorted {
		counts[c.Rank] = append(counts[c.Rank], c)
	}

	var result [][]card.Card
	switch previous.Type {
	case PlaySingle:
		for _, c := range sorted {
			if int(c.Rank) > previous.Weight {
				result = append(result, []card.Card{c})
			}
		}
	case PlayPair:
		for rank, cards := range counts {
			if len(cards) >= 2 && int(rank) > previous.Weight {
				result = append(result, []card.Card{cards[0], cards[1]})
			}
		}
	case PlayTriple:
		for rank, cards := range counts {
			if len(cards) >= 3 && int(rank) > previous.Weight {
				result = append(result, []card.Card{cards[0], cards[1], cards[2]})
			}
		}
	case PlayBomb:
		for rank, cards := range counts {
			if len(cards) == 4 && int(rank) > previous.Weight {
				result = append(result, cards)
			}
		}
	case PlayRocket:
		return nil
	}

	if previous.Type != PlayRocket {
		for _, cards := range counts {
			if len(cards) == 4 {
				result = append(result, cards)
			}
		}
		if counts[card.RankSJ] != nil && counts[card.RankBJ] != nil {
			result = append(result, []card.Card{counts[card.RankSJ][0], counts[card.RankBJ][0]})
		}
	}

	return result
}

func cardIDs(cards []card.Card) []string {
	ids := make([]string, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}
