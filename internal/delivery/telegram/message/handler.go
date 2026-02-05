package message

import (
	"log/slog"

	"github.com/jus1d/kypidbot/internal/usecase"
)

type Handler struct {
	Registration *usecase.Registration
	Log          *slog.Logger
}
