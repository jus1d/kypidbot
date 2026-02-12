package command

import (
	"context"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) CloseRegistration(c tele.Context) error {
	err := h.Settings.Set(context.Background(), "registration_closed", "true")
	if err != nil {
		slog.Error("failed to close registration", sl.Err(err))
		return c.Send("Ошибка при закрытии регистрации")
	}
	return c.Send(messages.M.Admin.RegistrationClosed)
}

func (h *Handler) OpenRegistration(c tele.Context) error {
	err := h.Settings.Set(context.Background(), "registration_closed", "false")
	if err != nil {
		slog.Error("failed to open registration", sl.Err(err))
		return c.Send("Ошибка при открытии регистрации")
	}
	return c.Send(messages.M.Admin.RegistrationOpened)
}
