package main

import (
	"backupagent/config"
	"backupagent/internal/server"
	"flag"
	"log"
)

func main() {

	flag.Parse()
	
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

    srv,err := server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	
    log.Fatal(srv.Run())
}

