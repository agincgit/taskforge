package main

import (
	"log"

	"github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/config"
)

func main() {
	cfg := config.GetConfig("config.json")
	srv, err := taskforge.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	log.Fatal(srv.Start())
}
