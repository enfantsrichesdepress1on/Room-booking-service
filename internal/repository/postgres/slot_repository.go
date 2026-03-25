package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"room-booking-service/internal/models"
	"room-booking-service/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct{ db *pgxpool.Pool }

func NewSlotRepository(db *pgxpool.Pool) *SlotRepository { return &SlotRepository{db: db} }

func (r *SlotRepository) CreateBatch(ctx context.Context, slots []models.Slot) error {
	batch := &pgx.Batch{}
	for _, slot := range slots {
		batch.Queue(`INSERT INTO slots (id,room_id,start_time,end_time) VALUES ($1,$2,$3,$4) ON CONFLICT (room_id,start_time,end_time) DO NOTHING`, slot.ID, slot.RoomID, slot.Start.UTC(), slot.End.UTC())
	}
	results := r.db.SendBatch(ctx, batch)
	defer results.Close()
	for range slots {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("insert slot batch: %w", err)
		}
	}
	return nil
}

func (r *SlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]models.Slot, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	rows, err := r.db.Query(ctx, `SELECT s.id,s.room_id,s.start_time,s.end_time,s.created_at
		FROM slots s
		LEFT JOIN bookings b ON b.slot_id = s.id AND b.status='active'
		WHERE s.room_id=$1 AND s.start_time >= $2 AND s.start_time < $3 AND b.id IS NULL
		ORDER BY s.start_time`, roomID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Slot, 0)
	for rows.Next() {
		var slot models.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End, &slot.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, slot)
	}
	return items, rows.Err()
}

func (r *SlotRepository) ExistsForRange(ctx context.Context, roomID string, start, end time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM slots WHERE room_id=$1 AND start_time >= $2 AND start_time < $3)`, roomID, start.UTC(), end.UTC()).Scan(&exists)
	return exists, err
}

func (r *SlotRepository) GetByID(ctx context.Context, slotID string) (models.Slot, error) {
	var slot models.Slot
	err := r.db.QueryRow(ctx, `SELECT id,room_id,start_time,end_time,created_at FROM slots WHERE id=$1`, slotID).
		Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End, &slot.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Slot{}, service.ErrSlotNotFound
	}
	return slot, err
}
