package conference

import (
	"context"
	"fmt"
	"net/url"

	"room-booking-service/internal/models"
)

type MockClient struct{ BaseURL string }

func (m MockClient) CreateLink(ctx context.Context, slot models.Slot, userID string) (string, error) {
	_ = ctx
	u, err := url.Parse(m.BaseURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("slotId", slot.ID)
	q.Set("userId", userID)
	u.RawQuery = q.Encode()
	return fmt.Sprintf("%s", u.String()), nil
}
