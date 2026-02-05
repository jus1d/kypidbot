package message

import (
	"context"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Text(c tele.Context) error {
	sender := c.Sender()

	state, err := h.Registration.GetState(context.Background(), sender.ID)
	if err != nil {
		h.Log.Error("get state", "err", err)
		return nil
	}

	if state != "awaiting_about" {
		return nil
	}

	if err := h.Registration.SetAbout(context.Background(), sender.ID, c.Text()); err != nil {
		h.Log.Error("set about", "err", err)
		return nil
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, "awaiting_time"); err != nil {
		h.Log.Error("set state", "err", err)
		return nil
	}

	binaryStr, err := h.Registration.GetTimeRanges(context.Background(), sender.ID)
	if err != nil {
		h.Log.Error("get time ranges", "err", err)
		return nil
	}

	selected := domain.BinaryToSet(binaryStr)

	return c.Send(view.Msg("about_received", "message"), view.TimeKeyboard(selected))
}
