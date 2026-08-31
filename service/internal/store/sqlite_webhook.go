package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func webhookTargetSelectSQL() string {
	return `SELECT id, url, secret, enabled,
		COALESCE(created, ''), COALESCE(updated, '')
		FROM webhook_targets`
}

func scanWebhookTarget(s interface{ Scan(...any) error }) (*WebhookTargetRow, error) {
	var w WebhookTargetRow
	var enabled int
	if err := s.Scan(&w.ID, &w.URL, &w.Secret, &enabled, &w.Created, &w.Updated); err != nil {
		return nil, err
	}
	w.Enabled = enabled != 0
	return &w, nil
}

func scanWebhookTargets(rows *sql.Rows) ([]WebhookTargetRow, error) {
	out := make([]WebhookTargetRow, 0)
	for rows.Next() {
		w, err := scanWebhookTarget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, rows.Err()
}

func (s *SQLite) ListWebhookTargets(ctx context.Context) ([]WebhookTargetRow, error) {
	rows, err := s.db.QueryContext(ctx, webhookTargetSelectSQL()+` ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	targets, err := scanWebhookTargets(rows)
	if err != nil {
		return nil, err
	}
	for i := range targets {
		events, loadErr := s.loadWebhookEvents(ctx, targets[i].ID)
		if loadErr != nil {
			return nil, loadErr
		}
		targets[i].Events = events
	}
	return targets, nil
}

func (s *SQLite) FindWebhookTarget(ctx context.Context, id int) (*WebhookTargetRow, error) {
	row := s.db.QueryRowContext(ctx, webhookTargetSelectSQL()+` WHERE id = ?`, id)
	w, err := scanWebhookTarget(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	events, loadErr := s.loadWebhookEvents(ctx, w.ID)
	if loadErr != nil {
		return nil, loadErr
	}
	w.Events = events
	return w, nil
}

func (s *SQLite) EnabledTargetsForEvent(ctx context.Context, event string) ([]WebhookTargetRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.url, t.secret, t.enabled,
			COALESCE(t.created, ''), COALESCE(t.updated, '')
		 FROM webhook_targets t
		 INNER JOIN webhook_events e ON e.webhook_target_id = t.id
		 WHERE e.event = ? AND t.enabled = 1`, event)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWebhookTargets(rows)
}

func (s *SQLite) CreateWebhookTarget(ctx context.Context, url, secret string, events []string, enabled bool) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	id, err := insertWebhookTargetTx(ctx, tx, url, secret, events, enabled)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func insertWebhookTargetTx(ctx context.Context, tx *sql.Tx, url, secret string, events []string, enabled bool) (int, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := tx.ExecContext(ctx,
		`INSERT INTO webhook_targets (url, secret, enabled, created, updated) VALUES (?, ?, ?, ?, ?)`,
		url, secret, boolToInt(enabled), now, now)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := insertWebhookEventsTx(ctx, tx, int(lid), events); err != nil {
		return 0, err
	}
	return int(lid), nil
}

func (s *SQLite) UpdateWebhookTarget(ctx context.Context, id int, url, secret string, events []string, enabled bool, clearSecret bool) error {
	cur, err := s.FindWebhookTarget(ctx, id)
	if err != nil || cur == nil {
		return sql.ErrNoRows
	}
	if url != "" {
		cur.URL = url
	}
	if !clearSecret && secret != "" {
		cur.Secret = secret
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	if _, err := tx.ExecContext(ctx,
		`UPDATE webhook_targets SET url = ?, secret = ?, enabled = ?, updated = ? WHERE id = ?`,
		cur.URL, cur.Secret, boolToInt(enabled), now, id); err != nil {
		return err
	}
	if err := replaceWebhookEventsTx(ctx, tx, id, events); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLite) DeleteWebhookTarget(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM webhook_targets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) loadWebhookEvents(ctx context.Context, targetID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT event FROM webhook_events WHERE webhook_target_id = ? ORDER BY event`, targetID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	events := make([]string, 0)
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func replaceWebhookEventsTx(ctx context.Context, tx *sql.Tx, targetID int, events []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM webhook_events WHERE webhook_target_id = ?`, targetID); err != nil {
		return err
	}
	return insertWebhookEventsTx(ctx, tx, targetID, events)
}

func insertWebhookEventsTx(ctx context.Context, tx *sql.Tx, targetID int, events []string) error {
	for _, e := range events {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO webhook_events (webhook_target_id, event) VALUES (?, ?)`, targetID, e); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLite) InsertWebhookDelivery(ctx context.Context, row WebhookDeliveryRow) (int, error) {
	firedAt := row.FiredAt
	if firedAt == "" {
		firedAt = time.Now().UTC().Format(time.RFC3339)
	}
	var targetID any
	if row.WebhookTargetID > 0 {
		targetID = row.WebhookTargetID
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO webhook_deliveries (webhook_target_id, event, url, success, http_status, error_message, fired_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		targetID, row.Event, row.URL, boolToInt(row.Success), row.HTTPStatus, row.ErrorMessage, firedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) ListWebhookDeliveries(ctx context.Context, limit int) ([]WebhookDeliveryRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, COALESCE(webhook_target_id, 0), event, url, success, http_status, error_message, fired_at
		 FROM webhook_deliveries
		 ORDER BY fired_at DESC, id DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []WebhookDeliveryRow{}
	for rows.Next() {
		var row WebhookDeliveryRow
		var success int
		if err := rows.Scan(&row.ID, &row.WebhookTargetID, &row.Event, &row.URL, &success, &row.HTTPStatus, &row.ErrorMessage, &row.FiredAt); err != nil {
			return nil, err
		}
		row.Success = success != 0
		out = append(out, row)
	}
	return out, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
