package icws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

// Session describes a session connected to a PureConnect server
type Session struct {
	ID               string                  `json:"id"`
	Token            string                  `json:"token"`
	Cookies          []*http.Cookie          `json:"cookies"`
	Timezone         string                  `json:"timezone"` // ???
	APIRoot          *url.URL                `json:"-"`
	User             User                    `json:"user"`
	Status           SessionStatus           `json:"status"`
	Features         []SessionFeature        `json:"features"`
	Logger           *logger.Logger `json:"-"`
	SessionOptions
}

// SessionOptions describes the options of a Session
//
// To give a Logger to the Session, pass it to the Context
//
// The Context will be passed to the TokenUpdated chan (if any) when the Token changes
// allowing application to pass data through it.
type SessionOptions struct {
	Context      context.Context   `json:"-"`
	Servers      []*url.URL        `json:"-"`
	UserID       string            `json:"-"`
	Password     string            `json:"-"`
	Application  string            `json:"applicationName"`
	Language     string            `json:"language"`
	TokenUpdated chan UpdatedToken `json:"-"`
}

// UpdatedToken describes the event sent to a chan letting applications know about new Token
type UpdatedToken struct {
	Token     string          `json:"token"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Context   context.Context `json:"context"`
}

// SessionFeature describes a feature supported by PureConnect Servers
type SessionFeature struct {
	Name    string `json:"featureId"`
	Version int    `json:"version"`
}

// NewSession creates a new Session
//
// Warning: The Session is NOT connected
//
// To give a logger.Logger to the Session, pass it to the context.Context
func NewSession(options SessionOptions) *Session {
	log, err := logger.FromContext(options.Context)
	if err != nil {
		log = logger.Create("ICWS", &logger.NilStream{})
	}
	log = log.Child("session", "session")
	if len(options.Language) == 0 {
		options.Language = "en-us"
	}
	return &Session{
		User:           User{ID: options.UserID},
		Status:         DisconnectedStatus,
		SessionOptions: options,
		Logger:         log,
	}
}

// IsConnected tells if the Session is connected to a PureConnect server
func (session Session) IsConnected() bool {
	return session.Status == ConnectedStatus || session.Status == DisconnectingStatus || session.Status == ChangingStatus
}

// Connect connects to a PureConnect Server
//
// If the Session is currently connected, nothing is done
func (session *Session) Connect() (err error) {
	log := session.Logger.Child(nil, "connect")
	if session.IsConnected() || session.Status == ConnectingStatus {
		log.Tracef("Session is already connected or connecting")
		return nil
	}
	session.Status = ConnectingStatus
	serverIndex := 0
	nextIndex := func(index int, currentServer *url.URL) (int, error) {
		for index++; index < len(session.Servers); index++ {
			if currentServer.Host != session.Servers[index].Host {
				return index, nil
			}
		}
		return 0, errors.HTTPServiceUnavailable.WithStack()
	}
	for {
		var endpoint *url.URL

		server := session.Servers[serverIndex]
		session.APIRoot, _ = server.Parse("/icws")
		endpoint, err = session.APIRoot.Parse("/connection")
		if err != nil {
			log.Errorf("Failed to create endpoint: %s/connection", server.String())
			serverIndex, err = nextIndex(serverIndex, server)
			if err != nil {
				break // We should return an error to the caller now...
			}
			continue
		}

		log.Debugf("Connecting to %s (endpoint: %s)", server, endpoint)
		results := struct {
			Token             string           `json:"csrfToken"`
			SessionID         string           `json:"sessionId"`
			Alternates        []string         `json:"alternateHostList"`
			Server            string           `json:"icServer"`
			UserID            string           `json:"userID"`
			DisplayName       string           `json:"userDisplayName"`
			PasswordExpiredIn int              `json:"daysUntilPasswordExpiration"`
			Features          []SessionFeature `json:"features"`
		}{}

		err = session.sendPost("/connection?include=features",
			struct {
				Type        string `json:"__type"`
				Application string `json:"applicationName"`
				UserID      string `json:"userID"`
				Password    string `json:"password"`
			}{
				Type:        "urn:inin.com:connection:icAuthConnectionRequestSettings",
				Application: session.Application,
				UserID:      session.UserID,
				Password:    session.Password,
			},
			&results,
		)
		if errors.Is(err, errors.HTTPServiceUnavailable) {
			// TODO: On HTTP 503, we receive a list of alternate hosts that we should connect to
			// We also need to reset the serverIndex after getting the new list
			serverIndex, err = nextIndex(serverIndex, server)
			if err != nil {
				break // We should return an error to the caller now...
			}
			continue
		} else if err != nil {
			log.Errorf("Failed to connect to %s", endpoint, err)
			serverIndex, err = nextIndex(serverIndex, server)
			if err != nil {
				break // We should return an error to the caller now...
			}
			continue
		}
		session.ID = results.SessionID
		session.Token = results.Token
		if session.TokenUpdated != nil {
			log.Tracef("Sending new Token to chan")
			session.TokenUpdated <- UpdatedToken{
				Token:   session.Token,
				Context: session.Context,
			}
		}
		if len(results.Alternates) > 0 {
			session.Servers = make([]*url.URL, len(results.Alternates))
			for i := 0; i < len(results.Alternates); i++ {
				session.Servers[i] = core.Must(url.Parse(fmt.Sprintf("%s://%s:%s", server.Scheme, results.Alternates[i], server.Port()))).(*url.URL)
			}
		}
		session.User.ID = results.UserID
		session.User.DisplayName = results.DisplayName
		session.Status = ConnectedStatus
		session.Features = results.Features
		session.Logger = session.Logger.Record("session", session.ID)

		return nil
	}
	session.Status = DisconnectedStatus
	return err
}

// Disconnect disconnects the Session from PureConnect
func (session *Session) Disconnect() error {
	log := session.Logger.Child(nil, "disconnect")
	if !session.IsConnected() || session.Status == DisconnectingStatus {
		return nil
	}
	session.Status = DisconnectingStatus
	err := session.sendDelete("/connection")
	log.Debugf("Disconnected from %s", session.APIRoot.Host)
	if err == nil {
		session.Status = DisconnectedStatus
		session.ID = ""
	}
	return err
}

// HasSupport tells if the Session supports the given PureConnect feature
func (session Session) HasSupport(featureName string) bool {
	featureName = strings.ToLower(featureName)
	for _, feature := range session.Features {
		if feature.Name == featureName {
			return true
		}
	}
	return false
}

// HasSupport tells if the Session supports the given PureConnect feature
func (session Session) HasSupportWithAtLeastVersion(featureName string, minimumVersion int) bool {
	featureName = strings.ToLower(featureName)
	for _, feature := range session.Features {
		if feature.Name == featureName {
			return feature.Version >= minimumVersion
		}
	}
	return false
}

// String gets a text representation
//
// implements fmt.Stringer
func (session Session) String() string {
	builder := strings.Builder{}
	builder.WriteString("Session ")
	if len(session.ID) > 0 {
		builder.WriteString("#")
		builder.WriteString(session.ID)
	}
	if session.APIRoot != nil {
		builder.WriteString(" to ")
		builder.WriteString(session.APIRoot.Host)
	}
	if len(session.User.String()) > 0 {
		builder.WriteString(" as ")
		builder.WriteString(session.User.String())
	}
	builder.WriteString(" ")
	builder.WriteString(session.Status.String())
	return builder.String()
}

// MarshalJSON marshals into JSON
//
// implements json.Marshaler
func (session Session) MarshalJSON() ([]byte, error) {
	type surrogate Session
	servers := make([]*core.URL, len(session.Servers))
	for i := 0; i < len(session.Servers); i++ {
		servers[i] = (*core.URL)(session.Servers[i])
	}
	data, err := json.Marshal(struct {
		surrogate
		APIRoot *core.URL
		Servers []*core.URL
	}{
		surrogate: surrogate(session),
		APIRoot:   (*core.URL)(session.APIRoot),
		Servers:   servers,
	})
	return data, errors.JSONMarshalError.Wrap(err)
}
