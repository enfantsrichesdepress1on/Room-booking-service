package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"room-booking-service/internal/app"
	"room-booking-service/internal/config"
	"room-booking-service/internal/repository/memory"
)

func TestCreateRoomScheduleBookingFlow(t *testing.T) {
	store := memory.NewStore()
	ts := httptest.NewServer(app.BuildHandler(config.Load(), memory.UserRepo{S: store}, memory.RoomRepo{S: store}, memory.ScheduleRepo{S: store}, memory.SlotRepo{S: store}, memory.BookingRepo{S: store}))
	defer ts.Close()

	adminToken := dummyLogin(t, ts, "admin")
	userToken := dummyLogin(t, ts, "user")
	roomID := createRoom(t, ts, adminToken)
	createSchedule(t, ts, adminToken, roomID)
	date := nextWeekdayUTC(3).Format("2006-01-02")
	slots := listSlots(t, ts, userToken, roomID, date)
	if len(slots) == 0 {
		t.Fatal("expected available slots")
	}
	bookingID := createBooking(t, ts, userToken, slots[0]["id"].(string))
	if bookingID == "" {
		t.Fatal("empty booking id")
	}
}

func TestCancelBookingFlow(t *testing.T) {
	store := memory.NewStore()
	ts := httptest.NewServer(app.BuildHandler(config.Load(), memory.UserRepo{S: store}, memory.RoomRepo{S: store}, memory.ScheduleRepo{S: store}, memory.SlotRepo{S: store}, memory.BookingRepo{S: store}))
	defer ts.Close()

	adminToken := dummyLogin(t, ts, "admin")
	userToken := dummyLogin(t, ts, "user")
	roomID := createRoom(t, ts, adminToken)
	createSchedule(t, ts, adminToken, roomID)
	slots := listSlots(t, ts, userToken, roomID, nextWeekdayUTC(3).Format("2006-01-02"))
	bookingID := createBooking(t, ts, userToken, slots[0]["id"].(string))

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/bookings/"+bookingID+"/cancel", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var out map[string]map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["booking"]["status"] != "cancelled" {
		t.Fatalf("expected cancelled, got %+v", out)
	}
}

func dummyLogin(t *testing.T, ts *httptest.Server, role string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"role": role})
	resp, err := http.Post(ts.URL+"/dummyLogin", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var out map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out["token"]
}

func createRoom(t *testing.T, ts *httptest.Server, token string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"name": "Room A", "capacity": 6})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/rooms/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var out map[string]map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out["room"]["id"].(string)
}

func createSchedule(t *testing.T, ts *httptest.Server, token, roomID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"daysOfWeek": []int{3}, "startTime": "10:00", "endTime": "12:00"})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/rooms/"+roomID+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func listSlots(t *testing.T, ts *httptest.Server, token, roomID, date string) []map[string]any {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/rooms/"+roomID+"/slots/list?date="+date, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var out map[string][]map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out["slots"]
}

func createBooking(t *testing.T, ts *httptest.Server, token, slotID string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"slotId": slotID, "createConferenceLink": true})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/bookings/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var out map[string]map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out["booking"]["id"].(string)
}

func nextWeekdayUTC(target int) time.Time {
	now := time.Now().UTC()
	for i := 1; i <= 14; i++ {
		candidate := now.AddDate(0, 0, i)
		wd := int(candidate.Weekday())
		if wd == 0 {
			wd = 7
		}
		if wd == target {
			return candidate
		}
	}
	return now.AddDate(0, 0, 1)
}
