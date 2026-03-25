package service

import (
	"context"
	"time"

	"room-booking-service/internal/models"

	"github.com/google/uuid"
)

type BookingService struct {
	slots      SlotRepository
	bookings   BookingRepository
	conference ConferenceClient
	clock      Clock
}

func NewBookingService(slots SlotRepository, bookings BookingRepository, conference ConferenceClient) *BookingService {
	return &BookingService{slots: slots, bookings: bookings, conference: conference, clock: realClock{}}
}

func (s *BookingService) WithClock(clock Clock) *BookingService { s.clock = clock; return s }

func (s *BookingService) Create(ctx context.Context, slotID, userID string, createConferenceLink bool) (models.Booking, error) {
	slot, err := s.slots.GetByID(ctx, slotID)
	if err != nil {
		return models.Booking{}, err
	}
	if slot.Start.Before(s.clock.Now()) {
		return models.Booking{}, ErrInvalidRequest
	}
	var link *string
	if createConferenceLink {
		generated, err := s.conference.CreateLink(ctx, slot, userID)
		if err != nil {
			return models.Booking{}, err
		}
		link = &generated
	}
	return s.bookings.Create(ctx, models.Booking{ID: uuid.NewString(), SlotID: slotID, UserID: userID, Status: models.BookingStatusActive, ConferenceLink: link})
}

func (s *BookingService) Cancel(ctx context.Context, bookingID, userID string) (models.Booking, error) {
	booking, err := s.bookings.GetByID(ctx, bookingID)
	if err != nil {
		return models.Booking{}, err
	}
	if booking.UserID != userID {
		return models.Booking{}, ErrForbidden
	}
	if booking.Status == models.BookingStatusCancelled {
		return booking, nil
	}
	return s.bookings.UpdateStatus(ctx, bookingID, models.BookingStatusCancelled)
}

func (s *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, models.Pagination, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return nil, models.Pagination{}, ErrInvalidRequest
	}
	items, total, err := s.bookings.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	return items, models.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *BookingService) ListMyFuture(ctx context.Context, userID string) ([]models.Booking, error) {
	return s.bookings.ListByUserFuture(ctx, userID, s.clock.Now())
}

// compile-time use for imports
var _ = time.Now
