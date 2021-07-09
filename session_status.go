package icws

// SessionStatus reflects the status of the Session
type SessionStatus uint32

const (
	UnknownStatus SessionStatus = iota
	ConnectedStatus
	ConnectingStatus
	DisconnectedStatus
	DisconnectingStatus
	ChangingStatus
)

// MarshalText marshals this SessionStatus into a textual form
//
// implements encoding.TextMarshaler
func (status SessionStatus) MarshalText() ([]byte, error) {
	switch status {
	case ConnectedStatus:
		return []byte("Connected"), nil
	case ConnectingStatus:
		return []byte("Connecting"), nil
	case DisconnectedStatus:
		return []byte("Disconnected"), nil
	case DisconnectingStatus:
		return []byte("Disconnecting"), nil
	case ChangingStatus:
		return []byte("Changing Status"), nil
	default:
		return []byte("Unknown"), nil
	}
}

// String gets a text representation
//
// implements fmt.Stringer
func (status SessionStatus) String() string {
	switch status {
	case ConnectedStatus:
		return " (Connected)"
	case ConnectingStatus:
		return " (Connecting)"
	case DisconnectedStatus:
		return " (Disconnected)"
	case DisconnectingStatus:
		return " (Disconnecting)"
	case ChangingStatus:
		return " (Changing Status)"
	default:
		return " (Unknown)"
	}
}
