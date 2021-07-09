package icws

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

type StatusMessageMessage struct {
	AddedMessages   []StatusMessage `json:"statusMessagesAdded"`
	ChangedMessages []StatusMessage `json:"statusMessagesChanged"`
	RemovedMessages []string        `json:"statusMessagesRemoved"`
	IsDelta         bool            `json:"isDelta"`
}

func init() {
	messageRegistry.Add(StatusMessageMessage{})
}

// GetType tells the JSON type
//
// implements core.TypeCarrier
func (subscription StatusMessageMessage) GetType() string {
	return "urn:inin.com:status:statusMessagesMessage"
}

// Subscribe subscribe a Session to this type of messages
//
// implements Subscriber
func (subscription StatusMessageMessage) Subscribe(session *Session, payload interface{}) error {
	return session.sendPut("/messaging/subscriptions/status/status-messages", payload, nil)
}

// Subscribe subscribe a Session to this type of messages
//
// implements Unsubscriber
func (subscription StatusMessageMessage) Unsubscribe(session *Session) error {
	return session.sendDelete("/messaging/subscriptions/status/status-messages")
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message StatusMessageMessage) MarshalJSON() ([]byte, error) {
	type surrogate StatusMessageMessage
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
func (message *StatusMessageMessage) UnmarshalJSON(payload []byte) (err error) {
	type surrogate StatusMessageMessage
	var inner struct {
		Type string `json:"__type"`
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	if inner.Type != (StatusMessageMessage{}.GetType()) {
		return errors.JSONUnmarshalError.Wrap(errors.ArgumentInvalid.With("__type", inner.Type))
	}
	*message = StatusMessageMessage(inner.surrogate)
	return nil
}
