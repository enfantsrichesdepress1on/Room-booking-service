package service

import (
	"context"
	"time"

	"room-booking-service/internal/models"
)

type SlotService struct {
	rooms     RoomRepository
	slots     SlotRepository
	schedules *ScheduleService
}

func NewSlotService(rooms RoomRepository, slots SlotRepository, schedules *ScheduleService) *SlotService {
	return &SlotService{rooms: rooms, slots: slots, schedules: schedules}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID string, date time.Time) ([]models.Slot, error) {
	exists, err := s.rooms.Exists(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRoomNotFound
	}
	if err := s.schedules.EnsureSlotsForDate(ctx, roomID, date.UTC()); err != nil {
		return nil, err
	}
	return s.slots.ListAvailableByRoomAndDate(ctx, roomID, date.UTC())
}
