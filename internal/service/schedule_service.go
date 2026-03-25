package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"room-booking-service/internal/models"

	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type ScheduleService struct {
	rooms      RoomRepository
	schedules  ScheduleRepository
	slots      SlotRepository
	windowDays int
	clock      Clock
}

func NewScheduleService(rooms RoomRepository, schedules ScheduleRepository, slots SlotRepository, windowDays int) *ScheduleService {
	return &ScheduleService{rooms: rooms, schedules: schedules, slots: slots, windowDays: windowDays, clock: realClock{}}
}

func (s *ScheduleService) WithClock(clock Clock) *ScheduleService { s.clock = clock; return s }

func (s *ScheduleService) Create(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (models.Schedule, error) {
	exists, err := s.rooms.Exists(ctx, roomID)
	if err != nil {
		return models.Schedule{}, err
	}
	if !exists {
		return models.Schedule{}, ErrRoomNotFound
	}
	if err := validateSchedule(daysOfWeek, startTime, endTime); err != nil {
		return models.Schedule{}, err
	}
	days := append([]int(nil), daysOfWeek...)
	sort.Ints(days)
	schedule, err := s.schedules.Create(ctx, models.Schedule{ID: uuid.NewString(), RoomID: roomID, DaysOfWeek: days, StartTime: startTime, EndTime: endTime})
	if err != nil {
		return models.Schedule{}, err
	}
	if err := s.ensureSlots(ctx, schedule, s.clock.Now(), s.clock.Now().AddDate(0, 0, s.windowDays)); err != nil {
		return models.Schedule{}, err
	}
	return schedule, nil
}

func (s *ScheduleService) EnsureSlotsForDate(ctx context.Context, roomID string, date time.Time) error {
	schedule, err := s.schedules.GetByRoomID(ctx, roomID)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	exists, err := s.slots.ExistsForRange(ctx, roomID, start, end)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.ensureSlots(ctx, schedule, start, end)
}

func validateSchedule(daysOfWeek []int, startTime, endTime string) error {
	if len(daysOfWeek) == 0 {
		return ErrInvalidRequest
	}
	seen := map[int]bool{}
	for _, day := range daysOfWeek {
		if day < 1 || day > 7 || seen[day] {
			return ErrInvalidRequest
		}
		seen[day] = true
	}
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return ErrInvalidRequest
	}
	end, err := time.Parse("15:04", endTime)
	if err != nil || !end.After(start) {
		return ErrInvalidRequest
	}
	if int(end.Sub(start).Minutes()) < 30 {
		return ErrInvalidRequest
	}
	if int(end.Sub(start).Minutes())%30 != 0 {
		return ErrInvalidRequest
	}
	return nil
}

func (s *ScheduleService) ensureSlots(ctx context.Context, schedule models.Schedule, from, to time.Time) error {
	startOfDay := time.Date(from.UTC().Year(), from.UTC().Month(), from.UTC().Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(to.UTC().Year(), to.UTC().Month(), to.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if to.After(endOfDay) {
		endOfDay = endOfDay.Add(24 * time.Hour)
	}
	slots := make([]models.Slot, 0)
	for day := startOfDay; day.Before(endOfDay); day = day.Add(24 * time.Hour) {
		weekday := isoWeekday(day)
		if !containsInt(schedule.DaysOfWeek, weekday) {
			continue
		}
		partsStart := strings.Split(schedule.StartTime, ":")
		partsEnd := strings.Split(schedule.EndTime, ":")
		startHour, startMin := mustAtoi(partsStart[0]), mustAtoi(partsStart[1])
		endHour, endMin := mustAtoi(partsEnd[0]), mustAtoi(partsEnd[1])
		slotStart := time.Date(day.Year(), day.Month(), day.Day(), startHour, startMin, 0, 0, time.UTC)
		rangeEnd := time.Date(day.Year(), day.Month(), day.Day(), endHour, endMin, 0, 0, time.UTC)
		for slotStart.Before(rangeEnd) {
			slotEnd := slotStart.Add(30 * time.Minute)
			if slotEnd.After(rangeEnd) {
				break
			}
			slots = append(slots, models.Slot{ID: deterministicSlotID(schedule.RoomID, slotStart), RoomID: schedule.RoomID, Start: slotStart, End: slotEnd})
			slotStart = slotEnd
		}
	}
	if len(slots) == 0 {
		return nil
	}
	return s.slots.CreateBatch(ctx, slots)
}

func deterministicSlotID(roomID string, start time.Time) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("%s|%s", roomID, start.UTC().Format(time.RFC3339)))).String()
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func containsInt(items []int, target int) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
func mustAtoi(s string) int { var v int; fmt.Sscanf(s, "%d", &v); return v }
