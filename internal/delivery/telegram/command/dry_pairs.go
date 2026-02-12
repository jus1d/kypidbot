package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/stickers"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) DryPairs(c tele.Context) error {
	ctx := context.Background()
	sticker := &tele.Sticker{File: tele.File{FileID: stickers.Thinking}}
	stickerMsg, err := h.Bot.Send(c.Chat(), sticker)
	if err != nil {
		slog.Error("send sticker", sl.Err(err))
	}

	pairs, err := h.Matching.DryMatch(ctx)
	if stickerMsg != nil {
		_ = h.Bot.Delete(stickerMsg)
	}
	if err != nil {
		slog.Error("dry match", sl.Err(err))
		return c.Send(messages.M.Command.Pairs.Error)
	}

	if len(pairs) == 0 {
		return c.Send(messages.M.Command.Pairs.NotFound)
	}

	var sb strings.Builder
	for _, p := range pairs {
		sb.WriteString(fmt.Sprintf("%s x %s\n",
			messages.Mention(p.DillTelegramID, p.DillFirstName, p.DillUsername),
			messages.Mention(p.DoeTelegramID, p.DoeFirstName, p.DoeUsername),
		))
	}

	return c.Send(sb.String())
}
