package chat

import (
	"testing"
)

func newTestChatStore(t *testing.T) *Store {
	t.Helper()
	db := newTestDB(t)
	return NewStore(db)
}

func TestSaveMessageAndListMessages(t *testing.T) {
	store := newTestChatStore(t)

	if _, err := store.SaveMessage("team-standup", "user-1", "hello"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := store.SaveMessage("team-standup", "user-2", "hi there"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	messages, err := store.ListMessages("team-standup", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Body != "hello" || messages[1].Body != "hi there" {
		t.Fatalf("expected messages in insertion order, got %+v", messages)
	}
}

func TestSaveMessageRejectsEmptyBody(t *testing.T) {
	store := newTestChatStore(t)

	if _, err := store.SaveMessage("team-standup", "user-1", "   "); err == nil {
		t.Fatal("expected an error for an empty message body")
	}
}

func TestListMessagesOnlyReturnsMatchingRoom(t *testing.T) {
	store := newTestChatStore(t)

	if _, err := store.SaveMessage("room-a", "user-1", "in room a"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := store.SaveMessage("room-b", "user-1", "in room b"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	messages, err := store.ListMessages("room-a", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 1 || messages[0].Body != "in room a" {
		t.Fatalf("expected only room-a messages, got %+v", messages)
	}
}
