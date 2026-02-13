package domain

import (
	"context"
	"time"
)

type ConfirmationState string

const (
	StateNotConfirmed ConfirmationState = "not_confirmed"
	StateConfirmed    ConfirmationState = "confirmed"
	StateCancelled    ConfirmationState = "cancelled"
	StateArrived      ConfirmationState = "arrived"
)

type Meeting struct {
	ID             int64
	DillID         int64
	DoeID          int64
	PairScore      float64
	IsFullmatch    bool
	PlaceID        *int64
	Time           *time.Time
	DillState      ConfirmationState
	DoeState       ConfirmationState
	UsersNotified  bool
	DillCantFind   bool
	DoeCantFind    bool
}

type MeetingRepository interface {
	SaveMeeting(ctx context.Context, m *Meeting) error
	GetMeetingByID(ctx context.Context, id int64) (*Meeting, error)
	GetMeetingByUsers(ctx context.Context, dillID, doeID int64) (*Meeting, error)
	GetRegularMeetings(ctx context.Context) ([]Meeting, error)
	GetFullMeetings(ctx context.Context) ([]Meeting, error)
	AssignPlaceAndTime(ctx context.Context, id int64, placeID int64, time time.Time) error
	UpdateState(ctx context.Context, meetingID int64, isDill bool, state ConfirmationState) error
	ClearMeetings(ctx context.Context) error
	GetMeetingsStartingIn(ctx context.Context, interval time.Duration) ([]Meeting, error)
	MarkNotified(ctx context.Context, meetingID int64) error
	SetCantFind(ctx context.Context, meetingID int64, isDill bool) error
	GetArrivedMeetingID(ctx context.Context, telegramID int64) (int64, error)
	GetMeetingStats(ctx context.Context) (MeetingStats, error)
}

type MeetingStats struct {
	Total     uint
	Confirmed uint
	Cancelled uint
	Pending   uint
}
