package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jus1d/kypidbot/internal/domain"
)

type MeetingNotification struct {
	MeetingID int64
	DillID    int64
	DoeID     int64
	Place     string
	Time      time.Time
}

type FullMatchNotification struct {
	DillTelegramID int64
	DoeTelegramID  int64
	DillFirstName  string
	DoeFirstName   string
}

type MeetResult struct {
	Meetings    []MeetingNotification
	FullMatches []FullMatchNotification
}

type Meeting struct {
	users    domain.UserRepository
	places   domain.PlaceRepository
	meetings domain.MeetingRepository
}

func NewMeeting(users domain.UserRepository, places domain.PlaceRepository, meetings domain.MeetingRepository) *Meeting {
	return &Meeting{
		users:    users,
		places:   places,
		meetings: meetings,
	}
}

var (
	ErrNoPlaces = errors.New("matching: no places")
	ErrNoPairs  = errors.New("matching: no pairs")
)

func (m *Meeting) CreateMeetings(ctx context.Context) (*MeetResult, error) {
	regularMeetings, err := m.meetings.GetRegularMeetings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get regular meetings: %w", err)
	}

	fullMeetings, err := m.meetings.GetFullMeetings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get full meetings: %w", err)
	}

	if len(regularMeetings) == 0 && len(fullMeetings) == 0 {
		return nil, ErrNoPairs
	}

	places, err := m.places.GetAllPlaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("get places: %w", err)
	}

	if len(places) == 0 && len(regularMeetings) > 0 {
		return nil, ErrNoPlaces
	}

	var result MeetResult

	for _, mt := range regularMeetings {
		dill, err := m.users.GetUser(ctx, mt.DillID)
		if err != nil {
			return nil, fmt.Errorf("get dill: %w", err)
		}
		doe, err := m.users.GetUser(ctx, mt.DoeID)
		if err != nil {
			return nil, fmt.Errorf("get doe: %w", err)
		}

		if dill == nil || doe == nil {
			continue
		}

		place := places[rand.Intn(len(places))]
		timeIntersection := domain.CalculateTimeIntersection(dill.TimeRanges, doe.TimeRanges)
		meetingTime := domain.PickRandomTime(timeIntersection)

		// Format: YYYY-MM-DD HH:MM
		full := fmt.Sprintf("%d-02-14 %s", time.Now().Year(), meetingTime)
		layout := "2006-01-02 15:04"

		loc, err := time.LoadLocation("Europe/Samara")
		if err != nil {
			return nil, fmt.Errorf("load location: %w", err)
		}

		t, err := time.ParseInLocation(layout, full, loc)
		if err != nil {
			return nil, fmt.Errorf("parse time %q: %w", full, err)
		}

		if err := m.meetings.AssignPlaceAndTime(ctx, mt.ID, place.ID, t); err != nil {
			return nil, fmt.Errorf("assign place and time: %w", err)
		}

		result.Meetings = append(result.Meetings, MeetingNotification{
			MeetingID: mt.ID,
			DillID:    dill.TelegramID,
			DoeID:     doe.TelegramID,
			Place:     place.Description,
			Time:      t,
		})
	}

	for _, mt := range fullMeetings {
		dill, err := m.users.GetUser(ctx, mt.DillID)
		if err != nil {
			return nil, fmt.Errorf("get dill: %w", err)
		}
		doe, err := m.users.GetUser(ctx, mt.DoeID)
		if err != nil {
			return nil, fmt.Errorf("get doe: %w", err)
		}

		if dill == nil || doe == nil {
			continue
		}

		result.FullMatches = append(result.FullMatches, FullMatchNotification{
			DillTelegramID: dill.TelegramID,
			DoeTelegramID:  doe.TelegramID,
			DillFirstName:  dill.FirstName,
			DoeFirstName:   doe.FirstName,
		})
	}

	return &result, nil
}

func (m *Meeting) ConfirmMeeting(ctx context.Context, meetingID int64, telegramID int64) (bool, *domain.Meeting, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, nil, err
	}

	isDill := meeting.DillID == telegramID
	isDoe := meeting.DoeID == telegramID

	if !isDill && !isDoe {
		return false, nil, nil
	}

	if err := m.meetings.UpdateState(ctx, meetingID, isDill, domain.StateConfirmed); err != nil {
		return false, nil, err
	}

	updated, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil {
		return false, nil, err
	}

	bothConfirmed := updated.DillState == domain.StateConfirmed && updated.DoeState == domain.StateConfirmed
	return bothConfirmed, updated, nil
}

func (m *Meeting) CancelMeeting(ctx context.Context, meetingID int64, telegramID int64) (bool, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, err
	}

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return false, err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
	if err != nil {
		return false, err
	}

	if dill != nil && dill.TelegramID == telegramID {
		return true, m.meetings.UpdateState(ctx, meetingID, true, domain.StateCancelled)
	}
	if doe != nil && doe.TelegramID == telegramID {
		return true, m.meetings.UpdateState(ctx, meetingID, false, domain.StateCancelled)
	}

	return false, nil
}

func (m *Meeting) BothConfirmed(ctx context.Context, meetingID int64) (bool, *domain.Meeting, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, nil, err
	}
	return meeting.DillState == domain.StateConfirmed && meeting.DoeState == domain.StateConfirmed, meeting, nil
}

func (m *Meeting) GetPartnerTelegramID(ctx context.Context, meetingID int64, telegramID int64) (int64, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return 0, err
	}

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return 0, err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
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

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return "", err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
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

func (m *Meeting) GetPartner(ctx context.Context, meetingID int64, telegramID int64) (*domain.User, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return nil, err
	}

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return nil, err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
	if err != nil {
		return nil, err
	}

	if dill != nil && dill.TelegramID == telegramID {
		return doe, nil
	}
	if doe != nil && doe.TelegramID == telegramID {
		return dill, nil
	}

	return nil, nil
}

func (m *Meeting) SetArrived(ctx context.Context, meetingID int64, telegramID int64) error {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return err
	}

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
	if err != nil {
		return err
	}

	if dill != nil && dill.TelegramID == telegramID {
		return m.meetings.UpdateState(ctx, meetingID, true, domain.StateArrived)
	}
	if doe != nil && doe.TelegramID == telegramID {
		return m.meetings.UpdateState(ctx, meetingID, false, domain.StateArrived)
	}

	return nil
}

func (m *Meeting) GetArrivedMeetingID(ctx context.Context, telegramID int64) (int64, error) {
	return m.meetings.GetArrivedMeetingID(ctx, telegramID)
}

func (m *Meeting) SetCantFind(ctx context.Context, meetingID int64, telegramID int64) (bool, error) {
	meeting, err := m.meetings.GetMeetingByID(ctx, meetingID)
	if err != nil || meeting == nil {
		return false, err
	}

	dill, err := m.users.GetUser(ctx, meeting.DillID)
	if err != nil {
		return false, err
	}
	doe, err := m.users.GetUser(ctx, meeting.DoeID)
	if err != nil {
		return false, err
	}

	isDill := dill != nil && dill.TelegramID == telegramID
	isDoe := doe != nil && doe.TelegramID == telegramID

	if !isDill && !isDoe {
		return false, nil
	}

	if err := m.meetings.SetCantFind(ctx, meetingID, isDill); err != nil {
		return false, err
	}

	if isDill {
		return meeting.DoeCantFind, nil
	}
	return meeting.DillCantFind, nil
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
