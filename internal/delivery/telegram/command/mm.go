package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

const mmSticker = "CAACAgIAAxkBAANtaYKDDtR5d1478iPkCrZr2xnZOpMAAgIBAAJWnb0KTuJsgctA5P84BA"

func (h *Handler) MM(c tele.Context) error {
	sticker := &tele.Sticker{File: tele.File{FileID: mmSticker}}
	stickerMsg, err := h.Bot.Send(c.Chat(), sticker)
	if err != nil {
		slog.Error("send sticker", sl.Err(err))
	}

	result, err := h.Matching.RunMatch(context.Background())
	if err != nil {
		if stickerMsg != nil {
			_ = h.Bot.Delete(stickerMsg)
		}
		slog.Error("run match", sl.Err(err))
		return c.Send(messages.M.Matching.Errors.NotEnoughUsers)
	}

	if stickerMsg != nil {
		_ = h.Bot.Delete(stickerMsg)
	}

	fullInfo := ""
	if result.FullMatchCount > 0 {
		fullInfo = fmt.Sprintf("\n\nполных совпадений (без общего времени): %d", result.FullMatchCount)
	}

	if err := c.Send(messages.Format(messages.M.Matching.Success.Matched, map[string]string{
		"pairs":     fmt.Sprintf("%d", result.PairsCount),
		"users":     fmt.Sprintf("%d", result.UsersCount),
		"full_info": fullInfo,
	})); err != nil {
		slog.Error("send match result", sl.Err(err))
	}

	meetResult, err := h.Meeting.CreateMeetings(context.Background())
	if err != nil {
		slog.Error("create meetings", sl.Err(err))
		if errors.Is(err, usecase.ErrNoPairs) {
			for _, id := range result.UnmatchedIDs {
				_, sendErr := h.Bot.Send(&tele.User{ID: id}, messages.M.Matching.Success.NotMatched)
				if sendErr != nil {
					slog.Error("send not matched", sl.Err(sendErr), "telegram_id", id)
				}
			}
			return c.Send(messages.M.Matching.Errors.NoPairs)
		}
		if errors.Is(err, usecase.ErrNoPlaces) {
			return c.Send(messages.M.Matching.Errors.NoPlaces)
		}
		return c.Send(fmt.Sprintf("Ошибка при создании встреч: %v", err))
	}

	count := 0

	for _, m := range meetResult.Meetings {
		content := fmt.Sprintf("%s\n%s", messages.M.Meeting.Invite.Message, messages.M.Meeting.Invite.WaitConfirmation)
		message := messages.Format(content, map[string]string{
			"place": m.Place,
			"time":  domain.Timef(m.Time),
		})

		kb := view.MeetingKeyboard(fmt.Sprintf("%d", m.MeetingID))

		_, err := h.Bot.Send(&tele.User{ID: m.DillID}, message, kb)
		if err != nil {
			slog.Error("send meeting to dill", sl.Err(err), "telegram_id", m.DillID)
		}

		_, err = h.Bot.Send(&tele.User{ID: m.DoeID}, message, kb)
		if err != nil {
			slog.Error("send meeting to doe", sl.Err(err), "telegram_id", m.DoeID)
		}

		count++
	}

	for _, fm := range meetResult.FullMatches {
		dillMsg := messages.Format(messages.M.Meeting.Special.FullMatchNoTime, map[string]string{
			"partner_mention": messages.Mention(fm.DoeTelegramID, fm.DoeFirstName),
		})

		doeMsg := messages.Format(messages.M.Meeting.Special.FullMatchNoTime, map[string]string{
			"partner_mention": messages.Mention(fm.DillTelegramID, fm.DillFirstName),
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

	for _, id := range result.UnmatchedIDs {
		_, err := h.Bot.Send(&tele.User{ID: id}, messages.M.Matching.Success.NotMatched)
		if err != nil {
			slog.Error("send not matched", sl.Err(err), "telegram_id", id)
		}
	}

	return c.Send(messages.Format(messages.M.Matching.Success.MeetingsSent, map[string]string{
		"count": fmt.Sprintf("%d", count),
	}))
}
