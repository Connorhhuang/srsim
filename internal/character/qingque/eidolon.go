package qingque

import (
	"github.com/simimpact/srsim/pkg/engine/event"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/engine/modifier"
	"github.com/simimpact/srsim/pkg/engine/prop"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

const (
	E1      key.Modifier = "qingque-e1"
	Autarky key.Modifier = "qingque-e4"
)

func init() {
	modifier.Register(E1, modifier.Config{
		Listeners: modifier.Listeners{
			OnBeforeHit: func(mod *modifier.ModifierInstance, e event.HitStartEvent) {
				if e.Hit.AttackType != model.AttackType_ULT {
					return
				}
				e.Hit.Attacker.AddProperty(prop.AllDamagePercent, 0.1)
			},
		},
	})
}
func (c *char) e2() {
	if c.info.Eidolon >= 2 {
		c.engine.ModifyEnergy(c.id, 1)
	}
}
func (c *char) initEidolons() {
	if c.info.Eidolon >= 1 {
		c.engine.AddModifier(c.id, info.Modifier{
			Name:   E1,
			Source: c.id,
		})
	}
}
