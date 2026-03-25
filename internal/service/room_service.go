package service

import (
	"context"
	"strings"

	"room-booking-service/internal/models"

	"github.com/google/uuid"
)

type RoomService struct {
	rooms RoomRepository
}

func NewRoomService(rooms RoomRepository) *RoomService { return &RoomService{rooms: rooms} }

func (s *RoomService) Create(ctx context.Context, name string, description *string, capacity *int) (models.Room, error) {
	if strings.TrimSpace(name) == "" {
		return models.Room{}, ErrInvalidRequest
	}
	return s.rooms.Create(ctx, models.Room{ID: uuid.NewString(), Name: strings.TrimSpace(name), Description: description, Capacity: capacity})
}

func (s *RoomService) List(ctx context.Context) ([]models.Room, error) { return s.rooms.List(ctx) }
