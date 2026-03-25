CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    role TEXT NOT NULL CHECK (role IN ('admin','user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    capacity INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week SMALLINT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (room_id, start_time, end_time)
);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('active','cancelled')),
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_booking_per_slot
    ON bookings (slot_id)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_slots_room_start
    ON slots (room_id, start_time);

CREATE INDEX IF NOT EXISTS idx_bookings_user_status
    ON bookings (user_id, status);

CREATE INDEX IF NOT EXISTS idx_bookings_slot_id
    ON bookings (slot_id);
