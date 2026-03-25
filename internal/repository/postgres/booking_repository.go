package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"room-booking-service/internal/models"
	"room-booking-service/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct{ db *pgxpool.Pool }

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository { return &BookingRepository{db: db} }

func (r *BookingRepository) Create(ctx context.Context, booking models.Booking) (models.Booking, error) {
	var created models.Booking
	err := r.db.QueryRow(ctx, `INSERT INTO bookings (id,slot_id,user_id,status,conference_link) VALUES ($1,$2,$3,$4,$5) RETURNING id,slot_id,user_id,status,conference_link,created_at`, booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink).
		Scan(&created.ID, &created.SlotID, &created.UserID, &created.Status, &created.ConferenceLink, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return models.Booking{}, service.ErrSlotAlreadyBooked
			case "23503":
				return models.Booking{}, service.ErrSlotNotFound
			}
		}
		return models.Booking{}, fmt.Errorf("create booking: %w", err)
	}
	return created, nil
}

func (r *BookingRepository) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, `SELECT id,slot_id,user_id,status,conference_link,created_at FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]models.Booking, 0)
	for rows.Next() {
		var booking models.Booking
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, booking)
	}
	return items, total, rows.Err()
}

func (r *BookingRepository) ListByUserFuture(ctx context.Context, userID string, now time.Time) ([]models.Booking, error) {
	rows, err := r.db.Query(ctx, `SELECT b.id,b.slot_id,b.user_id,b.status,b.conference_link,b.created_at
		FROM bookings b JOIN slots s ON s.id=b.slot_id
		WHERE b.user_id=$1 AND s.start_time >= $2
		ORDER BY s.start_time`, userID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Booking, 0)
	for rows.Next() {
		var booking models.Booking
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, booking)
	}
	return items, rows.Err()
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID string) (models.Booking, error) {
	var booking models.Booking
	err := r.db.QueryRow(ctx, `SELECT id,slot_id,user_id,status,conference_link,created_at FROM bookings WHERE id=$1`, bookingID).
		Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Booking{}, service.ErrBookingNotFound
	}
	return booking, err
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) (models.Booking, error) {
	var booking models.Booking
	err := r.db.QueryRow(ctx, `UPDATE bookings SET status=$2 WHERE id=$1 RETURNING id,slot_id,user_id,status,conference_link,created_at`, bookingID, status).
		Scan(&booking.ID, &booking.SlotID, &booking.UserID, &booking.Status, &booking.ConferenceLink, &booking.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Booking{}, service.ErrBookingNotFound
	}
	return booking, err
}
