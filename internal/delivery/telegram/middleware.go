package telegram

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

func (b *Bot) AdminOnly(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		isAdmin, err := b.users.IsAdmin(context.Background(), c.Sender().ID)
		if err != nil || !isAdmin {
			return nil
		}
		return next(c)
	}
}
