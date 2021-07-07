package icws

// StatusMessage describes a Status Message
type StatusMessage struct {
	ID string `json:"statusId"`
}

// GetID tells the ID
//
// implements Identifiable
func (status StatusMessage) GetID() string {
	return status.ID
}
