package service

import (
	"context"
	"time"

	"room-booking-service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByID(ctx context.Context, id string) (models.User, error)
	Upsert(ctx context.Context, user models.User) error
}

type RoomRepository interface {
	Create(ctx context.Context, room models.Room) (models.Room, error)
	List(ctx context.Context) ([]models.Room, error)
	Exists(ctx context.Context, roomID string) (bool, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error)
	GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error)
}

type SlotRepository interface {
	CreateBatch(ctx context.Context, slots []models.Slot) error
	ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]models.Slot, error)
	ExistsForRange(ctx context.Context, roomID string, start, end time.Time) (bool, error)
	GetByID(ctx context.Context, slotID string) (models.Slot, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking models.Booking) (models.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error)
	ListByUserFuture(ctx context.Context, userID string, now time.Time) ([]models.Booking, error)
	GetByID(ctx context.Context, bookingID string) (models.Booking, error)
	UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) (models.Booking, error)
}

type ConferenceClient interface {
	CreateLink(ctx context.Context, slot models.Slot, userID string) (string, error)
}
