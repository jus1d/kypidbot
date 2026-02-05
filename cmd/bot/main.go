package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/delivery/telegram"
	"github.com/jus1d/kypidbot/internal/repository/postgres"
	"github.com/jus1d/kypidbot/internal/usecase"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := config.MustLoad()

	if err := telegram.LoadMessages("messages.yaml"); err != nil {
		log.Error("load messages", "err", err)
		os.Exit(1)
	}

	db, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		log.Error("open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepo(db)
	pairRepo := postgres.NewPairRepo(db)
	placeRepo := postgres.NewPlaceRepo(db)
	meetingRepo := postgres.NewMeetingRepo(db)

	registration := usecase.NewRegistration(userRepo)
	admin := usecase.NewAdmin(userRepo)
	matching := usecase.NewMatching(userRepo, pairRepo, cfg.OllamaURL)
	meeting := usecase.NewMeeting(userRepo, pairRepo, placeRepo, meetingRepo)

	bot, err := telegram.NewBot(
		cfg.TelegramToken,
		registration,
		admin,
		matching,
		meeting,
		userRepo,
		log,
	)
	if err != nil {
		log.Error("create bot", "err", err)
		os.Exit(1)
	}

	bot.Setup()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go bot.Start()

	<-stop
	log.Info("shutting down")
	bot.Stop()
}
