package store

import (
	"context"
	"sort"
)

type memoryNotificationDeliveries struct {
	nextID int
	rows   []NotificationDeliveryRow
}

func (m *Memory) notificationDeliveryState() *memoryNotificationDeliveries {
	if m.notificationDeliveries == nil {
		m.notificationDeliveries = &memoryNotificationDeliveries{}
	}
	return m.notificationDeliveries
}

func (m *Memory) InsertNotificationDelivery(_ context.Context, row NotificationDeliveryRow) (int, error) {
	st := m.notificationDeliveryState()
	st.nextID++
	row.ID = st.nextID
	if row.SentAt == "" {
		row.SentAt = familyNow()
	}
	if row.RecipientMemberID != 0 && row.RecipientDisplayName == "" {
		if member, err := m.GetMemberByID(context.Background(), row.RecipientMemberID); err == nil && member != nil {
			row.RecipientDisplayName = member.DisplayName
		}
	}
	st.rows = append(st.rows, row)
	return row.ID, nil
}

func (m *Memory) ListNotificationDeliveries(_ context.Context, limit int) ([]NotificationDeliveryRow, error) {
	st := m.notificationDeliveryState()
	out := append([]NotificationDeliveryRow(nil), st.rows...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].SentAt != out[j].SentAt {
			return out[i].SentAt > out[j].SentAt
		}
		return out[i].ID > out[j].ID
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
