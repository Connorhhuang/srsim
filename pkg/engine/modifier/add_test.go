package modifier

import (
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/simimpact/srsim/pkg/engine/event"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/engine/prop"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/mock"
	"github.com/simimpact/srsim/pkg/model"
	"github.com/stretchr/testify/assert"
)

func NewTestManagerForAdd(t *testing.T) (*Manager, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	engine := mock.NewMockEngineWithEvents(mockCtrl)
	engine.EXPECT().IsValid(gomock.Any()).Return(true).AnyTimes()

	manager := &Manager{
		engine:  engine,
		targets: make(map[key.TargetID]activeModifiers),
	}
	return manager, mockCtrl
}

func TestIgnoreResist(t *testing.T) {
	manager, mockCtrl := NewTestManager(t)
	defer mockCtrl.Finish()

	target := key.TargetID(1)
	mod := info.Modifier{
		Name: key.Modifier("Test"),
	}

	chance, resist, err := manager.attemptResist(target, mod, []model.BehaviorFlag{})
	assert.NoError(t, err)
	assert.Equal(t, -1.0, chance)
	assert.False(t, resist)
}

func TestResistModifier(t *testing.T) {
	manager, mockCtrl := NewTestManagerWithEvents(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)
	engine := manager.engine.(*mock.MockEngine)
	rand := rand.New(rand.NewSource(1))
	engine.EXPECT().Rand().Return(rand).AnyTimes()

	BChance := 0.05
	EHR := 0.01
	ERES := 0.3
	DRES := 0.5
	flags := []model.BehaviorFlag{model.BehaviorFlag_STAT_CTRL}

	target := key.TargetID(1)
	targetStats := mock.NewEmptyStats(target)
	engine.EXPECT().Stats(gomock.Eq(target)).Return(targetStats).Times(1)
	targetStats.AddProperty(prop.EffectRES, ERES)
	targetStats.AddDebuffRES(model.BehaviorFlag_STAT_CTRL, DRES)

	source := key.TargetID(2)
	sourceStats := mock.NewEmptyStats(source)
	sourceStats.AddProperty(prop.EffectHitRate, EHR)
	engine.EXPECT().Stats(gomock.Eq(source)).Return(sourceStats).Times(1)

	name := key.Modifier("TestResistModifier")
	mod := info.Modifier{
		Name:   name,
		Source: source,
		Chance: BChance,
	}

	expectedChance := BChance * (1 + EHR) * (1 - ERES) * (1 - DRES)

	engine.Events().ModifierResisted.Subscribe(func(event event.ModifierResistedEvent) {
		assert.Equal(t, target, event.Target)
		assert.Equal(t, source, event.Source)
		assert.Equal(t, name, event.Modifier)
		assert.Equal(t, expectedChance, event.Chance)
		assert.Equal(t, BChance, event.BaseChance)
		assert.Equal(t, EHR, event.EHR)
		assert.Equal(t, ERES, event.EffectRES)
		assert.Equal(t, DRES, event.DebuffRES)
	})

	chance, resist, err := manager.attemptResist(target, mod, flags)
	assert.NoError(t, err)
	assert.Equal(t, expectedChance, chance)
	assert.True(t, resist)
}

func TestFailedResist(t *testing.T) {
	manager, mockCtrl := NewTestManagerWithEvents(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)
	engine := manager.engine.(*mock.MockEngine)
	rand := rand.New(rand.NewSource(1))
	engine.EXPECT().Rand().Return(rand).AnyTimes()

	BChance := 2.0
	EHR := 0.01
	ERES := 0.3
	DRES := 0.5
	flags := []model.BehaviorFlag{model.BehaviorFlag_STAT_CTRL}

	target := key.TargetID(1)
	targetStats := mock.NewEmptyStats(target)
	engine.EXPECT().Stats(gomock.Eq(target)).Return(targetStats).Times(1)
	targetStats.AddProperty(prop.EffectRES, ERES)
	targetStats.AddDebuffRES(model.BehaviorFlag_STAT_CTRL, DRES)

	source := key.TargetID(2)
	sourceStats := mock.NewEmptyStats(source)
	sourceStats.AddProperty(prop.EffectHitRate, EHR)
	engine.EXPECT().Stats(gomock.Eq(source)).Return(sourceStats).Times(1)

	name := key.Modifier("TestResistModifier")
	mod := info.Modifier{
		Name:   name,
		Source: source,
		Chance: BChance,
	}

	expectedChance := BChance * (1 + EHR) * (1 - ERES) * (1 - DRES)

	engine.Events().ModifierResisted.Subscribe(func(event event.ModifierResistedEvent) {
		assert.Fail(t, "Event should never be emitted (modifier should not be resisted)")
	})

	chance, resist, err := manager.attemptResist(target, mod, flags)
	assert.NoError(t, err)
	assert.Equal(t, expectedChance, chance)
	assert.False(t, resist)
}

func TestAddInvalidTarget(t *testing.T) {
	manager, mockCtrl := NewTestManager(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)
	engine := manager.engine.(*mock.MockEngine)

	target := key.TargetID(1)
	engine.EXPECT().IsValid(target).Return(false).AnyTimes()

	mod := info.Modifier{
		Name: key.Modifier("TestAddInvalidTarget"),
	}

	added, err := manager.AddModifier(target, mod)
	assert.ErrorContains(t, err, "invalid target")
	assert.False(t, added)
}

func TestAddInvalidSource(t *testing.T) {
	manager, mockCtrl := NewTestManager(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)
	engine := manager.engine.(*mock.MockEngine)

	target := key.TargetID(1)
	firstCheck := engine.EXPECT().IsValid(target).Return(true)
	engine.EXPECT().IsValid(key.TargetID(0)).Return(false).After(firstCheck)

	mod := info.Modifier{
		Name: key.Modifier("TestAddInvalidSource"),
	}

	added, err := manager.AddModifier(target, mod)
	assert.ErrorContains(t, err, "invalid source")
	assert.False(t, added)
}

