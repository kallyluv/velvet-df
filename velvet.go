package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/oomph-ac/oomph"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"time"
	"velvet/db"
	"velvet/handlers"
	"velvet/utils"
	"velvet/utils/worldmanager"
)

var logger = logrus.New()

func startServer() {
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.Level = logrus.InfoLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}
	config.WorldConfig = func(def world.Config) world.Config {
		//def.ReadOnly = true
		def.RandomTickSpeed = 0
		def.Generator = nil
		def.PortalDestination = nil
		return def
	}

	srv := server.New(&config.Config, logger)
	srv.SetName("Velvet")
	srv.Allow(allower{})
	srv.CloseOnProgramEnd()
	srv.World().SetSpawn(cube.Pos{263, 66, 257})
	for _, e := range srv.World().Entities() {
		srv.World().RemoveEntity(e)
		fmt.Println("Removed entity: " + e.Name())
	}
	if err := srv.Start(); err != nil {
		logger.Fatalln(err)
	}

	if p, err := resource.Compile("resources/pack.zip"); err == nil {
		srv.AddResourcePack(p) // todo: p.WithContentKey
	}

	utils.Srv = srv
	utils.WorldMG = worldmanager.New(srv, "worlds/", logger)
	utils.Started = time.Now().Unix()

	if files, err := ioutil.ReadDir("worlds"); err == nil {
		for _, f := range files {
			if f.Name() != "world" && f.Name() != "world.zip" {
				err = utils.WorldMG.LoadWorld(f.Name(), f.Name(), world.Overworld)
				if err != nil {
					fmt.Println("Error loading world " + f.Name() + ": " + err.Error())
				} else {
					fmt.Println("Loaded world: " + f.Name())
					if w, ok := utils.WorldMG.World(f.Name()); ok {
						switch f.Name() {
						case utils.Config.World.NoDebuff:
							w.StopRaining()
							w.StopWeatherCycle()
						}
					}
				}
			}
		}
	}

	w := srv.World()
	w.SetDefaultGameMode(world.GameModeSurvival)
	w.SetTickRange(0)
	w.SetTime(0)
	w.StopTime()

	// AntiCheat start
	if config.Oomph.Enabled {
		go func() {
			ac := oomph.New(logger, config.Oomph.Address)
			if err := ac.Listen(srv, config.Server.Name, config.Resources.Required); err != nil {
				panic(err)
			}
			for {
				p, err := ac.Accept()
				if err != nil {
					return
				}
				p.Handle(handlers.NewACHandler(p))
			}
		}()
	}
	// AntiCheat end

	for srv.Accept(handleJoin) {

	}
}

// handleJoin processes a players join.
func handleJoin(p *player.Player) {
	s, err := db.LoadSession(p)
	if err != nil {
		p.Disconnect(fmt.Sprintf("There was an error loading your account, create a ticket at %s\nError: %s", utils.Config.Discord.Invite, err.Error()))
	}
	p.Handle(handlers.NewPlayerHandler(p, s))
}
