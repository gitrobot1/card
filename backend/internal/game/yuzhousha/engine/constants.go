package engine

const (
	MinPlayers = 2
	MaxPlayers = 4

	PhasePlaying  = "playing"
	PhaseResponse = "response"
	PhaseFinished = "finished"

	StepPrepare = "prepare"
	StepDraw    = "draw"
	StepPlay    = "play"
	StepDiscard = "discard"

	DefaultMaxHP    = 4
	InitialHandSize = 4
	DrawPerTurn     = 2
	TurnTimeoutSec  = 35

	CardSha  = "sha"
	CardShan = "shan"
	CardTao  = "tao"
	CardJiu  = "jiu"

	CardGuoHe   = "guohe"
	CardTanNang = "tannang"
	CardNanMan  = "nanman"
	CardWanJian = "wanjian"
	CardJueDou  = "juedou"
	CardLeBu    = "lebu"
	CardBingLiang = "bingliang"
	CardShanDian  = "shandian"
	CardWuGu      = "wugu"
	CardTaoYuan = "taoyuan"
	CardWuZhong = "wuzhong"
	CardWuxiek    = "wuxiek"

	ResponseModeCard       = "card"
	ResponseModeWuxiekTrick = "wuxiek_trick"
	ResponseModeWuxiekLebu  = "wuxiek_lebu"
	ResponseModeWuxiekBingliang = "wuxiek_bingliang"
	ResponseModeWuxiekShandian  = "wuxiek_shandian"
	ResponseModePeekDeck        = "peek_deck"
	ResponseModeWuguPick        = "wugu_pick"
	ResponseModeGuanYuFollow = "guanyu_follow"
	ResponseModeQilinBow     = "qilin_bow"

	CardWeapon1    = "weapon_1"
	CardWeapon2    = "weapon_2"
	CardWeapon3    = "weapon_3"
	CardWeapon4    = "weapon_4"
	CardWeapon5    = "weapon_5"
	CardArmor      = "armor"
	CardPlusHorse  = "plus_horse"
	CardMinusHorse = "minus_horse"

	EquipWeapon     = "weapon"
	EquipArmor      = "armor"
	EquipPlusHorse  = "plus_horse"
	EquipMinusHorse = "minus_horse"
)
