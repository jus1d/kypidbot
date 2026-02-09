package callback

import (
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

type Handler struct {
	Registration *usecase.Registration
	Meeting      *usecase.Meeting
	Users        domain.UserRepository
	UserMessages domain.UserMessageRepository
	Bot          *tele.Bot
}

func (h *Handler) DeleteAndSend(c tele.Context, what any, opts ...any) error {
	_ = c.Delete()
	return c.Send(what, opts...)
}
