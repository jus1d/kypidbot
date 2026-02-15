package message

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

func (h *Handler) Text(c tele.Context) error {
	sender := c.Sender()

	state, err := h.Registration.GetState(context.Background(), sender.ID)
	if err != nil {
		slog.Error("get state", sl.Err(err))
		return nil
	}

	switch state {
	case domain.UserStateAwaitingAppearance:
		return h.handleAppearance(c, sender)
	case domain.UserStateAwaitingAbout:
		return h.handleAbout(c, sender)
	case domain.UserStateAwaitingSupport:
		return h.handleSupport(c, sender)
	case domain.UserStateAwaitingFeedback:
		return h.handleFeedback(c, sender)
	}

	return nil
}

func (h *Handler) handleAbout(c tele.Context, sender *tele.User) error {
	if err := h.Registration.SetAbout(context.Background(), sender.ID, c.Text()); err != nil {
		slog.Error("set about", sl.Err(err))
		return nil
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, domain.UserStateAwaitingTime); err != nil {
		slog.Error("set state", sl.Err(err))
		return nil
	}

	binaryStr, err := h.Registration.GetTimeRanges(context.Background(), sender.ID)
	if err != nil {
		slog.Error("get time ranges", sl.Err(err))
		return nil
	}

	selected := domain.BinaryToSet(binaryStr)

	return c.Send(messages.M.Profile.Schedule.Request, view.TimeKeyboard(selected))
}

func (h *Handler) handleAppearance(c tele.Context, sender *tele.User) error {
	meetingID, err := h.Meeting.GetArrivedMeetingID(context.Background(), sender.ID)
	if err != nil || meetingID == 0 {
		slog.Error("get arrived meeting id", sl.Err(err))
		return nil
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, domain.UserStateCompleted); err != nil {
		slog.Error("set state", sl.Err(err))
		return nil
	}

	partnerID, err := h.Meeting.GetPartnerTelegramID(context.Background(), meetingID, sender.ID)
	if err != nil {
		slog.Error("get partner telegram id", sl.Err(err))
		return nil
	}

	if partnerID != 0 {
		kb := view.CantFindKeyboard(fmt.Sprintf("%d", meetingID))
		msg := messages.Format(messages.M.Notifications.ArrivedPartner, map[string]string{
			"description": c.Text(),
		})
		if _, err := h.Bot.Send(&tele.User{ID: partnerID}, msg, kb); err != nil {
			slog.Error("send appearance to partner", sl.Err(err), "partner_id", partnerID)
		}
	}

	return nil
}

func (h *Handler) handleSupport(c tele.Context, sender *tele.User) error {
	if err := h.Registration.SetState(context.Background(), sender.ID, domain.UserStateCompleted); err != nil {
		slog.Error("set state", sl.Err(err))
		return nil
	}

	admins, err := h.Users.GetAdmins(context.Background())
	if err != nil {
		slog.Error("get admins", sl.Err(err))
		return nil
	}

	content := messages.Format(messages.M.Command.Support.Ticket, map[string]string{
		"mention":     messages.Mention(sender.ID, sender.FirstName, sender.Username),
		"description": c.Text(),
	})

	for _, admin := range admins {
		if _, err := h.Bot.Send(&tele.User{ID: admin.TelegramID}, content); err != nil {
			slog.Error("send support to admin", sl.Err(err), "admin_id", admin.TelegramID)
		}
	}

	return c.Send(messages.M.Command.Support.ProblemSent)
}

func (h *Handler) handleFeedback(c tele.Context, sender *tele.User) error {
	if err := h.Feedback.Save(context.Background(), sender.ID, c.Text()); err != nil {
		slog.Error("save feedback", sl.Err(err))
		return nil
	}

	if err := h.Registration.SetState(context.Background(), sender.ID, domain.UserStateCompleted); err != nil {
		slog.Error("set state after feedback", sl.Err(err))
		return nil
	}

	return c.Send(messages.M.Feedback.ThankYou)
}
