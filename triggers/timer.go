package triggers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KarlGW/azfunc/data"
)

// Timer represent a Timer trigger.
type Timer struct {
	ScheduleStatus TimerScheduleStatus
	Metadata       Metadata
	Schedule       TimerSchedule
	IsPastDue      bool
}

// TimerOptions contains options for a Timer trigger.
type TimerOptions struct{}

// TimerOption is a function that sets options on a Timer trigger.
type TimerOption func(o *TimerOptions)

// TimerSchedule represents the Schedule field from the incoming
// request.
type TimerSchedule struct {
	AdjustForDST bool
}

// TimerScheduleStatus represents the ScheduleStatus field from
// the incoming request.
type TimerScheduleStatus struct {
	Last        time.Time
	Next        time.Time
	LastUpdated time.Time
}

// Parse together with Data satisfies the Triggerable interface. It
// is a no-op method. A timer trigger contains no other data needed
// to be parsed. Use the fields directly.
func (t Timer) Parse(v any) error {
	return nil
}

// Data together with Parse satisfies the Triggerable interface. It
// is a no-op method. A timer trigger contains no other data needed
// to be parsed. Use the fields directly.
func (t Timer) Data() data.Raw {
	return nil
}

// NewTimer creates and returns a Timer trigger from the provided
// *http.Request. The name on the trigger in function.json must
// be "timer".
func NewTimer(r *http.Request, options ...TimerOption) (*Timer, error) {
	opts := TimerOptions{}
	for _, option := range options {
		option(&opts)
	}

	var t timerTrigger
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, ErrTriggerPayloadMalformed
	}
	defer r.Body.Close()

	return &Timer{
		Schedule:       t.Data.Timer.Schedule,
		ScheduleStatus: t.Data.Timer.ScheduleStatus,
		IsPastDue:      t.Data.Timer.IsPastDue,
		Metadata:       t.Metadata,
	}, nil
}

// timerTrigger is the incoming request from the Function host.
type timerTrigger struct {
	Data struct {
		Timer struct {
			ScheduleStatus TimerScheduleStatus
			Schedule       TimerSchedule
			IsPastDue      bool
		} `json:"timer"`
	}
	Metadata Metadata
}
