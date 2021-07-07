package icws

type StationSettings interface {
	Connect(session *Session) error
	Disconnect(session *Session) error
}

// ConnectStation connects to a Station
func (session *Session) ConnectStation(settings StationSettings) error {
	return settings.Connect(session)
}