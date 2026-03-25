package main

import (
	"log"
	"net/http"

	"room-booking-service/internal/app"
	"room-booking-service/internal/config"
)

func main() {
	cfg := config.Load()
	handler, err := app.BuildPostgresHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
