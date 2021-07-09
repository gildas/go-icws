package icws

import (
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

type StationSettings interface {
	Connect(session *Session) error
	Disconnect(session *Session) error
	core.TypeCarrier
}

// ConnectStation connects to a Station
func (session *Session) ConnectStation(settings StationSettings) error {
	return settings.Connect(session)
}

var stationSettingsRegistry = core.TypeRegistry{}

// UnmarshalStationSettings unmarshals from a JSON payload
func UnmarshalStationSettings(payload []byte) (StationSettings, error) {
	value, err := stationSettingsRegistry.UnmarshalJSON(payload, "__type")
	if err != nil {
		return nil, errors.JSONUnmarshalError.Wrap(err)
	}
	return value.(StationSettings), nil
}
