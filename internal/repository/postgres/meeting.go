package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jus1d/kypidbot/internal/domain"
)

type MeetingRepo struct {
	db *sql.DB
}

func NewMeetingRepo(d *DB) *MeetingRepo {
	return &MeetingRepo{db: d.db}
}

func (r *MeetingRepo) SaveMeeting(ctx context.Context, m *domain.Meeting) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO meetings (pair_id, place_id, time)
		VALUES ($1, $2, $3)
		RETURNING id`,
		m.PairID, m.PlaceID, m.Time,
	).Scan(&id)
	return id, err
}

func (r *MeetingRepo) GetMeetingByID(ctx context.Context, id int64) (*domain.Meeting, error) {
	var m domain.Meeting

	err := r.db.QueryRowContext(ctx, `
		SELECT id, pair_id, place_id, time, dill_confirmed, doe_confirmed, dill_cancelled, doe_cancelled
		FROM meetings WHERE id = $1`, id).Scan(
		&m.ID, &m.PairID, &m.PlaceID, &m.Time,
		&m.DillConfirmed, &m.DoeConfirmed, &m.DillCancelled, &m.DoeCancelled,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MeetingRepo) ConfirmMeeting(ctx context.Context, meetingID int64, isDill bool) error {
	col := "doe_confirmed"
	if isDill {
		col = "dill_confirmed"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE meetings SET `+col+` = TRUE WHERE id = $1`, meetingID)
	return err
}

func (r *MeetingRepo) CancelMeeting(ctx context.Context, meetingID int64, isDill bool) error {
	col := "doe_cancelled"
	if isDill {
		col = "dill_cancelled"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE meetings SET `+col+` = TRUE WHERE id = $1`, meetingID)
	return err
}
