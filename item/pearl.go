package item

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Pearl is an edited usable for ender pearls.
type Pearl struct{}

var pearlConf = entity.ProjectileBehaviourConfig{
	Gravity: 0.085,
	Drag: 0.01,
	Particle: particle.EndermanTeleport{},
	Sound: sound.Teleport{},
	Hit: teleport,
}

// Use ...
func (v Pearl) Use(w *world.World, user item.User, ctx *item.UseContext) bool {
	e := entity.Config{Behaviour: pearlConf.New(user.(world.Entity))}.New(entity.EnderPearlType{}, entity.EyePosition(user.(world.Entity)))
	e.SetVelocity(user.Rotation().Vec3().Mul(1.5))
	w.AddEntity(e)

	w.PlaySound(user.Position(), sound.ItemThrow{})
	ctx.SubtractFromCount(1)
	return true
}

type teleporter interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
	entity.Living
}

func teleport(e *entity.Ent, target trace.Result) {
	if user, ok := e.Behaviour().(*entity.ProjectileBehaviour).Owner().(teleporter); ok {
		e.World().PlaySound(user.Position(), sound.Teleport{})
		user.Teleport(target.Position())
		user.Hurt(5, entity.FallDamageSource{})
	}
}