package engine

const (
	MinPlayers = 2
	MaxPlayers = 4

	PhasePlaying   = "playing"
	PhaseResponse  = "response"
	PhaseHPChange  = "hp_change" // 血量变化阶段
	PhaseFinished  = "finished"

	StepStart    = "start"
	StepPrepare  = "prepare"
	StepJudge    = "judge"
	StepDraw     = "draw"
	StepPlay     = "play"
	StepDiscard  = "discard"
	StepFinish   = "finish"

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
	CardHuoGong   = "huogong"
	CardTieSuo    = "tiesuo"
	CardJieDao    = "jiedao"   // 借刀杀人

	// 属性杀
	CardShaFire    = "sha_fire"
	CardShaThunder = "sha_thunder"

	// 伤害类型
	DamageTypeNormal  = "normal"
	DamageTypeFire    = "fire"
	DamageTypeThunder = "thunder"

	// 朱雀羽扇
	CardWeapon7 = "weapon_7"
	// 雌雄双股剑
	CardWeapon8 = "weapon_8"
	// 贯石斧
	CardWeapon9 = "weapon_9"

	// 锦囊作用域
	TrickScopeSingle = "single"
	TrickScopeAoe    = "aoe"

	ResponseModeCard       = "card"
	ResponseModeWuxiekTrick = "wuxiek_trick"
	ResponseModeWuxiekLebu  = "wuxiek_lebu"
	ResponseModeWuxiekBingliang = "wuxiek_bingliang"
	ResponseModeWuxiekShandian  = "wuxiek_shandian"
	ResponseModeWuxiekGuose     = "wuxiek_guose"
	ResponseModePeekDeck        = "peek_deck"
	ResponseModeWuguPick          = "wugu_pick"
	ResponseModeTaoYuanAnnounce   = "taoyuan_announce"
	ResponseModeGuanYuFollow   = "guanyu_follow"
	ResponseModeQilinBow     = "qilin_bow"
	ResponseModeWeapon8     = "weapon_8"
	ResponseModeWeapon9     = "weapon_9"
	ResponseModeGuoHe       = "guohe"
	ResponseModeTanNang     = "tannang"
	ResponseModeJieDao      = "jiedao"       // 借刀杀人：被借刀者选择出杀或给武器

	CardWeapon1    = "weapon_1"
	CardWeapon2    = "weapon_2"
	CardWeapon3    = "weapon_3"
	CardWeapon4    = "weapon_4"
	CardWeapon5    = "weapon_5"
	CardWeapon6    = "weapon_6"
	CardWeapon10   = "weapon_10"  // 丈八蛇矛
	CardArmor       = "armor"        // 八卦阵
	CardArmorVine   = "armor_vine"   // 藤甲
	CardArmorRenwang = "armor_renwang" // 仁王盾
	CardArmorBaiyin  = "armor_baiyin"  // 白银狮子
	CardPlusHorse  = "plus_horse"
	CardMinusHorse = "minus_horse"

	EquipWeapon     = "weapon"
	EquipArmor      = "armor"
	EquipPlusHorse  = "plus_horse"
	EquipMinusHorse = "minus_horse"
)
