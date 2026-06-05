package engine

import (
	"fmt"
	"math/rand"
	"time"

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
	case mode.Solo3v3:
		return setupSolo3v3(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	case mode.SoloIdentity5:
		return setupSoloIdentity5(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	case mode.SoloIdentity8:
		return setupSoloIdentity8(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	default:
		return setupSolo1v1(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
	}
}

func NewSolo1v1(id, humanName, humanCharID, aiCharID string) (*Game, error) {
	return setupSolo1v1(soloStartParams{
		gameID: id, humanName: humanName, humanCharID: humanCharID, aiCharID: aiCharID,
	})
}

// NewOnline1v1 创建双人在线 1v1 对局（无 AI）。
func NewOnline1v1(id string, names [2]string, charIDs [2]string) (*Game, error) {
	for i, charID := range charIDs {
		if charID == "" {
			return nil, fmt.Errorf("player %d character required", i)
		}
		if err := validateCharacterIDStatic(charID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode1v1, charID); err != nil {
			return nil, err
		}
	}
	ch0 := buildCharacter(charIDs[0])
	ch1 := buildCharacter(charIDs[1])
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode1v1,
	}
	g.Players = []Player{
		{Index: 0, Name: names[0], IsAI: false, Character: ch0, MaxHP: ch0.MaxHP, HP: ch0.MaxHP},
		{Index: 1, Name: names[1], IsAI: false, Character: ch1, MaxHP: ch1.MaxHP, HP: ch1.MaxHP},
	}
	return finishSoloSetup(g, fmt.Sprintf("%s 先手，请出牌", names[0]))
}

// NewOnline2v2 创建四人在线 2v2 对局（无 AI）。座位：0 下、1 敌左、2 队友、3 敌右。
func NewOnline2v2(id string, names [4]string, charIDs [4]string) (*Game, error) {
	used := map[string]bool{}
	for i, charID := range charIDs {
		if charID == "" {
			return nil, fmt.Errorf("player %d character required", i)
		}
		if err := validateCharacterIDStatic(charID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode2v2, charID); err != nil {
			return nil, err
		}
		if used[charID] {
			return nil, fmt.Errorf("duplicate hero in 2v2 lineup: %s", charID)
		}
		used[charID] = true
	}
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode2v2,
	}
	g.Players = make([]Player, 4)
	for i := range names {
		ch := buildCharacter(charIDs[i])
		g.Players[i] = Player{
			Index: i, Name: names[i], IsAI: false,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("2v2：%s 先手（十字阵）", names[0]))
}

// NewOnline3pChain 创建三人在线杀上保下对局（无 AI）。座位：0 下、1 下家、2 上家。
func NewOnline3pChain(id string, names [3]string, charIDs [3]string) (*Game, error) {
	used := map[string]bool{}
	for i, charID := range charIDs {
		if charID == "" {
			return nil, fmt.Errorf("player %d character required", i)
		}
		if err := validateCharacterIDStatic(charID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode3pChain, charID); err != nil {
			return nil, err
		}
		if used[charID] {
			return nil, fmt.Errorf("duplicate hero in 3p chain lineup: %s", charID)
		}
		used[charID] = true
	}
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode3pChain,
	}
	g.Players = make([]Player, 3)
	for i := range names {
		ch := buildCharacter(charIDs[i])
		g.Players[i] = Player{
			Index: i, Name: names[i], IsAI: false,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("杀上保下：%s 先手", names[0]))
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

func NewSolo3v3(id, humanName, humanCharID string) (*Game, error) {
	return setupSolo3v3(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func setupSolo3v3(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(Mode3v3, humanCharID); err != nil {
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
	aiChars := [5]string{}
	for i := range aiChars {
		aiChars[i] = pickAI()
	}

	g := &Game{
		ID:          p.gameID,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode3v3,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName + "·主帅", false, humanCharID},
		{1, "冷前锋·左", true, aiChars[0]},
		{2, "冷主帅", true, aiChars[1]},
		{3, "冷前锋·右", true, aiChars[2]},
		{4, "暖前锋·左", true, aiChars[3]},
		{5, "暖前锋·右", true, aiChars[4]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("3v3：%s 担任暖色主帅，击败冷色主帅获胜", p.humanName))
}

// NewSolo3v3WithHeroes creates a 6-player 3v3 game with explicit hero IDs per seat.
// Seats: 0 warm commander, 1 cold fwd, 2 cold commander, 3 cold fwd, 4 warm fwd, 5 warm fwd.
func NewSolo3v3WithHeroes(id string, seatHeroes [6]string) (*Game, error) {
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(Mode3v3, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in 3v3 lineup: %s", heroID)
		}
		used[heroID] = true
	}
	g := &Game{
		ID:          id,
		HumanPlayer: 0,
		Phase:       PhasePlaying,
		Mode:        Mode3v3,
	}
	roles := []struct {
		seat int
		name string
		char string
	}{
		{0, "暖主帅", seatHeroes[0]},
		{1, "冷前锋·左", seatHeroes[1]},
		{2, "冷主帅", seatHeroes[2]},
		{3, "冷前锋·右", seatHeroes[3]},
		{4, "暖前锋·左", seatHeroes[4]},
		{5, "暖前锋·右", seatHeroes[5]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: true,
			Character: ch, MaxHP: ch.MaxHP, HP: ch.MaxHP,
		}
	}
	return finishSoloSetup(g, fmt.Sprintf("3v3：%s 担任暖色主帅（测试盘）", seatHeroes[0]))
}

func NewSoloIdentity5(id, humanName, humanCharID string) (*Game, error) {
	return setupSoloIdentity5(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func setupSoloIdentity5(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(ModeIdentity5, humanCharID); err != nil {
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
	aiChars := [4]string{}
	for i := range aiChars {
		aiChars[i] = pickAI()
	}
	shuffled := shuffleIdentityRoles(rand.New(rand.NewSource(time.Now().UnixNano())))
	identities, revealed := assignIdentity5Roles(shuffled)

	g := &Game{
		ID:               p.gameID,
		HumanPlayer:      0,
		Phase:            PhasePlaying,
		Mode:             ModeIdentity5,
		LordSeat:         0,
		Identities:       identities,
		RoleRevealed:     revealed,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName + "·主公", false, humanCharID},
		{1, "玩家·一", true, aiChars[0]},
		{2, "玩家·二", true, aiChars[1]},
		{3, "玩家·三", true, aiChars[2]},
		{4, "玩家·四", true, aiChars[3]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		maxHP := ch.MaxHP
		hp := maxHP
		if r.seat == g.LordSeat {
			maxHP++
			hp = maxHP
		}
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: maxHP, HP: hp,
		}
	}
	return finishIdentitySoloSetup(g, fmt.Sprintf("身份局：%s 担任主公，消灭反贼与内奸获胜", p.humanName))
}

// NewSoloIdentity5WithHeroes creates a 5-player identity game with explicit heroes and roles per seat.
func NewSoloIdentity5WithHeroes(id string, seatHeroes [5]string, seatRoles [5]string) (*Game, error) {
	roles := make([]string, len(seatRoles))
	copy(roles, seatRoles[:])
	if err := mode.ValidateIdentity5Roles(roles); err != nil {
		return nil, err
	}
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(ModeIdentity5, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in identity lineup: %s", heroID)
		}
		used[heroID] = true
	}
	lordSeat := -1
	for i, role := range roles {
		if role == mode.RoleLord {
			lordSeat = i
			break
		}
	}
	revealed := make([]bool, 5)
	for i, role := range roles {
		revealed[i] = role == mode.RoleLord
	}
	g := &Game{
		ID:               id,
		HumanPlayer:      0,
		Phase:            PhasePlaying,
		Mode:             ModeIdentity5,
		LordSeat:         lordSeat,
		Identities:       roles,
		RoleRevealed:     revealed,
	}
	names := [5]string{"主公", "一", "二", "三", "四"}
	g.Players = make([]Player, 5)
	for seat := 0; seat < 5; seat++ {
		ch := buildCharacter(seatHeroes[seat])
		maxHP := ch.MaxHP
		hp := maxHP
		if seat == lordSeat {
			maxHP++
			hp = maxHP
		}
		g.Players[seat] = Player{
			Index: seat, Name: "玩家·" + names[seat], IsAI: seat != 0,
			Character: ch, MaxHP: maxHP, HP: hp,
		}
	}
	return finishIdentitySoloSetup(g, fmt.Sprintf("身份局：%s 担任主公（测试盘）", seatHeroes[lordSeat]))
}

func NewSoloIdentity8(id, humanName, humanCharID string) (*Game, error) {
	return setupSoloIdentity8(soloStartParams{gameID: id, humanName: humanName, humanCharID: humanCharID})
}

func setupSoloIdentity8(p soloStartParams) (*Game, error) {
	humanCharID := p.humanCharID
	if humanCharID == "" {
		humanCharID = CharLiuBei
	}
	if err := validateCharacterIDStatic(humanCharID); err != nil {
		return nil, err
	}
	if err := ValidateHeroForMode(ModeIdentity8, humanCharID); err != nil {
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
	aiChars := [7]string{}
	for i := range aiChars {
		aiChars[i] = pickAI()
	}
	shuffled := shuffleIdentity8Roles(rand.New(rand.NewSource(time.Now().UnixNano())))
	identities, revealed := assignIdentity8Roles(shuffled)

	g := &Game{
		ID:           p.gameID,
		HumanPlayer:  0,
		Phase:        PhasePlaying,
		Mode:         ModeIdentity8,
		LordSeat:     0,
		Identities:   identities,
		RoleRevealed: revealed,
	}
	roles := []struct {
		seat int
		name string
		ai   bool
		char string
	}{
		{0, p.humanName + "·主公", false, humanCharID},
		{1, "玩家·一", true, aiChars[0]},
		{2, "玩家·二", true, aiChars[1]},
		{3, "玩家·三", true, aiChars[2]},
		{4, "玩家·四", true, aiChars[3]},
		{5, "玩家·五", true, aiChars[4]},
		{6, "玩家·六", true, aiChars[5]},
		{7, "玩家·七", true, aiChars[6]},
	}
	g.Players = make([]Player, len(roles))
	for _, r := range roles {
		ch := buildCharacter(r.char)
		maxHP := ch.MaxHP
		hp := maxHP
		if r.seat == g.LordSeat {
			maxHP++
			hp = maxHP
		}
		g.Players[r.seat] = Player{
			Index: r.seat, Name: r.name, IsAI: r.ai,
			Character: ch, MaxHP: maxHP, HP: hp,
		}
	}
	return finishIdentitySoloSetup(g, fmt.Sprintf("八人身份局：%s 担任主公，消灭反贼与内奸获胜", p.humanName))
}

// NewSoloIdentity8WithHeroes creates an 8-player identity game with explicit heroes and roles per seat.
func NewSoloIdentity8WithHeroes(id string, seatHeroes [8]string, seatRoles [8]string) (*Game, error) {
	roles := make([]string, len(seatRoles))
	copy(roles, seatRoles[:])
	if err := mode.ValidateIdentity8Roles(roles); err != nil {
		return nil, err
	}
	used := map[string]bool{}
	for _, heroID := range seatHeroes {
		if err := validateCharacterIDStatic(heroID); err != nil {
			return nil, err
		}
		if err := ValidateHeroForMode(ModeIdentity8, heroID); err != nil {
			return nil, err
		}
		if used[heroID] {
			return nil, fmt.Errorf("duplicate hero in identity lineup: %s", heroID)
		}
		used[heroID] = true
	}
	lordSeat := -1
	for i, role := range roles {
		if role == mode.RoleLord {
			lordSeat = i
			break
		}
	}
	revealed := make([]bool, 8)
	for i, role := range roles {
		revealed[i] = role == mode.RoleLord
	}
	g := &Game{
		ID:           id,
		HumanPlayer:  0,
		Phase:        PhasePlaying,
		Mode:         ModeIdentity8,
		LordSeat:     lordSeat,
		Identities:   roles,
		RoleRevealed: revealed,
	}
	names := [8]string{"主公", "一", "二", "三", "四", "五", "六", "七"}
	g.Players = make([]Player, 8)
	for seat := 0; seat < 8; seat++ {
		ch := buildCharacter(seatHeroes[seat])
		maxHP := ch.MaxHP
		hp := maxHP
		if seat == lordSeat {
			maxHP++
			hp = maxHP
		}
		g.Players[seat] = Player{
			Index: seat, Name: "玩家·" + names[seat], IsAI: seat != 0,
			Character: ch, MaxHP: maxHP, HP: hp,
		}
	}
	return finishIdentitySoloSetup(g, fmt.Sprintf("八人身份局：%s 担任主公（测试盘）", seatHeroes[lordSeat]))
}

func shuffleIdentityRoles(r *rand.Rand) [4]string {
	roles := []string{mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel}
	r.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })
	return [4]string{roles[0], roles[1], roles[2], roles[3]}
}

func assignIdentity5Roles(shuffled [4]string) ([]string, []bool) {
	identities := make([]string, 5)
	revealed := make([]bool, 5)
	identities[0] = mode.RoleLord
	revealed[0] = true
	for i := 0; i < 4; i++ {
		identities[i+1] = shuffled[i]
	}
	return identities, revealed
}

func shuffleIdentity8Roles(r *rand.Rand) [7]string {
	roles := []string{
		mode.RoleLoyalist, mode.RoleLoyalist,
		mode.RoleSpy,
		mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
	}
	r.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })
	return [7]string{roles[0], roles[1], roles[2], roles[3], roles[4], roles[5], roles[6]}
}

func assignIdentity8Roles(shuffled [7]string) ([]string, []bool) {
	identities := make([]string, 8)
	revealed := make([]bool, 8)
	identities[0] = mode.RoleLord
	revealed[0] = true
	for i := 0; i < 7; i++ {
		identities[i+1] = shuffled[i]
	}
	return identities, revealed
}

func finishIdentitySoloSetup(g *Game, message string) (*Game, error) {
	g.syncAllPlayerSkillsMeta()
	g.setupDeck()
	g.CurrentTurn = g.LordSeat
	g.beginTurn(nil)
	g.Message = message
	return g, nil
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
	g.syncAllPlayerSkillsMeta()
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
