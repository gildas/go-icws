package icws

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

type WorkStationSettings struct {
}

func init() {
	stationSettingsRegistry.Add(WorkStationSettings{})
}

// GetType tells the JSON type
//
// implements core.TypeCarrier
func (settings WorkStationSettings) GetType() string {
	return "urn:inin.com:connection:workstationSettings"
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (message WorkStationSettings) MarshalJSON() ([]byte, error) {
	type surrogate WorkStationSettings
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
func (message *WorkStationSettings) UnmarshalJSON(payload []byte) (err error) {
	type surrogate WorkStationSettings
	var inner struct {
		Type string `json:"__type"`
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	if inner.Type != (WorkStationSettings{}.GetType()) {
		return errors.JSONUnmarshalError.Wrap(errors.ArgumentInvalid.With("__type", inner.Type))
	}
	*message = WorkStationSettings(inner.surrogate)
	return nil
}
