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
	ID                   string                  `json:"id"`
	Token                string                  `json:"token"`
	Cookies              []*http.Cookie          `json:"cookies"`
	Timezone             string                  `json:"timezone"` // ???
	APIRoot              *url.URL                `json:"-"`
	Version              VersionInfo             `json:"pureconnectVersion"`
	User                 User                    `json:"user"`
	DefaultWorkstationID string                  `json:"defaultWorkstationId"`
	StationSettings      StationSettings         `json:"stationSettings"`
	Status               SessionStatus           `json:"status"`
	Features             []SessionFeature        `json:"features"`
	Subscriptions        map[string]Subscription `json:"-"`
	eventStream          *EventStream            `json:"-"`
	Logger               *logger.Logger          `json:"-"`
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
		Subscriptions:  map[string]Subscription{},
		eventStream:    NewEventStream(),
		Logger:         log,
	}
}

// GetID tells the ID
//
// implements Identifiable
func (session Session) GetID() string {
	return session.ID
}

// IsConnected tells if the Session is connected to a PureConnect server
func (session Session) IsConnected() bool {
	return session.Status == ConnectedStatus || session.Status == DisconnectingStatus || session.Status == ChangingStatus
}

// Events gives the EventSource chan to read for new Server-Sent Events from PureConnect
func (session Session) Events() chan EventSource {
	return session.eventStream.Events
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
			Token                string           `json:"csrfToken"`
			SessionID            string           `json:"sessionId"`
			Alternates           []string         `json:"alternateHostList"`
			Server               string           `json:"icServer"`
			UserID               string           `json:"userID"`
			DisplayName          string           `json:"userDisplayName"`
			PasswordExpiredIn    int              `json:"daysUntilPasswordExpiration"`
			DefaultWorkstationID *string          `json:"defaultWorkstationId"`
			Features             []SessionFeature `json:"features"`
			Version              VersionInfo      `json:"version"`
		}{}

		err = session.sendPost("/connection?include=features,default-workstation,version",
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
				session.Servers[i] = core.Must(url.Parse(fmt.Sprintf("%s://%s:%s", server.Scheme, results.Alternates[i], server.Port())))
			}
		}
		if results.DefaultWorkstationID != nil && len(*results.DefaultWorkstationID) > 0 {
			session.DefaultWorkstationID = *results.DefaultWorkstationID
		}
		session.User.ID = results.UserID
		session.User.DisplayName = results.DisplayName
		session.Status = ConnectedStatus
		session.Features = results.Features
		session.Logger = session.Logger.Record("session", session.ID)

		err = session.startMessageProcessing()
		if err != nil {
			return err
		}

		err = session.Subscribe(UserStatusMessage{}, UserStatusSubscription{
			UserIDs: IDList(session.User),
		})
		if err != nil {
			return err
		}
		return nil
	}
	session.Status = DisconnectedStatus
	return err
}

// Disconnect disconnects the Session from PureConnect
//
// All subscriptions are canceled prior to the disconnection.
// Also disconnected the Station, if any.
func (session *Session) Disconnect() error {
	log := session.Logger.Child(nil, "disconnect")
	if !session.IsConnected() || session.Status == DisconnectingStatus {
		return nil
	}
	var errs errors.MultiError
	session.Status = DisconnectingStatus
	for key, subscription := range session.Subscriptions {
		if err := subscription.Unsubscribe(session); err != nil {
			errs.Append(err)
		} else {
			log.Debugf("Unsubcribed from %s", subscription.GetType())
			delete(session.Subscriptions, key)
		}
	}
	if session.StationSettings != nil {
		if err := session.StationSettings.Disconnect(session); err != nil {
			errs.Append(err)
		} else {
			log.Debugf("Disconnected from station %s", session.StationSettings)
			session.StationSettings = nil
		}
	}
	if !errs.IsEmpty() {
		return errs.AsError()
	}

	session.stopMessageProcessing()
	log.Debugf("Message Processing stopped")

	errs.Append(session.sendDelete("/connection"))
	if !errs.IsEmpty() {
		log.Debugf("Disconnected from %s", session.APIRoot.Host)
		session.Status = DisconnectedStatus
		session.ID = ""
	}
	return errs.AsError()
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

func (session *Session) startMessageProcessing() error {
	if session.HasSupportWithAtLeastVersion("messaging", 2) { // Server-Sent Events are supported
		return session.eventStream.Connect(session, "/messaging/messages")
	}
	return errors.NotImplemented.WithStack()
}

func (session *Session) stopMessageProcessing() {
	if session.HasSupportWithAtLeastVersion("messaging", 2) { // Server-Sent Events are supported
		session.eventStream.Disconnect()
	}
}
