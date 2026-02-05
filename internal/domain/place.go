package domain

import "context"

type Place struct {
	ID          int64
	Description string
}

type PlaceRepository interface {
	SavePlace(ctx context.Context, description string) error
	GetAllPlaces(ctx context.Context) ([]Place, error)
}
