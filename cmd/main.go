package main

import (
	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Env"
	"log"
)

func main() {

	config := config{
		port: Env.GetInt("PORT", 8080),
	}

	app := server{
		config: config,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
