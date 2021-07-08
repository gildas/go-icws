package icws

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

// EventSource describe an Server-Sent Event
type EventSource struct {
	Type    string  `json:"__type"`
	ID      string  `json:"eventId"`
	Message Message `json:"message"`
}

// EventStream describes an EventSource processor
type EventStream struct {
	Events    chan EventSource // listen to this to process EventSource
	closeChan chan struct{}
	Logger    *logger.Logger
}

// NewEventStream creates a new EventStream
func NewEventStream() *EventStream {
	return &EventStream{
		Events:    make(chan EventSource),
		closeChan: make(chan struct{}),
	}
}

// Connect connects to the PureConnect Server-Sent Event Service of the Session
//
// Do not forget to call the closeEventStream when you are done
func (stream *EventStream) Connect(session *Session, path string) error {
	if stream.Logger == nil {
		stream.Logger = session.Logger.Child("stream", stream)
	}
	log := stream.Logger.Child(nil, "messageprocessing")

	endpoint, err := session.endpoint("/messaging/messages")
	if err != nil {
		return err
	}

	log.Tracef("HTTP %s %s", http.MethodGet, endpoint.String())
	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Set("UserAgent", "GENESYS ICWS GO Client v"+VERSION)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Accept-Language", session.Language)
	req.Header.Set("Use-Credentials", "include") // Equivalent to JavaScript: EventSource(url, { withCredentials: true })
	if len(session.Token) > 0 {
		req.Header.Set("ININ-ICWS-CSRF-Token", session.Token)
	}
	for _, cookie := range session.Cookies {
		req.AddCookie(cookie)
	}
	req.Close = true

	log.Tracef("Request Headers: %#v", req.Header)
	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Tracef("Response %s in %s", res.Status, duration)
	log.Tracef("Response Headers: %#v", res.Header)

	// The EventSource processor
	go func() {
		defer close(stream.Events)

		// Closes the response's body when the stream closes
		go func() {
			<-stream.closeChan
			res.Body.Close()
		}()

		scanner := bufio.NewScanner(res.Body)

		event := EventSource{}
		data := strings.Builder{}

		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				log.Debugf("Unmarshaling: %s", data.String())
				event.Message, err = UnmarshalMessage([]byte(data.String()))
				if err != nil {
					log.Errorf("Unknown Message: %s", data.String(), err)
					continue
				}
				// send the EventSource to the chan for processing by the application
				stream.Events <- event
				event = EventSource{} // Create a new EventSource to fill in
				data  = strings.Builder{}
			}

			field, value := stream.analyzeLine(line)
			switch field {
			// See: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
			case "comment":
				log.Tracef("Comment: %s", value)
			case "ping":
				if core.GetEnvAsBool("TRACE_PING", false) {
					log.Tracef("Received a ping")
				}
			case "id":
				event.ID = value
			case "event":
				event.Type = value
			case "data":
				data.WriteString(value)
			case "retry":
				log.Warnf("Not Implemented, Event Retry: %s", value)
				// This is the timeout in ms to use when reconnecting to PureConnect's SSE if the connection gets closed
			default:
				if len(value) > 0 {
					log.Tracef("Ignoring %s: %#+v", field, value)
				}
			}
		}
		if err != nil && errors.Is(scanner.Err(), net.ErrClosed) {
			log.Errorf("Failed to scan", scanner.Err())
		}
		res.Body.Close()
	}()

	return nil
}

// Disconnect disconnects the EventStream
func (stream EventStream) Disconnect() {
	stream.closeChan <- struct{}{}
}

func (stream EventStream) analyzeLine(line string) (field string, value string) {
	if string(line) == ":ping" {
		return "ping", ""
	}
	pos := strings.Index(line, ":")
	if pos == 0 {
		return "comment", line
	}
	if pos > -1 {
		return string(line[:pos]), strings.TrimSpace(line[pos+1:])
	}
	return "unknown", line
}
