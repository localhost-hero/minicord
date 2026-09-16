package chat

import "testing"

func TestHubDeliversPublishedMessageToSubscriber(t *testing.T) {
	hub := NewHub()
	ch := hub.Subscribe("team-standup")
	defer hub.Unsubscribe("team-standup", ch)

	hub.Publish("team-standup", Message{Room: "team-standup", Body: "hello"})

	select {
	case received := <-ch:
		if received.Body != "hello" {
			t.Fatalf("unexpected message body: %q", received.Body)
		}
	default:
		t.Fatal("expected a message to be delivered")
	}
}

func TestHubDoesNotDeliverToOtherRooms(t *testing.T) {
	hub := NewHub()
	ch := hub.Subscribe("room-a")
	defer hub.Unsubscribe("room-a", ch)

	hub.Publish("room-b", Message{Room: "room-b", Body: "hello"})

	select {
	case <-ch:
		t.Fatal("did not expect a message from a different room")
	default:
	}
}

func TestHubUnsubscribeClosesChannel(t *testing.T) {
	hub := NewHub()
	ch := hub.Subscribe("team-standup")
	hub.Unsubscribe("team-standup", ch)

	if _, open := <-ch; open {
		t.Fatal("expected the channel to be closed after unsubscribe")
	}
}
