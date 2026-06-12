package auth

import "time"

// TokenStore is the interface for token revocation storage.
// Implementations must be safe for concurrent use.
type TokenStore interface {
	// Revoke adds a token to the revocation set, expiring at the given time.
	Revoke(tokenStr string, expiresAt time.Time)
	// IsRevoked returns true if the token has been revoked and has not yet expired.
	IsRevoked(tokenStr string) bool
}
