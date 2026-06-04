package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

// GameEvent 对局事件（推送给前端）。
type GameEvent struct {
	Type        string `json:"type"`
	PlayerIndex int    `json:"player_index"`
	TargetIndex int    `json:"target_index"`
	Card        *Card  `json:"card,omitempty"`
	Message     string `json:"message,omitempty"`
	Damage      int    `json:"damage,omitempty"`
	Heal        int    `json:"heal,omitempty"`
	Amount      int    `json:"amount,omitempty"`
	SkillID     string `json:"skill_id,omitempty"`
}

// Character 角色元数据。
type Character struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	MaxHP         int                 `json:"max_hp"`
	Kingdom       string              `json:"kingdom,omitempty"`
	SkillIDs      []string            `json:"skill_ids,omitempty"`
	Skills        []SkillMeta         `json:"skills,omitempty"`
	DefaultSkinID string              `json:"default_skin_id,omitempty"`
	SkinID        string              `json:"skin_id,omitempty"`
	Display       *skill.HeroDisplay  `json:"display,omitempty"`
}

// Card 基础牌。
type Card struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Suit  string `json:"suit,omitempty"`
	Rank  int    `json:"rank,omitempty"`
	Label string `json:"label,omitempty"`
	Name  string `json:"name"`
}

// Player 对局中的玩家状态。
type Player struct {
	Index           int       `json:"index"`
	Name            string    `json:"name"`
	IsAI            bool      `json:"is_ai"`
	Character       Character `json:"character"`
	HP              int       `json:"hp"`
	MaxHP           int       `json:"max_hp"`
	Hand            []Card    `json:"hand,omitempty"`
	HandCount       int       `json:"hand_count"`
	ShaUsedThisTurn     bool      `json:"sha_used_this_turn"`
	ShaExtraUsedThisTurn bool     `json:"sha_extra_used_this_turn,omitempty"`
	SkipPlay        bool      `json:"skip_play"`
	SkipDraw        bool      `json:"skip_draw"`
	Drunk           bool      `json:"drunk"`
	Weapon          *Card     `json:"weapon,omitempty"`
	Armor           *Card     `json:"armor,omitempty"`
	PlusHorse       *Card     `json:"plus_horse,omitempty"`
	MinusHorse      *Card     `json:"minus_horse,omitempty"`
	JudgeArea       []Card         `json:"judge_area,omitempty"`
	SkillCounters   map[string]int `json:"skill_counters,omitempty"`
}

type PlayTarget struct {
	SeatIndex int
	Zone      string
	CardID    string
}

// PendingCombat 等待目标响应的杀、锦囊或无懈可击窗口。
type PendingCombat struct {
	SourceIndex  int    `json:"source_index"`
	TargetIndex  int    `json:"target_index"`
	ReturnIndex  int    `json:"return_index"`
	EffectTarget int    `json:"effect_target,omitempty"`
	Card         Card   `json:"card"`
	RequiredKind string `json:"required_kind,omitempty"`
	Damage       int    `json:"damage,omitempty"`
	ResponseMode string `json:"response_mode,omitempty"`
	TargetZone   string `json:"target_zone,omitempty"`
	TargetCardID string `json:"target_card_id,omitempty"`
	AllowWuxiek   bool   `json:"allow_wuxiek,omitempty"`
	ResponsesNeeded int  `json:"responses_needed,omitempty"`
	BaguaUsed     bool   `json:"bagua_used,omitempty"`
	IgnoreArmor   bool   `json:"ignore_armor,omitempty"`
	RevealedCards []Card `json:"revealed_cards,omitempty"`
	WuguPickSeat  int    `json:"wugu_pick_seat,omitempty"`
	SkillID       string `json:"skill_id,omitempty"`
	JijiangLord   int    `json:"jijiang_lord,omitempty"`
	JijiangUse    bool   `json:"jijiang_use,omitempty"`
	TieqiPending  bool   `json:"tieqi_pending,omitempty"`
	ShaUnblockable bool  `json:"sha_unblockable,omitempty"`
	FankuiRemaining int  `json:"fankui_remaining,omitempty"`
	FankuiResumeMode string `json:"fankui_resume_mode,omitempty"`
	FankuiResumeCard Card  `json:"fankui_resume_card,omitempty"`
	FankuiReturnIndex int `json:"fankui_return_index,omitempty"`
	JudgeCard      Card   `json:"judge_card,omitempty"`
	JudgeReason    string `json:"judge_reason,omitempty"`
	GuicaiResume   string `json:"guicai_resume,omitempty"`
	GuicaiJudgeSeat int `json:"guicai_judge_seat,omitempty"`
	GanglieOwner    int `json:"ganglie_owner,omitempty"`
	GanglieIndex    int `json:"ganglie_index,omitempty"` // 剩余刚烈次数，用于区分连续触发
	LuanwuSha       bool `json:"luanwu_sha,omitempty"`
	LuanwuOwner     int  `json:"luanwu_owner,omitempty"`
	AoeQueue        []int `json:"aoe_queue,omitempty"`
	TuxiRemaining   int `json:"tuxi_remaining,omitempty"`
	YijiGiveRemaining int `json:"yiji_give_remaining,omitempty"`
	SavedPending   *PendingCombat `json:"-"`
}

// PlayerPublic 对外公开的玩家信息。
type PlayerPublic struct {
	Index           int       `json:"index"`
	Name            string    `json:"name"`
	IsAI            bool      `json:"is_ai"`
	Team            int       `json:"team,omitempty"`
	Character       Character `json:"character"`
	HP              int       `json:"hp"`
	MaxHP           int       `json:"max_hp"`
	HandCount       int       `json:"hand_count"`
	ShaUsedThisTurn     bool      `json:"sha_used_this_turn"`
	ShaExtraUsedThisTurn bool     `json:"sha_extra_used_this_turn,omitempty"`
	SkipPlay        bool      `json:"skip_play"`
	SkipDraw        bool      `json:"skip_draw"`
	Drunk           bool      `json:"drunk"`
	Weapon          *Card     `json:"weapon,omitempty"`
	Armor           *Card     `json:"armor,omitempty"`
	PlusHorse       *Card     `json:"plus_horse,omitempty"`
	MinusHorse      *Card     `json:"minus_horse,omitempty"`
	JudgeArea       []Card         `json:"judge_area,omitempty"`
	SkillCounters   map[string]int `json:"skill_counters,omitempty"`
	Hand            []Card         `json:"hand,omitempty"`
}
