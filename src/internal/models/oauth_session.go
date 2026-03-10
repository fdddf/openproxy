package models

import "time"

// OauthSession model (for storing OAuth2 session data including PKCE parameters)
type OauthSession struct {
	CommonFields

	SessionID           string    `json:"session_id"`
	ProviderID          uint      `json:"provider_id"`
	State               string    `json:"state"`
	CodeVerifier        string    `json:"code_verifier"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
	ExpiresAt           time.Time `json:"expires_at"`
}
