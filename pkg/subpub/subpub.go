package subpub

import (
	"context"
	"errors"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPubBroker struct {
	subscribers map[string][]*subscription
	mutex       sync.RWMutex
	is_closed   bool
	wg          sync.WaitGroup
}

type subscription struct {
	subject string
	handler MessageHandler
	bus     *SubPubBroker
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

func NewSubPub() SubPub {
	return &SubPubBroker{
		subscribers: make(map[string][]*subscription),
	}
}
func (s *subscription) Unsubscribe() {
	s.bus.unsubscribe(s)
}
func (s *SubPubBroker) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sub := &subscription{
		subject: subject,
		handler: cb,
		bus:     s,
	}
	if s.is_closed {
		return nil, errors.New("subpub is closed")
	}
	s.subscribers[subject] = append(s.subscribers[subject], sub)
	return sub, nil
}

func (s *SubPubBroker) unsubscribe(sub *subscription) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	subs, ok := s.subscribers[sub.subject]
	if !ok {
		return
	}
	for i, candidate := range subs {
		if candidate == sub {
			s.subscribers[sub.subject] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(s.subscribers[sub.subject]) == 0 {
		delete(s.subscribers, sub.subject)
	}
}

func (s *SubPubBroker) Publish(subject string, msg interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.is_closed {
		return errors.New("subpub is closed")
	}
	subs, ok := s.subscribers[subject]
	if !ok {
		return nil
	}
	for _, sub := range subs {
		s.wg.Add(1)
		go func(sub *subscription) {
			defer s.wg.Done()
			sub.handler(msg)
		}(sub)
	}
	return nil
}

func (s *SubPubBroker) Close(ctx context.Context) error {
	s.mutex.Lock()
	if s.is_closed {
		s.mutex.Unlock()
		return nil
	}
	s.is_closed = true
	s.mutex.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
