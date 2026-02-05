package telegram

import (
	"log/slog"
	"time"

	"github.com/jus1d/kypidbot/internal/delivery/telegram/callback"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/command"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/message"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/view"
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
	log          *slog.Logger
}

func LoadMessages(path string) error {
	return view.LoadMessages(path)
}

func NewBot(
	token string,
	registration *usecase.Registration,
	admin *usecase.Admin,
	matching *usecase.Matching,
	meeting *usecase.Meeting,
	users domain.UserRepository,
	log *slog.Logger,
) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
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
		log:          log,
	}, nil
}

func (b *Bot) Setup() {
	cmd := &command.Handler{
		Registration: b.registration,
		Admin:        b.admin,
		Matching:     b.matching,
		Meeting:      b.meeting,
		Bot:          b.bot,
		Log:          b.log,
	}

	cb := &callback.Handler{
		Registration: b.registration,
		Meeting:      b.meeting,
		Users:        b.users,
		Bot:          b.bot,
		Log:          b.log,
	}

	msg := &message.Handler{
		Registration: b.registration,
		Log:          b.log,
	}

	btnSexMale := tele.Btn{Unique: "sex_male"}
	btnSexFemale := tele.Btn{Unique: "sex_female"}
	btnTime := tele.Btn{Unique: "time"}
	btnConfirmTime := tele.Btn{Unique: "confirm_time"}
	btnResubmit := tele.Btn{Unique: "resubmit"}
	btnConfirmMeeting := tele.Btn{Unique: "confirm_meeting"}
	btnCancelMeeting := tele.Btn{Unique: "cancel_meeting"}

	b.bot.Handle("/start", cmd.Start)
	b.bot.Handle("/match", cmd.Match, b.AdminOnly)
	b.bot.Handle("/meet", cmd.Meet, b.AdminOnly)
	b.bot.Handle("/promote", cmd.Promote, b.AdminOnly)
	b.bot.Handle("/demote", cmd.Demote, b.AdminOnly)

	b.bot.Handle(&btnSexMale, cb.Sex)
	b.bot.Handle(&btnSexFemale, cb.Sex)
	b.bot.Handle(&btnTime, cb.Time)
	b.bot.Handle(&btnConfirmTime, cb.ConfirmTime)
	b.bot.Handle(&btnResubmit, cb.Resubmit)
	b.bot.Handle(&btnConfirmMeeting, cb.ConfirmMeeting)
	b.bot.Handle(&btnCancelMeeting, cb.CancelMeeting)

	b.bot.Handle(tele.OnText, msg.Text)
	b.bot.Handle(tele.OnSticker, msg.Sticker, b.AdminOnly)
}

func (b *Bot) Start() {
	b.log.Info("starting bot")
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}
