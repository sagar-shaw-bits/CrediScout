package ask

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	asksrv "github.com/sagar-shaw-bits/CrediScout/internal/services/ask"
)

type AskEndpointHandler struct {
	service asksrv.Service
	log     *slog.Logger
}

func NewAskEndpointHandler(service asksrv.Service, log *slog.Logger) *AskEndpointHandler {
	if log == nil {
		log = slog.Default()
	}
	return &AskEndpointHandler{service: service, log: log}
}

func (h *AskEndpointHandler) AskEndpointHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req asksrv.AskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := h.service.Ask(c.Request.Context(), req)
		if err != nil {
			if errors.Is(err, asksrv.ErrEmptyQuestion) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			h.log.Error("ask failed", "err", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to answer question"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
