package callback

import (
	"context"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) ConfirmTime(c tele.Context) error {
	sender := c.Sender()

	if err := h.Registration.SetState(context.Background(), sender.ID, "completed"); err != nil {
		h.Log.Error("set state", "err", err)
		return c.Respond()
	}

	return c.Edit(view.Msg("completed", "message"), view.ResubmitKeyboard())
}

func (h *Handler) Time(c tele.Context) error {
	sender := c.Sender()
	timeRange := c.Callback().Data

	binaryStr, err := h.Registration.GetTimeRanges(context.Background(), sender.ID)
	if err != nil {
		h.Log.Error("get time ranges", "err", err)
		return c.Respond()
	}

	selected := domain.BinaryToSet(binaryStr)

	if selected[timeRange] {
		delete(selected, timeRange)
	} else {
		selected[timeRange] = true
	}

	newBinary := domain.SetToBinary(selected)
	if err := h.Registration.SaveTimeRanges(context.Background(), sender.ID, newBinary); err != nil {
		h.Log.Error("save time ranges", "err", err)
		return c.Respond()
	}

	return c.Edit(view.Msg("about_received", "message"), view.TimeKeyboard(selected))
}
