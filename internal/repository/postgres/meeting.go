package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jus1d/kypidbot/internal/domain"
)

type MeetingRepo struct {
	db *sql.DB
}

func NewMeetingRepo(d *DB) *MeetingRepo {
	return &MeetingRepo{db: d.db}
}

func (r *MeetingRepo) SaveMeeting(ctx context.Context, m *domain.Meeting) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO meetings (dill_id, doe_id, pair_score, is_fullmatch)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		m.DillID, m.DoeID, m.PairScore, m.IsFullmatch,
	).Scan(&m.ID)
}

func (r *MeetingRepo) GetMeetingByID(ctx context.Context, id int64) (*domain.Meeting, error) {
	var m domain.Meeting

	err := r.db.QueryRowContext(ctx, `
		SELECT id, dill_id, doe_id, pair_score, is_fullmatch,
		       place_id, time, dill_state, doe_state, users_notified,
		       dill_cant_find, doe_cant_find
		FROM meetings WHERE id = $1`, id).Scan(
		&m.ID, &m.DillID, &m.DoeID, &m.PairScore, &m.IsFullmatch,
		&m.PlaceID, &m.Time, &m.DillState, &m.DoeState, &m.UsersNotified,
		&m.DillCantFind, &m.DoeCantFind,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MeetingRepo) GetMeetingByUsers(ctx context.Context, dillID, doeID int64) (*domain.Meeting, error) {
	var m domain.Meeting

	err := r.db.QueryRowContext(ctx, `
		SELECT id, dill_id, doe_id, pair_score, is_fullmatch,
		       place_id, time, dill_state, doe_state, users_notified,
		       dill_cant_find, doe_cant_find
		FROM meetings WHERE dill_id = $1 AND doe_id = $2`, dillID, doeID).Scan(
		&m.ID, &m.DillID, &m.DoeID, &m.PairScore, &m.IsFullmatch,
		&m.PlaceID, &m.Time, &m.DillState, &m.DoeState, &m.UsersNotified,
		&m.DillCantFind, &m.DoeCantFind,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MeetingRepo) GetRegularMeetings(ctx context.Context) ([]domain.Meeting, error) {
	return r.getMeetingsByFullmatch(ctx, false)
}

func (r *MeetingRepo) GetFullMeetings(ctx context.Context) ([]domain.Meeting, error) {
	return r.getMeetingsByFullmatch(ctx, true)
}

func (r *MeetingRepo) getMeetingsByFullmatch(ctx context.Context, fullmatch bool) ([]domain.Meeting, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, dill_id, doe_id, pair_score, is_fullmatch,
		       place_id, time, dill_state, doe_state, users_notified,
		       dill_cant_find, doe_cant_find
		FROM meetings WHERE is_fullmatch = $1`, fullmatch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		if err := rows.Scan(
			&m.ID, &m.DillID, &m.DoeID, &m.PairScore, &m.IsFullmatch,
			&m.PlaceID, &m.Time, &m.DillState, &m.DoeState, &m.UsersNotified,
			&m.DillCantFind, &m.DoeCantFind,
		); err != nil {
			return nil, err
		}
		meetings = append(meetings, m)
	}
	return meetings, rows.Err()
}

func (r *MeetingRepo) AssignPlaceAndTime(ctx context.Context, id int64, placeID int64, time time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE meetings SET place_id = $1, time = $2 WHERE id = $3`,
		placeID, time, id)
	return err
}

func (r *MeetingRepo) UpdateState(ctx context.Context, meetingID int64, isDill bool, state domain.ConfirmationState) error {
	col := "doe_state"
	if isDill {
		col = "dill_state"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE meetings SET `+col+` = $1 WHERE id = $2`, state, meetingID)
	return err
}

func (r *MeetingRepo) ClearMeetings(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM meetings`)
	return err
}

func (r *MeetingRepo) GetMeetingsStartingIn(ctx context.Context, interval time.Duration) ([]domain.Meeting, error) {
	secs := fmt.Sprintf("%ds", int(interval.Seconds()))
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, dill_id, doe_id, pair_score, is_fullmatch,
		       place_id, time, dill_state, doe_state, users_notified,
		       dill_cant_find, doe_cant_find
		FROM meetings WHERE time >= NOW() AND time <= NOW() + $1::interval AND users_notified = FALSE`, secs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		if err := rows.Scan(
			&m.ID, &m.DillID, &m.DoeID, &m.PairScore, &m.IsFullmatch,
			&m.PlaceID, &m.Time, &m.DillState, &m.DoeState, &m.UsersNotified,
			&m.DillCantFind, &m.DoeCantFind,
		); err != nil {
			return nil, err
		}
		meetings = append(meetings, m)
	}
	return meetings, rows.Err()
}

func (r *MeetingRepo) MarkNotified(ctx context.Context, meetingID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE meetings SET users_notified = TRUE WHERE id = $1`, meetingID)
	return err
}

func (r *MeetingRepo) SetCantFind(ctx context.Context, meetingID int64, isDill bool) error {
	col := "doe_cant_find"
	if isDill {
		col = "dill_cant_find"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE meetings SET `+col+` = TRUE WHERE id = $1`, meetingID)
	return err
}

func (r *MeetingRepo) GetArrivedMeetingID(ctx context.Context, telegramID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT m.id FROM meetings m
		JOIN users d ON m.dill_id = d.telegram_id
		JOIN users e ON m.doe_id = e.telegram_id
		WHERE (d.telegram_id = $1 AND m.dill_state = 'arrived')
		   OR (e.telegram_id = $1 AND m.doe_state = 'arrived')
		LIMIT 1`, telegramID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

func (r *MeetingRepo) GetMeetingStats(ctx context.Context) (domain.MeetingStats, error) {
	var s domain.MeetingStats
	err := r.db.QueryRowContext(ctx, `SELECT
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE dill_state = 'confirmed' AND doe_state = 'confirmed') AS confirmed,
		COUNT(*) FILTER (WHERE dill_state = 'cancelled' OR doe_state = 'cancelled') AS cancelled,
		COUNT(*) FILTER (WHERE dill_state != 'cancelled' AND doe_state != 'cancelled'
			AND NOT (dill_state = 'confirmed' AND doe_state = 'confirmed')) AS pending
		FROM meetings`).Scan(&s.Total, &s.Confirmed, &s.Cancelled, &s.Pending)
	if err != nil {
		return domain.MeetingStats{}, err
	}
	return s, nil
}
