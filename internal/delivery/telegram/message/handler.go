package message

import (
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

type Handler struct {
	Registration *usecase.Registration
	Meeting      *usecase.Meeting
	Users        domain.UserRepository
	Feedback     domain.FeedbackRepository
	Bot          *tele.Bot
}
