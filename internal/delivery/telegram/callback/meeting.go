package callback

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) ConfirmMeeting(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		slog.Error("parse meeting id", sl.Err(err), "data", data)
		return c.Respond()
	}

	telegramID := c.Sender().ID

	ok, err := h.Meeting.ConfirmMeeting(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("confirm meeting", sl.Err(err))
		return c.Respond()
	}
	if !ok {
		return c.Respond()
	}

	_ = c.Respond()

	origmsg := c.Message()

	partnerID, err := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("get partner telegram id", sl.Err(err))
		return nil
	}

	both, meeting, err := h.Meeting.BothConfirmed(context.Background(), meetingID)
	if err != nil {
		slog.Error("check both confirmed", sl.Err(err))
		return nil
	}

	if !both {
		place := ""
		timeStr := ""
		if meeting != nil && meeting.PlaceID != nil && meeting.Time != nil {
			place, err = h.Meeting.GetPlaceDescription(context.Background(), *meeting.PlaceID)
			slog.Error("get place description", sl.Err(err))
			timeStr = *meeting.Time
		}

		content := messages.Format(
			messages.M.Meeting.Invite.Message+"\n"+messages.M.Meeting.Status.Confirmed,
			map[string]string{"place": place, "time": timeStr},
		)

		cancelkb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))
		if _, err := h.Bot.Edit(origmsg, content, cancelkb); err != nil {
			slog.Error("edit confirmation message", sl.Err(err))
		}

		h.storeMessageID(fmt.Sprintf("original_msg_%d_%d", meetingID, telegramID), origmsg.ID)

		if partnerID != 0 {
			msg, err := h.Bot.Send(&tele.User{ID: partnerID}, messages.M.Meeting.Status.PartnerConfirmed)
			if err != nil {
				slog.Error("send partner confirmed", sl.Err(err), "partner_id", partnerID)
			} else {
				h.storeMessageID(fmt.Sprintf("partner_msg_%d_%d", meetingID, partnerID), msg.ID)
			}
		}
		return nil
	}

	if meeting != nil && meeting.PlaceID != nil && meeting.Time != nil {
		place, _ := h.Meeting.GetPlaceDescription(context.Background(), *meeting.PlaceID)

		finalMessage := messages.Format(messages.M.Meeting.Status.BothConfirmed, map[string]string{
			"place": place,
			"time":  *meeting.Time,
		})

		cancelkb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))

		if err := h.DeleteAndSend(c, finalMessage, cancelkb); err != nil {
			slog.Error("send both confirmed to user", sl.Err(err))
		}

		partnerNotifID := h.getMessageID(fmt.Sprintf("partner_msg_%d_%d", meetingID, telegramID))
		if partnerNotifID != 0 {
			_ = h.Bot.Delete(&tele.Message{Chat: &tele.Chat{ID: telegramID}, ID: partnerNotifID})
		}

		if partnerID != 0 {
			partnerOriginalID := h.getMessageID(fmt.Sprintf("original_msg_%d_%d", meetingID, partnerID))
			if partnerOriginalID != 0 {
				_ = h.Bot.Delete(&tele.Message{Chat: &tele.Chat{ID: partnerID}, ID: partnerOriginalID})
			}

			_, err := h.Bot.Send(&tele.User{ID: partnerID}, finalMessage, cancelkb)
			if err != nil {
				slog.Error("send both confirmed to partner", sl.Err(err))
			}
		}
	}

	return nil
}

func (h *Handler) CancelMeeting(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		slog.Error("parse meeting id", sl.Err(err), "data", data)
		return c.Respond()
	}

	telegramID := c.Sender().ID

	ok, err := h.Meeting.CancelMeeting(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("cancel meeting", sl.Err(err))
		return c.Respond()
	}
	if !ok {
		return c.Respond()
	}

	partnerUsername, _ := h.Meeting.GetPartnerUsername(context.Background(), meetingID, telegramID)
	if partnerUsername == "" {
		partnerUsername = "unknown"
	}

	if err := h.DeleteAndSend(c, messages.Format(messages.M.Meeting.Status.Cancelled, map[string]string{
		"partner_username": partnerUsername,
	})); err != nil {
		slog.Error("send cancelled message", sl.Err(err))
	}

	userUsername, _ := h.Users.GetUserUsername(context.Background(), telegramID)
	if userUsername == "" {
		userUsername = "unknown"
	}

	partnerID, _ := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, telegramID)
	if partnerID != 0 {
		_, err := h.Bot.Send(&tele.User{ID: partnerID}, messages.Format(messages.M.Meeting.Status.PartnerCancelled, map[string]string{
			"partner_username": userUsername,
		}))
		if err != nil {
			slog.Error("send partner cancelled", sl.Err(err), "partner_id", partnerID)
		}
	}

	return nil
}
