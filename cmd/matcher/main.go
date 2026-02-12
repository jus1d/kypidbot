package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/jus1d/kypidbot/internal/domain"
	"github.com/jus1d/kypidbot/internal/infrastructure/ollama"
	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	"github.com/jus1d/kypidbot/internal/matcher"
	"github.com/jus1d/kypidbot/internal/repository/postgres"
)

const placeBuffer = 45 * time.Minute

func hasEarlySlots(intersection string) bool {
	for i := 0; i < 4 && i < len(intersection); i++ {
		if intersection[i] == '1' {
			return true
		}
	}
	return false
}

type outputUser struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	About      string `json:"about"`
}

type outputPair struct {
	Dill  outputUser   `json:"dill"`
	Doe   outputUser   `json:"doe"`
	Score float64      `json:"score"`
	Place *outputPlace `json:"place,omitempty"`
	Time  string       `json:"time,omitempty"`
}

type outputPlace struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Quality     int    `json:"quality"`
}

type output struct {
	Pairs       []outputPair `json:"pairs"`
	FullMatches []outputPair `json:"full_matches"`
	Unmatched   []outputUser `json:"unmatched"`
}

type placeBooking struct {
	placeID int64
	time    time.Time
}

func main() {
	outputPath := flag.String("o", "match-result.json", "output file path")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	c := config.MustLoad()

	db, err := postgres.New(&c.Postgres)
	if err != nil {
		slog.Error("postgresql: failed to connect", sl.Err(err))
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("postgresql: ok")

	ol := ollama.New(&c.Ollama)
	slog.Info("ollama: pulling model...", slog.String("model", c.Ollama.Model))
	if err := ol.PullModel(); err != nil {
		slog.Error("ollama: failed to pull model", sl.Err(err))
		os.Exit(1)
	}
	slog.Info("ollama: ok")

	ctx := context.Background()
	userRepo := postgres.NewUserRepo(db)
	placeRepo := postgres.NewPlaceRepo(db)

	users, err := userRepo.GetVerifiedUsers(ctx)
	if err != nil {
		slog.Error("failed to get users", sl.Err(err))
		os.Exit(1)
	}
	slog.Info("fetched users", slog.Int("count", len(users)))

	if len(users) < 2 {
		slog.Error("not enough users", slog.Int("count", len(users)))
		os.Exit(1)
	}

	places, err := placeRepo.GetAllPlaces(ctx)
	if err != nil {
		slog.Error("failed to get places", sl.Err(err))
		os.Exit(1)
	}
	slog.Info("fetched places", slog.Int("count", len(places)))

	sort.Slice(places, func(i, j int) bool {
		return places[i].Quality > places[j].Quality
	})

	matchUsers := make([]matcher.MatchUser, len(users))
	for i, u := range users {
		matchUsers[i] = matcher.MatchUser{
			Index:      i,
			Username:   u.Username,
			Sex:        u.Sex,
			About:      u.About,
			TimeRanges: u.TimeRanges,
		}
	}

	pairs, fullMatches, err := matcher.Match(matchUsers, ol)
	if err != nil {
		slog.Error("match failed", sl.Err(err))
		os.Exit(1)
	}

	matched := make(map[int]bool)
	for _, p := range pairs {
		matched[p.I] = true
		matched[p.J] = true
	}
	for _, fm := range fullMatches {
		matched[fm.I] = true
		matched[fm.J] = true
	}

	toUser := func(i int) outputUser {
		return outputUser{
			TelegramID: users[i].TelegramID,
			Username:   users[i].Username,
			FirstName:  users[i].FirstName,
			About:      users[i].About,
		}
	}

	loc, err := time.LoadLocation("Europe/Samara")
	if err != nil {
		slog.Error("failed to load location", sl.Err(err))
		os.Exit(1)
	}

	var bookings []placeBooking

	assignPlaceAndTime := func(dillIdx, doeIdx int) (*outputPlace, string) {
		intersection := domain.CalculateTimeIntersection(users[dillIdx].TimeRanges, users[doeIdx].TimeRanges)

		preferred := intersection
		if len(intersection) == 6 && hasEarlySlots(intersection) {
			preferred = intersection[:4] + "00"
		}

		for attempt := 0; attempt < 50; attempt++ {
			src := preferred
			if attempt >= 30 {
				src = intersection
			}
			timeStr := domain.PickRandomTime(src)

			full := fmt.Sprintf("%d-02-14 %s", time.Now().Year(), timeStr)
			meetingTime, err := time.ParseInLocation("2006-01-02 15:04", full, loc)
			if err != nil {
				slog.Warn("failed to parse time", sl.Err(err))
				return nil, ""
			}

			for _, place := range places {
				occupied := false
				for _, b := range bookings {
					if b.placeID == place.ID {
						diff := meetingTime.Sub(b.time)
						if diff < 0 {
							diff = -diff
						}
						if diff < placeBuffer {
							occupied = true
							break
						}
					}
				}
				if !occupied {
					bookings = append(bookings, placeBooking{placeID: place.ID, time: meetingTime})
					return &outputPlace{
						ID:          place.ID,
						Description: place.Description,
						Quality:     place.Quality,
					}, domain.Timef(meetingTime)
				}
			}
		}

		slog.Warn("no available place after 50 attempts")
		timeStr := domain.PickRandomTime(intersection)
		full := fmt.Sprintf("%d-02-14 %s", time.Now().Year(), timeStr)
		meetingTime, _ := time.ParseInLocation("2006-01-02 15:04", full, loc)
		return nil, domain.Timef(meetingTime)
	}

	result := output{
		Pairs:       make([]outputPair, 0, len(pairs)),
		FullMatches: make([]outputPair, 0, len(fullMatches)),
		Unmatched:   make([]outputUser, 0),
	}

	for _, p := range pairs {
		place, timeStr := assignPlaceAndTime(p.I, p.J)
		result.Pairs = append(result.Pairs, outputPair{
			Dill:  toUser(p.I),
			Doe:   toUser(p.J),
			Score: p.Score,
			Place: place,
			Time:  timeStr,
		})
	}

	for _, fm := range fullMatches {
		result.FullMatches = append(result.FullMatches, outputPair{
			Dill:  toUser(fm.I),
			Doe:   toUser(fm.J),
			Score: fm.Score,
		})
	}

	for i := range users {
		if !matched[i] {
			result.Unmatched = append(result.Unmatched, toUser(i))
		}
	}

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		slog.Error("failed to marshal result", sl.Err(err))
		os.Exit(1)
	}

	if err := os.WriteFile(*outputPath, data, 0644); err != nil {
		slog.Error("failed to write output file", sl.Err(err))
		os.Exit(1)
	}

	slog.Info("match completed",
		slog.Int("users", len(users)),
		slog.Int("pairs", len(pairs)),
		slog.Int("full_matches", len(fullMatches)),
		slog.Int("unmatched", len(result.Unmatched)),
		slog.String("output", *outputPath),
	)

	slog.Info("output written", slog.String("path", *outputPath))
}
