package icws

import (
	"encoding/json"
	"strings"
	"time"

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

// String gets a text representation
//
// implements fmt.Stringer
func (message UserStatus) String() string {
	sb := strings.Builder{}
	sb.WriteString(message.UserID)
	sb.WriteString(": ")
	sb.WriteString(message.StatusID)
	if message.IsLoggedIn {
		sb.WriteString(", logged in ")
	} else {
		sb.WriteString(", logged out")
	}
	return sb.String()
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message UserStatus) MarshalJSON() ([]byte, error) {
	type surrogate UserStatus
	data, err := json.Marshal(struct {
		surrogate
		OnPhoneChangedAt Time `json:"onPhoneChanged"`
		ChangedAt        Time `json:"statusChanged"`
	}{
		surrogate:        surrogate(message),
		OnPhoneChangedAt: Time(message.OnPhoneChangedAt),
		ChangedAt:        Time(message.ChangedAt),
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
		OnPhoneChangedAt Time `json:"onPhoneChanged"`
		ChangedAt        Time `json:"statusChanged"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*message = UserStatus(inner.surrogate)
	message.OnPhoneChangedAt = time.Time(inner.OnPhoneChangedAt)
	message.ChangedAt = time.Time(inner.ChangedAt)
	return nil
}
