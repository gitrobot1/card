package engine

import "testing"

// TestGuoHeCanTargetVineArmor 验证过河拆桥可以目标有藤甲的玩家（藤甲不阻挡过河拆桥）
func TestGuoHeCanTargetVineArmor(t *testing.T) {
	g, err := NewSolo1v1("vine-guoh-e", "甲", CharLvMeng, CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}

	// 给目标装备藤甲
	g.Players[1].Armor = &Card{ID: "vine-1", Kind: CardArmorVine, Name: "藤甲"}
	
	// 验证藤甲不会阻挡过河拆桥的目标选择
	blocked := g.vineBlocksTrick(1, CardGuoHe)
	if blocked {
		t.Fatal("vine armor should NOT block guohe")
	}

	// 验证目标选择正常工作
	g.Players[0].Hand = []Card{
		{ID: "guohe-1", Kind: CardGuoHe, Name: "过河拆桥"},
	}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = 0

	targets := g.validPlayTargets(0, CardGuoHe)
	found := false
	for _, t := range targets {
		if t == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("want target[1] with vine armor valid for guohe, got", targets)
	}
}

// TestTanNangCanTargetVineArmor 验证探囊取物可以目标有藤甲的玩家（藤甲不阻挡探囊取物）
func TestTanNangCanTargetVineArmor(t *testing.T) {
	g, err := NewSolo1v1("vine-tannang", "甲", CharLvMeng, CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}

	// 给目标装备藤甲
	g.Players[1].Armor = &Card{ID: "vine-1", Kind: CardArmorVine, Name: "藤甲"}
	
	// 验证藤甲不会阻挡探囊取物的目标选择
	blocked := g.vineBlocksTrick(1, CardTanNang)
	if blocked {
		t.Fatal("vine armor should NOT block tannang")
	}

	// 验证目标选择正常工作
	g.Players[0].Hand = []Card{
		{ID: "tannang-1", Kind: CardTanNang, Name: "探囊取物"},
	}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = 0

	targets := g.validPlayTargets(0, CardTanNang)
	found := false
	for _, t := range targets {
		if t == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("want target[1] with vine armor valid for tannang, got", targets)
	}
}

// TestVineArmorBlocksCorrectTricks 验证藤甲只阻挡南蛮入侵和万箭齐发
func TestVineArmorBlocksCorrectTricks(t *testing.T) {
	g, err := NewSolo1v1("vine-blocks", "甲", CharLvMeng, CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}

	// 给目标装备藤甲
	g.Players[1].Armor = &Card{ID: "vine-1", Kind: CardArmorVine, Name: "藤甲"}

	// 应该被阻挡的锦囊：南蛮入侵、万箭齐发
	shouldBlock := []string{CardNanMan, CardWanJian}
	for _, kind := range shouldBlock {
		if !g.vineBlocksTrick(1, kind) {
			t.Fatalf("vine armor should block %s", kind)
		}
	}

	// 不应该被阻挡的锦囊：决斗、火攻、过河拆桥、探囊取物等
	shouldNotBlock := []string{CardJueDou, CardHuoGong, CardGuoHe, CardTanNang, CardLeBu, CardBingLiang, CardTaoYuan, CardWuZhong}
	for _, kind := range shouldNotBlock {
		if g.vineBlocksTrick(1, kind) {
			t.Fatalf("vine armor should NOT block %s", kind)
		}
	}
}

// TestVineArmorBlocksNormalSha 验证藤甲让普通杀无效
func TestVineArmorBlocksNormalSha(t *testing.T) {
	g, err := NewSolo1v1("vine-sha", "甲", CharLvMeng, CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}

	// 给目标装备藤甲
	g.Players[1].Armor = &Card{ID: "vine-1", Kind: CardArmorVine, Name: "藤甲"}

	// 普通杀对装备藤甲的角色无效
	normalSha := Card{ID: "sha-1", Kind: CardSha, Name: "杀", DamageType: DamageTypeNormal}
	damage := g.applyDamage(0, 1, 1, normalSha, &[]GameEvent{})
	if damage != 0 {
		t.Fatal("vine armor should block normal sha damage")
	}

	// 火杀和雷杀不会被藤甲无效
	fireSha := Card{ID: "sha-2", Kind: CardSha, Name: "火杀", DamageType: DamageTypeFire}
	damage = g.applyDamage(0, 1, 1, fireSha, &[]GameEvent{})
	if damage == 0 {
		t.Fatal("vine armor should NOT block fire sha")
	}

	thunderSha := Card{ID: "sha-3", Kind: CardSha, Name: "雷杀", DamageType: DamageTypeThunder}
	g.Players[1].HP = 4 // 恢复血量
	damage = g.applyDamage(0, 1, 1, thunderSha, &[]GameEvent{})
	if damage == 0 {
		t.Fatal("vine armor should NOT block thunder sha")
	}
}
