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
    
    slog.Info("processing leaderboard command", 
        "user_id", sender.ID, 
        "username", sender.Username)
    
    fullLeaderboard, err := h.Registration.GetReferralLeaderboard(context.Background())
    if err != nil {
        slog.Error("failed to get referral leaderboard", 
            sl.Err(err), 
            "user_id", sender.ID)
        
        return c.Send(messages.M.Command.Leaderboard.Error)
    }
    
    if len(fullLeaderboard) == 0 {
        return c.Send(messages.M.Command.Leaderboard.Empty)
    }
    
    var userPosition int
    userInTop := false
    
    for i, entry := range fullLeaderboard {
        if entry.ReferrerID == sender.ID {
            userPosition = i + 1
            userInTop = i < 10
            break
        }
    }
    
    if userPosition == 0 {
        userInTop = false
    }
    
    var messageBuilder strings.Builder
    messageBuilder.WriteString(messages.M.Command.Leaderboard.Title)
    messageBuilder.WriteString("\n\n")
    
    for i := 0; i < len(fullLeaderboard) && i < 10; i++ {
        entry := fullLeaderboard[i]
        
        var emoji string
        switch i {
        case 0:
            emoji = "🥇"
        case 1:
            emoji = "🥈"
        case 2:
            emoji = "🥉"
        default:
            emoji = fmt.Sprintf("%d.", i+1)
        }
        
        displayName := h.formatDisplayName(entry.ReferrerID, entry.Username, entry.FirstName)
        
        messageBuilder.WriteString(fmt.Sprintf(
            "%s %s -- %d\n",
            emoji,
            displayName,
            entry.ReferralCount,
        ))
    }
    
    if !userInTop && userPosition > 0 && userPosition <= len(fullLeaderboard) {
        messageBuilder.WriteString("\n")

        userDisplayName := h.formatDisplayName(sender.ID, sender.Username, sender.FirstName)

        messageBuilder.WriteString(fmt.Sprintf(
            "%d. %s -- %d\n",
            userPosition,
            userDisplayName,
            fullLeaderboard[userPosition-1].ReferralCount,
        ))
    }
    
    messageBuilder.WriteString("\n" + messages.M.Command.Leaderboard.Footer)
    
    return c.Send(messageBuilder.String())
}

func (h *Handler) formatDisplayName(userID int64, username, firstName string) string {
    if username != "" {
        return "@" + username
    } else if firstName != "" {
        return firstName
    }
    return fmt.Sprintf("ID: %d", userID)
}