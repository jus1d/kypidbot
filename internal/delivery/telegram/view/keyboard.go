package view

import (
	"github.com/jus1d/kypidbot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

func SexKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	male := menu.Data(Msg("buttons", "sex", "male"), "sex_male")
	female := menu.Data(Msg("buttons", "sex", "female"), "sex_female")

	menu.Inline(
		menu.Row(male, female),
	)
	return menu
}

func TimeKeyboard(selected map[string]bool) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	var rows []tele.Row
	for i := 0; i < len(domain.TimeRanges); i += 2 {
		var row tele.Row
		for _, tr := range domain.TimeRanges[i:min(i+2, len(domain.TimeRanges))] {
			text := tr
			if selected[tr] {
				text = "> " + tr + " <"
			}
			btn := menu.Data(text, "time", tr)
			row = append(row, btn)
		}
		rows = append(rows, row)
	}

	if len(selected) > 0 {
		confirm := menu.Data(Msg("buttons", "confirm"), "confirm_time")
		rows = append(rows, menu.Row(confirm))
	}

	menu.Inline(rows...)
	return menu
}

func ResubmitKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	btn := menu.Data(Msg("buttons", "resubmit"), "resubmit")
	menu.Inline(menu.Row(btn))
	return menu
}

func MeetingKeyboard(meetingID string) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	confirm := menu.Data(Msg("meet", "confirm_button"), "confirm_meeting", meetingID)
	cancel := menu.Data(Msg("meet", "cancel_button"), "cancel_meeting", meetingID)

	menu.Inline(menu.Row(confirm, cancel))
	return menu
}

func CancelKeyboard(meetingID string) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	cancel := menu.Data(Msg("meet", "cancel_button"), "cancel_meeting", meetingID)
	menu.Inline(menu.Row(cancel))
	return menu
}
