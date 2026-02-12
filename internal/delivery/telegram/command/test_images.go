package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"

	"github.com/jus1d/kypidbot/internal/lib/logger/sl"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) TestImages(c tele.Context) error {
	ctx := context.Background()

	places, err := h.Places.GetAllPlaces(ctx)
	if err != nil {
		slog.Error("get all places", sl.Err(err))
		return c.Send("Ошибка при получении мест")
	}

	if len(places) == 0 {
		return c.Send("Нет мест в базе")
	}

	rand.Shuffle(len(places), func(i, j int) {
		places[i], places[j] = places[j], places[i]
	})

	count := 3
	if len(places) < count {
		count = len(places)
	}

	for _, p := range places[:count] {
		caption := fmt.Sprintf("<b>%s</b>\n\nКак добраться: %s", p.Description, p.Route)

		if p.PhotoURL == "" {
			if err := c.Send(caption); err != nil {
				slog.Error("send place text", sl.Err(err))
			}
			continue
		}

		reader, err := h.S3.GetPhoto(ctx, p.PhotoURL)
		if err != nil {
			slog.Error("get photo from s3", sl.Err(err))
			if err := c.Send(caption); err != nil {
				slog.Error("send place text fallback", sl.Err(err))
			}
			continue
		}

		photoBytes, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			slog.Error("read photo bytes", sl.Err(err))
			if err := c.Send(caption); err != nil {
				slog.Error("send place text fallback", sl.Err(err))
			}
			continue
		}

		photo := &tele.Photo{File: tele.FromReader(bytes.NewReader(photoBytes)), Caption: caption}
		if err := c.Send(photo); err != nil {
			slog.Error("send place photo", sl.Err(err))
		}
	}

	return nil
}