func TestAddUnregistered(t *testing.T) {
	manager, mockCtrl := NewTestManagerForAdd(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)

	type state struct {
		Mod float64
	}

	target := key.TargetID(1)
	name := key.Modifier("TestAddUnregistered")
	mod1 := info.Modifier{
		Name:   name,
		Source: target,
		State:  state{Mod: 5.0},
	}
	mod2 := info.Modifier{
		Name:   name,
		Source: target,
		State:  state{Mod: 1.0},
	}

	called := 0
	manager.engine.Events().ModifierAdded.Subscribe(func(event event.ModifierAddedEvent) {
		state := event.Modifier.State.(state)

		if called == 0 {
			assert.Equal(t, 5.0, state.Mod)
		} else {
			assert.Fail(t, "ModifierAddedEvent expected to only emit once")
		}
		called += 1
	})

	added, err := manager.AddModifier(target, mod1)
	assert.NoError(t, err)
	assert.True(t, added)
	added, err = manager.AddModifier(target, mod2)
	assert.NoError(t, err)
	assert.True(t, added)
	assert.Len(t, manager.targets[target], 1)
	assert.Equal(t, 1, called)
}

func TestAddMultiple(t *testing.T) {
	manager, mockCtrl := NewTestManagerForAdd(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)

	type state struct {
		Mod float64
	}

	target := key.TargetID(1)
	name := key.Modifier("TestAddMultiple")
	mod1 := info.Modifier{
		Name:   name,
		Source: target,
		State:  state{Mod: 5.0},
	}
	mod2 := info.Modifier{
		Name:   name,
		Source: target,
		State:  state{Mod: 1.0},
	}

	Register(name, Config{
		Stacking: Multiple,
	})

	called := 0
	manager.engine.Events().ModifierAdded.Subscribe(func(event event.ModifierAddedEvent) {
		state := event.Modifier.State.(state)

		if called == 0 {
			assert.Equal(t, 5.0, state.Mod)
		} else {
			assert.Equal(t, 1.0, state.Mod)
		}
		called += 1
	})

	added, err := manager.AddModifier(target, mod1)
	assert.NoError(t, err)
	assert.True(t, added)
	added, err = manager.AddModifier(target, mod2)
	assert.NoError(t, err)
	assert.True(t, added)
	assert.Len(t, manager.targets[target], 2)
	assert.Equal(t, 2, called)
}

func TestAddRefresh(t *testing.T) {
	manager, mockCtrl := NewTestManagerForAdd(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)

	target := key.TargetID(1)
	name := key.Modifier("TestAddRefresh")
	mod1 := info.Modifier{
		Name:     name,
		Duration: 3,
	}
	mod2 := info.Modifier{
		Name:     name,
		Duration: 5,
	}

	Register(name, Config{
		Stacking: Refresh,
	})

	addedCalls := 0
	manager.engine.Events().ModifierAdded.Subscribe(func(event event.ModifierAddedEvent) {
		if addedCalls == 0 {
			assert.Equal(t, 3, event.Modifier.Duration)
		}
		addedCalls += 1
	})

	extendedCalls := 0
	manager.engine.Events().ModifierExtendedDuration.Subscribe(func(event event.ModifierExtendedDurationEvent) {
		if extendedCalls == 0 {
			assert.Equal(t, 3, event.OldValue)
			assert.Equal(t, 5, event.NewValue)
		}
		extendedCalls += 1
	})

	added, err := manager.AddModifier(target, mod1)
	assert.NoError(t, err)
	assert.True(t, added)
	added, err = manager.AddModifier(target, mod2)
	assert.NoError(t, err)
	assert.True(t, added)
	assert.Len(t, manager.targets[target], 1)
	assert.Equal(t, 1, addedCalls)
	assert.Equal(t, 1, extendedCalls)
	assert.Equal(t, 5, manager.targets[target][0].duration)
}

func TestAddProlong(t *testing.T) {
	manager, mockCtrl := NewTestManagerForAdd(t)
	defer mockCtrl.Finish()
	defer FailOnPanic(t)

	target := key.TargetID(1)
	name := key.Modifier("TestAddProlong")
	mod1 := info.Modifier{
		Name:     name,
		Duration: 3,
	}
	mod2 := info.Modifier{
		Name:     name,
		Duration: 5,
	}

	Register(name, Config{
		Stacking: Prolong,
	})

	addedCalls := 0
	manager.engine.Events().ModifierAdded.Subscribe(func(event event.ModifierAddedEvent) {
		if addedCalls == 0 {
			assert.Equal(t, 3, event.Modifier.Duration)
		}
		addedCalls += 1
	})

	extendedCalls := 0
	manager.engine.Events().ModifierExtendedDuration.Subscribe(func(event event.ModifierExtendedDurationEvent) {
		if extendedCalls == 0 {
			assert.Equal(t, 3, event.OldValue)
			assert.Equal(t, 8, event.NewValue)
		}
		extendedCalls += 1
	})

	added, err := manager.AddModifier(target, mod1)
	assert.NoError(t, err)
	assert.True(t, added)
	added, err = manager.AddModifier(target, mod2)
	assert.NoError(t, err)
	assert.True(t, added)
	assert.Len(t, manager.targets[target], 1)
	assert.Equal(t, 1, addedCalls)
	assert.Equal(t, 1, extendedCalls)
	assert.Equal(t, 8, manager.targets[target][0].duration)
}
