package postgres

import (
	"context"
	"database/sql"

	"github.com/jus1d/kypidbot/internal/domain"
)

type PlaceRepo struct {
	db *sql.DB
}

func NewPlaceRepo(d *DB) *PlaceRepo {
	return &PlaceRepo{db: d.db}
}

func (r *PlaceRepo) SavePlace(ctx context.Context, description string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO places (description) VALUES ($1)`, description)
	return err
}

func (r *PlaceRepo) GetPlace(ctx context.Context, placeID int64) (*domain.Place, error) {
	var p domain.Place
	err := r.db.QueryRowContext(ctx,
		`SELECT id, description, photo_url, route, quality FROM places WHERE id = $1`,
		placeID).Scan(&p.ID, &p.Description, &p.PhotoURL, &p.Route, &p.Quality)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlaceRepo) GetAllPlaces(ctx context.Context) ([]domain.Place, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, description, photo_url, route, quality FROM places ORDER BY quality DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []domain.Place
	for rows.Next() {
		var p domain.Place
		if err := rows.Scan(&p.ID, &p.Description, &p.PhotoURL, &p.Route, &p.Quality); err != nil {
			return nil, err
		}
		places = append(places, p)
	}
	return places, rows.Err()
}
