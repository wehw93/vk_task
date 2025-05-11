package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestGoodSubscribePublish(t *testing.T) {
	bus := NewSubPub()
	var wg sync.WaitGroup
	wg.Add(1)
	_, err := bus.Subscribe("test", func(msg interface{}) {
		if msg != "hello" {
			t.Errorf("Expected 'hello', got %v", msg)
		}
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	err = bus.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	wg.Wait()
}

func TestUnsubscribePublish(t *testing.T) {
	bus := NewSubPub()
	called := false
	sub, err := bus.Subscribe("test", func(msg interface{}) {
		called = true
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	sub.Unsubscribe()
	err = bus.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if called {
		t.Error("Handler was called after unsubscribe")
	}
}

func TestCloseBus(t *testing.T) {
	bus := NewSubPub()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	_, err := bus.Subscribe("test", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	err = bus.Publish("test", "hello")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	err = bus.Close(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	wg.Wait()
}
func TestPublishToNonexistentTopic(t *testing.T) {
    bus := NewSubPub()
    err := bus.Publish("nonexistent", "message")
    if err != nil {
        t.Errorf("Publish to nonexistent topic should not return error, got: %v", err)
    }
}

func TestMultipleSubscriptionsUnsubscribe(t *testing.T) {
    bus := NewSubPub()
    var (
        callCount int
        mu        sync.Mutex
    )
    subs := make([]Subscription, 5)
    for i := 0; i < 5; i++ {
        sub, err := bus.Subscribe("test", func(msg interface{}) {
            mu.Lock()
            callCount++
            mu.Unlock()
        })
        if err != nil {
            t.Fatalf("Subscribe failed: %v", err)
        }
        subs[i] = sub
    }
    for i := 0; i < 3; i++ {
        subs[i].Unsubscribe()
    }
    err := bus.Publish("test", "hello")
    if err != nil {
        t.Fatalf("Publish failed: %v", err)
    }
    time.Sleep(100 * time.Millisecond)
    if callCount != 2 {
        t.Errorf("Expected 2 handlers to be called, got %d", callCount)
    }
}

func TestConcurrentAccess(t *testing.T) {
    bus := NewSubPub()
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            sub, err := bus.Subscribe("test", func(msg interface{}) {})
            if err != nil {
                t.Errorf("Subscribe failed: %v", err)
                return
            }
            err = bus.Publish("test", i)
            if err != nil {
                t.Errorf("Publish failed: %v", err)
            }
            sub.Unsubscribe()
        }(i)
    }
    wg.Wait()
    err := bus.Close(context.Background())
    if err != nil {
        t.Errorf("Close failed: %v", err)
    }
}