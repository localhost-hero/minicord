// Package adminsession issues and validates signed, stateless admin session
// tokens without requiring server-side session storage.
package adminsession

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DefaultSessionTTL is the lifetime of an admin session token when no override is given.
const DefaultSessionTTL = 24 * time.Hour

// CheckCredentials reports whether the given username and password match the
// configured admin account. passwordHash must be a bcrypt hash.
func CheckCredentials(username, password, expectedUsername, passwordHash string) bool {
	if subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) != 1 {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

// NewToken issues a signed session token that expires after ttl. A zero ttl
// uses DefaultSessionTTL; a negative ttl produces an already-expired token,
// which is useful for tests.
func NewToken(secret string, ttl time.Duration) string {
	if ttl == 0 {
		ttl = DefaultSessionTTL
	}
	expiry := time.Now().Add(ttl).Unix()
	payload := strconv.FormatInt(expiry, 10)
	return payload + "." + sign(secret, payload)
}

// ValidateToken reports whether a session token is well-formed, correctly
// signed, and not expired.
func ValidateToken(secret, token string) bool {
	dot := indexOfDot(token)
	if dot < 0 {
		return false
	}
	payload := token[:dot]
	signature := token[dot+1:]

	expected := sign(secret, payload)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) != 1 {
		return false
	}

	expiry, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return false
	}

	return time.Now().Unix() < expiry
}

func indexOfDot(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
