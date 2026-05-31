package uno

import (
	"fmt"
	"math/rand"
)

type Color string

const (
	ColorRed    Color = "red"
	ColorYellow Color = "yellow"
	ColorGreen  Color = "green"
	ColorBlue   Color = "blue"
	ColorWild   Color = "wild"
)

var PlayColors = []Color{ColorRed, ColorYellow, ColorGreen, ColorBlue}

type Value string

const (
	ValueSkip    Value = "skip"
	ValueReverse Value = "reverse"
	ValueDraw2   Value = "draw2"
	ValueWild    Value = "wild"
	ValueWild4   Value = "wild4"
)

type Card struct {
	ID    string `json:"id"`
	Color Color  `json:"color"`
	Value string `json:"value"`
	Label string `json:"label"`
}

func cardLabel(color Color, value string) string {
	switch Value(value) {
	case ValueSkip:
		return "跳过"
	case ValueReverse:
		return "反转"
	case ValueDraw2:
		return "+2"
	case ValueWild:
		return "变色"
	case ValueWild4:
		return "+4"
	default:
		return value
	}
}

func NewDeck108() []Card {
	var deck []Card
	seq := 0
	add := func(color Color, value string, count int) {
		for i := 0; i < count; i++ {
			seq++
			deck = append(deck, Card{
				ID:    fmt.Sprintf("uno-%d", seq),
				Color: color,
				Value: value,
				Label: cardLabel(color, value),
			})
		}
	}
	for _, color := range PlayColors {
		add(color, "0", 1)
		for n := 1; n <= 9; n++ {
			add(color, fmt.Sprintf("%d", n), 2)
		}
		add(color, string(ValueSkip), 2)
		add(color, string(ValueReverse), 2)
		add(color, string(ValueDraw2), 2)
	}
	add(ColorWild, string(ValueWild), 4)
	add(ColorWild, string(ValueWild4), 4)
	return deck
}

func ShuffleDeck(deck []Card) []Card {
	out := append([]Card(nil), deck...)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func IsWildCard(c Card) bool {
	return c.Color == ColorWild || Value(c.Value) == ValueWild || Value(c.Value) == ValueWild4
}

func IsActionValue(v string) bool {
	switch Value(v) {
	case ValueSkip, ValueReverse, ValueDraw2, ValueWild, ValueWild4:
		return true
	default:
		return false
	}
}
