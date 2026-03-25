package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"room-booking-service/internal/config"
	"room-booking-service/internal/db"
	"room-booking-service/internal/models"
	pgrepo "room-booking-service/internal/repository/postgres"
	"room-booking-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositories(t *testing.T) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("set RUN_DB_TESTS=1 to run postgres integration tests")
	}

	cfg := config.Load()
	pool, err := db.NewPool(cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	cleanupDatabase(t, pool)

	users := pgrepo.NewUserRepository(pool)
	rooms := pgrepo.NewRoomRepository(pool)
	schedules := pgrepo.NewScheduleRepository(pool)
	slots := pgrepo.NewSlotRepository(pool)
	bookings := pgrepo.NewBookingRepository(pool)

	authSvc := service.NewAuthService(users)
	admin, err := authSvc.EnsureDummyUser(context.Background(), models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	user, err := authSvc.EnsureDummyUser(context.Background(), models.RoleUser)
	if err != nil {
		t.Fatal(err)
	}
	if admin.ID == user.ID {
		t.Fatal("dummy users must differ")
	}

	roomSvc := service.NewRoomService(rooms)
	room, err := roomSvc.Create(context.Background(), "Integration room", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	targetDate := now.Add(24 * time.Hour)
	weekday := int(targetDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	schSvc := service.NewScheduleService(rooms, schedules, slots, 3).WithClock(testClock{t: now})
	if _, err := schSvc.Create(context.Background(), room.ID, []int{weekday}, "10:00", "11:00"); err != nil {
		t.Fatal(err)
	}

	slotSvc := service.NewSlotService(rooms, slots, schSvc)
	available, err := slotSvc.ListAvailable(context.Background(), room.ID, targetDate)
	if err != nil {
		t.Fatal(err)
	}
	if len(available) == 0 {
		t.Fatal("expected generated slots")
	}

	bookingSvc := service.NewBookingService(slots, bookings, noOpConferenceClient{}).WithClock(testClock{t: targetDate.Add(-1 * time.Hour)})
	created, err := bookingSvc.Create(context.Background(), available[0].ID, user.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := bookingSvc.Cancel(context.Background(), created.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != models.BookingStatusCancelled {
		t.Fatalf("unexpected status: %s", cancelled.Status)
	}
}

type noOpConferenceClient struct{}

func (noOpConferenceClient) CreateLink(ctx context.Context, slot models.Slot, userID string) (string, error) {
	_ = ctx
	_ = slot
	_ = userID
	return "", nil
}

type testClock struct{ t time.Time }

func (c testClock) Now() time.Time { return c.t }

func cleanupDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	stmts := []string{
		`DELETE FROM bookings`,
		`DELETE FROM slots`,
		`DELETE FROM schedules`,
		`DELETE FROM rooms`,
		`DELETE FROM users`,
	}
	for _, stmt := range stmts {
		if _, err := pool.Exec(context.Background(), stmt); err != nil {
			t.Fatalf("cleanup failed on %q: %v", stmt, err)
		}
	}
}
