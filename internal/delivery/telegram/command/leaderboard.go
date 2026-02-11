package command

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) Leaderboard(c tele.Context) error {
	sender := c.Sender()

	leaderboard, err := h.Registration.GetReferralLeaderboard(context.Background())
	if err != nil {
		slog.Error("failed to get referral leaderboard", sl.Err(err))
		return c.Send(messages.M.Command.Leaderboard.Error)
	}

	if len(leaderboard) == 0 {
		return c.Send(messages.M.Command.Leaderboard.Empty)
	}

	var userPosition int
	userInTop := false

	for i, entry := range leaderboard {
		if entry.ReferrerID == sender.ID {
			userPosition = i + 1
			userInTop = i < 10
			break
		}
	}

	if userPosition == 0 {
		userInTop = false
	}

	var sb strings.Builder
	sb.WriteString(messages.M.Command.Leaderboard.Title)
	sb.WriteString("\n\n")

	medals := map[int]string{0: " 🥇", 1: " 🥈", 2: " 🥉"}

	for i := 0; i < len(leaderboard) && i < 10; i++ {
		entry := leaderboard[i]
		displayName := formatDisplayName(entry.ReferrerID, entry.Username, entry.FirstName)
		medal := medals[i]

		s := fmt.Sprintf("%d. %s -- %d%s\n", i+1, displayName, entry.ReferralCount, medal)
		sb.WriteString(s)
	}

	if !userInTop && userPosition > 0 && userPosition <= len(leaderboard) {
		sb.WriteString("\n")

		userDisplayName := formatDisplayName(sender.ID, sender.Username, sender.FirstName)

		s := fmt.Sprintf("%d. %s -- %d\n", userPosition, userDisplayName, leaderboard[userPosition-1].ReferralCount)
		sb.WriteString(s)
	}

	sb.WriteString("\n" + messages.M.Command.Leaderboard.Footer)

	return c.Send(sb.String())
}

func formatDisplayName(userID int64, username, firstName string) string {
	if username != "" {
		return "@" + username
	}

	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, userID, firstName)
}
