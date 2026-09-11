// Package livekit issues short-lived LiveKit access tokens for authorized room participants.
package livekit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// DefaultTokenTTL is the lifetime of an issued access token when no override is given.
const DefaultTokenTTL = 10 * time.Minute

type header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type videoGrant struct {
	Room         string `json:"room"`
	RoomJoin     bool   `json:"roomJoin"`
	CanPublish   bool   `json:"canPublish"`
	CanSubscribe bool   `json:"canSubscribe"`
}

type claims struct {
	Issuer     string     `json:"iss"`
	Subject    string     `json:"sub"`
	IssuedAt   int64      `json:"iat"`
	NotBefore  int64      `json:"nbf"`
	Expiration int64      `json:"exp"`
	JWTID      string     `json:"jti"`
	Video      videoGrant `json:"video"`
}

// NewAccessToken builds a signed LiveKit access token that grants the identity join,
// publish, and subscribe access to the given room for the provided time-to-live.
func NewAccessToken(apiKey, apiSecret, identity, room string, ttl time.Duration) (string, error) {
	if apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("livekit: api key and api secret must not be empty")
	}
	if identity == "" || room == "" {
		return "", fmt.Errorf("livekit: identity and room must not be empty")
	}
	if ttl <= 0 {
		ttl = DefaultTokenTTL
	}

	now := time.Now().UTC()
	tokenClaims := claims{
		Issuer:     apiKey,
		Subject:    identity,
		IssuedAt:   now.Unix(),
		NotBefore:  now.Unix(),
		Expiration: now.Add(ttl).Unix(),
		JWTID:      identity,
		Video: videoGrant{
			Room:         room,
			RoomJoin:     true,
			CanPublish:   true,
			CanSubscribe: true,
		},
	}

	headerSegment, err := encodeSegment(header{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", fmt.Errorf("livekit: encode header: %w", err)
	}

	claimsSegment, err := encodeSegment(tokenClaims)
	if err != nil {
		return "", fmt.Errorf("livekit: encode claims: %w", err)
	}

	signingInput := headerSegment + "." + claimsSegment
	signature := sign(signingInput, apiSecret)

	return signingInput + "." + signature, nil
}

func encodeSegment(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func sign(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
