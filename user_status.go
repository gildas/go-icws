package icws

import (
	"encoding/json"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// UserStatus describes a User Status
type UserStatus struct {
	UserID           string    `json:"userId"`
	StatusID         string    `json:"statusId"`
	IsLoggedIn       bool      `json:"loggedIn"`
	IsOnPhone        bool      `json:"onPhone"`
	OnPhoneChangedAt time.Time `json:"-"`
	ChangedAt        time.Time `json:"-"`
	Servers          []string  `json:"icServers"`
	Stations         []string  `json:"stations"`
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message UserStatus) MarshalJSON() ([]byte, error) {
	type surrogate UserStatus
	data, err := json.Marshal(struct {
		surrogate
		OnPhoneChangedAt core.Time `json:"onPhoneChanged"`
		ChangedAt        core.Time `json:"statusChanged"`
	}{
		surrogate:        surrogate(message),
		OnPhoneChangedAt: core.Time(message.OnPhoneChangedAt),
		ChangedAt:        core.Time(message.ChangedAt),
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON unmarshals from JSON
//
// implements json.Unmarshaler
func (message *UserStatus) UnmarshalJSON(payload []byte) (err error) {
	type surrogate UserStatus
	var inner struct {
		surrogate
		OnPhoneChangedAt string `json:"onPhoneChanged"`
		ChangedAt        string `json:"statusChanged"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*message = UserStatus(inner.surrogate)
	message.OnPhoneChangedAt, err = time.Parse("20060102T150405Z", inner.OnPhoneChangedAt)
	if err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	message.ChangedAt, err = time.Parse("20060102T150405Z", inner.ChangedAt)
	if err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return nil
}
