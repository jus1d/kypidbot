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
    
    leaderboard, err := h.Registration.GetReferralLeaderboard(context.Background(), 10)
    if err != nil {
        slog.Error("failed to get referral leaderboard", 
            sl.Err(err), 
            "user_id", sender.ID)
        
        return c.Send(messages.M.Command.Leaderboard.Error)
    }
    
    userReferralCount, err := h.Registration.GetUserReferralCount(context.Background(), sender.ID)
    if err != nil {
        slog.Error("failed to get user referral count", 
            sl.Err(err), 
            "user_id", sender.ID)
        userReferralCount = 0
    }
    
    userPosition, err := h.Registration.GetUserLeaderboardPosition(context.Background(), sender.ID)
    if err != nil {
        slog.Error("failed to get user leaderboard position", 
            sl.Err(err), 
            "user_id", sender.ID)
        userPosition = 0
    }
    
    userInTop := false
    for _, entry := range leaderboard {
        if entry.ReferrerID == sender.ID {
            userInTop = true
            break
        }
    }

	userDisplayName := h.formatDisplayName(sender.ID, sender.Username, sender.FirstName)
    
    var messageBuilder strings.Builder
    
    messageBuilder.WriteString(messages.M.Command.Leaderboard.Title)
    messageBuilder.WriteString("\n\n")
    
    if len(leaderboard) == 0 {
        messageBuilder.WriteString(messages.M.Command.Leaderboard.Empty)
    } else {
        for i, entry := range leaderboard {
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
                "%s %s — %d\n",
                emoji,
                displayName,
                entry.ReferralCount,
            ))
        }
		
        if !userInTop && userPosition > 0 && userReferralCount > 0 {
            messageBuilder.WriteString("\n")
            messageBuilder.WriteString(fmt.Sprintf(
                "%d. %s — %d\n",
                userPosition,
                userDisplayName,
                userReferralCount,
            ))
        }
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