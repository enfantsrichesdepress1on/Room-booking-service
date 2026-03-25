package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"room-booking-service/internal/auth"
	"room-booking-service/internal/middleware"
	"room-booking-service/internal/models"
	"room-booking-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	jwt        *auth.JWTManager
	authSvc    *service.AuthService
	roomSvc    *service.RoomService
	schSvc     *service.ScheduleService
	slotSvc    *service.SlotService
	bookingSvc *service.BookingService
}

func NewHandler(jwt *auth.JWTManager, authSvc *service.AuthService, roomSvc *service.RoomService, schSvc *service.ScheduleService, slotSvc *service.SlotService, bookingSvc *service.BookingService) *Handler {
	return &Handler{jwt: jwt, authSvc: authSvc, roomSvc: roomSvc, schSvc: schSvc, slotSvc: slotSvc, bookingSvc: bookingSvc}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/_info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/dummyLogin", h.dummyLogin)
	r.Post("/register", h.register)
	r.Post("/login", h.login)

	r.Group(func(pr chi.Router) {
		pr.Use(middleware.Auth(h.jwt))
		pr.Get("/rooms/list", h.roomsList)
		pr.With(middleware.RequireRole(string(models.RoleAdmin))).Post("/rooms/create", h.roomsCreate)
		pr.With(middleware.RequireRole(string(models.RoleAdmin))).Post("/rooms/{roomId}/schedule/create", h.scheduleCreate)
		pr.Get("/rooms/{roomId}/slots/list", h.slotsList)
		pr.With(middleware.RequireRole(string(models.RoleUser))).Post("/bookings/create", h.bookingCreate)
		pr.With(middleware.RequireRole(string(models.RoleAdmin))).Get("/bookings/list", h.bookingsList)
		pr.With(middleware.RequireRole(string(models.RoleUser))).Get("/bookings/my", h.bookingsMy)
		pr.With(middleware.RequireRole(string(models.RoleUser))).Post("/bookings/{bookingId}/cancel", h.bookingCancel)
	})
	return r
}

func (h *Handler) dummyLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	user, err := h.authSvc.EnsureDummyUser(r.Context(), models.Role(strings.ToLower(req.Role)))
	if err != nil {
		writeError(w, err)
		return
	}
	token, err := h.jwt.Generate(user.ID, string(user.Role))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string      `json:"email"`
		Password string      `json:"password"`
		Role     models.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	user, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	token, err := h.jwt.Generate(user.ID, string(user.Role))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) roomsList(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomSvc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}

func (h *Handler) roomsCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Capacity    *int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	room, err := h.roomSvc.Create(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"room": room})
}

func (h *Handler) scheduleCreate(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	var req struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	schedule, err := h.schSvc.Create(r.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"schedule": schedule})
}

func (h *Handler) slotsList(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	slots, err := h.slotSvc.ListAvailable(r.Context(), roomID, date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"slots": slots})
}

func (h *Handler) bookingCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SlotID               string `json:"slotId"`
		CreateConferenceLink bool   `json:"createConferenceLink"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SlotID == "" {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	claims := middleware.GetClaims(r.Context())
	booking, err := h.bookingSvc.Create(r.Context(), req.SlotID, claims.UserID, req.CreateConferenceLink)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"booking": booking})
}

func (h *Handler) bookingsList(w http.ResponseWriter, r *http.Request) {
	page, err := parseIntWithDefault(r.URL.Query().Get("page"), 1)
	if err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	pageSize, err := parseIntWithDefault(r.URL.Query().Get("pageSize"), 20)
	if err != nil {
		writeError(w, service.ErrInvalidRequest)
		return
	}
	bookings, pagination, err := h.bookingSvc.ListAll(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings, "pagination": pagination})
}

func (h *Handler) bookingsMy(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	bookings, err := h.bookingSvc.ListMyFuture(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings})
}

func (h *Handler) bookingCancel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	booking, err := h.bookingSvc.Cancel(r.Context(), chi.URLParam(r, "bookingId"), claims.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"booking": booking})
}

func parseIntWithDefault(v string, fallback int) (int, error) {
	if v == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
