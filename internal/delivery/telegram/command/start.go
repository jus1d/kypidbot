package command

import (
	"context"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Start(c tele.Context) error {
	sender := c.Sender()

	prevState, err := h.Registration.GetState(context.Background(), sender.ID)
	if err != nil {
		slog.Error("get state", sl.Err(err))
		return nil
	}

	err = h.Registration.SaveUser(context.Background(), &domain.User{
		TelegramID:   sender.ID,
		Username:     sender.Username,
		FirstName:    sender.FirstName,
		LastName:     sender.LastName,
		IsBot:        sender.IsBot,
		LanguageCode: sender.LanguageCode,
		IsPremium:    sender.IsPremium,
	})
	if err != nil {
		slog.Error("save user", sl.Err(err))
		return nil
	}

	if payload := c.Message().Payload; payload != "" && prevState == "start" {
		referrer, err := h.Registration.GetUserByReferralCode(context.Background(), payload)
		if err != nil {
			slog.Error("get referrer by code", sl.Err(err))
		} else if referrer != nil && referrer.TelegramID != sender.ID {
			if err := h.Registration.SetReferrer(context.Background(), sender.ID, referrer.TelegramID); err != nil {
				slog.Error("set referrer", sl.Err(err))
			}
		}
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, "awaiting_sex"); err != nil {
		slog.Error("set state", sl.Err(err))
		return nil
	}

	if err := c.Send(messages.M.Start.Welcome, view.HowItWorksKeyboard()); err != nil {
		return err
	}

	return c.Send(messages.M.Profile.Sex.AskNew, view.SexKeyboard())
}
