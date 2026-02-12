package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) SendInvites(c tele.Context) error {
	ctx := context.Background()
	meetResult, err := h.Meeting.GetMeetingsForInvites(ctx)
	if err != nil {
		slog.Error("get meetings for invites", sl.Err(err))
		if errors.Is(err, usecase.ErrNoPairs) {
			return c.Send(messages.M.Matching.Errors.NoPairs)
		}
		return c.Send(fmt.Sprintf("Ошибка: %v", err))
	}

	count := 0

	for _, m := range meetResult.Meetings {
		content := fmt.Sprintf("%s\n%s", messages.M.Meeting.Invite.Message, messages.M.Meeting.Invite.WaitConfirmation)
		message := messages.Format(content, map[string]string{
			"place": m.Place,
			"route": m.Route,
			"time":  domain.Timef(m.Time),
		})

		kb := view.MeetingKeyboard(fmt.Sprintf("%d", m.MeetingID))

		if m.PhotoURL != "" {
			photoReader, err := h.S3.GetPhoto(ctx, m.PhotoURL)
			if err != nil {
				slog.Error("get photo from s3", sl.Err(err))
				_, err := h.Bot.Send(&tele.User{ID: m.DillID}, message, kb)
				if err != nil {
					slog.Error("send meeting to dill", sl.Err(err), "telegram_id", m.DillID)
				}
				_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, message, kb)
				if err != nil {
					slog.Error("send meeting to doe", sl.Err(err), "telegram_id", m.DoeID)
				}
			} else {
				photoBytes, err := io.ReadAll(photoReader)
				photoReader.Close()
				if err != nil {
					slog.Error("read photo bytes", sl.Err(err))
					_, err := h.Bot.Send(&tele.User{ID: m.DillID}, message, kb)
					if err != nil {
						slog.Error("send meeting to dill", sl.Err(err), "telegram_id", m.DillID)
					}
					_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, message, kb)
					if err != nil {
						slog.Error("send meeting to doe", sl.Err(err), "telegram_id", m.DoeID)
					}
				} else {
					dillPhoto := &tele.Photo{File: tele.FromReader(bytes.NewReader(photoBytes)), Caption: message}
					_, err := h.Bot.Send(&tele.User{ID: m.DillID}, dillPhoto, kb)
					if err != nil {
						slog.Error("send photo to dill", sl.Err(err), "telegram_id", m.DillID)
					}

					doePhoto := &tele.Photo{File: tele.FromReader(bytes.NewReader(photoBytes)), Caption: message}
					_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, doePhoto, kb)
					if err != nil {
						slog.Error("send photo to doe", sl.Err(err), "telegram_id", m.DoeID)
					}
				}
			}
		} else {
			_, err := h.Bot.Send(&tele.User{ID: m.DillID}, message, kb)
			if err != nil {
				slog.Error("send meeting to dill", sl.Err(err), "telegram_id", m.DillID)
			}
			_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, message, kb)
			if err != nil {
				slog.Error("send meeting to doe", sl.Err(err), "telegram_id", m.DoeID)
			}
		}

		count++
	}

	for _, fm := range meetResult.FullMatches {
		dillMsg := messages.Format(messages.M.Meeting.Special.FullMatchNoTime, map[string]string{
			"partner_mention": messages.Mention(fm.DoeTelegramID, fm.DoeFirstName, fm.DoeUsername),
		})

		doeMsg := messages.Format(messages.M.Meeting.Special.FullMatchNoTime, map[string]string{
			"partner_mention": messages.Mention(fm.DillTelegramID, fm.DillFirstName, fm.DillUsername),
		})

		_, err := h.Bot.Send(&tele.User{ID: fm.DillTelegramID}, dillMsg)
		if err != nil {
			slog.Error("send full match to dill", sl.Err(err), "telegram_id", fm.DillTelegramID)
		}

		_, err = h.Bot.Send(&tele.User{ID: fm.DoeTelegramID}, doeMsg)
		if err != nil {
			slog.Error("send full match to doe", sl.Err(err), "telegram_id", fm.DoeTelegramID)
		}

		count++
	}

	unmatchedIDs, err := h.Meeting.GetUnmatchedUserIDs(ctx)
	if err != nil {
		slog.Error("get unmatched users", sl.Err(err))
	} else {
		for _, id := range unmatchedIDs {
			_, err := h.Bot.Send(&tele.User{ID: id}, messages.M.Matching.Success.NotMatched)
			if err != nil {
				slog.Error("send not matched", sl.Err(err), "telegram_id", id)
			}
		}
	}

	return c.Send(messages.Format(messages.M.Matching.Success.MeetingsSent, map[string]string{
		"count": fmt.Sprintf("%d", count),
	}))
}
