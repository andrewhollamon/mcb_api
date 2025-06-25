package api

import (
	"fmt"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/uuidservice"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

func SetupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		fmt.Printf("Error trying to disabled trusted proxies: %v", err)
		os.Exit(1)
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
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		return "unknown"
	} else {
		return serverName
	}
}

func getServerIp() string {
	serverIp := os.Getenv("SERVER_IP")
	if serverIp == "" {
		return "unknown"
	} else {
		return serverIp
	}
}

func getStatus(c *gin.Context) {
	// check caching DB lookup service
}

func checkboxCheck(c *gin.Context) {
	checkboxNbr, userUuid, err := validateParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	requestUuid, err := uuidservice.NewRequestUuid()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	payload := queueservice.CheckboxActionPayload{
		Action:      "check",
		CheckboxNbr: checkboxNbr,
		UserUuid:    userUuid.String(),
		RequestUuid: requestUuid.String(),
		RequestTime: time.Now(),
		Hostname:    getServerName(),
		Hostip:      getServerIp(),
	}

	err = queueservice.ProduceCheckboxActionMessage(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(200, gin.H{
		"request_uuid": requestUuid,
		"result":       "success",
	})
}

func checkboxUncheck(c *gin.Context) {

}
