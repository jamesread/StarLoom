package store

import (
	"context"
	"database/sql"
)

func (s *SQLite) ListChoreNotificationSubscriptions(ctx context.Context, subscriberMemberID int) ([]ChoreNotificationSubscriptionRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, family_id, subscriber_member_id, child_member_id, chore_id, created_at
		FROM chore_notification_subscriptions
		WHERE subscriber_member_id = ?
		ORDER BY COALESCE(child_member_id, 0), COALESCE(chore_id, 0), id`,
		subscriberMemberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ChoreNotificationSubscriptionRow, 0)
	for rows.Next() {
		var row ChoreNotificationSubscriptionRow
		var childID, choreID sql.NullInt64
		if err := rows.Scan(&row.ID, &row.FamilyID, &row.SubscriberMemberID, &childID, &choreID, &row.CreatedAt); err != nil {
			return nil, err
		}
		if childID.Valid {
			v := int(childID.Int64)
			row.ChildMemberID = &v
		}
		if choreID.Valid {
			v := int(choreID.Int64)
			row.ChoreID = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *SQLite) ReplaceChoreNotificationSubscriptions(ctx context.Context, familyID, subscriberMemberID int, subs []ChoreNotificationSubscriptionRow) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chore_notification_subscriptions WHERE subscriber_member_id = ?`, subscriberMemberID); err != nil {
		return err
	}
	for i := range subs {
		var childID, choreID any
		if subs[i].ChildMemberID != nil {
			childID = *subs[i].ChildMemberID
		}
		if subs[i].ChoreID != nil {
			choreID = *subs[i].ChoreID
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO chore_notification_subscriptions (family_id, subscriber_member_id, child_member_id, chore_id)
			VALUES (?, ?, ?, ?)`,
			familyID, subscriberMemberID, childID, choreID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) MatchingChoreNotificationSubscribers(ctx context.Context, familyID, childMemberID, choreID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT subscriber_member_id
		FROM chore_notification_subscriptions
		WHERE family_id = ?
		  AND (child_member_id IS NULL OR child_member_id = ?)
		  AND (chore_id IS NULL OR chore_id = ?)`,
		familyID, childMemberID, choreID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
