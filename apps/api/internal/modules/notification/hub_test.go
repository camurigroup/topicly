package notification

import "testing"

func TestHubPublishReachesOnlyTargetUser(t *testing.T) {
	h := NewHub()
	a, unsubA := h.Subscribe("a")
	defer unsubA()
	b, unsubB := h.Subscribe("b")
	defer unsubB()

	h.Publish("a", StreamEvent{Kind: "notification", Unread: 3})

	select {
	case ev := <-a:
		if ev.Unread != 3 {
			t.Fatalf("unread = %d, want 3", ev.Unread)
		}
	default:
		t.Fatal("subscriber a did not receive the event")
	}
	select {
	case <-b:
		t.Fatal("subscriber b must not receive a's event")
	default:
	}
}

func TestHubUnsubscribeReleasesUser(t *testing.T) {
	h := NewHub()
	_, unsub := h.Subscribe("a")
	if !h.Connected("a") {
		t.Fatal("expected a to be connected")
	}
	unsub()
	unsub() // idempotent
	if h.Connected("a") {
		t.Fatal("expected a to be disconnected")
	}
	h.Publish("a", StreamEvent{Kind: "sync"}) // must not panic
}

func TestHubDropsWhenConsumerIsSlow(t *testing.T) {
	h := NewHub()
	_, unsub := h.Subscribe("a")
	defer unsub()
	for i := 0; i < 100; i++ { // buffer is 16; publisher must never block
		h.Publish("a", StreamEvent{Kind: "sync"})
	}
}
