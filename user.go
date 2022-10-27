package icws

import (
	"net/http"
)

// User describes a PureConnect User
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
	SelfUri     string `json:"uri"`
	License     LicenseProperties
}

type userConfiguration struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
	SelfUri     string `json:"uri"`
}

type userRecord struct {
	UserConfiguration         userConfiguration `json:"configurationId"`
	NTDomainUser              string            `json:"ntDomainUser"`
	MWIEnabled                bool              `json:"mwiEnabled"`
	MWIMode                   int               `json:"mwiMode"`
	PagerActive               bool              `json:"pagerActive"`
	OutlookIntegrationEnabled bool              `json:"outlookIntegrationEnabled"`
	LicenseProperties         LicenseProperties `json:"licenseProperties"`
	CreatedAt                 Time              `json:"createdDate"`
	ModifiedAt                Time              `json:"modifiedDate"`
}

// GetID tells the ID
//
// implements Identifiable
func (user User) GetID() string {
	return user.ID
}

// GetUsers retrieves a list of Users
func (session *Session) GetUsers() ([]User, error) {
	data := struct {
		Items []userRecord `json:"items"`
	}{}
	err := session.sendGet("/configuration/users", &data)
	users := make([]User, len(data.Items))
	for i := 0; i < len(data.Items); i++ {
		users[i] = User{
			ID:          data.Items[i].UserConfiguration.ID,
			DisplayName: data.Items[i].UserConfiguration.DisplayName,
			SelfUri:     data.Items[i].UserConfiguration.SelfUri,
		}
	}
	return users, err
}

// GetUsers retrieves a list of Users
func (session *Session) GetUsersWithOptions(options QueryOptions) ([]User, error) {
	// If there is a select option, users are sent back 200 at a time
	// See: https://help.genesys.com/developer/cic/docs/icws/webhelp/icws/(sessionId)/configuration/users/index.htm#get
	data := struct {
		Items []userRecord `json:"items"`
	}{}
	headers := map[string]string{}
	users := []User{}
	for userRange := NewRange("items"); !userRange.IsAtEnd(); {
		userRange.ToMap(headers)
		response, err := session.send(
			http.MethodGet,
			"/configuration/users",
			headers,
			options.AsQueryParameters(),
			nil,
			&data,
		)
		if err != nil {
			return []User{}, err
		}

		for _, item := range data.Items {
			user := User{
				ID:          item.UserConfiguration.ID,
				DisplayName: item.UserConfiguration.DisplayName,
				SelfUri:     item.UserConfiguration.SelfUri,
				License:     item.LicenseProperties,
			}
			users = append(users, user)
		}
		userRange = GetRangeFromHeader(response.Headers)
		session.Logger.Tracef("New Range: %#+v", userRange)
	}
	return users, nil
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
