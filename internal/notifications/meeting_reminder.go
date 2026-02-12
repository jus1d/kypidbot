package notifications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (n *Notificator) MeetingReminder(ctx context.Context) error {
	list, err := n.meetings.GetMeetingsStartingIn(ctx, n.config.DateUpcomingIn)
	if err != nil {
		return err
	}

	for _, m := range list {
		log := slog.With(slog.Int64("meeting_id", m.ID))

		if m.UsersNotified {
			continue
		}

		if m.DillState != domain.StateConfirmed || m.DoeState != domain.StateConfirmed {
			continue
		}

		if m.PlaceID == nil || m.Time == nil {
			continue
		}

		dill, err := n.users.GetUser(ctx, m.DillID)
		if err != nil {
			log.Error("notifications: get dill", sl.Err(err))
			continue
		}

		doe, err := n.users.GetUser(ctx, m.DoeID)
		if err != nil {
			log.Error("notifications: get doe", sl.Err(err))
			continue
		}

		if dill == nil || doe == nil {
			continue
		}

		msg := messages.M.Notifications.MeetingSoon

		kb := view.ArrivedKeyboard(fmt.Sprintf("%d", m.ID))

		if _, err := n.bot.Send(&tele.User{ID: dill.TelegramID}, msg, kb); err != nil {
			log.Error("notifications: send to dill", sl.Err(err), slog.Int64("telegram_id", dill.TelegramID))
		}

		if _, err := n.bot.Send(&tele.User{ID: doe.TelegramID}, msg, kb); err != nil {
			log.Error("notifications: send to doe", sl.Err(err), slog.Int64("telegram_id", doe.TelegramID))
		}

		if err := n.meetings.MarkNotified(ctx, m.ID); err != nil {
			log.Error("notifications: mark notified", sl.Err(err))
		}
	}

	return nil
}
