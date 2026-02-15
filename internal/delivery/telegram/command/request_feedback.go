package command

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) RequestFeedback(c tele.Context) error {
	ctx := context.Background()

	telegramIDs, err := h.Meeting.GetTelegramIDsForFeedbackRequest(ctx)
	if err != nil {
		slog.Error("get telegram ids for feedback request", sl.Err(err))
		return c.Send("Ошибка при получении списка пользователей")
	}

	if len(telegramIDs) == 0 {
		return c.Send("Нет подходящих пар (оба подтвердили и хотя бы один отметился на месте).")
	}

	feedbackMsg := messages.M.Feedback.Request

	sent := 0
	for _, id := range telegramIDs {
		if err := h.Registration.SetState(ctx, id, domain.UserStateAwaitingFeedback); err != nil {
			slog.Error("set state awaiting_feedback", sl.Err(err), "telegram_id", id)
			continue
		}
		if _, err := h.Bot.Send(&tele.User{ID: id}, feedbackMsg); err != nil {
			slog.Error("send feedback request", sl.Err(err), "telegram_id", id)
			continue
		}
		sent++
	}

	return c.Send(fmt.Sprintf("Запрос отзыва отправлен %d пользователям.", sent))
}
