package domain

import "context"

type Place struct {
	ID          int64
	Description string
	PhotoURL    string
	Route       string
	Quality     int
}

type PlaceRepository interface {
	SavePlace(ctx context.Context, description string) error
	GetAllPlaces(ctx context.Context) ([]Place, error)
	GetPlace(ctx context.Context, placeID int64) (*Place, error)
}
