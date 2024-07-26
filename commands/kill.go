package commands

import (
	"velvet/utils"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
)

type Kill struct {
	Targets cmd.Optional[[]cmd.Target] `cmd:"victim"`
}

func (t Kill) Run(source cmd.Source, _ *cmd.Output) {
	p := source.(*player.Player)
	targets := t.Targets.LoadOr(nil)
	if len(targets) > 0 {
		if len(targets) > 1 {
			if p.XUID() != utils.Config.Staff.Owner.XUID {
				p.Message(NoPermission)
				return
			}
			p.Messagef("§cYou have killed §d%v §cpeople.", len(targets))
			return
		}
		if tg, ok := targets[0].(*player.Player); ok {
			tg.Hurt(tg.MaxHealth(), entity.VoidDamageSource{})
			p.Messagef("§cYou have killed %v.", tg.Name())
		}
		return
	}
	p.Hurt(p.MaxHealth(), entity.VoidDamageSource{})
	p.Message("§cYou have killed yourself.")
}

func (Kill) Allow(s cmd.Source) bool { return checkAdmin(s) }
