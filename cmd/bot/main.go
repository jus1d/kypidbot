package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/delivery/telegram"
	"github.com/jus1d/kypidbot/internal/lib/logger/daily"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/notifications"
	"github.com/jus1d/kypidbot/internal/repository/postgres"
	"github.com/jus1d/kypidbot/internal/usecase"
)

func main() {
	c := config.MustLoad()

	var level slog.Level
	switch c.Env {
	case config.EnvProduction:
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}

	writer := daily.NewLogsWriter("logs", c.Env)
	multi := io.MultiWriter(os.Stdout, writer)
	handler := slog.NewJSONHandler(multi, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))

	slog.Info("starting...", slog.String("env", c.Env))

	db, err := postgres.New(&c.Postgres)
	if err != nil {
		slog.Error("postgresql: failed to connect", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("postgresql: ok")

	userRepo := postgres.NewUserRepo(db)
	placeRepo := postgres.NewPlaceRepo(db)
	meetingRepo := postgres.NewMeetingRepo(db)

	registration := usecase.NewRegistration(userRepo)
	admin := usecase.NewAdmin(userRepo)
	matching := usecase.NewMatching(userRepo, meetingRepo, &c.Ollama)
	meeting := usecase.NewMeeting(userRepo, placeRepo, meetingRepo)

	bot, err := telegram.NewBot(
		c.Bot.Token,
		registration,
		admin,
		matching,
		meeting,
		userRepo,
	)
	if err != nil {
		slog.Error("failed to create the bot", sl.Err(err))
		os.Exit(1)
	}

	bot.Setup()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notificator := notifications.New(&c.Notifications, bot.TeleBot(), userRepo, placeRepo, meetingRepo)
	notificator.Register(notificator.MeetingReminder)
	notificator.Register(notificator.RegisterReminder)
	notificator.Register(notificator.InviteReminder)

	go notificator.Run(ctx)
	go bot.Start()
	slog.Info("notifications: ok", slog.String("poll_interval", c.Notifications.PollInterval.String()))

	<-stop
	slog.Info("bot: shutting down...")
	bot.Stop()
}
