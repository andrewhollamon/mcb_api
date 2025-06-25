package api

import (
	"fmt"
	error0 "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
	"strings"
)

func validateCheckboxNumber(c *gin.Context) (int, error0.APIError) {
	checkboxNbrStr := strings.TrimSpace(c.Param("checkboxNbr"))
	if checkboxNbrStr == "" {
		return 0, error0.FromGinContext(c, "MISSING_CHECKBOX_NUMBER",
			"Checkbox number parameter is required", 400)
	}

	checkboxNbr, err := strconv.Atoi(checkboxNbrStr)
	if err != nil {
		return 0, error0.FromGinContext(c, "INVALID_CHECKBOX_NUMBER",
			fmt.Sprintf("Checkbox number '%s' is not a valid integer", checkboxNbrStr), 400).
			WithContext("provided_value", checkboxNbrStr)
	}

	if checkboxNbr < 0 || checkboxNbr >= 1000000 {
		return 0, invalidCheckboxNumber(checkboxNbr)
	}

	return checkboxNbr, nil
}

func validateUserUUID(c *gin.Context) (uuid.UUID, error0.APIError) {
	userUuidStr := strings.TrimSpace(c.Param("userUuid"))
	if userUuidStr == "" {
		return uuid.Nil, error0.FromGinContext(c, "MISSING_USER_UUID",
			"User UUID parameter is required", 400)
	}

	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return uuid.Nil, invalidUserUUID(userUuidStr)
	}

	return userUuid, nil
}

func validateParams(c *gin.Context) (int, uuid.UUID, error0.APIError) {
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

func invalidCheckboxNumber(checkboxNbr int) error0.APIError {
	return error0.NewValidationError("INVALID_CHECKBOX_NUMBER",
		fmt.Sprintf("Checkbox number %d is invalid. Must be between 0 and 999999", checkboxNbr)).
		WithContext("checkbox_number", checkboxNbr).
		WithContext("valid_range", "0-999999")
}

func invalidUserUUID(userUuid string) error0.APIError {
	return error0.NewValidationError("INVALID_USER_UUID",
		fmt.Sprintf("User UUID '%s' is not a valid UUID", userUuid)).
		WithContext("provided_uuid", userUuid)
}

// from Claude Code errors package

func ExampleUsageInHandlers() {
	r := gin.New()

	r.Use(error0.TraceIDMiddleware())
	r.Use(error0.ErrorHandlerMiddleware())
	r.Use(error0.CORSMiddleware())

	r.GET("/api/v1/checkbox/:checkboxNbr/status", func(c *gin.Context) {
		checkboxNbr, err := validateCheckboxNumber(c)
		if err != nil {
			err.Respond(c)
			return
		}

		c.JSON(200, gin.H{
			"checkbox_number": checkboxNbr,
			"status":          "checked",
		})
	})

	r.POST("/api/v1/checkbox/:checkboxNbr/check/:userUuid", func(c *gin.Context) {
		checkboxNbr, userUuid, err := validateParams(c)
		if err != nil {
			err.Respond(c)
			return
		}

		c.JSON(200, gin.H{
			"checkbox_number": checkboxNbr,
			"user_uuid":       userUuid.String(),
			"action":          "checked",
		})
	})

	r.POST("/api/v1/checkbox/:checkboxNbr/uncheck/:userUuid", func(c *gin.Context) {
		checkboxNbr, userUuid, err := validateParams(c)
		if err != nil {
			err.Respond(c)
			return
		}

		c.JSON(200, gin.H{
			"checkbox_number": checkboxNbr,
			"user_uuid":       userUuid.String(),
			"action":          "unchecked",
		})
	})
}
