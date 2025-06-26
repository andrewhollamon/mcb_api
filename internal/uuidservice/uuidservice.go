package uuidservice

import (
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/google/uuid"
)

// wrapping the uuidv7 gen in a service may be total overkill here, but doing it in case:
// 1. Want to swap out to different provider in the future
// 2. Want to produce different kinds of UUIDs for different purposes, ie requestUuids may not need to be v7s

func newUuid() (uuid.UUID, error) {
	return uuid.NewV7()
}

func NewClientUuid() (uuid.UUID, apierror.APIError) {
	u, err := newUuid()
	if err != nil {
		return u, apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, "could not generate UUID for ClientUuid")
	}
	return u, nil
}

func NewRequestUuid() (uuid.UUID, apierror.APIError) {
	u, err := newUuid()
	if err != nil {
		return u, apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, "could not generate UUID for RequestUuid")
	}
	return u, nil
}
