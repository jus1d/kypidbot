package command

import (
	"context"
	"errors"
	"strings"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Promote(c tele.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Send(view.Msg("promote", "usage"))
	}

	username := strings.TrimPrefix(args[0], "@")

	err := h.Admin.Promote(context.Background(), username)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return c.Send(view.Msgf(map[string]string{"username": username}, "promote", "user_not_found"))
		case errors.Is(err, usecase.ErrAlreadyAdmin):
			return c.Send(view.Msgf(map[string]string{"username": username}, "promote", "already_admin"))
		default:
			h.Log.Error("promote", "err", err)
			return nil
		}
	}

	return c.Send(view.Msgf(map[string]string{"username": username}, "promote", "success"))
}

func (h *Handler) Demote(c tele.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Send(view.Msg("demote", "usage"))
	}

	username := strings.TrimPrefix(args[0], "@")

	err := h.Admin.Demote(context.Background(), username)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return c.Send(view.Msgf(map[string]string{"username": username}, "demote", "user_not_found"))
		case errors.Is(err, usecase.ErrNotAdmin):
			return c.Send(view.Msgf(map[string]string{"username": username}, "demote", "not_admin"))
		default:
			h.Log.Error("demote", "err", err)
			return nil
		}
	}

	return c.Send(view.Msgf(map[string]string{"username": username}, "demote", "success"))
}
