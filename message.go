package icws

import (
	"encoding/json"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

type Message interface {
	core.TypeCarrier
}

type messageWrapper struct {
	Message Message
}

var messageRegistry = core.TypeRegistry{}

// UnmarshalMessage unmarshals from a JSON payload
func UnmarshalMessage(payload []byte) (Message, error) {
	wrapper := messageWrapper{}
	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return nil, err // err is already decorated by the UnmarshalJSON funcs
	}
	return wrapper.Message, nil
}

// UnmarshalJSON unmarshals from JSON
//
// implements json.Unmarshaler
func (wrapper *messageWrapper) UnmarshalJSON(payload []byte) (err error) {
	value, err := messageRegistry.UnmarshalJSON(payload, "__type")
	if err == nil {
		wrapper.Message = value.(Message)
	}
	return errors.WithStack(err)
}
