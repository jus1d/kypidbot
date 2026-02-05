package command

import (
	"context"
	"fmt"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	tele "gopkg.in/telebot.v3"
)

const matchSticker = "CAACAgIAAxkBAANtaYKDDtR5d1478iPkCrZr2xnZOpMAAgIBAAJWnb0KTuJsgctA5P84BA"

func (h *Handler) Match(c tele.Context) error {
	sticker := &tele.Sticker{File: tele.File{FileID: matchSticker}}
	stickerMsg, err := h.Bot.Send(c.Chat(), sticker)
	if err != nil {
		h.Log.Error("send sticker", "err", err)
	}

	result, err := h.Matching.RunMatch(context.Background())
	if err != nil {
		if stickerMsg != nil {
			_ = h.Bot.Delete(stickerMsg)
		}
		h.Log.Error("run match", "err", err)
		return c.Send(view.Msg("match", "not_enough_users"))
	}

	if stickerMsg != nil {
		_ = h.Bot.Delete(stickerMsg)
	}

	fullInfo := ""
	if result.FullMatchCount > 0 {
		fullInfo = fmt.Sprintf("\n\nполных совпадений (без общего времени): %d", result.FullMatchCount)
	}

	return c.Send(view.Msgf(map[string]string{
		"pairs":     fmt.Sprintf("%d", result.PairsCount),
		"users":     fmt.Sprintf("%d", result.UsersCount),
		"full_info": fullInfo,
	}, "match", "success"))
}
