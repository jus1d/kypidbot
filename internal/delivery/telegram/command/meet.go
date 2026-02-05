package command

import (
	"context"
	"fmt"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Meet(c tele.Context) error {
	result, err := h.Meeting.CreateMeetings(context.Background())
	if err != nil {
		h.Log.Error("create meetings", "err", err)
		if err.Error() == "no pairs" {
			return c.Send(view.Msg("meet", "no_pairs"))
		}
		if err.Error() == "no places" {
			return c.Send(view.Msg("meet", "no_places"))
		}
		return nil
	}

	count := 0

	for _, m := range result.Meetings {
		message := view.Msgf(map[string]string{
			"place": m.Place,
			"time":  m.Time,
		}, "meet", "notification")

		kb := view.MeetingKeyboard(fmt.Sprintf("%d", m.MeetingID))

		_, err := h.Bot.Send(&tele.User{ID: m.DillID}, message, kb)
		if err != nil {
			h.Log.Error("send meeting to dill", "err", err, "telegram_id", m.DillID)
		}

		_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, message, kb)
		if err != nil {
			h.Log.Error("send meeting to doe", "err", err, "telegram_id", m.DoeID)
		}

		count++
	}

	for _, fm := range result.FullMatches {
		dillMsg := view.Msgf(map[string]string{
			"partner_username": fm.DoeUsername,
		}, "meet", "full_match")

		doeMsg := view.Msgf(map[string]string{
			"partner_username": fm.DillUsername,
		}, "meet", "full_match")

		_, err := h.Bot.Send(&tele.User{ID: fm.DillTelegramID}, dillMsg)
		if err != nil {
			h.Log.Error("send full match to dill", "err", err, "telegram_id", fm.DillTelegramID)
		}

		_, err = h.Bot.Send(&tele.User{ID: fm.DoeTelegramID}, doeMsg)
		if err != nil {
			h.Log.Error("send full match to doe", "err", err, "telegram_id", fm.DoeTelegramID)
		}

		count++
	}

	return c.Send(view.Msgf(map[string]string{
		"count": fmt.Sprintf("%d", count),
	}, "meet", "success"))
}
