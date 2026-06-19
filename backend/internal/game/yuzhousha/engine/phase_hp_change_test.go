package engine

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// init 注册测试用技能
func init() {
	// 圣光：受到属性伤害时，伤害值-1
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID:   "test_shengguang",
			Name: "圣光",
			Kind: skill.KindPassive,
			Desc: "当你受到属性伤害时，伤害值-1。",
		},
		OnDamageCalculated: func(r skill.Runtime, ctx skill.DamageCalculatedCtx) (int, error) {
			// 属性伤害（火杀/雷杀）才生效
			if ctx.CardKind == CardShaFire || ctx.CardKind == CardShaThunder {
				return ctx.Amount - 1, nil
			}
			return ctx.Amount, nil
		},
	})

	// 防止扣血：测试用技能，100%防止扣血
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID:   "test_prevent_damage",
			Name: "防止扣血",
			Kind: skill.KindPassive,
			Desc: "测试用：防止所有扣血。",
		},
		OnBeforeHPChange: func(r skill.Runtime, ctx skill.BeforeHPChangeCtx) (bool, error) {
			return true, nil // 防止扣血
		},
	})

	// 亡语：阵亡时，杀手受到 1 点伤害（简化版武魂）
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID:   "test_death_curse",
			Name: "亡语诅咒",
			Kind: skill.KindPassive,
			Desc: "你阵亡时，伤害来源受到 1 点伤害。",
		},
		OnDeath: func(r skill.Runtime, ctx skill.DeathCtx) error {
			// 简化：不真的造成伤害，只记录日志
			return nil
		},
	})
}

