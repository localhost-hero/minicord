package rooms

import "testing"

func TestValidateRoomNameAcceptsValidName(t *testing.T) {
	if err := ValidateRoomName("team-standup_1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateRoomNameRejectsTooShort(t *testing.T) {
	if err := ValidateRoomName("ab"); err == nil {
		t.Fatal("expected an error for a too-short room name")
	}
}

func TestValidateRoomNameRejectsTooLong(t *testing.T) {
	name := ""
	for i := 0; i < 65; i++ {
		name += "a"
	}
	if err := ValidateRoomName(name); err == nil {
		t.Fatal("expected an error for a too-long room name")
	}
}

func TestValidateRoomNameRejectsInvalidCharacters(t *testing.T) {
	if err := ValidateRoomName("room with spaces"); err == nil {
		t.Fatal("expected an error for a room name with spaces")
	}
}

func TestValidateIdentityAcceptsValidIdentity(t *testing.T) {
	if err := ValidateIdentity("user-42"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateIdentityRejectsInvalidCharacters(t *testing.T) {
	if err := ValidateIdentity("user/42"); err == nil {
		t.Fatal("expected an error for an identity with a slash")
	}
}
