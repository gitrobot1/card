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
	Success     bool   `json:"success,omitempty"`
}

// PlayerGameStats 单玩家对局统计（游戏结束时填充，预留扩展）。
// 目前只记录基础数据，后续可扩展击杀数、伤害量、治疗量等。
type PlayerGameStats struct {
	Seat         int    `json:"seat"`
	Name          string `json:"name"`
	CharacterID   string `json:"character_id"`
	IsWinner      bool   `json:"is_winner"`
	DamageDealt   int    `json:"damage_dealt,omitempty"`   // 造成的总伤害
	DamageTaken   int    `json:"damage_taken,omitempty"`   // 受到的总伤害
	HealDone      int    `json:"heal_done,omitempty"`      // 治疗量
	KillCount     int    `json:"kill_count,omitempty"`      // 击杀数
	SurvivalRank  int    `json:"survival_rank,omitempty"`  // 存活排名（1=最先阵亡）
}

// GameOverStats 游戏结束时的统计数据（预留，暂未全部填充）。
// 前端可在结算界面展示 MVP、伤害统计等信息。
type GameOverStats struct {
	WinnerIndex   int                `json:"winner_index"`
	WinnerTeam    int                `json:"winner_team,omitempty"`
	Reason        string             `json:"reason,omitempty"` // damage | hp_loss | timeout | surrender
	PlayerStats   []PlayerGameStats `json:"player_stats,omitempty"`
}

// Character 角色元数据。
type Character struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	MaxHP         int                 `json:"max_hp"`
	Kingdom       string              `json:"kingdom,omitempty"`
	Gender        string              `json:"gender,omitempty"` // male | female
	SkillIDs      []string            `json:"skill_ids,omitempty"`
	Skills        []SkillMeta         `json:"skills,omitempty"`
	DefaultSkinID string              `json:"default_skin_id,omitempty"`
	SkinID        string              `json:"skin_id,omitempty"`
	Display       *skill.HeroDisplay  `json:"display,omitempty"`
}

// Card 基础牌。
type Card struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Suit       string `json:"suit,omitempty"`
	Rank       int    `json:"rank,omitempty"`
	Label      string `json:"label,omitempty"`
	Name       string `json:"name"`
	TrickScope string `json:"trick_scope,omitempty"`
	DamageType string `json:"damage_type,omitempty"`
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
	CampCards       []Card         `json:"camp_cards,omitempty"`
	SkillCounters   map[string]int `json:"skill_counters,omitempty"`
}

type PlayTarget struct {
	SeatIndex       int
	SecondSeatIndex int // 铁索连环双目标的第二个目标（0 表示无）
	Zone            string
	CardID          string
}

// PendingCombat 等待目标响应的杀、锦囊或无懈可击窗口。
type PendingCombat struct {
	SourceIndex  int    `json:"source_index"`
	TargetIndex  int    `json:"target_index"`
	ReturnIndex  int    `json:"return_index"`
	EffectTarget int    `json:"effect_target,omitempty"`
	SecondTargetIndex int `json:"second_target_index,omitempty"` // 铁索连环双目标
	Card         Card   `json:"card"`
	RequiredKind string `json:"required_kind,omitempty"`
	Damage       int    `json:"damage,omitempty"`
	ResponseMode string `json:"response_mode,omitempty"`
	TargetZone   string `json:"target_zone,omitempty"`
	TargetCardID string `json:"target_card_id,omitempty"`
	AllowWuxiek   bool   `json:"allow_wuxiek,omitempty"`
	TaoYuanQueue  bool   `json:"-"` // 桃园结义队列标记
	ResponsesNeeded int  `json:"responses_needed,omitempty"`
	BaguaUsed     bool   `json:"bagua_used,omitempty"`
	IgnoreArmor   bool   `json:"ignore_armor,omitempty"`
	RevealedCards    []Card `json:"revealed_cards,omitempty"`
	WuguRevealedAll  []Card `json:"wugu_revealed_all,omitempty"` // 五谷初始亮牌完整列表（框展示用，不变）
	WuguPickSeat     int    `json:"wugu_pick_seat,omitempty"`
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
	PojunMax        int `json:"pojun_max,omitempty"`
	PojunPlaced     int `json:"pojun_placed,omitempty"`
	PojunRemaining  int `json:"pojun_remaining,omitempty"`
	SavedPending     *PendingCombat `json:"-"`
	// 改判队列：按座位顺序收集的候选人及当前索引
	ModifyCandidates []int `json:"-"`
	ModifyIndex      int   `json:"-"`

	// Extra 通用扩展字段，用于技能逻辑传递中间状态
	Extra map[string]int `json:"extra,omitempty"`

	// 响应队列：按照三国杀规则管理响应顺序
	ResponseQueue []int `json:"response_queue,omitempty"` // 响应队列，按顺序排列
	ResponseIndex int   `json:"response_index,omitempty"` // 当前响应者在队列中的索引
	
	// 语义字段（v0.1）：优先于 SourceIndex/TargetIndex 推导；由 FillPendingRoles 填充。
	ActorSeat   int    `json:"actor_seat"`
	SubjectSeat int    `json:"subject_seat,omitempty"`
	OriginSeat  int    `json:"origin_seat,omitempty"`
	WindowKind  string `json:"window_kind,omitempty"`

	// 无懈可击链：记录所有打出的无懈可击顺序（最后一张是最新的）
	WuxiekChain []WuxiekEntry `json:"-"`
}

// WuxiekEntry 无懈可击链中的一条记录
type WuxiekEntry struct {
	Seat int    // 谁打出的
	Card Card   // 打出的无懈可击牌
}

// PlayerPublic 对外公开的玩家信息。
type PlayerPublic struct {
	Index           int       `json:"index"`
	Name            string    `json:"name"`
	IsAI            bool      `json:"is_ai"`
	Team            int       `json:"team"`
	Identity        string    `json:"identity,omitempty"`
	IdentityRevealed bool     `json:"identity_revealed,omitempty"`
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
	CampCards       []Card         `json:"camp_cards,omitempty"`
	SkillCounters   map[string]int `json:"skill_counters,omitempty"`
	Hand            []Card         `json:"hand,omitempty"`
}