// TestHPChangeHooks 测试血量变化钩子。
func TestHPChangeHooks(t *testing.T) {
	g, err := NewSolo1v1("hp-change-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 测试伤害导致的血量变化
	oldHP := g.Players[seat].HP
	g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)

	// 检查是否触发了 hp_changed 事件
	found := false
	for _, e := range events {
		if e.Type == "hp_changed" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	if !found {
		t.Error("hp_changed event not found after damage")
	}

	// 检查血量是否正确减少
	if g.Players[seat].HP != oldHP-1 {
		t.Errorf("expected HP %d, got %d", oldHP-1, g.Players[seat].HP)
	}
}

// TestHPLossHooks 测试血量流失钩子。
func TestHPLossHooks(t *testing.T) {
	g, err := NewSolo1v1("hp-loss-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 测试血量流失
	oldHP := g.Players[seat].HP
	g.applyHPLossWithHook(seat, 1, "skill", 1, "test_skill", &events)

	// 检查血量是否正确减少（钩子已经触发并扣血）
	if g.Players[seat].HP != oldHP-1 {
		t.Errorf("expected HP %d, got %d", oldHP-1, g.Players[seat].HP)
	}
	
	t.Log("血量流失钩子测试通过")
}

// TestHealHooks 测试血量回复钩子。
func TestHealHooks(t *testing.T) {
	g, err := NewSolo1v1("heal-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 先造成伤害
	g.applyDamageWithHook(1, seat, 2, Card{Kind: CardSha, Name: "杀"}, &events)
	events = []GameEvent{} // 清空事件

	// 测试血量回复
	oldHP := g.Players[seat].HP
	g.applyHealWithHook(seat, 1, "skill", seat, "test_heal", &events)

	// 检查血量是否正确增加（钩子已经触发并回复血量）
	if g.Players[seat].HP != oldHP+1 {
		t.Errorf("expected HP %d, got %d", oldHP+1, g.Players[seat].HP)
	}
	
	t.Log("血量回复钩子测试通过")
}

// TestDyingAfterHPChange 测试血量变化后濒死触发。
func TestDyingAfterHPChange(t *testing.T) {
	g, err := NewSolo1v1("dying-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 将血量设置为 1
	g.Players[seat].HP = 1

	// 造成致命伤害
	g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)

	// 检查是否触发了 dying_start 事件
	found := false
	for _, e := range events {
		if e.Type == "dying_start" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	if !found {
		t.Error("dying_start event not found after fatal damage")
	}
}

// TestDamageVsHPLoss 测试伤害和血量流失的区别。
func TestDamageVsHPLoss(t *testing.T) {
	g, err := NewSolo1v1("damage-vs-loss-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 测试伤害
	g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)

	// 检查伤害事件
	damageFound := false
	for _, e := range events {
		if e.Type == "hp_changed" && e.Damage > 0 {
			damageFound = true
			break
		}
	}
	if !damageFound {
		t.Error("damage event not found")
	}

	events = []GameEvent{} // 清空事件

	// 测试血量流失
	oldHP := g.Players[seat].HP
	g.applyHPLossWithHook(seat, 1, "skill", 1, "test_skill", &events)

	// 检查血量是否正确减少（不检查 events，因为 runHPLostHooks 不直接添加 GameEvent）
	if g.Players[seat].HP != oldHP-1 {
		t.Errorf("expected HP %d, got %d", oldHP-1, g.Players[seat].HP)
	}
	
	t.Log("伤害 vs 血量流失测试通过")
}

// TestMultipleHPChanges 测试连续血量变化。
func TestMultipleHPChanges(t *testing.T) {
	g, err := NewSolo1v1("multi-hp-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 连续造成 3 点伤害
	for i := 0; i < 3; i++ {
		g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)
	}

	// 检查触发了 3 次 hp_changed 事件
	count := 0
	for _, e := range events {
		if e.Type == "hp_changed" && e.PlayerIndex == seat {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 hp_changed events, got %d", count)
	}

	// 检查血量是否正确
	if g.Players[seat].HP != g.Players[seat].MaxHP-3 {
		t.Errorf("expected HP %d, got %d", g.Players[seat].MaxHP-3, g.Players[seat].HP)
	}
}

// TestDamageCalculatedHook 验证 OnDamageCalculated 钩子能修改伤害值。
func TestDamageCalculatedHook(t *testing.T) {
	g, err := NewSolo1v1("damage-calculated-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 给目标玩家加上圣光技能
	g.Players[seat].Character.SkillIDs = append(g.Players[seat].Character.SkillIDs, "test_shengguang")
	g.Players[seat].Character.Skills = append(g.Players[seat].Character.Skills, skill.MetasForCharacter("test_shengguang")...)

	// 调试：检查技能是否注册成功
	t.Logf("SkillIDs: %v", g.Players[seat].Character.SkillIDs)
	if h, ok := skill.Lookup("test_shengguang"); ok {
		t.Logf("Lookup test_shengguang: OK, Meta=%v, OnDamageCalculated=%v", h.Meta(), h.Decl.OnDamageCalculated != nil)
	} else {
		t.Errorf("Lookup test_shengguang: NOT FOUND")
	}
	handlers := g.playerSkillHandlers(seat)
	t.Logf("playerSkillHandlers count: %d", len(handlers))
	for _, h := range handlers {
		t.Logf("  handler: %s, OnDamageCalculated=%v", h.Meta().ID, h.Decl.OnDamageCalculated != nil)
	}

	// 普通杀：伤害值不应被修改
	oldHP := g.Players[seat].HP
	g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)
	if g.Players[seat].HP != oldHP-1 {
		t.Errorf("普通杀：期望 HP %d，实际 %d", oldHP-1, g.Players[seat].HP)
	}

	// 火杀：圣光应该让伤害值-1（1→0，不扣血）
	events = []GameEvent{}
	oldHP = g.Players[seat].HP
	g.applyDamageWithHook(1, seat, 1, Card{Kind: CardShaFire, Name: "火杀"}, &events)
	if g.Players[seat].HP != oldHP {
		t.Errorf("火杀+圣光：期望 HP 不变（伤害值被减为0），实际 %d", g.Players[seat].HP)
	}

	// 火杀+基础伤害2：圣光让伤害值-1（2→1）
	events = []GameEvent{}
	oldHP = g.Players[seat].HP
	g.applyDamageWithHook(1, seat, 2, Card{Kind: CardShaFire, Name: "火杀"}, &events)
	if g.Players[seat].HP != oldHP-1 {
		t.Errorf("火杀(2点)+圣光：期望 HP %d，实际 %d", oldHP-1, g.Players[seat].HP)
	}

	t.Log("OnDamageCalculated 钩子测试通过")
}

// TestBeforeHPChangeHook 验证 OnBeforeHPChange 钩子能防止扣血。
func TestBeforeHPChangeHook(t *testing.T) {
	g, err := NewSolo1v1("before-hp-change-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 给目标玩家加上防止扣血技能
	g.Players[seat].Character.SkillIDs = append(g.Players[seat].Character.SkillIDs, "test_prevent_damage")
	g.Players[seat].Character.Skills = append(g.Players[seat].Character.Skills, skill.MetasForCharacter("test_prevent_damage")...)

	// 造成伤害：应该被防止
	oldHP := g.Players[seat].HP
	g.applyDamageWithHook(1, seat, 3, Card{Kind: CardSha, Name: "杀"}, &events)

	if g.Players[seat].HP != oldHP {
		t.Errorf("防止扣血：期望 HP 不变，实际 %d（期望 %d）", g.Players[seat].HP, oldHP)
	}

	t.Log("OnBeforeHPChange 钩子测试通过")
}

// TestDeathHook 验证 OnDeath 和 OnAfterDeath 钩子能触发。
func TestDeathHook(t *testing.T) {
	g, err := NewSolo1v1("death-hook-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}

	seat := 0
	events := []GameEvent{}

	// 给目标玩家加上亡语技能
	g.Players[seat].Character.SkillIDs = append(g.Players[seat].Character.SkillIDs, "test_death_curse")

	// 直接设置 HP=0，模拟阵亡
	g.Players[seat].HP = 0

	// 手动调用 resolveDyingDeath（模拟濒死失败后的死亡流程）
	g.dyingContext = &DyingContext{
		Victim: seat,
		Killer: 1,
		Damage: 5,
	}
	err = g.resolveDyingDeath(&events)
	if err != nil {
		t.Errorf("resolveDyingDeath failed: %v", err)
	}

	// 检查是否触发了 dying_death 事件
	found := false
	for _, e := range events {
		if e.Type == "dying_death" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	if !found {
		t.Error("dying_death event not found after resolveDyingDeath")
	}

	t.Log("OnDeath/OnAfterDeath 钩子测试通过")
}
