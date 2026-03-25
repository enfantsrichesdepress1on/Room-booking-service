package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"room-booking-service/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAuthError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	msg := "internal server error"
	switch {
	case errors.Is(err, service.ErrInvalidRequest), errors.Is(err, service.ErrEmailAlreadyUsed):
		status = http.StatusBadRequest
		code = "INVALID_REQUEST"
		msg = err.Error()
	case errors.Is(err, service.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "UNAUTHORIZED"
		msg = err.Error()
	case errors.Is(err, service.ErrForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"
		msg = err.Error()
	case errors.Is(err, service.ErrRoomNotFound):
		status = http.StatusNotFound
		code = "ROOM_NOT_FOUND"
		msg = err.Error()
	case errors.Is(err, service.ErrSlotNotFound):
		status = http.StatusNotFound
		code = "SLOT_NOT_FOUND"
		msg = err.Error()
	case errors.Is(err, service.ErrBookingNotFound):
		status = http.StatusNotFound
		code = "BOOKING_NOT_FOUND"
		msg = err.Error()
	case errors.Is(err, service.ErrSlotAlreadyBooked):
		status = http.StatusConflict
		code = "SLOT_ALREADY_BOOKED"
		msg = err.Error()
	case errors.Is(err, service.ErrScheduleExists):
		status = http.StatusConflict
		code = "SCHEDULE_EXISTS"
		msg = "schedule for this room already exists and cannot be changed"
	}
	writeAuthError(w, status, code, msg)
}
