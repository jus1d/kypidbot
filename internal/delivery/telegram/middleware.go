package telegram

import (
	"context"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

func (b *Bot) RegistrationGuard(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		ctx := context.Background()

		value, _ := b.settings.Get(ctx, "registration_closed")
		closed := value == "true"
		if !closed {
			return next(c)
		}

		isAdmin, _ := b.users.IsAdmin(ctx, c.Sender().ID)
		if isAdmin {
			return next(c)
		}

		user, err := b.users.GetUser(ctx, c.Sender().ID)
		if err == nil && user != nil {
			switch user.State {
			case domain.UserStateAwaitingAppearance, domain.UserStateAwaitingSupport:
				return next(c)
			case domain.UserStateCompleted:
				return c.Send(messages.M.Registration.ClosedRegistered)
			}
		}

		return c.Send(messages.M.Registration.Closed)
	}
}

func (b *Bot) AdminOnly(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		isAdmin, err := b.users.IsAdmin(context.Background(), c.Sender().ID)
		if err != nil || !isAdmin {
			return nil
		}
		return next(c)
	}
}

func LogUpdates(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		sender := c.Sender()
		log := slog.With(
			slog.Int64("telegram_id", sender.ID),
			slog.String("username", sender.Username),
		)

		switch {
		case c.Callback() != nil:
			cb := c.Callback()
			log.Debug("updated recieved", slog.String("kind", "callback"), slog.String("unique", cb.Unique), slog.String("data", cb.Data))
		case c.Message() != nil:
			msg := c.Message()
			switch {
			case msg.Text != "" && msg.Text[0] == '/':
				log.Debug("update recieved", slog.String("kind", "command"), slog.String("command", msg.Text))
			case msg.Sticker != nil:
				log.Debug("update recieved", slog.String("kind", "sticker"))
			case msg.Text != "":
				log.Debug("update recieved", slog.String("kind", "text_message"), slog.String("content", msg.Text))
			}
		case c.Update().MessageReaction != nil:
			log.Debug("update recieved", slog.String("kind", "reaction"))
		}

		return next(c)
	}
}
