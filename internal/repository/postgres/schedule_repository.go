package postgres

import (
	"context"
	"errors"
	"fmt"

	"room-booking-service/internal/models"
	"room-booking-service/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct{ db *pgxpool.Pool }

func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository { return &ScheduleRepository{db: db} }

func (r *ScheduleRepository) Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	var created models.Schedule
	err := r.db.QueryRow(ctx, `INSERT INTO schedules (id,room_id,days_of_week,start_time,end_time) VALUES ($1,$2,$3,$4,$5) RETURNING id,room_id,days_of_week,to_char(start_time,'HH24:MI'),to_char(end_time,'HH24:MI'),created_at`, schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).
		Scan(&created.ID, &created.RoomID, &created.DaysOfWeek, &created.StartTime, &created.EndTime, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Schedule{}, service.ErrScheduleExists
		}
		return models.Schedule{}, fmt.Errorf("create schedule: %w", err)
	}
	return created, nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error) {
	var schedule models.Schedule
	err := r.db.QueryRow(ctx, `SELECT id,room_id,days_of_week,to_char(start_time,'HH24:MI'),to_char(end_time,'HH24:MI'),created_at FROM schedules WHERE room_id=$1`, roomID).
		Scan(&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime, &schedule.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Schedule{}, service.ErrNotFound
	}
	if err != nil {
		return models.Schedule{}, fmt.Errorf("get schedule: %w", err)
	}
	return schedule, nil
}
