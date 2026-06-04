package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

type soloStartParams struct {
	gameID      string
	humanName   string
	humanCharID string
	aiCharID    string // 1v1 only: empty = random
}

func NewSolo(id, humanName, humanCharID, gameMode string) (*Game, error) {
	modeID := mode.NormalizeID(gameMode)
	if _, ok := mode.Lookup(modeID); !ok {
		return nil, fmt.Errorf("unknown mode: %s", gameMode)
	}
	switch modeID {
	case mode.Solo2v2:
		return setupSolo2v2(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	case mode.Solo3pChain:
		return setupSolo3pChain(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	case mode.Solo3pDdz:
		return setupSolo3pDdz(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	default:
		return setupSolo1v1(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	}
}

func NewSolo1v1(id, humanName, humanCharID, aiCharID string) (*Game, error) {
	return setupSolo1v1(soloStartParams{
		gameID: id, humanName: humanName, humanCharID: humanCharID, aiCharID: aiCharID,
	})
}

func NewSolo2v2(id, humanName, humanCharID string) (*Game, error) {
	return setupSolo2v2(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func NewSolo3pChain(id, humanName, humanCharID string) (*Game, error) {
	return setupSolo3pChain(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func NewSolo3pDdz(id, humanName, humanCharID string) (*Game, error) {
	return setupSolo3pDdz(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func setupSolo3pDdz(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(Mode3pDdz, humanCharID); err != nil {
		return nil, err
	}
	used := map[string]bool{humanCharID: true}
	pickAI := func() string {
		for tries := 0; tries < 64; tries++ {
			c := RandomAICharacter("")
			if !used[c] {
				used[c] = true
				return c
			}
		}
		for _, h := range HeroesCatalog() {
			if !used[h.ID] {
				used[h.ID] = true
				return h.ID
			}
		}
		return CharGuanYu
	}
	left := pickAI()
	right := pickAI()

	g := &Game{
		ID:           p.gameID,
		HumanPlayer:  0,
		Phase:        PhasePlaying,
		Mode:         Mode3pDdz,
		LandlordSeat: 0,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName + "·地主", false, humanCharID},
		{1, "农民·左", true, left},
		{2, "农民·右", true, right},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("斗地主：%s 担任地主，对抗两名农民", p.humanName))
}

func setupSolo3pChain(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(Mode3pChain, humanCharID); err != nil {
		return nil, err
	}
	used := map[string]bool{humanCharID: true}
	pickAI := func() string {
		for tries := 0; tries < 64; tries++ {
			c := RandomAICharacter("")
			if !used[c] {
				used[c] = true
				return c
			}
		}
		for _, h := range HeroesCatalog() {
			if !used[h.ID] {
				used[h.ID] = true
				return h.ID
			}
		}
		return CharGuanYu
	}
	left := pickAI()
	right := pickAI()

	g := &Game{
		ID:          p.gameID,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode3pChain,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName, false, humanCharID},
		{1, "下家·左", true, left},
		{2, "上家·右", true, right},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("杀上保下：%s 先手（左=下家需保护，右=上家需击杀）", p.humanName))
}

func setupSolo1v1(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(Mode1v1, humanCharID); err != nil {
		return nil, err
	}
	aiCharID := p.aiCharID
	if aiCharID == "" {
		aiCharID = RandomAICharacter(humanCharID)
	}
	if err := validateCharacterIDStatic(aiCharID); err != nil {
		return nil, err
	}
	humChar := buildCharacter(humanCharID)
	aiChar := buildCharacter(aiCharID)

	g := &Game{
		ID:          p.gameID,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode1v1,
	}
	g.Players = []Player{
		{Index: 0, Name: p.humanName, IsAI: false, Character: humChar, MaxHP: humChar.MaxHP, HP: humChar.MaxHP},
		{Index: 1, Name: "电脑", IsAI: true, Character: aiChar, MaxHP: aiChar.MaxHP, HP: aiChar.MaxHP},
	}
	return finishSoloSetup(g, fmt.Sprintf("%s 先手，请出牌", p.humanName))
}

func setupSolo2v2(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(Mode2v2, humanCharID); err != nil {
		return nil, err
	}
	used := map[string]bool{humanCharID: true}
	pickAI := func() string {
		for tries := 0; tries < 64; tries++ {
			c := RandomAICharacter("")
			if !used[c] {
				used[c] = true
				return c
			}
		}
		for _, h := range HeroesCatalog() {
			if !used[h.ID] {
				used[h.ID] = true
				return h.ID
			}
		}
		return CharGuanYu
	}
	enemy1 := pickAI()
	ally := pickAI()
	enemy2 := pickAI()

	g := &Game{
		ID:          p.gameID,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode2v2,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName, false, humanCharID},
		{1, "敌将·左", true, enemy1},
		{2, "队友", true, ally},
		{3, "敌将·右", true, enemy2},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("2v2：%s 先手（十字阵：你-下，队友-上，敌将在两侧）", p.humanName))
}

// NewSolo2v2WithHeroes creates a 4-player 2v2 game with explicit hero IDs per seat.
// Seats: 0 bottom, 1 left enemy, 2 top ally, 3 right enemy.
func NewSolo2v2WithHeroes(id string, seatHeroes [4]string) (*Game, error) {
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode2v2, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in 2v2 lineup: %s", heroID)
		}
		used[heroID] = true
	}
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode2v2,
	}
	roles := []struct {
		seat int
		name string
		char string
	}{
		{0, "甲", seatHeroes[0]},
		{1, "敌将·左", seatHeroes[1]},
		{2, "队友", seatHeroes[2]},
		{3, "敌将·右", seatHeroes[3]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: true,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("2v2：%s 先手（测试盘）", seatHeroes[0]))
}

// NewSolo3pChainWithHeroes creates a 3-player chain game with explicit hero IDs per seat.
// Seats: 0 human/bottom, 1 lower (protect), 2 upper (mark).
func NewSolo3pChainWithHeroes(id string, seatHeroes [3]string) (*Game, error) {
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode3pChain, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in 3p chain lineup: %s", heroID)
		}
		used[heroID] = true
	}
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode3pChain,
	}
	roles := []struct {
		seat int
		name string
		char string
	}{
		{0, "甲", seatHeroes[0]},
		{1, "下家·左", seatHeroes[1]},
		{2, "上家·右", seatHeroes[2]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: true,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("杀上保下：%s 先手（测试盘）", seatHeroes[0]))
}

// NewSolo3pDdzWithHeroes creates a 3-player ddz game with explicit hero IDs per seat.
// Seats: 0 landlord, 1 farmer left, 2 farmer right.
func NewSolo3pDdzWithHeroes(id string, seatHeroes [3]string) (*Game, error) {
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode3pDdz, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in 3p ddz lineup: %s", heroID)
		}
		used[heroID] = true
	}
	g := &Game{
		ID:           id,
		HumanPlayer:  0,
		Phase:        PhasePlaying,
		Mode:         Mode3pDdz,
		LandlordSeat: 0,
	}
	roles := []struct {
		seat int
		name string
		char string
	}{
		{0, "地主", seatHeroes[0]},
		{1, "农民·左", seatHeroes[1]},
		{2, "农民·右", seatHeroes[2]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: true,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("斗地主：%s 担任地主（测试盘）", seatHeroes[0]))
}

func finishSoloSetup(g *Game, message string) (*Game, error) {
	g.setupDeck()
	g.CurrentTurn = 0
	g.beginTurn(nil)
	g.Message = message
	return g, nil
}

// ModesCatalog returns registered mode metadata for API clients.
func ModesCatalog() []mode.Meta {
	return mode.All()
}

// ModeMeta is the public API shape for mode metadata.
type ModeMeta = mode.Meta
