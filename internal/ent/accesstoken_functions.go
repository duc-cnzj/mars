package ent

import (
	"time"
)

// Expired token is expired.
func (at *AccessToken) Expired() bool {
	return time.Now().After(at.ExpiredAt)
}
