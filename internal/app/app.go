package app

import (
	"net/http"

	"room-booking-service/internal/auth"
	"room-booking-service/internal/conference"
	"room-booking-service/internal/config"
	"room-booking-service/internal/db"
	"room-booking-service/internal/httpapi"
	pgrepo "room-booking-service/internal/repository/postgres"
	"room-booking-service/internal/service"
)

func BuildHandler(cfg config.Config, userRepo service.UserRepository, roomRepo service.RoomRepository, schedRepo service.ScheduleRepository, slotRepo service.SlotRepository, bookingRepo service.BookingRepository) http.Handler {
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo)
	roomSvc := service.NewRoomService(roomRepo)
	schSvc := service.NewScheduleService(roomRepo, schedRepo, slotRepo, cfg.SlotWindowDays)
	slotSvc := service.NewSlotService(roomRepo, slotRepo, schSvc)
	bookingSvc := service.NewBookingService(slotRepo, bookingRepo, conference.MockClient{BaseURL: cfg.ConferenceBaseURL})
	return httpapi.NewHandler(jwtManager, authSvc, roomSvc, schSvc, slotSvc, bookingSvc).Routes()
}

func BuildPostgresHandler(cfg config.Config) (http.Handler, error) {
	pool, err := db.NewPool(cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(nil, pool); err != nil {
		return nil, err
	}
	return BuildHandler(cfg,
		pgrepo.NewUserRepository(pool),
		pgrepo.NewRoomRepository(pool),
		pgrepo.NewScheduleRepository(pool),
		pgrepo.NewSlotRepository(pool),
		pgrepo.NewBookingRepository(pool),
	), nil
}
