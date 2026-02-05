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

func (r *PlaceRepo) GetAllPlaces(ctx context.Context) ([]domain.Place, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, description FROM places`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []domain.Place
	for rows.Next() {
		var p domain.Place
		if err := rows.Scan(&p.ID, &p.Description); err != nil {
			return nil, err
		}
		places = append(places, p)
	}
	return places, rows.Err()
}
