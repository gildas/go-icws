package icws

// StatusMessage describes a Status Message
type StatusMessage struct {
	ID              string `json:"statusId"`
	SystemID        string `json:"systemId"`
	Text            string `json:"messageText"`
	IconURI         string `json:"iconUri"`
	GroupTag        string `json:"groupTag"`
	CanHaveDate     bool   `json:"canHaveDate"`
	CanHaveTime     bool   `json:"canHaveTime"`
	IsDoNotDisturb  bool   `json:"isDoNotDisturbStatus"`
	IsSelectable    bool   `json:"isSelectableStatus"`
	IsPersistent    bool   `json:"isPersistentStatus"`
	IsForward       bool   `json:"isForwardStatus"`
	IsAfterCallWork bool   `json:"isAfterCallWorkStatus"`
	IsACD           bool   `json:"isACDStatus"`
	IsAllowFollowUp bool   `json:"isAllowFollowUpStatus"`
}

// GetID tells the ID
//
// implements Identifiable
func (status StatusMessage) GetID() string {
	return status.ID
}
