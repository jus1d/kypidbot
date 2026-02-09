package command

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Invite(c tele.Context) error {
	sender := c.Sender()

	user, err := h.Registration.GetUser(context.Background(), sender.ID)
	if err != nil {
		slog.Error("get user", sl.Err(err))
		return nil
	}
	if user == nil {
		return nil
	}

	code := user.ReferralCode
	if code == "" {
		code, err = domain.GenerateReferralCode()
		if err != nil {
			slog.Error("generate referral code", sl.Err(err))
			return nil
		}

		if err := h.Registration.SetReferralCode(context.Background(), sender.ID, code); err != nil {
			slog.Error("set referral code", sl.Err(err))
			return nil
		}
	}

	link := fmt.Sprintf("https://t.me/%s?start=%s", h.Bot.Me.Username, code)
	text := messages.Format(messages.M.Notifications.Invite, map[string]string{"link": link})

	return c.Send(text)
}
