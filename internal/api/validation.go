package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
	"strings"
)

func validateCheckboxNumber(c *gin.Context) (int, error) {
	checkboxNbrStr := strings.TrimSpace(c.Param("checkboxNbr"))
	if checkboxNbrStr == "" {
		return 0, errors.New("validation error: Missing checkbox number parameter")
	}

	checkboxNbr, err := strconv.Atoi(checkboxNbrStr)
	if err != nil {
		return 0, fmt.Errorf("validation error: Checkbox number '%d' is not a valid integer", checkboxNbr)
	}

	if checkboxNbr <= 0 || checkboxNbr > 1000000 {
		return 0, fmt.Errorf("validation error: Checkbox number '%d' is out of range (1 - 1000000)", checkboxNbr)
	}

	return checkboxNbr, nil
}

func validateUserUUID(c *gin.Context) (uuid.UUID, error) {
	userUuidStr := strings.TrimSpace(c.Param("userUuid"))
	if userUuidStr == "" {
		return uuid.Nil, fmt.Errorf("validation error: Missing user UUID parameter")
	}

	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation error: User UUID '%v' is not a valid UUID", userUuidStr)
	}

	return userUuid, nil
}

func validateParams(c *gin.Context) (int, uuid.UUID, error) {
	checkboxNbr, err := validateCheckboxNumber(c)
	if err != nil {
		return 0, uuid.Nil, err
	}

	userUuid, err := validateUserUUID(c)
	if err != nil {
		return 0, uuid.Nil, err
	}

	return checkboxNbr, userUuid, nil
}
