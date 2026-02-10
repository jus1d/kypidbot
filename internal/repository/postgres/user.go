package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jus1d/kypidbot/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(d *DB) *UserRepo {
	return &UserRepo{db: d.db}
}

func (r *UserRepo) SaveUser(ctx context.Context, u *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name, is_bot, language_code, is_premium)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			is_bot = EXCLUDED.is_bot,
			language_code = EXCLUDED.language_code,
			is_premium = EXCLUDED.is_premium`,
		u.TelegramID, u.Username, u.FirstName, u.LastName,
		u.IsBot, u.LanguageCode, u.IsPremium,
	)
	return err
}

func (r *UserRepo) GetUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
		       referral_code, referrer_id, created_at
		FROM users WHERE telegram_id = $1`, telegramID)
	return scanUser(row)
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
		       referral_code, referrer_id, created_at
		FROM users WHERE username = $1`, username)
	return scanUser(row)
}

func (r *UserRepo) GetUserByReferralCode(ctx context.Context, code string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
		       referral_code, referrer_id, created_at
		FROM users WHERE referral_code = $1`, code)
	return scanUser(row)
}

func (r *UserRepo) GetUserState(ctx context.Context, telegramID int64) (domain.UserState, error) {
	var state domain.UserState
	err := r.db.QueryRowContext(ctx,
		`SELECT state FROM users WHERE telegram_id = $1`, telegramID).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UserStateStart, nil
	}
	return state, err
}

func (r *UserRepo) SetUserState(ctx context.Context, telegramID int64, state domain.UserState) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET state = $1 WHERE telegram_id = $2`, state, telegramID)
	return err
}

func (r *UserRepo) SetUserSex(ctx context.Context, telegramID int64, sex string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET sex = $1 WHERE telegram_id = $2`, sex, telegramID)
	return err
}

func (r *UserRepo) SetUserAbout(ctx context.Context, telegramID int64, about string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET about = $1 WHERE telegram_id = $2`, about, telegramID)
	return err
}

func (r *UserRepo) GetTimeRanges(ctx context.Context, telegramID int64) (string, error) {
	var tr string
	err := r.db.QueryRowContext(ctx,
		`SELECT time_ranges FROM users WHERE telegram_id = $1`, telegramID).Scan(&tr)
	if errors.Is(err, sql.ErrNoRows) {
		return "000000", nil
	}
	return tr, err
}

func (r *UserRepo) SaveTimeRanges(ctx context.Context, telegramID int64, timeRanges string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET time_ranges = $1 WHERE telegram_id = $2`, timeRanges, telegramID)
	return err
}

func (r *UserRepo) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	var isAdmin bool
	err := r.db.QueryRowContext(ctx,
		`SELECT is_admin FROM users WHERE telegram_id = $1`, telegramID).Scan(&isAdmin)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *UserRepo) SetAdmin(ctx context.Context, telegramID int64, isAdmin bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_admin = $1 WHERE telegram_id = $2`, isAdmin, telegramID)
	return err
}

func (r *UserRepo) SetReferralCode(ctx context.Context, telegramID int64, code string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET referral_code = $1 WHERE telegram_id = $2`, code, telegramID)
	return err
}

func (r *UserRepo) SetReferrer(ctx context.Context, telegramID int64, referrerID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET referrer_id = $1 WHERE telegram_id = $2`, referrerID, telegramID)
	return err
}

func (r *UserRepo) GetVerifiedUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
		       referral_code, referrer_id, created_at
		FROM users WHERE state = 'completed' AND opted_out = FALSE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepo) GetAdmins(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
		       referral_code, referrer_id, created_at
		FROM users WHERE is_admin = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepo) GetUserUsername(ctx context.Context, telegramID int64) (string, error) {
	var username sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT username FROM users WHERE telegram_id = $1`, telegramID).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return username.String, nil
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var username, firstName, lastName, languageCode, sex, referralCode sql.NullString
	var referrerID sql.NullInt64

	err := row.Scan(
		&u.TelegramID, &username, &firstName, &lastName,
		&u.IsBot, &languageCode, &u.IsPremium, &sex, &u.About,
		&u.State, &u.RegistrationNotified, &u.InviteNotified, &u.TimeRanges, &u.IsAdmin, &u.OptedOut,
		&referralCode, &referrerID, &u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.Username = username.String
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.LanguageCode = languageCode.String
	u.Sex = sex.String
	u.ReferralCode = referralCode.String
	if referrerID.Valid {
		u.ReferrerID = &referrerID.Int64
	}

	return &u, nil
}

