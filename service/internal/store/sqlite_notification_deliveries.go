package store

import (
	"context"
)

func (s *SQLite) InsertNotificationDelivery(ctx context.Context, row NotificationDeliveryRow) (int, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_deliveries (family_id, recipient_member_id, notification_type, title, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?)`,
		nullInt(row.FamilyID),
		nullInt(row.RecipientMemberID),
		row.NotificationType,
		row.Title,
		boolInt(row.Success),
		row.ErrorMessage,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s *SQLite) ListNotificationDeliveries(ctx context.Context, limit int) ([]NotificationDeliveryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT nd.id, COALESCE(nd.family_id, 0), COALESCE(nd.recipient_member_id, 0),
		       COALESCE(fm.display_name, ''), nd.notification_type, nd.title, nd.success,
		       nd.error_message, nd.sent_at
		FROM notification_deliveries nd
		LEFT JOIN family_members fm ON fm.id = nd.recipient_member_id
		ORDER BY nd.sent_at DESC, nd.id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]NotificationDeliveryRow, 0)
	for rows.Next() {
		var row NotificationDeliveryRow
		var success int
		if err := rows.Scan(
			&row.ID, &row.FamilyID, &row.RecipientMemberID, &row.RecipientDisplayName,
			&row.NotificationType, &row.Title, &success, &row.ErrorMessage, &row.SentAt,
		); err != nil {
			return nil, err
		}
		row.Success = success != 0
		out = append(out, row)
	}
	return out, rows.Err()
}

func nullInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
