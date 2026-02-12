package command

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Remind(c tele.Context) error {
	users, err := h.Registration.GetUnregisteredUsers(context.Background())
	if err != nil {
		slog.Error("get unregistered users", sl.Err(err))
		return c.Send("Ошибка при получении пользователей")
	}

	if len(users) == 0 {
		return c.Send(messages.M.Command.Remind.NoUsers)
	}

	count := 0
	for _, u := range users {
		if u.Sex != "female" {
			continue
		}

		_, err := h.Bot.Send(&tele.User{ID: u.TelegramID}, messages.M.Notifications.Remind)
		if err != nil {
			slog.Error("send remind", sl.Err(err), "telegram_id", u.TelegramID)
			continue
		}
		count++
	}

	return c.Send(messages.Format(messages.M.Command.Remind.Sent, map[string]string{
		"count": fmt.Sprintf("%d", count),
	}))
}
