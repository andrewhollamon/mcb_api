package api

import (
	"fmt"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/dbservice"
	"net/http"
	"time"

	"github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/logging"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/tracing"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/uuidservice"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func SetupRouter() *gin.Engine {
	// Create router without default middleware
	r := gin.New()

	// Add our custom middleware
	r.Use(tracing.RequestIDMiddleware())
	r.Use(logging.DetailedRequestLoggingMiddleware())
	r.Use(apierror.ErrorHandlingMiddleware())

	err := r.SetTrustedProxies(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error trying to disable trusted proxies")
		return nil
	}

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/api/v1/checkbox/:checkboxNbr/status", getStatus)
	r.POST("/api/v1/checkbox/:checkboxNbr/check/:userUuid", checkboxCheck)
	r.POST("/api/v1/checkbox/:checkboxNbr/uncheck/:userUuid", checkboxUncheck)

	return r
}

func getServerName() string {
	return config.GetStringWithDefault("SERVER_NAME", "unknown")
}

func getServerIp() string {
	return config.GetStringWithDefault("SERVER_IP", "unknown")
}

func getStatus(c *gin.Context, checkboxNbr int) (bool, time.Time, apierror.APIError) {
	checked, lastUpdated, err := dbservice.GetCheckboxStatus(c, checkboxNbr)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get checkbox status for checkbox %d", checkboxNbr)
		return false, time.UnixMilli(0), apierror.WrapWithCodeFromConstants(err, apierror.ErrDatabaseError, fmt.Sprintf("failed to get checkbox status for checkbox %d", checkboxNbr))
	}

	return checked, lastUpdated, nil
}

func checkboxCheck(c *gin.Context) {
	logging.LogAPICall(c, "checkbox_check", map[string]interface{}{
		"checkbox_nbr": c.Param("checkboxNbr"),
		"user_uuid":    c.Param("userUuid"),
	})

	checkboxNbr, userUuid, err := validateParams(c)
	if err != nil {
		apiErr := apierror.ValidationError(err.Error())
		apierror.AbortWithAPIError(c, apiErr)
		return
	}

	requestUuid, err := uuidservice.NewRequestUuid()
	if err != nil {
		apiErr := apierror.InternalError("Failed to generate request UUID")
		apierror.AbortWithAPIError(c, apiErr.WithStackTrace())
		return
	}

	payload := queueservice.CheckboxActionPayload{
		Action:      "check",
		CheckboxNbr: checkboxNbr,
		UserUuid:    userUuid.String(),
		RequestUuid: requestUuid.String(),
		RequestTime: time.Now(),
		UserIp:      c.RemoteIP(),
		ApiServer:   getServerName(),
	}

	// Use context-aware queue service call
	ctx := tracing.PropagateTraceID(c)
	_, apiErr := queueservice.PublishCheckboxAction(ctx, payload)
	if apiErr != nil {
		apierror.AbortWithAPIError(c, apiErr.WithStackTrace())
		return
	}

	response := gin.H{
		"request_uuid": requestUuid,
		"result":       "success",
	}

	logging.LogAPIResponse(c, "checkbox_check", http.StatusOK, response)
	c.JSON(http.StatusOK, response)
}

func checkboxUncheck(c *gin.Context) {
	logging.LogAPICall(c, "checkbox_uncheck", map[string]interface{}{
		"checkbox_nbr": c.Param("checkboxNbr"),
		"user_uuid":    c.Param("userUuid"),
	})

	checkboxNbr, userUuid, err := validateParams(c)
	if err != nil {
		apiErr := apierror.ValidationError(err.Error())
		apierror.AbortWithAPIError(c, apiErr)
		return
	}

	requestUuid, err := uuidservice.NewRequestUuid()
	if err != nil {
		apiErr := apierror.InternalError("Failed to generate request UUID")
		apierror.AbortWithAPIError(c, apiErr.WithStackTrace())
		return
	}

	payload := queueservice.CheckboxActionPayload{
		Action:      "uncheck",
		CheckboxNbr: checkboxNbr,
		UserUuid:    userUuid.String(),
		RequestUuid: requestUuid.String(),
		RequestTime: time.Now(),
		UserIp:      c.RemoteIP(),
		ApiServer:   getServerName(),
	}

	// Use context-aware queue service call
	ctx := tracing.PropagateTraceID(c)
	_, apiErr := queueservice.PublishCheckboxAction(ctx, payload)
	if apiErr != nil {
		apierror.AbortWithAPIError(c, apiErr.WithStackTrace())
		return
	}

	response := gin.H{
		"request_uuid": requestUuid,
		"result":       "success",
	}

	logging.LogAPIResponse(c, "checkbox_uncheck", http.StatusOK, response)
	c.JSON(http.StatusOK, response)
}
