package icws

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

type LicenseMessage struct {
	Licenses []License `json:"licenseAssignedStatusList"`
	IsDelta  bool      `json:"isDelta"`
}

// LicenseSubscription  describes a License Subscription Request
type LicenseSubscription struct {
	Licenses    []string `json:"licenseList"`
}

func init() {
	messageRegistry.Add(LicenseMessage{})
}

// GetType tells the JSON type
//
// implements core.TypeCarrier
func (message LicenseMessage) GetType() string {
	return "urn:inin.com:status:licenseMessage"
}

// Subscribe subscribe a Session to this type of messages
//
// implements Subscriber
func (message LicenseMessage) Subscribe(session *Session, payload interface{}) error {
	return session.sendPut("/messaging/subscriptions/licenses", payload, nil)
}

// Subscribe subscribe a Session to this type of messages
//
// implements Unsubscriber
func (message LicenseMessage) Unsubscribe(session *Session) error {
	return session.sendDelete("/messaging/subscriptions/licenses")
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message LicenseMessage) MarshalJSON() ([]byte, error) {
	type surrogate LicenseMessage
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
func (message *LicenseMessage) UnmarshalJSON(payload []byte) (err error) {
	type surrogate LicenseMessage
	var inner struct {
		Type string `json:"__type"`
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	if inner.Type != (LicenseMessage{}.GetType()) {
		return errors.JSONUnmarshalError.Wrap(errors.ArgumentInvalid.With("__type", inner.Type))
	}
	*message = LicenseMessage(inner.surrogate)
	return nil
}
