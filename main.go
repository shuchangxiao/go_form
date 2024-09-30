package main

import (
	"veripTest/config"
	"veripTest/router"
)

func main() {
	config.Init()
	engine := router.SetRouter()
	if config.Cf.App.Port == "" {
		config.Cf.App.Port = ":8080"
	}
	engine.Run(config.Cf.App.Port)
}
