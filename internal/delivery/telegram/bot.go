package telegram

import (
	"context"
	"log/slog"
	"time"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/config/messages"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/callback"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/command"
	"github.com/jus1d/kypidbot/internal/delivery/telegram/message"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/infrastructure/s3"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/usecase"
	"github.com/jus1d/kypidbot/internal/version"
	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	env          string
	bot          *tele.Bot
	registration *usecase.Registration
	admin        *usecase.Admin
	matching     *usecase.Matching
	meeting      *usecase.Meeting
	users        domain.UserRepository
	userMessages domain.UserMessageRepository
	settings     domain.SettingsRepository
	places       domain.PlaceRepository
	s3           *s3.Client
}

func NewBot(env string, token string, registration *usecase.Registration, admin *usecase.Admin, matching *usecase.Matching, meeting *usecase.Meeting, users domain.UserRepository, userMessages domain.UserMessageRepository, settings domain.SettingsRepository, places domain.PlaceRepository, s3Client *s3.Client) (*Bot, error) {
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
		env:          env,
		bot:          bot,
		registration: registration,
		admin:        admin,
		matching:     matching,
		meeting:      meeting,
		users:        users,
		userMessages: userMessages,
		settings:     settings,
		places:       places,
		s3:           s3Client,
	}, nil
}

func (b *Bot) Setup() {
	cmd := &command.Handler{
		Registration: b.registration,
		Admin:        b.admin,
		Matching:     b.matching,
		Meeting:      b.meeting,
		Settings:     b.settings,
		Places:       b.places,
		Bot:          b.bot,
		S3:           b.s3,
	}

	cb := &callback.Handler{
		Registration: b.registration,
		Admin:        b.admin,
		Meeting:      b.meeting,
		Users:        b.users,
		UserMessages: b.userMessages,
		Bot:          b.bot,
		S3:           b.s3,
	}

	msg := &message.Handler{
		Registration: b.registration,
		Meeting:      b.meeting,
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
	btnArrivedMeeting := tele.Btn{Unique: "arrived_meeting"}
	btnCantFindPartner := tele.Btn{Unique: "cant_find_partner"}
	btnOptOut := tele.Btn{Unique: "opt_out"}
	btnRefreshAdmin := tele.Btn{Unique: "refresh_admin"}

	b.bot.Use(LogUpdates)

	b.bot.Handle("/start", cmd.Start, b.RegistrationGuard)
	b.bot.Handle("/invite", cmd.Invite, b.RegistrationGuard)
	b.bot.Handle("/leaderboard", cmd.Leaderboard)
	b.bot.Handle("/about", cmd.About)
	b.bot.Handle("/support", cmd.Support)

	b.bot.Handle("/matchpairs", cmd.MatchPairs, b.AdminOnly)
	b.bot.Handle("/sendinvites", cmd.SendInvites, b.AdminOnly)
	b.bot.Handle("/drypairs", cmd.DryPairs, b.AdminOnly)
	b.bot.Handle("/promote", cmd.Promote, b.AdminOnly)
	b.bot.Handle("/demote", cmd.Demote, b.AdminOnly)
	b.bot.Handle("/admin", cmd.AdminPanel, b.AdminOnly)
	b.bot.Handle("/remind", cmd.Remind, b.AdminOnly)
	b.bot.Handle("/closeregistration", cmd.CloseRegistration, b.AdminOnly)
	b.bot.Handle("/openregistration", cmd.OpenRegistration, b.AdminOnly)
	b.bot.Handle("/testimages", cmd.TestImages, b.AdminOnly)

	b.bot.Handle(&btnSexMale, cb.Sex, b.RegistrationGuard)
	b.bot.Handle(&btnSexFemale, cb.Sex, b.RegistrationGuard)
	b.bot.Handle(&btnTime, cb.Time, b.RegistrationGuard)
	b.bot.Handle(&btnConfirmTime, cb.ConfirmTime, b.RegistrationGuard)
	b.bot.Handle(&btnResubmit, cb.Resubmit, b.RegistrationGuard)
	b.bot.Handle(&btnConfirmMeeting, cb.ConfirmMeeting)
	b.bot.Handle(&btnCancelMeeting, cb.CancelMeeting)
	b.bot.Handle(&btnCancelSupport, cb.CancelSupport)
	b.bot.Handle(&btnHowItWorks, cb.HowItWorks)
	b.bot.Handle(&btnArrivedMeeting, cb.ArrivedAtMeeting)
	b.bot.Handle(&btnCantFindPartner, cb.CantFindPartner)
	b.bot.Handle(&btnOptOut, cb.OptOut)
	b.bot.Handle(&btnRefreshAdmin, cb.RefreshAdmin, b.AdminOnly)

	b.bot.Handle(tele.OnText, msg.Text, b.RegistrationGuard)
	b.bot.Handle(tele.OnSticker, msg.Sticker, b.AdminOnly)
}

func (b *Bot) Start(ctx context.Context) {
	// notify admins on bot restart in production
	if b.env == config.EnvProduction {
		admins, err := b.users.GetAdmins(ctx)
		if err != nil {
			slog.Error("get admins", sl.Err(err))
		} else {
			for _, admin := range admins {
				content := messages.Format(messages.M.Admin.StartedLog, map[string]string{
					"branch": version.Branch,
					"commit": version.Commit,
				})
				_, _ = b.bot.Send(&tele.User{ID: admin.TelegramID}, content)
			}
		}

	}

	slog.Info("bot: ok", slog.String("username", b.bot.Me.Username))
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}

func (b *Bot) TeleBot() *tele.Bot {
	return b.bot
}
