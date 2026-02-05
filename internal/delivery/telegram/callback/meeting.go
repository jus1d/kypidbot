package callback

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) ConfirmMeeting(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		h.Log.Error("parse meeting id", "err", err, "data", data)
		return c.Respond()
	}

	telegramID := c.Sender().ID

	ok, err := h.Meeting.ConfirmMeeting(context.Background(), meetingID, telegramID)
	if err != nil {
		h.Log.Error("confirm meeting", "err", err)
		return c.Respond()
	}
	if !ok {
		return c.Respond()
	}

	kb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))
	originalText := c.Message().Text
	newText := originalText + "\n\n" + view.Msg("meet", "confirmed")
	if err := c.Edit(newText, kb); err != nil {
		h.Log.Error("edit message", "err", err)
	}

	partnerID, err := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, telegramID)
	if err != nil {
		h.Log.Error("get partner telegram id", "err", err)
		return nil
	}

	if partnerID != 0 {
		_, err := h.Bot.Send(&tele.User{ID: partnerID}, view.Msg("meet", "partner_confirmed"))
		if err != nil {
			h.Log.Error("send partner confirmed", "err", err, "partner_id", partnerID)
		}
	}

	both, meeting, err := h.Meeting.BothConfirmed(context.Background(), meetingID)
	if err != nil {
		h.Log.Error("check both confirmed", "err", err)
		return nil
	}

	if both && meeting != nil {
		placeDesc, _ := h.Meeting.GetPlaceDescription(context.Background(), meeting.PlaceID)

		finalMessage := view.Msgf(map[string]string{
			"place": placeDesc,
			"time":  meeting.Time,
		}, "meet", "both_confirmed")

		cancelKb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))

		_, err := h.Bot.Send(&tele.User{ID: telegramID}, finalMessage, cancelKb)
		if err != nil {
			h.Log.Error("send both confirmed to user", "err", err)
		}

		if partnerID != 0 {
			_, err := h.Bot.Send(&tele.User{ID: partnerID}, finalMessage, cancelKb)
			if err != nil {
				h.Log.Error("send both confirmed to partner", "err", err)
			}
		}
	}

	return nil
}

func (h *Handler) CancelMeeting(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		h.Log.Error("parse meeting id", "err", err, "data", data)
		return c.Respond()
	}

	telegramID := c.Sender().ID

	ok, err := h.Meeting.CancelMeeting(context.Background(), meetingID, telegramID)
	if err != nil {
		h.Log.Error("cancel meeting", "err", err)
		return c.Respond()
	}
	if !ok {
		return c.Respond()
	}

	partnerUsername, _ := h.Meeting.GetPartnerUsername(context.Background(), meetingID, telegramID)
	if partnerUsername == "" {
		partnerUsername = "unknown"
	}

	if err := c.Edit(view.Msgf(map[string]string{
		"partner_username": partnerUsername,
	}, "meet", "cancelled")); err != nil {
		h.Log.Error("edit message", "err", err)
	}

	userUsername, _ := h.Users.GetUserUsername(context.Background(), telegramID)
	if userUsername == "" {
		userUsername = "unknown"
	}

	partnerID, _ := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, telegramID)
	if partnerID != 0 {
		_, err := h.Bot.Send(&tele.User{ID: partnerID}, view.Msgf(map[string]string{
			"partner_username": userUsername,
		}, "meet", "partner_cancelled"))
		if err != nil {
			h.Log.Error("send partner cancelled", "err", err, "partner_id", partnerID)
		}
	}

	return nil
}
