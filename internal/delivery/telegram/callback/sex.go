package callback

import (
	"context"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Sex(c tele.Context) error {
	sender := c.Sender()
	cb := c.Callback()

	sex := "female"
	if cb.Unique == "sex_male" {
		sex = "male"
	}

	if err := h.Registration.SetSex(context.Background(), sender.ID, sex); err != nil {
		h.Log.Error("set sex", "err", err)
		return c.Respond()
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, "awaiting_about"); err != nil {
		h.Log.Error("set state", "err", err)
		return c.Respond()
	}

	return c.Edit(view.Msg("sex_selected"))
}
