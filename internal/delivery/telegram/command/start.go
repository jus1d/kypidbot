package command

import (
	"context"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Start(c tele.Context) error {
	sender := c.Sender()

	err := h.Registration.SaveUser(context.Background(), &domain.User{
		TelegramID:   sender.ID,
		Username:     sender.Username,
		FirstName:    sender.FirstName,
		LastName:     sender.LastName,
		IsBot:        sender.IsBot,
		LanguageCode: sender.LanguageCode,
		IsPremium:    sender.IsPremium,
	})
	if err != nil {
		h.Log.Error("save user", "err", err)
		return nil
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, "awaiting_sex"); err != nil {
		h.Log.Error("set state", "err", err)
		return nil
	}

	if err := c.Send(view.Msg("start", "welcome")); err != nil {
		return err
	}

	return c.Send(view.Msg("start", "ask_sex", "new"), view.SexKeyboard())
}
