package engine

// ============================================================================
// 伤害事件（参考 noname 04-damage-system.md）
// 伤害是完整的 GameEvent：damageBegin → damage(扣血) → dying(濒死) → damageEnd
// ============================================================================

// DamageEvent 伤害事件参数。
// 参考 noname: player.damage(num, nature, source)
type DamageEvent struct {
	Source int    // 伤害来源座位号
	Target int    // 受伤者座位号
	Amount int    // 伤害值
	Nature string // 伤害属性（fire/thunder/poison/ice/normal）
	Card   Card   // 造成伤害的牌
	// 控制标记
	NoDying     bool // 跳过濒死检测（参考 noname nodying）
	NoTrigger   bool // 跳过技能触发（参考 noname notrigger）
	Unreal      bool // 视为伤害，不扣血（参考 noname unreal）
	IgnoreArmor bool // 青釭剑无视防具（参考 noname: qinggang2 / unequip2）
}

// ============================================================================
// 伤害 GameEvent 创建（参考 noname: player.damage() → game.createEvent("damage")）
// ============================================================================

// StartDamageEvent 创建并启动伤害事件。
// 参考 noname:
//
//	damage: function() {
//	    step 0-3: damageBegin1~4
//	    step 4: changeHp(-num) + trigger("damage")
//	    step 5: if hp<=0 → player.dying(event)
//	    step 6: trigger("damageSource")
//	}
func (g *Game) StartDamageEvent(params DamageEvent, events *[]GameEvent) {
	if g.IsFinished() || params.Target < 0 || params.Target >= len(g.Players) {
		return
	}
	if params.Amount <= 0 {
		params.Amount = 1
	}

	damageEv := g.NewPlayerEvent("damage", params.Target)
	damageEv.Source = params.Source
	damageEv.Card = &params.Card
	damageEv.Num = params.Amount
	damageEv.Nature = params.Nature
	damageEv.NoTrigger = params.NoTrigger

	// OnBefore: damageBegin（参考 noname: step 0-3, damageBegin1~4）
	// 藤甲等防具可在此取消伤害
	damageEv.OnBefore = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		if params.Unreal {
			// 视为伤害跳过前置检查（参考 noname: if (unreal) goto(4)）
			ev.FinishEvent()
			return nil
		}
		// TODO: trigger("damageBegin1") → trigger("damageBegin4")
		return nil
	}

	// Content: 核心扣血 + 自动濒死（参考 noname: step 4-5）
	damageEv.Content = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		if params.Unreal {
			// 视为伤害不扣血（参考 noname: changeHp 在 unreal 时跳过）
			return nil
		}

		target := params.Target
		p := &g.Players[target]
		oldHP := p.HP

		// 调整伤害值（藤甲加伤、白银狮子减伤等）
		actualDamage := g.adjustDamageAmount(params.Source, target, params.Amount, params.Card, false, params.IgnoreArmor)
		// 白银狮子：伤害值 > 1 时锁定为 1（青釭剑可穿透）
		if !params.IgnoreArmor {
			g.baiyinReduceDamage(target, &actualDamage)
		}
		if actualDamage <= 0 {
			// 伤害被减为 0（参考 noname: trigger("damageZero")）
			return nil
		}

		// 扣血（参考 noname: player.changeHp(-num)）
		g.applyDamage(params.Source, target, actualDamage, params.Card, evs)

		// 血量变化钩子
		if p.HP != oldHP {
			g.handleHPChange(HPChangeContext{
				Seat: target, OldHP: oldHP, NewHP: p.HP,
				Delta: p.HP - oldHP, Reason: "damage",
				Source: params.Source, Damage: actualDamage,
			}, evs)
		}

		// 自动濒死检查（参考 noname: step 5, if hp<=0 → player.dying(event)）
		// 濒死通过 afterDamageApplied 接入现有系统（后续迁移到 GameEvent 子事件）
		if !params.NoDying && p.HP <= 0 {
			resume := DamageResume{}
			g.afterDamageApplied(params.Source, target, actualDamage, params.Card, resume, evs)
		}

		return nil
	}

	// OnEnd: damageEnd（参考 noname: 刚烈、反馈等在此触发）
	damageEv.OnEnd = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		// TODO: trigger("damageEnd") → 刚烈、反馈等卖血技
		return nil
	}

	// OnAfter: damageSource（参考 noname: step 6）
	damageEv.OnAfter = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		// TODO: trigger("damageSource")
		return nil
	}

	g.StartEvent(damageEv, events)
}

// ============================================================================
// 保持向后兼容：ApplyDamageAndCheckDeath 改为调用 StartDamageEvent
// ============================================================================

// applyDamageAndCheckDeathImpl 是 ApplyDamageAndCheckDeath 的新实现。
// 在 phase_hp_change.go 中的 ApplyDamageAndCheckDeath 方法将调用此函数。
func (g *Game) applyDamageAndCheckDeathImpl(source, target, amount int, damageCard Card, resume DamageResume, events *[]GameEvent) bool {
	if amount <= 0 || target < 0 || target >= len(g.Players) {
		return false
	}

	g.StartDamageEvent(DamageEvent{
		Source: source,
		Target: target,
		Amount: amount,
		Card:   damageCard,
		Nature: damageCard.DamageType,
	}, events)

	return g.Players[target].HP <= 0 && !g.IsFinished()
}
