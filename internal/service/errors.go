package service

import "errors"

var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("not found")
	ErrRoomNotFound      = errors.New("room not found")
	ErrSlotNotFound      = errors.New("slot not found")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrScheduleExists    = errors.New("schedule already exists")
	ErrEmailAlreadyUsed  = errors.New("email already exists")
)
