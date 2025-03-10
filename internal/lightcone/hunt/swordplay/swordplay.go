package swordplay

import (
	"github.com/simimpact/srsim/pkg/engine"
	"github.com/simimpact/srsim/pkg/engine/equip/lightcone"
	"github.com/simimpact/srsim/pkg/engine/event"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/engine/modifier"
	"github.com/simimpact/srsim/pkg/engine/prop"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

const (
	Check  key.Modifier = "swordplay_check"
	Buff   key.Modifier = "swordplay_buff"
	Target key.Modifier = "swordplay_target"
)

// For each time the wearer hits the same target, DMG dealt increases by
// 8/10/12/14/16%, stacking up to 5 time(s). This effect will be dispelled when
// the wearer changes targets.
func init() {
	lightcone.Register(key.Swordplay, lightcone.Config{
		CreatePassive: Create,
		Rarity:        4,
		Path:          model.Path_HUNT,
		Promotions:    promotions,
	})

	modifier.Register(Check, modifier.Config{
		Listeners: modifier.Listeners{
			OnBeforeHit: onBeforeHit,
			OnAfterHit:  onAfterHit,
		},
	})

	modifier.Register(Buff, modifier.Config{
		StatusType:        model.StatusType_STATUS_BUFF,
		Stacking:          modifier.ReplaceBySource,
		MaxCount:          5,
		CountAddWhenStack: 1,
		Listeners: modifier.Listeners{
			OnAdd: buffOnAdd,
		},
	})

	modifier.Register(Target, modifier.Config{
		Stacking: modifier.ReplaceBySource,
	})
}

func Create(engine engine.Engine, owner key.TargetID, lc info.LightCone) {
	engine.AddModifier(owner, info.Modifier{
		Name:   Check,
		Source: owner,
		State:  0.06 + 0.02*float64(lc.Imposition),
	})
}

func onBeforeHit(mod *modifier.ModifierInstance, e event.HitStartEvent) {
	if !hasModifierFromSource(mod.Engine(), e.Defender, mod.Owner(), Target) {
		if mod.Engine().HasModifier(mod.Owner(), Buff) {
			stacks := mod.Engine().GetModifiers(mod.Owner(), Buff)[0].Count
			e.Hit.Attacker.AddProperty(prop.AllDamagePercent, -mod.State().(float64)*stacks)
			mod.Engine().RemoveModifier(mod.Owner(), Buff)
		}

		for _, enemy := range mod.Engine().Enemies() {
			mod.Engine().RemoveModifierFromSource(enemy, mod.Owner(), Target)
		}

		mod.Engine().AddModifier(e.Defender, info.Modifier{
			Name:   Target,
			Source: mod.Owner(),
		})
	}
}

func hasModifierFromSource(engine engine.Engine, target, source key.TargetID, key key.Modifier) bool {
	for _, mod := range engine.GetModifiers(target, key) {
		if mod.Source == source {
			return true
		}
	}
	return false
}

func onAfterHit(mod *modifier.ModifierInstance, e event.HitEndEvent) {
	if mod.Engine().HasModifier(mod.Owner(), Buff) {
		stacks := mod.Engine().GetModifiers(mod.Owner(), Buff)[0].Count
		if stacks == 5 {
			return
		}
	}

	mod.Engine().AddModifier(mod.Owner(), info.Modifier{
		Name:   Buff,
		Source: mod.Owner(),
		State:  mod.State(),
	})
}

func buffOnAdd(mod *modifier.ModifierInstance) {
	amt := mod.State().(float64)
	mod.AddProperty(prop.AllDamagePercent, amt*mod.Count())
}
