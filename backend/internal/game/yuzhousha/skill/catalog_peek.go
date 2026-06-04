package skill

// catalogPeekDeckSkills 准备阶段「看牌堆顶并分配」类技能（声明式注册）。
func catalogPeekDeckSkills() []Decl {
	return []Decl{
		{
			Meta: Meta{
				ID: IDGuanxing, Name: "观星", Kind: KindActive,
				Desc: "准备阶段，你可以观看牌堆顶 X 张牌（X 为存活角色数且至多为 5），然后将其中任意数量的牌以任意顺序置于牌堆顶，其余置于牌堆底。",
			},
			PeekDeck: &PeekDeckConfig{
				MaxPeekCap:  5,
				AIPartition: guanxingAIPartition,
			},
			CanActivate: canActivatePreparePeek(IDGuanxing),
			Activate:    activatePreparePeek(IDGuanxing),
			AIPriority:  preparePeekAIPriority(IDGuanxing),
			AIActivate:  activatePreparePeekAI(IDGuanxing),
		},
	}
}

func canActivatePreparePeek(skillID string) func(Runtime, int) bool {
	return func(r Runtime, seat int) bool {
		if !r.HasSkill(seat, skillID) {
			return false
		}
		if r.Phase() != "playing" || r.TurnStep() != "prepare" || r.CurrentTurn() != seat {
			return false
		}
		h, ok := Lookup(skillID)
		if !ok || h.PeekDeckConfig() == nil {
			return false
		}
		return PeekCountFor(r, seat, h) > 0
	}
}

func activatePreparePeek(skillID string) func(Runtime, int, ActivateReq) error {
	return func(r Runtime, seat int, _ ActivateReq) error {
		return r.StartPeekDeck(seat, skillID)
	}
}

func activatePreparePeekAI(skillID string) func(Runtime, int) error {
	return func(r Runtime, seat int) error {
		return r.StartPeekDeck(seat, skillID)
	}
}

func preparePeekAIPriority(skillID string) func(Runtime, int) int {
	return func(r Runtime, seat int) int {
		if !canActivatePreparePeek(skillID)(r, seat) {
			return 0
		}
		return 50
	}
}

func guanxingAIPartition(_ Runtime, _ int, cards []PeekCardView) (topIDs, bottomIDs []string) {
	preferTop := map[string]int{
		"tao":  100,
		"shan": 80,
		"sha":  20,
		"jiu":  60,
	}
	type scored struct {
		id    string
		score int
	}
	items := make([]scored, 0, len(cards))
	for _, c := range cards {
		score := preferTop[c.Kind]
		if score == 0 {
			score = 40
		}
		items = append(items, scored{id: c.ID, score: score})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].score > items[i].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	split := len(items) / 2
	if split < 1 && len(items) > 1 {
		split = 1
	}
	for i, it := range items {
		if i < split {
			topIDs = append(topIDs, it.id)
		} else {
			bottomIDs = append(bottomIDs, it.id)
		}
	}
	return topIDs, bottomIDs
}