func scanUserFromRows(rows *sql.Rows) (*domain.User, error) {
	var u domain.User
	var username, firstName, lastName, languageCode, sex, referralCode sql.NullString
	var referrerID sql.NullInt64

	err := rows.Scan(
		&u.TelegramID, &username, &firstName, &lastName,
		&u.IsBot, &languageCode, &u.IsPremium, &sex, &u.About,
		&u.State, &u.RegistrationNotified, &u.InviteNotified, &u.TimeRanges, &u.IsAdmin, &u.OptedOut,
		&referralCode, &referrerID, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.Username = username.String
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.LanguageCode = languageCode.String
	u.Sex = sex.String
	u.ReferralCode = referralCode.String
	if referrerID.Valid {
		u.ReferrerID = &referrerID.Int64
	}

	return &u, nil
}

func (r *UserRepo) GetReferralLeaderboard(ctx context.Context) ([]domain.ReferralLeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT
            users.referrer_id,
            COUNT(*) as referral_count,
            u.username,
            u.first_name
        FROM users
        JOIN users u ON users.referrer_id = u.telegram_id
        WHERE users.referrer_id IS NOT NULL
          AND u.is_admin = FALSE
        GROUP BY users.referrer_id, u.username, u.first_name, u.created_at
        ORDER BY COUNT(*) DESC, u.created_at ASC
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []domain.ReferralLeaderboardEntry
	for rows.Next() {
		var entry domain.ReferralLeaderboardEntry
		var username, firstName sql.NullString

		err := rows.Scan(
			&entry.ReferrerID,
			&entry.ReferralCount,
			&username,
			&firstName,
		)
		if err != nil {
			return nil, err
		}

		entry.Username = username.String
		entry.FirstName = firstName.String
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, rows.Err()
}

func (r *UserRepo) MarkNotified(ctx context.Context, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET registration_notified = TRUE WHERE telegram_id = $1`, telegramID)
	return err
}

func (r *UserRepo) GetNotCompleted(ctx context.Context, interval time.Duration) ([]domain.User, error) {
	secs := fmt.Sprintf("%ds", int(interval.Seconds()))
	rows, err := r.db.QueryContext(ctx, `SELECT telegram_id, username, first_name, last_name, is_bot,
	       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
	       referral_code, referrer_id, created_at
	FROM users WHERE now() - created_at > $1::interval AND registration_notified = FALSE AND state <> 'completed'`, secs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepo) GetForInviteReminder(ctx context.Context, interval time.Duration) ([]domain.User, error) {
	secs := fmt.Sprintf("%ds", int(interval.Seconds()))
	rows, err := r.db.QueryContext(ctx, `SELECT telegram_id, username, first_name, last_name, is_bot,
	       language_code, is_premium, sex, about, state, registration_notified, invite_notified, time_ranges, is_admin, opted_out,
	       referral_code, referrer_id, created_at
	FROM users WHERE now() - created_at > $1::interval AND invite_notified = FALSE AND is_admin = FALSE`, secs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepo) MarkInviteNotified(ctx context.Context, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET invite_notified = TRUE WHERE telegram_id = $1`, telegramID)
	return err
}

func (r *UserRepo) SetOptedOut(ctx context.Context, telegramID int64, optedOut bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET opted_out = $1 WHERE telegram_id = $2`, optedOut, telegramID)
	return err
}

func (r *UserRepo) GetLastRegisteredCount(ctx context.Context) (daily uint, weekly uint, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT
		COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '1 day')  AS users_last_24h,
		COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days') AS users_last_7d
		FROM users`)
	err = row.Scan(&daily, &weekly)
	if err != nil {
		return 0, 0, err
	}

	return daily, weekly, nil
}

func (r *UserRepo) GetSexCounts(ctx context.Context) (males uint, females uint, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT
		COUNT(*) FILTER (WHERE sex = 'male') AS males,
		COUNT(*) FILTER (WHERE sex = 'female') AS females
		FROM users WHERE state = 'completed'`)
	err = row.Scan(&males, &females)
	if err != nil {
		return 0, 0, err
	}
	return males, females, nil
}
