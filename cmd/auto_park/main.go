package main

import (
	autopark "auto_park"
	_ "auto_park/docs"
	"auto_park/internal/config"
	"auto_park/pkg/postgres"
	"log"
)

// @title Auto Park API
// @version 1.0
// @description Modular monolith API for auto park system
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load error: ", err)
	}

	db, err := postgres.Connect(cfg.DB)
	if err != nil {
		log.Fatal("db connect error: ", err)
	}
	defer postgres.Close()

	r := autopark.NewRouter(cfg, db)
	addr := ":" + cfg.Service.Port
	log.Printf("%s started on %s", cfg.Service.Name, addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
