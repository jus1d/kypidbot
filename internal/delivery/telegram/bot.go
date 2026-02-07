package telegram

import (
	"log/slog"
	"time"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/callback"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/command"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/message"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/usecase"
	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	bot          *tele.Bot
	registration *usecase.Registration
	admin        *usecase.Admin
	matching     *usecase.Matching
	meeting      *usecase.Meeting
	users        domain.UserRepository
}

func NewBot(token string, registration *usecase.Registration, admin *usecase.Admin, matching *usecase.Matching, meeting *usecase.Meeting, users domain.UserRepository) (*Bot, error) {
	pref := tele.Settings{
		Token:     token,
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:          bot,
		registration: registration,
		admin:        admin,
		matching:     matching,
		meeting:      meeting,
		users:        users,
	}, nil
}

func (b *Bot) Setup() {
	cmd := &command.Handler{
		Registration: b.registration,
		Admin:        b.admin,
		Matching:     b.matching,
		Meeting:      b.meeting,
		Bot:          b.bot,
	}

	cb := &callback.Handler{
		Registration: b.registration,
		Meeting:      b.meeting,
		Users:        b.users,
		Bot:          b.bot,
	}

	msg := &message.Handler{
		Registration: b.registration,
		Users:        b.users,
		Bot:          b.bot,
	}

	btnSexMale := tele.Btn{Unique: "sex_male"}
	btnSexFemale := tele.Btn{Unique: "sex_female"}
	btnTime := tele.Btn{Unique: "time"}
	btnConfirmTime := tele.Btn{Unique: "confirm_time"}
	btnResubmit := tele.Btn{Unique: "resubmit"}
	btnConfirmMeeting := tele.Btn{Unique: "confirm_meeting"}
	btnCancelMeeting := tele.Btn{Unique: "cancel_meeting"}
	btnCancelSupport := tele.Btn{Unique: "cancel_support"}
	btnHowItWorks := tele.Btn{Unique: "how_it_works"}

	b.bot.Use(LogUpdates)

	b.bot.Handle("/start", cmd.Start)
	b.bot.Handle("/invite", cmd.Invite)
	b.bot.Handle("/mm", cmd.MM, b.AdminOnly)
	b.bot.Handle("/promote", cmd.Promote, b.AdminOnly)
	b.bot.Handle("/demote", cmd.Demote, b.AdminOnly)
	b.bot.Handle("/about", cmd.About)
	b.bot.Handle("/support", cmd.Support)

	b.bot.Handle(&btnSexMale, cb.Sex)
	b.bot.Handle(&btnSexFemale, cb.Sex)
	b.bot.Handle(&btnTime, cb.Time)
	b.bot.Handle(&btnConfirmTime, cb.ConfirmTime)
	b.bot.Handle(&btnResubmit, cb.Resubmit)
	b.bot.Handle(&btnConfirmMeeting, cb.ConfirmMeeting)
	b.bot.Handle(&btnCancelMeeting, cb.CancelMeeting)
	b.bot.Handle(&btnCancelSupport, cb.CancelSupport)
	b.bot.Handle(&btnHowItWorks, cb.HowItWorks)

	b.bot.Handle(tele.OnText, msg.Text)
	b.bot.Handle(tele.OnSticker, msg.Sticker, b.AdminOnly)
}

func (b *Bot) Start() {
	slog.Info("bot: ok")
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}
