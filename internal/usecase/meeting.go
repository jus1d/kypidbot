package usecase

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jus1d/kypidbot/internal/domain"
)

type MeetingNotification struct {
	MeetingID int64
	DillID    int64
	DoeID     int64
	Place     string
	Time      string
}

type FullMatchNotification struct {
	DillTelegramID int64
	DoeTelegramID  int64
	DillUsername    string
	DoeUsername     string
}

type MeetResult struct {
	Meetings    []MeetingNotification
	FullMatches []FullMatchNotification
}

type Meeting struct {
	users    domain.UserRepository
	pairs    domain.PairRepository
	places   domain.PlaceRepository
	meetings domain.MeetingRepository
}

func NewMeeting(
	users domain.UserRepository,
	pairs domain.PairRepository,
	places domain.PlaceRepository,
	meetings domain.MeetingRepository,
) *Meeting {
	return &Meeting{
		users:    users,
		pairs:    pairs,
		places:   places,
		meetings: meetings,
	}
}

func (m *Meeting) CreateMeetings(ctx context.Context) (*MeetResult, error) {
	regularPairs, err := m.pairs.GetRegularPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get regular pairs: %w", err)
	}

	fullPairs, err := m.pairs.GetFullPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get full pairs: %w", err)
	}

	if len(regularPairs) == 0 && len(fullPairs) == 0 {
		return nil, fmt.Errorf("no pairs")
	}

	places, err := m.places.GetAllPlaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("get places: %w", err)
	}

	if len(places) == 0 && len(regularPairs) > 0 {
		return nil, fmt.Errorf("no places")
	}

	var result MeetResult

	for _, pair := range regularPairs {
		place := places[rand.Intn(len(places))]
		meetingTime := domain.PickRandomTime(pair.TimeIntersection)

		meetingID, err := m.meetings.SaveMeeting(ctx, &domain.Meeting{
			PairID:  pair.ID,
			PlaceID: place.ID,
			Time:    meetingTime,
		})
		if err != nil {
			return nil, fmt.Errorf("save meeting: %w", err)
		}

		dill, err := m.users.GetUserByID(ctx, pair.DillID)
		if err != nil {
			return nil, fmt.Errorf("get dill: %w", err)
		}
		doe, err := m.users.GetUserByID(ctx, pair.DoeID)
		if err != nil {
			return nil, fmt.Errorf("get doe: %w", err)
		}

		if dill != nil && doe != nil {
			result.Meetings = append(result.Meetings, MeetingNotification{
				MeetingID: meetingID,
				DillID:    dill.TelegramID,
				DoeID:     doe.TelegramID,
				Place:     place.Description,
				Time:      meetingTime,
			})
		}
	}

	for _, pair := range fullPairs {
		dill, err := m.users.GetUserByID(ctx, pair.DillID)
		if err != nil {
			return nil, fmt.Errorf("get dill: %w", err)
		}
		doe, err := m.users.GetUserByID(ctx, pair.DoeID)
		if err != nil {
			return nil, fmt.Errorf("get doe: %w", err)
		}

		if dill != nil && doe != nil {
			result.FullMatches = append(result.FullMatches, FullMatchNotification{
				DillTelegramID: dill.TelegramID,
				DoeTelegramID:  doe.TelegramID,
				DillUsername:    dill.Username,
				DoeUsername:     doe.Username,
			})
		}
	}

	return &result, nil
}

func (m *Meeting) ConfirmMeeting(ctx context.Context, meetingID int64, telegramID int64) (bool, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, err
	}

	pair, err := m.pairs.GetPairByID(ctx, meeting.PairID)
	if err != nil || pair == nil {
		return false, err
	}

	dill, err := m.users.GetUserByID(ctx, pair.DillID)
	if err != nil {
		return false, err
	}
	doe, err := m.users.GetUserByID(ctx, pair.DoeID)
	if err != nil {
		return false, err
	}

	if dill != nil && dill.TelegramID == telegramID {
		return true, m.meetings.ConfirmMeeting(ctx, meetingID, true)
	}
	if doe != nil && doe.TelegramID == telegramID {
		return true, m.meetings.ConfirmMeeting(ctx, meetingID, false)
	}

	return false, nil
}

func (m *Meeting) CancelMeeting(ctx context.Context, meetingID int64, telegramID int64) (bool, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, err
	}

	pair, err := m.pairs.GetPairByID(ctx, meeting.PairID)
	if err != nil || pair == nil {
		return false, err
	}

	dill, err := m.users.GetUserByID(ctx, pair.DillID)
	if err != nil {
		return false, err
	}
	doe, err := m.users.GetUserByID(ctx, pair.DoeID)
	if err != nil {
		return false, err
	}

	if dill != nil && dill.TelegramID == telegramID {
		return true, m.meetings.CancelMeeting(ctx, meetingID, true)
	}
	if doe != nil && doe.TelegramID == telegramID {
		return true, m.meetings.CancelMeeting(ctx, meetingID, false)
	}

	return false, nil
}

func (m *Meeting) BothConfirmed(ctx context.Context, meetingID int64) (bool, *domain.Meeting, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, nil, err
	}
	return meeting.DillConfirmed && meeting.DoeConfirmed, meeting, nil
}

func (m *Meeting) GetPartnerTelegramID(ctx context.Context, meetingID int64, telegramID int64) (int64, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return 0, err
	}

	pair, err := m.pairs.GetPairByID(ctx, meeting.PairID)
	if err != nil || pair == nil {
		return 0, err
	}

	dill, err := m.users.GetUserByID(ctx, pair.DillID)
	if err != nil {
		return 0, err
	}
	doe, err := m.users.GetUserByID(ctx, pair.DoeID)
	if err != nil {
		return 0, err
	}

	if dill != nil && dill.TelegramID == telegramID {
		if doe != nil {
			return doe.TelegramID, nil
		}
	}
	if doe != nil && doe.TelegramID == telegramID {
		if dill != nil {
			return dill.TelegramID, nil
		}
	}

	return 0, nil
}

func (m *Meeting) GetPartnerUsername(ctx context.Context, meetingID int64, telegramID int64) (string, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return "", err
	}

	pair, err := m.pairs.GetPairByID(ctx, meeting.PairID)
	if err != nil || pair == nil {
		return "", err
	}

	dill, err := m.users.GetUserByID(ctx, pair.DillID)
	if err != nil {
		return "", err
	}
	doe, err := m.users.GetUserByID(ctx, pair.DoeID)
	if err != nil {
		return "", err
	}

	if dill != nil && dill.TelegramID == telegramID {
		if doe != nil {
			return doe.Username, nil
		}
	}
	if doe != nil && doe.TelegramID == telegramID {
		if dill != nil {
			return dill.Username, nil
		}
	}

	return "", nil
}

func (m *Meeting) GetPlaceDescription(ctx context.Context, placeID int64) (string, error) {
	places, err := m.places.GetAllPlaces(ctx)
	if err != nil {
		return "", err
	}
	for _, p := range places {
		if p.ID == placeID {
			return p.Description, nil
		}
	}
	return "", nil
}
