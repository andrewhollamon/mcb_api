package api

import (
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func wsGetAllCheckboxes(ctx *gin.Context) {
	var upgrader = websocket.Upgrader{}

	w, r := ctx.Writer, ctx.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		apiErr := apierror.InternalError("Failed to generate client UUID")
		apierror.AbortWithAPIError(ctx, apiErr.WithStackTrace())
		return
	}
	defer conn.Close()

	for {

	}
}
