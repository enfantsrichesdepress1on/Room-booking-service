package main

import (
	"context"
	"log"

	"room-booking-service/internal/config"
	"room-booking-service/internal/db"
	pgrepo "room-booking-service/internal/repository/postgres"
	"room-booking-service/internal/service"
)

func main() {
	cfg := config.Load()
	pool, err := db.NewPool(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Migrate(context.Background(), pool); err != nil {
		log.Fatal(err)
	}
	roomRepo := pgrepo.NewRoomRepository(pool)
	roomSvc := service.NewRoomService(roomRepo)
	room, err := roomSvc.Create(context.Background(), "Main Room", strPtr("Seeded room"), intPtr(8))
	if err != nil {
		log.Fatal(err)
	}
	schSvc := service.NewScheduleService(roomRepo, pgrepo.NewScheduleRepository(pool), pgrepo.NewSlotRepository(pool), cfg.SlotWindowDays)
	if _, err := schSvc.Create(context.Background(), room.ID, []int{1, 2, 3, 4, 5}, "09:00", "18:00"); err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded room %s", room.ID)
}
func strPtr(s string) *string { return &s }
func intPtr(v int) *int       { return &v }
