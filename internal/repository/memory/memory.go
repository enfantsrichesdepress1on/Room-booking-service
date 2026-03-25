package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"room-booking-service/internal/models"
	"room-booking-service/internal/service"
)

type Store struct {
	mu        sync.Mutex
	Users     map[string]models.User
	Rooms     map[string]models.Room
	Schedules map[string]models.Schedule
	Slots     map[string]models.Slot
	Bookings  map[string]models.Booking
}

func NewStore() *Store {
	return &Store{
		Users: map[string]models.User{}, Rooms: map[string]models.Room{}, Schedules: map[string]models.Schedule{}, Slots: map[string]models.Slot{}, Bookings: map[string]models.Booking{},
	}
}

func (s *Store) Create(ctx context.Context, user models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	user.CreatedAt = &now
	s.Users[user.ID] = user
	return user, nil
}
func (s *Store) GetByEmail(ctx context.Context, email string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return models.User{}, service.ErrNotFound
}
func (s *Store) GetByID(ctx context.Context, id string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.Users[id]
	if !ok {
		return models.User{}, service.ErrNotFound
	}
	return u, nil
}
func (s *Store) Upsert(ctx context.Context, user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Users[user.ID] = user
	return nil
}

func (s *Store) CreateRoom(ctx context.Context, room models.Room) (models.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	room.CreatedAt = &now
	s.Rooms[room.ID] = room
	return room, nil
}
func (s *Store) List(ctx context.Context) ([]models.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]models.Room, 0, len(s.Rooms))
	for _, r := range s.Rooms {
		items = append(items, r)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}
func (s *Store) Exists(ctx context.Context, roomID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Rooms[roomID]
	return ok, nil
}

func (s *Store) CreateSchedule(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Schedules[schedule.RoomID]; ok {
		return models.Schedule{}, service.ErrScheduleExists
	}
	s.Schedules[schedule.RoomID] = schedule
	return schedule, nil
}
func (s *Store) GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sched, ok := s.Schedules[roomID]
	if !ok {
		return models.Schedule{}, service.ErrNotFound
	}
	return sched, nil
}

func (s *Store) CreateBatch(ctx context.Context, slots []models.Slot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, slot := range slots {
		s.Slots[slot.ID] = slot
	}
	return nil
}
func (s *Store) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]models.Slot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	items := make([]models.Slot, 0)
	for _, slot := range s.Slots {
		if slot.RoomID != roomID || slot.Start.Before(start) || !slot.Start.Before(end) {
			continue
		}
		busy := false
		for _, b := range s.Bookings {
			if b.SlotID == slot.ID && b.Status == models.BookingStatusActive {
				busy = true
				break
			}
		}
		if !busy {
			items = append(items, slot)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	return items, nil
}
func (s *Store) ExistsForRange(ctx context.Context, roomID string, start, end time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, slot := range s.Slots {
		if slot.RoomID == roomID && !slot.Start.Before(start) && slot.Start.Before(end) {
			return true, nil
		}
	}
	return false, nil
}
func (s *Store) GetByIDSlot(ctx context.Context, slotID string) (models.Slot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	slot, ok := s.Slots[slotID]
	if !ok {
		return models.Slot{}, service.ErrSlotNotFound
	}
	return slot, nil
}

func (s *Store) CreateBooking(ctx context.Context, booking models.Booking) (models.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.Bookings {
		if b.SlotID == booking.SlotID && b.Status == models.BookingStatusActive {
			return models.Booking{}, service.ErrSlotAlreadyBooked
		}
	}
	now := time.Now().UTC()
	booking.CreatedAt = &now
	s.Bookings[booking.ID] = booking
	return booking, nil
}
func (s *Store) ListAllBookings(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]models.Booking, 0, len(s.Bookings))
	for _, b := range s.Bookings {
		items = append(items, b)
	}
	total := len(items)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	start := (page - 1) * pageSize
	if start > len(items) {
		return []models.Booking{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}
func (s *Store) ListByUserFuture(ctx context.Context, userID string, now time.Time) ([]models.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]models.Booking, 0)
	for _, b := range s.Bookings {
		slot := s.Slots[b.SlotID]
		if b.UserID == userID && !slot.Start.Before(now) {
			items = append(items, b)
		}
	}
	return items, nil
}
func (s *Store) GetByIDBooking(ctx context.Context, bookingID string) (models.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.Bookings[bookingID]
	if !ok {
		return models.Booking{}, service.ErrBookingNotFound
	}
	return b, nil
}
func (s *Store) UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) (models.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.Bookings[bookingID]
	if !ok {
		return models.Booking{}, service.ErrBookingNotFound
	}
	b.Status = status
	s.Bookings[bookingID] = b
	return b, nil
}

// adapters

type UserRepo struct{ S *Store }

func (r UserRepo) Create(ctx context.Context, user models.User) (models.User, error) {
	return r.S.Create(ctx, user)
}
func (r UserRepo) GetByEmail(ctx context.Context, email string) (models.User, error) {
	return r.S.GetByEmail(ctx, email)
}
func (r UserRepo) GetByID(ctx context.Context, id string) (models.User, error) {
	return r.S.GetByID(ctx, id)
}
func (r UserRepo) Upsert(ctx context.Context, user models.User) error { return r.S.Upsert(ctx, user) }

type RoomRepo struct{ S *Store }

func (r RoomRepo) Create(ctx context.Context, room models.Room) (models.Room, error) {
	return r.S.CreateRoom(ctx, room)
}
func (r RoomRepo) List(ctx context.Context) ([]models.Room, error) { return r.S.List(ctx) }
func (r RoomRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	return r.S.Exists(ctx, roomID)
}

type ScheduleRepo struct{ S *Store }

func (r ScheduleRepo) Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	return r.S.CreateSchedule(ctx, schedule)
}
func (r ScheduleRepo) GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error) {
	return r.S.GetByRoomID(ctx, roomID)
}

type SlotRepo struct{ S *Store }

func (r SlotRepo) CreateBatch(ctx context.Context, slots []models.Slot) error {
	return r.S.CreateBatch(ctx, slots)
}
func (r SlotRepo) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]models.Slot, error) {
	return r.S.ListAvailableByRoomAndDate(ctx, roomID, date)
}
func (r SlotRepo) ExistsForRange(ctx context.Context, roomID string, start, end time.Time) (bool, error) {
	return r.S.ExistsForRange(ctx, roomID, start, end)
}
func (r SlotRepo) GetByID(ctx context.Context, slotID string) (models.Slot, error) {
	return r.S.GetByIDSlot(ctx, slotID)
}

type BookingRepo struct{ S *Store }

func (r BookingRepo) Create(ctx context.Context, booking models.Booking) (models.Booking, error) {
	return r.S.CreateBooking(ctx, booking)
}
func (r BookingRepo) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	return r.S.ListAllBookings(ctx, page, pageSize)
}
func (r BookingRepo) ListByUserFuture(ctx context.Context, userID string, now time.Time) ([]models.Booking, error) {
	return r.S.ListByUserFuture(ctx, userID, now)
}
func (r BookingRepo) GetByID(ctx context.Context, bookingID string) (models.Booking, error) {
	return r.S.GetByIDBooking(ctx, bookingID)
}
func (r BookingRepo) UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) (models.Booking, error) {
	return r.S.UpdateStatus(ctx, bookingID, status)
}
