package adminsession

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestCheckCredentialsAcceptsMatchingUsernameAndPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("s3cret-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckCredentials("admin", "s3cret-pass", "admin", string(hash)) {
		t.Fatal("expected matching credentials to be accepted")
	}
}

func TestCheckCredentialsRejectsWrongUsernameOrPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("s3cret-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if CheckCredentials("someone-else", "s3cret-pass", "admin", string(hash)) {
		t.Fatal("expected mismatched username to be rejected")
	}
	if CheckCredentials("admin", "wrong-pass", "admin", string(hash)) {
		t.Fatal("expected mismatched password to be rejected")
	}
}

func TestNewTokenProducesAValidatableToken(t *testing.T) {
	token := NewToken("secret", time.Minute)
	if !ValidateToken("secret", token) {
		t.Fatal("expected freshly issued token to validate")
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	token := NewToken("secret", time.Minute)
	if ValidateToken("other-secret", token) {
		t.Fatal("expected token signed with a different secret to be rejected")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	token := NewToken("secret", -time.Minute)
	if ValidateToken("secret", token) {
		t.Fatal("expected an expired token to be rejected")
	}
}

func TestValidateTokenRejectsMalformedToken(t *testing.T) {
	if ValidateToken("secret", "not-a-valid-token") {
		t.Fatal("expected a malformed token to be rejected")
	}
}
