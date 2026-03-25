package postgres

import (
	"context"

	"room-booking-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct{ db *pgxpool.Pool }

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository { return &RoomRepository{db: db} }

func (r *RoomRepository) Create(ctx context.Context, room models.Room) (models.Room, error) {
	var created models.Room
	err := r.db.QueryRow(ctx, `INSERT INTO rooms (id,name,description,capacity) VALUES ($1,$2,$3,$4) RETURNING id,name,description,capacity,created_at`, room.ID, room.Name, room.Description, room.Capacity).
		Scan(&created.ID, &created.Name, &created.Description, &created.Capacity, &created.CreatedAt)
	return created, err
}

func (r *RoomRepository) List(ctx context.Context) ([]models.Room, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,description,capacity,created_at FROM rooms ORDER BY created_at, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Room, 0)
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, room)
	}
	return items, rows.Err()
}

func (r *RoomRepository) Exists(ctx context.Context, roomID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM rooms WHERE id=$1)`, roomID).Scan(&exists)
	return exists, err
}
