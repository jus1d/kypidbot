package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/stickers"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) MatchPairs(c tele.Context) error {
	ctx := context.Background()
	sticker := &tele.Sticker{File: tele.File{FileID: stickers.Thinking}}
	stickerMsg, err := h.Bot.Send(c.Chat(), sticker)
	if err != nil {
		slog.Error("send sticker", sl.Err(err))
	}

	result, err := h.Matching.RunMatch(ctx)
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

	meetResult, err := h.Meeting.CreateMeetings(ctx)
	if err != nil {
		slog.Error("create meetings", sl.Err(err))
		if errors.Is(err, usecase.ErrNoPairs) {
			return c.Send(messages.M.Matching.Errors.NoPairs)
		}
		if errors.Is(err, usecase.ErrNoPlaces) {
			return c.Send(messages.M.Matching.Errors.NoPlaces)
		}
		return c.Send(fmt.Sprintf("Ошибка при создании встреч: %v", err))
	}

	return c.Send(fmt.Sprintf("Пары распределены и встречи созданы: %d обычных, %d полных совпадений, %d без пары",
		len(meetResult.Meetings), len(meetResult.FullMatches), len(result.UnmatchedIDs)))
}
