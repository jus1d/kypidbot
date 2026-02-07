package command

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

const (
	referralCodeLen     = 8
	referralCodeCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
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
		code, err = generateReferralCode()
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
	text := messages.Format(messages.M.Command.Invite, map[string]string{"link": link})

	return c.Send(text)
}

func generateReferralCode() (string, error) {
	b := make([]byte, referralCodeLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(referralCodeCharset))))
		if err != nil {
			return "", err
		}
		b[i] = referralCodeCharset[n.Int64()]
	}
	return string(b), nil
}
