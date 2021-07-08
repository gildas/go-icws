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
	_, err := session.send(http.MethodPost, path, nil, nil, payload, results)
	return err
}

func (session *Session) sendGet(path string, results interface{}) error {
	_, err := session.send(http.MethodGet, path, nil, nil, nil, results)
	return err
}

func (session *Session) sendPut(path string, payload interface{}, results interface{}) error {
	_, err := session.send(http.MethodPut, path, nil, nil, payload, results)
	return err
}

func (session *Session) sendDelete(path string) error {
	_, err := session.send(http.MethodDelete, path, nil, nil, nil, nil)
	return err
}

func (session *Session) send(method, path string, headers map[string]string, queryParameters map[string]string, payload interface{}, results interface{}) (response *request.ContentReader, err error) {
	log := session.Logger.Child(nil, "send_"+strings.ToLower(method))

	if !session.IsConnected() && session.Status != ConnectingStatus && len(session.Token) == 0 {
		if err = session.Connect(); err != nil {
			return nil, err
		}
	}
	endpoint, err := session.endpoint(path)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Accept-Language"] = session.Language
	if len(session.Token) > 0 {
		headers["ININ-ICWS-CSRF-Token"] = session.Token
	}
	response, err = request.Send(&request.Options{
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
		return response, err
	}
	if results != nil {
		log.Tracef("Results: %+#v", results)
	}
	if len(response.Cookies) > 0 {
		session.Cookies = response.Cookies
	}
	return response, nil
}
