package domain

import "context"

type Meeting struct {
	ID            int64
	PairID        int64
	PlaceID       int64
	Time          string
	DillConfirmed bool
	DoeConfirmed  bool
	DillCancelled bool
	DoeCancelled  bool
}

type MeetingRepository interface {
	SaveMeeting(ctx context.Context, m *Meeting) (int64, error)
	GetMeetingByID(ctx context.Context, id int64) (*Meeting, error)
	ConfirmMeeting(ctx context.Context, meetingID int64, isDill bool) error
	CancelMeeting(ctx context.Context, meetingID int64, isDill bool) error
}
