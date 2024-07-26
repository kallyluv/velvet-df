package entity

import (
	"image/color"
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	transform
	yaw, pitch float64

	age int
	close bool

	owner world.Entity

	potion potion.Potion
	ent *entity.Ent

	t world.EntityType
	b *entity.ProjectileBehaviour
	c *entity.ProjectileBehaviourConfig
}

// Thank you Tal!

// NewSplashPotion ...
func NewSplashPotion(pos, vel mgl64.Vec3, yaw, pitch float64, t potion.Potion, owner world.Entity) *SplashPotion {
	conf := &entity.ProjectileBehaviourConfig{
			Gravity:           0.06,
			Drag:              0.01, //0.0025
			Hit: func(e *entity.Ent, result trace.Result) {
				aabb := cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125).Translate(result.Position())
				w := e.World()
				colour := color.RGBA{R: 0x38, G: 0x5d, B: 0xc6, A: 0xff}
				if effects := t.Effects(); len(effects) > 0 {
					colour, _ = effect.ResultingColour(effects)

					ignore := func(e world.Entity) bool {
						_, living := e.(entity.Living)
						_, splashpotion := e.(*SplashPotion)
						return !living || splashpotion
					}

					for _, e := range w.EntitiesWithin(aabb.GrowVec3(mgl64.Vec3{6, 4.5, 6}), ignore) {
						pos := e.Position()
						if !e.Type().BBox(e).Translate(pos).IntersectsWith(aabb.GrowVec3(mgl64.Vec3{3, 2.125, 3})) {
							continue
						}

						dist := pos.Sub(result.Position()).Len()
						if dist > 4 {
							continue
						}

						f := 1 - dist/4
						if entityResult, ok := result.(trace.EntityResult); ok && entityResult.Entity() == e {
							f = 1
						}
						distMultiplier := 0.59
						if e.(*player.Player).Name() != owner.(*player.Player).Name() {
							distMultiplier = math.Max(math.Min(1-dist/3.9, 0.6), 0.48)
						}

						splashed := e.(entity.Living)
						for _, eff := range effects {
							if eff.Type() == (effect.InstantHealth{}) && eff.Level() == 2 {
								splashed.Heal(float64(int(4)<<(eff.Level()-1))*distMultiplier*1.75, effect.InstantHealingSource{})
								continue
							}
							if p, ok := eff.Type().(effect.PotentType); ok {
								splashed.AddEffect(effect.NewInstant(p.WithPotency(f), eff.Level()))
								continue
							}

							dur := time.Duration(float64(eff.Duration()) * 0.75 * f)
							if dur < time.Second {
								continue
							}
							splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
						}
					}
				} else if t == potion.Water() {
					if blockResult, ok := result.(*trace.BlockResult); ok {
						pos := blockResult.BlockPosition().Side(blockResult.Face())
						if _, ok := w.Block(pos).(block.Fire); ok {
							w.SetBlock(pos, block.Air{}, nil)
						}

						for _, f := range cube.HorizontalFaces() {
							h := pos.Side(f)
							if _, ok := w.Block(h).(block.Fire); ok {
								w.SetBlock(h, block.Air{}, nil)
							}
						}
					}
				}

				w.AddParticle(result.Position(), particle.Splash{Colour: colour})
				w.PlaySound(result.Position(), sound.GlassBreak{})
			},
		}
	b := conf.New(owner)
	entconf := &entity.Config{
		Behaviour: b,
	}
	ent := entconf.New(entity.SplashPotionType{}, pos)
	s := &SplashPotion{
		yaw:   yaw,
		pitch: pitch,
		owner: owner,

		potion: t,
		ent: ent,

		t: entity.SplashPotionType{},
	}
	s.transform = newTransform(s, pos)
	s.vel = vel
	return s
}

// Name ...
func (s *SplashPotion) Name() string {
	return "Splash Potion"
}

// EncodeEntity ...
func (s *SplashPotion) EncodeEntity() string {
	return "minecraft:splash_potion"
}

// BBox ...
func (s *SplashPotion) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Rotation ...
func (s *SplashPotion) Rotation() cube.Rotation {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cube.Rotation{s.yaw, s.pitch}
}

// Type returns the type of potion the splash potion will grant effects for when thrown.
func (s *SplashPotion) Type() world.EntityType {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.t
}

// Tick ...
func (s *SplashPotion) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m := s.ent.Behaviour().Tick(s.ent)
	s.yaw, s.pitch = m.Rotation().Yaw(), m.Rotation().Pitch()
	s.pos, s.vel = m.Position(), m.Velocity()
	s.mu.Unlock()

	s.age++
	m.Send()

	if m.Position()[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(e world.Entity) bool {
	_, ok := e.(entity.Living)
	return !ok || e == s || (s.age < 5 && e == s.owner)
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// Own ...
func (s *SplashPotion) Own(owner world.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}
