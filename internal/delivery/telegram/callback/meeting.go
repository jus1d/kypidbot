package callback

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
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

	both, meeting, err := h.Meeting.ConfirmMeeting(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("confirm meeting", sl.Err(err))
		return c.Respond()
	}
	if meeting == nil {
		return c.Respond()
	}

	_ = c.Respond()

	origmsg := c.Message()

	partnerID, err := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("get partner telegram id", sl.Err(err))
		return nil
	}

	if !both {
		if meeting.PlaceID == nil || meeting.Time == nil {
			slog.Error("meeting data incomplete", "meeting_id", meetingID)
			return nil
		}

		place, err := h.Meeting.GetPlace(context.Background(), *meeting.PlaceID)
		if err != nil {
			slog.Error("get place", sl.Err(err))
			return nil
		}

		content := messages.Format(
			messages.M.Meeting.Invite.Message+"\n"+messages.M.Meeting.Status.Confirmed,
			map[string]string{"place": place.Description, "route": place.Route, "time": domain.Timef(*meeting.Time)},
		)

		cancelkb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))

		if origmsg.Photo != nil {
			if _, err := h.Bot.EditCaption(origmsg, content, cancelkb); err != nil {
				slog.Error("edit photo caption", sl.Err(err))
			}
		} else {
			if _, err := h.Bot.Edit(origmsg, content, cancelkb); err != nil {
				slog.Error("edit confirmation message", sl.Err(err))
			}
		}

		_ = h.UserMessages.StoreMessageID(context.Background(), meetingID, telegramID, "original_msg", origmsg.ID)

		if partnerID != 0 {
			msg, err := h.Bot.Send(&tele.User{ID: partnerID}, messages.M.Meeting.Status.PartnerConfirmed)
			if err != nil {
				slog.Error("send partner confirmed", sl.Err(err), "partner_id", partnerID)
			} else {
				_ = h.UserMessages.StoreMessageID(context.Background(), meetingID, partnerID, "partner_msg", msg.ID)
			}
		}
		return nil
	}

	if meeting.PlaceID != nil && meeting.Time != nil {
		place, _ := h.Meeting.GetPlace(context.Background(), *meeting.PlaceID)
		if place == nil {
			slog.Error("get place", "place_id", *meeting.PlaceID)
			return nil
		}

		finalMessage := messages.Format(messages.M.Meeting.Status.BothConfirmed, map[string]string{
			"place": place.Description,
			"route": place.Route,
			"time":  domain.Timef(*meeting.Time),
		})

		cancelkb := view.CancelKeyboard(fmt.Sprintf("%d", meetingID))

		_ = c.Delete()

		if place.PhotoURL != "" {
			reader, err := h.S3.GetPhoto(context.Background(), place.PhotoURL)
			if err != nil {
				slog.Error("get photo from s3", sl.Err(err))
				if err := c.Send(finalMessage, cancelkb); err != nil {
					slog.Error("send both confirmed to user", sl.Err(err))
				}
			} else {
				defer reader.Close()
				photo := &tele.Photo{File: tele.FromReader(reader), Caption: finalMessage}
				if err := c.Send(photo, cancelkb); err != nil {
					slog.Error("send photo to user", sl.Err(err))
				}
			}
		} else {
			if err := c.Send(finalMessage, cancelkb); err != nil {
				slog.Error("send both confirmed to user", sl.Err(err))
			}
		}

		partnerNotifID, _ := h.UserMessages.GetMessageID(context.Background(), meetingID, telegramID, "partner_msg")
		if partnerNotifID != 0 {
			_ = h.Bot.Delete(&tele.Message{Chat: &tele.Chat{ID: telegramID}, ID: partnerNotifID})
		}

		if partnerID != 0 {
			partnerOriginalID, _ := h.UserMessages.GetMessageID(context.Background(), meetingID, partnerID, "original_msg")
			if partnerOriginalID != 0 {
				_ = h.Bot.Delete(&tele.Message{Chat: &tele.Chat{ID: partnerID}, ID: partnerOriginalID})
			}

			if place.PhotoURL != "" {
				reader, err := h.S3.GetPhoto(context.Background(), place.PhotoURL)
				if err != nil {
					slog.Error("get photo from s3 for partner", sl.Err(err))
					if _, err := h.Bot.Send(&tele.User{ID: partnerID}, finalMessage, cancelkb); err != nil {
						slog.Error("send both confirmed to partner", sl.Err(err))
					}
				} else {
					defer reader.Close()
					photo := &tele.Photo{File: tele.FromReader(reader), Caption: finalMessage}
					if _, err := h.Bot.Send(&tele.User{ID: partnerID}, photo, cancelkb); err != nil {
						slog.Error("send photo to partner", sl.Err(err))
					}
				}
			} else {
				if _, err := h.Bot.Send(&tele.User{ID: partnerID}, finalMessage, cancelkb); err != nil {
					slog.Error("send both confirmed to partner", sl.Err(err))
				}
			}
		}
	}

	return nil
}

