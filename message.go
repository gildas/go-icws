package icws

import (
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

type Message interface {
	core.TypeCarrier
}

var messageRegistry = core.TypeRegistry{}

// UnmarshalMessage unmarshals from a JSON payload
func UnmarshalMessage(payload []byte) (Message, error) {
	value, err := messageRegistry.UnmarshalJSON(payload, "__type")
	if err != nil {
		return nil, errors.JSONUnmarshalError.Wrap(err)
	}
	return value.(Message), nil
}
