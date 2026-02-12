package messages

var M Messages

type Messages struct {
	Start         StartSection         `yaml:"start" env-required:"true"`
	Notifications NotificationsSection `yaml:"notifications" env-required:"true"`
	Profile       ProfileSection       `yaml:"profile" env-required:"true"`
	Command       CommandSection       `yaml:"command" env-required:"true"`
	Registration  RegistrationSection  `yaml:"registration" env-required:"true"`
	UI            UISection            `yaml:"ui" env-required:"true"`
	Service       ServiceSection       `yaml:"service" env-required:"true"`
	Admin         AdminSection         `yaml:"admin" env-required:"true"`
	Error         ErrorSection         `yaml:"error" env-required:"true"`
	Matching      MatchingSection      `yaml:"matching" env-required:"true"`
	Meeting       MeetingSection       `yaml:"meeting" env-required:"true"`
}

type BotSection struct {
	Start        StartSection        `yaml:"start" env-required:"true"`
	Profile      ProfileSection      `yaml:"profile" env-required:"true"`
	Registration RegistrationSection `yaml:"registration" env-required:"true"`
}

type StartSection struct {
	Welcome string `yaml:"welcome" env-required:"true"`
}

type NotificationsSection struct {
	Remind         string `yaml:"remind" env-required:"true"`
	Registration   string `yaml:"registration" env-required:"true"`
	Invite         string `yaml:"invite" env-required:"true"`
	MeetingSoon    string `yaml:"meeting_soon" env-required:"true"`
	ArrivedAsk     string `yaml:"arrived_ask" env-required:"true"`
	ArrivedPartner string `yaml:"arrived_partner" env-required:"true"`
	CantFindNoted  string `yaml:"cant_find_noted" env-required:"true"`
	CantFindBoth   string `yaml:"cant_find_both" env-required:"true"`
}

type ProfileSection struct {
	Sex      SexOnboardingSection `yaml:"sex" env-required:"true"`
	About    AboutSection         `yaml:"about" env-required:"true"`
	Schedule ScheduleSection      `yaml:"schedule" env-required:"true"`
}

type SexOnboardingSection struct {
	AskNew   string `yaml:"ask_new" env-required:"true"`
	AskRetry string `yaml:"ask_retry" env-required:"true"`
}

type AboutSection struct {
	Request  string `yaml:"request" env-required:"true"`
	Accepted string `yaml:"accepted" env-required:"true"`
}

type ScheduleSection struct {
	Request string `yaml:"request" env-required:"true"`
}

type CommandSection struct {
	About       string             `yaml:"about" env-required:"true"`
	Support     SupportSection     `yaml:"support" env-required:"true"`
	Leaderboard LeaderboardSection `yaml:"leaderboard" env-required:"true"`
	Admin       string             `yaml:"admin_panel" env-required:"true"`
	Remind      RemindSection      `yaml:"remind" env-required:"true"`
	Pairs       PairsSection       `yaml:"pairs" env-required:"true"`
}

type RemindSection struct {
	Sent    string `yaml:"sent" env-required:"true"`
	NoUsers string `yaml:"no_users" env-required:"true"`
}

type PairsSection struct {
	NotFound string `yaml:"not_found" env-required:"true"`
	Error    string `yaml:"error" env-required:"true"`
}

type SupportSection struct {
	Request     string `yaml:"request" env-required:"true"`
	ProblemSent string `yaml:"problem_sent" env-required:"true"`
	Ticket      string `yaml:"ticket" env-required:"true"`
	Cancelled   string `yaml:"cancelled" env-required:"true"`
}

type RegistrationSection struct {
	Completed        string `yaml:"completed" env-required:"true"`
	Closed           string `yaml:"closed" env-required:"true"`
	ClosedRegistered string `yaml:"closed_registered" env-required:"true"`
}

type UISection struct {
	Buttons ButtonsSection `yaml:"buttons" env-required:"true"`
	Chosen  string         `yaml:"chosen" env-required:"true"`
}

