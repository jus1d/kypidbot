package postgres

import (
	"context"
	"database/sql"
	"errors"

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
		SELECT id, telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, time_ranges, is_admin
		FROM users WHERE telegram_id = $1`, telegramID)
	return scanUser(row)
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, time_ranges, is_admin
		FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, time_ranges, is_admin
		FROM users WHERE username = $1`, username)
	return scanUser(row)
}

func (r *UserRepo) GetUserState(ctx context.Context, telegramID int64) (string, error) {
	var state string
	err := r.db.QueryRowContext(ctx,
		`SELECT state FROM users WHERE telegram_id = $1`, telegramID).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return "start", nil
	}
	return state, err
}

func (r *UserRepo) SetUserState(ctx context.Context, telegramID int64, state string) error {
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

func (r *UserRepo) GetVerifiedUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, is_bot,
		       language_code, is_premium, sex, about, state, time_ranges, is_admin
		FROM users WHERE state = 'completed'`)
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
	var username, firstName, lastName, languageCode, sex sql.NullString

	err := row.Scan(
		&u.ID, &u.TelegramID, &username, &firstName, &lastName,
		&u.IsBot, &languageCode, &u.IsPremium, &sex, &u.About,
		&u.State, &u.TimeRanges, &u.IsAdmin,
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

	return &u, nil
}

func scanUserFromRows(rows *sql.Rows) (*domain.User, error) {
	var u domain.User
	var username, firstName, lastName, languageCode, sex sql.NullString

	err := rows.Scan(
		&u.ID, &u.TelegramID, &username, &firstName, &lastName,
		&u.IsBot, &languageCode, &u.IsPremium, &sex, &u.About,
		&u.State, &u.TimeRanges, &u.IsAdmin,
	)
	if err != nil {
		return nil, err
	}

	u.Username = username.String
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.LanguageCode = languageCode.String
	u.Sex = sex.String

	return &u, nil
}
