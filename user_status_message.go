package icws

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

type UserStatusMessage struct {
	UserStatuses []UserStatus `json:"userStatusList"`
	IsDelta      bool         `json:"isDelta"`
}

// UserStatusSubscription  describes a UserStatus Subscription Request
type UserStatusSubscription struct {
	UserIDs    []string `json:"userIds"`
	Properties []string `json:"userStatusProperties,omitempty"`
}

func init() {
	messageRegistry.Add(UserStatusMessage{})
}

// GetType tells the JSON type
//
// implements core.TypeCarrier
func (message UserStatusMessage) GetType() string {
	return "urn:inin.com:status:userStatusMessage"
}

// Subscribe subscribe a Session to this type of messages
//
// implements Subscriber
func (message UserStatusMessage) Subscribe(session *Session, payload interface{}) error {
	return session.sendPut("/messaging/subscriptions/status/user-statuses", payload, nil)
}

// Subscribe subscribe a Session to this type of messages
//
// implements Unsubscriber
func (message UserStatusMessage) Unsubscribe(session *Session) error {
	return session.sendDelete("/messaging/subscriptions/status/user-statuses")
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message UserStatusMessage) MarshalJSON() ([]byte, error) {
	type surrogate UserStatusMessage
	data, err := json.Marshal(struct {
		Type string `json:"__type"`
		surrogate
	}{
		Type:      message.GetType(),
		surrogate: surrogate(message),
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON unmarshals from JSON
//
// implements json.Unmarshaler
func (message *UserStatusMessage) UnmarshalJSON(payload []byte) (err error) {
	type surrogate UserStatusMessage
	var inner struct {
		Type string `json:"__type"`
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	if inner.Type != (UserStatusMessage{}.GetType()) {
		return errors.JSONUnmarshalError.Wrap(errors.ArgumentInvalid.With("__type", inner.Type))
	}
	*message = UserStatusMessage(inner.surrogate)
	return nil
}