type ButtonsSection struct {
	Sex            SexButtons `yaml:"sex" env-required:"true"`
	Confirm        string     `yaml:"confirm" env-required:"true"`
	Resubmit       string     `yaml:"resubmit" env-required:"true"`
	ConfirmMeeting string     `yaml:"confirm_meeting" env-required:"true"`
	CancelMeeting  string     `yaml:"cancel_meeting" env-required:"true"`
	CancelSupport  string     `yaml:"cancel_support" env-required:"true"`
	HowItWorks     string     `yaml:"how_it_works" env-required:"true"`
	Arrived        string     `yaml:"arrived" env-required:"true"`
	CantFind       string     `yaml:"cant_find" env-required:"true"`
	OptOut         string     `yaml:"opt_out" env-required:"true"`
	OptIn          string     `yaml:"opt_in" env-required:"true"`
}

type SexButtons struct {
	Male   string `yaml:"male" env-required:"true"`
	Female string `yaml:"female" env-required:"true"`
}

type ServiceSection struct {
	Sticker StickerSection `yaml:"sticker" env-required:"true"`
}

type StickerSection struct {
	Request string `yaml:"request" env-required:"true"`
}

type AdminSection struct {
	Promote            AdminCommand `yaml:"promote" env-required:"true"`
	Demote             AdminCommand `yaml:"demote" env-required:"true"`
	StartedLog         string       `yaml:"started_log" env-required:"true"`
	RegistrationClosed string       `yaml:"registration_closed" env-required:"true"`
	RegistrationOpened string       `yaml:"registration_opened" env-required:"true"`
}

type AdminCommand struct {
	Usage   string `yaml:"usage" env-required:"true"`
	Success string `yaml:"success" env-required:"true"`
}

type ErrorSection struct {
	UserNotFound         string `yaml:"user_not_found" env-required:"true"`
	AlreadyAdmin         string `yaml:"already_admin" env-required:"true"`
	NotAdmin             string `yaml:"not_admin" env-required:"true"`
	CannotDemoteYourself string `yaml:"cannot_demote_yourself" env-required:"true"`
}

type DemoteCommand struct {
	Usage        string `yaml:"usage" env-required:"true"`
	UserNotFound string `yaml:"user_not_found" env-required:"true"`
	AlreadyAdmin string `yaml:"already_admin" env-required:"true"`
	NotAdmin     string `yaml:"not_admin" env-required:"true"`
	Success      string `yaml:"success" env-required:"true"`
}

type MatchingSection struct {
	Errors  MatchingErrors  `yaml:"errors" env-required:"true"`
	Success MatchingSuccess `yaml:"success" env-required:"true"`
}

type MatchingErrors struct {
	NotEnoughUsers string `yaml:"not_enough_users" env-required:"true"`
	NoPairs        string `yaml:"no_pairs" env-required:"true"`
	NoPlaces       string `yaml:"no_places" env-required:"true"`
}

type MatchingSuccess struct {
	Matched      string `yaml:"matched" env-required:"true"`
	MeetingsSent string `yaml:"meetings_sent" env-required:"true"`
	NotMatched   string `yaml:"not_matched" env-required:"true"`
}

type MeetingSection struct {
	Invite  MeetingInviteSection  `yaml:"invite" env-required:"true"`
	Status  MeetingStatusSection  `yaml:"status" env-required:"true"`
	Special MeetingSpecialSection `yaml:"special" env-required:"true"`
}

type MeetingInviteSection struct {
	Message          string `yaml:"message" env-required:"true"`
	WaitConfirmation string `yaml:"wait_confirmation" env-required:"true"`
}

type MeetingStatusSection struct {
	Confirmed        string `yaml:"confirmed" env-required:"true"`
	PartnerConfirmed string `yaml:"partner_confirmed" env-required:"true"`
	BothConfirmed    string `yaml:"both_confirmed" env-required:"true"`
	Cancelled        string `yaml:"cancelled" env-required:"true"`
	PartnerCancelled string `yaml:"partner_cancelled" env-required:"true"`
}

type MeetingSpecialSection struct {
	FullMatchNoTime string `yaml:"full_match_no_time" env-required:"true"`
}

type LeaderboardSection struct {
	Title  string `yaml:"title" env-required:"true"`
	Empty  string `yaml:"empty" env-required:"true"`
	Footer string `yaml:"footer" env-required:"true"`
	Error  string `yaml:"error" env-required:"true"`
}
