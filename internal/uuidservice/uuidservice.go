package uuidservice

import (
	error0 "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/google/uuid"
)

// wrapping the uuidv7 gen in a service may be total overkill here, but doing it in case:
// 1. Want to swap out to different provider in the future
// 2. Want to produce different kinds of UUIDs for different purposes, ie requestUuids may not need to be v7s

func newUuid() (uuid.UUID, error) {
	return uuid.NewV7()
}

func NewClientUuid() (uuid.UUID, error0.APIError) {
	u, err := newUuid()
	if err != nil {
		return u, error0.NewInternalError("UUID Generation for NewClientUUID failed", err)
	}
	return u, nil
}

func NewRequestUuid() (uuid.UUID, error0.APIError) {
	u, err := newUuid()
	if err != nil {
		return u, error0.NewInternalError("UUID Generation for NewRequestUuid failed", err)
	}
	return u, nil
}
