package icws

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
)

func (session Session) endpoint(path string) (endpoint *url.URL, err error) {
	if session.APIRoot == nil {
		return nil, errors.CreationFailed.With("endpoint", path).Wrap(errors.ArgumentMissing.With("apiRoot"))
	}
	if len(session.ID) == 0 {
		endpoint, err = session.APIRoot.Parse(fmt.Sprintf("/icws%s", path))
	} else {
		endpoint, err = session.APIRoot.Parse(fmt.Sprintf("/icws/%s%s", session.ID, path))
	}
	return endpoint, errors.CreationFailed.With("endpoint", path).Wrap(err)
}

func (session *Session) sendPost(path string, payload interface{}, results interface{}) error {
	return session.send(http.MethodPost, path, nil, payload, results)
}

func (session *Session) sendGet(path string, results interface{}) error {
	return session.send(http.MethodGet, path, nil, nil, results)
}

func (session *Session) sendPut(path string, payload interface{}, results interface{}) error {
	return session.send(http.MethodPut, path, nil, payload, results)
}

func (session *Session) sendDelete(path string) error {
	return session.send(http.MethodDelete, path, nil, nil, nil)
}

func (session *Session) sendPostWithParameters(path string, queryParameters map[string]string, payload interface{}, results interface{}) error {
	return session.send(http.MethodPost, path, queryParameters, payload, results)
}

func (session *Session) sendGetWithParameters(path string, queryParameters map[string]string, results interface{}) error {
	return session.send(http.MethodGet, path, queryParameters, nil, results)
}

func (session *Session) sendPutWithParameters(path string, queryParameters map[string]string, payload interface{}, results interface{}) error {
	return session.send(http.MethodPut, path, queryParameters, payload, results)
}

func (session *Session) sendDeleteWithParameters(path string, queryParameters map[string]string) error {
	return session.send(http.MethodDelete, path, queryParameters, nil, nil)
}

func (session *Session) send(method, path string, queryParameters map[string]string, payload interface{}, results interface{}) (err error) {
	log := session.Logger.Child(nil, "send_"+strings.ToLower(method))

	if !session.IsConnected() && session.Status != ConnectingStatus && len(session.Token) == 0 {
		if err = session.Connect(); err != nil {
			return err
		}
	}
	endpoint, err := session.endpoint(path)
	if err != nil {
		return err
	}
	headers := map[string]string{"Accept-Language": session.Language}
	if len(session.Token) > 0 {
		headers["ININ-ICWS-CSRF-Token"] = session.Token
	}
	response, err := request.Send(&request.Options{
		Context:    session.Context,
		UserAgent:  "GENESYS ICWS GO Client v" + VERSION,
		Method:     method,
		URL:        endpoint,
		Headers:    headers,
		Cookies:    session.Cookies,
		Parameters: queryParameters,
		Payload:    payload,
		Logger:     log,
	}, results)
	if err != nil {
		// TODO: On HTTP 503, we receive a list of alternate hosts that we should connect to
		return err
	}
	if results != nil {
		log.Tracef("Results: %+#v", results)
	}
	if len(response.Cookies) > 0 {
		session.Cookies = response.Cookies
	}
	return nil
}
