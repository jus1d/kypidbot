package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Ugrumov(c tele.Context) error {
	ctx := context.Background()

	var dillID int64 = 780074874
	var doeID int64 = 1061574811
	var placeID int64 = 1

	meeting, err := h.Meeting.GetMeetingByUsers(ctx, dillID, doeID)
	if err != nil {
		slog.Error("get meeting by users", sl.Err(err))
		return c.Send(fmt.Sprintf("Ошибка: %v", err))
	}
	if meeting == nil {
		return c.Send("Встреча не найдена")
	}

	// loc, err := time.LoadLocation("Europe/Samara")
	// if err != nil {
	// 	slog.Error("load location", sl.Err(err))
	// 	return c.Send(fmt.Sprintf("Ошибка: %v", err))
	// }

	// meetingTime := time.Date(2026, 2, 14, 20, 0, 0, 0, loc)

	// if meeting.PlaceID == nil || meeting.Time == nil {
	// 	if err := h.Meeting.AssignPlaceAndTime(ctx, meeting.ID, placeID, meetingTime); err != nil {
	// 		slog.Error("assign place and time", sl.Err(err))
	// 		return c.Send(fmt.Sprintf("Ошибка: %v", err))
	// 	}
	// } else {
	// 	placeID = *meeting.PlaceID
	// 	meetingTime = *meeting.Time
	// }

	if meeting.Time == nil {
		return c.Send("Пустой meeting time")
	}

	place, err := h.Places.GetPlace(ctx, placeID)
	if err != nil || place == nil {
		slog.Error("get place", sl.Err(err))
		return c.Send(fmt.Sprintf("Ошибка: %v", err))
	}

	content := fmt.Sprintf("%s\n%s", messages.M.Meeting.Invite.Message, messages.M.Meeting.Invite.WaitConfirmation)
	message := messages.Format(content, map[string]string{
		"place": place.Description,
		"route": place.Route,
		"time":  domain.Timef(*meeting.Time),
	})

	kb := view.MeetingKeyboard(fmt.Sprintf("%d", meeting.ID))

	if place.PhotoURL != "" {
		photoReader, err := h.S3.GetPhoto(ctx, place.PhotoURL)
		if err != nil {
			slog.Error("get photo from s3", sl.Err(err))
			h.sendTextInvite(dillID, doeID, message, kb)
		} else {
			photoBytes, err := io.ReadAll(photoReader)
			photoReader.Close()
			if err != nil {
				slog.Error("read photo bytes", sl.Err(err))
				h.sendTextInvite(dillID, doeID, message, kb)
			} else {
				h.sendPhotoInvite(dillID, doeID, photoBytes, message, kb)
			}
		}
	} else {
		h.sendTextInvite(dillID, doeID, message, kb)
	}

	return c.Send("Приглашения отправлены для пары @booarr + @lizaagerasimova")
}

func (h *Handler) sendTextInvite(dillID, doeID int64, message string, kb *tele.ReplyMarkup) {
	if _, err := h.Bot.Send(&tele.User{ID: dillID}, message, kb); err != nil {
		slog.Error("send meeting to dill", sl.Err(err), "telegram_id", dillID)
	}
	if _, err := h.Bot.Send(&tele.User{ID: doeID}, message, kb); err != nil {
		slog.Error("send meeting to doe", sl.Err(err), "telegram_id", doeID)
	}
}

func (h *Handler) sendPhotoInvite(dillID, doeID int64, photoBytes []byte, message string, kb *tele.ReplyMarkup) {
	dillPhoto := &tele.Photo{File: tele.FromReader(bytes.NewReader(photoBytes)), Caption: message}
	if _, err := h.Bot.Send(&tele.User{ID: dillID}, dillPhoto, kb); err != nil {
		slog.Error("send photo to dill", sl.Err(err), "telegram_id", dillID)
	}

	doePhoto := &tele.Photo{File: tele.FromReader(bytes.NewReader(photoBytes)), Caption: message}
	if _, err := h.Bot.Send(&tele.User{ID: doeID}, doePhoto, kb); err != nil {
		slog.Error("send photo to doe", sl.Err(err), "telegram_id", doeID)
	}
}
