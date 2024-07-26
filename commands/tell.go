package commands

import (
	"velvet/console"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
)

type Tell struct {
	Target  []cmd.Target `cmd:"player"`
	Message cmd.Varargs  `cmd:"message"`
}

func (t Tell) Run(source cmd.Source, output *cmd.Output) {
	if len(t.Target) > 1 {
		output.Print("§cYou can only message one player at a time.")
		return
	}
	p, ok := t.Target[0].(*player.Player)
	if !ok {
		output.Printf(PlayerNotFound)
		return
	}
	name := "Server"
	if _, ok := source.(*console.CommandSender); ok {
		name = source.(*console.CommandSender).Name()
	}
	if _, ok := source.(*player.Player); ok {
		name = source.(*player.Player).Name()
	}
	p.Messagef("§7[§d%v §7-> §dYou§7]: §e%v", name, string(t.Message))
	output.Printf("§7[§dYou §7-> §d%v§7]: §e%v", p.Name(), string(t.Message))
}
