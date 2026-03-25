package service_test

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/conference"
	"room-booking-service/internal/models"
	"room-booking-service/internal/repository/memory"
	"room-booking-service/internal/service"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestScheduleCreatesSlots(t *testing.T) {
	store := memory.NewStore()
	roomSvc := service.NewRoomService(memory.RoomRepo{S: store})
	room, _ := roomSvc.Create(context.Background(), "A", nil, nil)
	base := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)
	schSvc := service.NewScheduleService(memory.RoomRepo{S: store}, memory.ScheduleRepo{S: store}, memory.SlotRepo{S: store}, 7).WithClock(fixedClock{t: base})
	if _, err := schSvc.Create(context.Background(), room.ID, []int{3}, "10:00", "11:00"); err != nil {
		t.Fatal(err)
	}
	slots, err := memory.SlotRepo{S: store}.ListAvailableByRoomAndDate(context.Background(), room.ID, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
}

func TestScheduleRejectsInvalidDay(t *testing.T) {
	store := memory.NewStore()
	store.Rooms["room"] = models.Room{ID: "room", Name: "Room"}
	schSvc := service.NewScheduleService(memory.RoomRepo{S: store}, memory.ScheduleRepo{S: store}, memory.SlotRepo{S: store}, 7)
	if _, err := schSvc.Create(context.Background(), "room", []int{0}, "10:00", "11:00"); err != service.ErrInvalidRequest {
		t.Fatalf("expected invalid request, got %v", err)
	}
}

func TestBookingCreateRejectsPastSlot(t *testing.T) {
	store := memory.NewStore()
	store.Rooms["room"] = models.Room{ID: "room", Name: "Room"}
	past := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	store.Slots["slot1"] = models.Slot{ID: "slot1", RoomID: "room", Start: past, End: past.Add(30 * time.Minute)}
	svc := service.NewBookingService(memory.SlotRepo{S: store}, memory.BookingRepo{S: store}, conference.MockClient{BaseURL: "https://x"}).WithClock(fixedClock{t: past.Add(2 * time.Hour)})
	if _, err := svc.Create(context.Background(), "slot1", "user", false); err != service.ErrInvalidRequest {
		t.Fatalf("expected invalid request, got %v", err)
	}
}

func TestBookingCancelIdempotent(t *testing.T) {
	store := memory.NewStore()
	store.Bookings["b1"] = models.Booking{ID: "b1", SlotID: "slot1", UserID: "user", Status: models.BookingStatusActive}
	svc := service.NewBookingService(memory.SlotRepo{S: store}, memory.BookingRepo{S: store}, conference.MockClient{BaseURL: "https://x"})
	first, err := svc.Cancel(context.Background(), "b1", "user")
	if err != nil || first.Status != "cancelled" {
		t.Fatalf("unexpected: %v %+v", err, first)
	}
	second, err := svc.Cancel(context.Background(), "b1", "user")
	if err != nil || second.Status != "cancelled" {
		t.Fatalf("unexpected: %v %+v", err, second)
	}
}

func TestBookingCannotCancelAnotherUsersBooking(t *testing.T) {
	store := memory.NewStore()
	store.Bookings["b1"] = models.Booking{ID: "b1", SlotID: "slot1", UserID: "owner", Status: models.BookingStatusActive}
	svc := service.NewBookingService(memory.SlotRepo{S: store}, memory.BookingRepo{S: store}, conference.MockClient{BaseURL: "https://x"})
	if _, err := svc.Cancel(context.Background(), "b1", "other"); err != service.ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestListMyFutureReturnsOnlyFuture(t *testing.T) {
	store := memory.NewStore()
	now := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)
	store.Slots["future"] = models.Slot{ID: "future", RoomID: "room", Start: now.Add(time.Hour), End: now.Add(90 * time.Minute)}
	store.Slots["past"] = models.Slot{ID: "past", RoomID: "room", Start: now.Add(-time.Hour), End: now.Add(-30 * time.Minute)}
	store.Bookings["b1"] = models.Booking{ID: "b1", SlotID: "future", UserID: "user", Status: models.BookingStatusActive}
	store.Bookings["b2"] = models.Booking{ID: "b2", SlotID: "past", UserID: "user", Status: models.BookingStatusActive}
	svc := service.NewBookingService(memory.SlotRepo{S: store}, memory.BookingRepo{S: store}, conference.MockClient{BaseURL: "https://x"}).WithClock(fixedClock{t: now})
	items, err := svc.ListMyFuture(context.Background(), "user")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "b1" {
		t.Fatalf("unexpected bookings: %+v", items)
	}
}
