package icws

// User describes a PureConnect User
type User struct {
	ID          string `json:"userID"`
	DisplayName string `json:"displayName,omitempty"`
	SelfUri     string `json:"uri"`
}

// GetUsers retrieves a list of Users
func (session *Session) GetUsers() ([]*User, error) {
	items := struct{
		Items []struct{
			User struct{
				ID          string `json:"id"`
				DisplayName string `json:"displayName,omitempty"`
				SelfUri     string `json:"uri"`
			} `json:"configurationId"`
		} `json:"items"`
	}{}
	err := session.sendGet("/configuration/users", &items)
	users := make([]*User, len(items.Items))
	for i := 0; i < len(items.Items); i++ {
		user := User(items.Items[i].User)
		users[i] = &user
	}
	return users, err
}

// String gets a text representation
//
// implements fmt.Stringer
func (user User) String() string {
	if len(user.DisplayName) > 0 {
		return user.DisplayName
	}
	return user.ID
}
