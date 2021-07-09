package icws

import "github.com/gildas/go-core"

type Subscription interface {
	core.TypeCarrier
	Subscribe(session *Session, payload interface{}) error
	Unsubscribe(session *Session) error
}

func (session *Session) Subscribe(subscriber Subscription, payload interface{}) error {
	err := subscriber.Subscribe(session, payload)
	if err == nil {
		session.Subscriptions[subscriber.GetType()] = subscriber
	}
	return err
}

func (session *Session) Unsubscribe(unsubscriber Subscription) error {
	err := unsubscriber.Unsubscribe(session)
	if err == nil {
		delete(session.Subscriptions, unsubscriber.GetType())
	}
	return err
}