func (h *Handler) ArrivedAtMeeting(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		slog.Error("parse meeting id", sl.Err(err), "data", data)
		return c.Respond()
	}

	_ = c.Respond()

	telegramID := c.Sender().ID

	if err := h.Meeting.SetArrived(context.Background(), meetingID, telegramID); err != nil {
		slog.Error("set arrived state", sl.Err(err))
		return nil
	}

	if err := h.Registration.SetState(context.Background(), telegramID, domain.UserStateAwaitingAppearance); err != nil {
		slog.Error("set awaiting_appearance state", sl.Err(err))
		return nil
	}

	_ = c.Delete()
	return c.Send(messages.M.Notifications.ArrivedAsk)
}

func (h *Handler) CantFindPartner(c tele.Context) error {
	data := c.Callback().Data
	meetingID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		slog.Error("parse meeting id", sl.Err(err), "data", data)
		return c.Respond()
	}

	_ = c.Respond()

	telegramID := c.Sender().ID

	bothCantFind, err := h.Meeting.SetCantFind(context.Background(), meetingID, telegramID)
	if err != nil {
		slog.Error("set cant find", sl.Err(err))
		return nil
	}

	if !bothCantFind {
		return c.Send(messages.M.Notifications.CantFindNoted)
	}

	partner, _ := h.Meeting.GetPartner(context.Background(), meetingID, telegramID)
	if partner != nil {
		if err := c.Send(messages.Format(messages.M.Notifications.CantFindBoth, map[string]string{
			"partner_mention": messages.Mention(partner.TelegramID, partner.FirstName, partner.Username),
		})); err != nil {
			slog.Error("send cant_find_both to user", sl.Err(err))
		}

		user, _ := h.Users.GetUser(context.Background(), telegramID)
		if user != nil {
			_, err := h.Bot.Send(&tele.User{ID: partner.TelegramID}, messages.Format(messages.M.Notifications.CantFindBoth, map[string]string{
				"partner_mention": messages.Mention(user.TelegramID, user.FirstName, user.Username),
			}))
			if err != nil {
				slog.Error("send cant_find_both to partner", sl.Err(err), "partner_id", partner.TelegramID)
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

	partner, _ := h.Meeting.GetPartner(context.Background(), meetingID, telegramID)

	partnerMention := "unknown"
	if partner != nil {
		partnerMention = messages.Mention(partner.TelegramID, partner.FirstName, partner.Username)
	}

	if err := h.DeleteAndSend(c, messages.Format(messages.M.Meeting.Status.Cancelled, map[string]string{
		"partner_mention": partnerMention,
	})); err != nil {
		slog.Error("send cancelled message", sl.Err(err))
	}

	if partner != nil {
		user, _ := h.Users.GetUser(context.Background(), telegramID)
		userMention := "unknown"
		if user != nil {
			userMention = messages.Mention(user.TelegramID, user.FirstName, user.Username)
		}

		_, err := h.Bot.Send(&tele.User{ID: partner.TelegramID}, messages.Format(messages.M.Meeting.Status.PartnerCancelled, map[string]string{
			"partner_mention": userMention,
		}))
		if err != nil {
			slog.Error("send partner cancelled", sl.Err(err), "partner_id", partner.TelegramID)
		}
	}

	return nil
}
