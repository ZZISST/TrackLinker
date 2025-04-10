package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/dto"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/service"
	"gitlab.com/etherealSent/tracklinker-backend/collection/pkg/response"
)

type CollectionHandler struct {
	service service.IListeningService
}

func NewCollectionHandler(
	service service.IListeningService,
) *CollectionHandler {
	return &CollectionHandler{
		service: service,
	}
}

// MakeListening godoc
//
//	@Summary	make listening
//	@Tags		collection
//	@Produce	json
//	@Param		_	   body		 dto.RequestMakeListening	true	"Body"
//	@Success    200    {object}  dto.ResponseMakeListening
//	@Router     /api/v1/action/listen [post]
func (h *CollectionHandler) MakeListening(c *gin.Context) {
	var req dto.RequestMakeListening

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err, "Invalid parameters")
		return
	}

	if req.Name == "" || len(req.Artists) == 0 || req.TimeEnd.After(time.Now()) || req.TimeStart.After(time.Now()) {
		response.Error(c, http.StatusBadRequest, errors.New("link is required"), "Invalid parameters")
		return
	}

	status, err := h.service.MakeListening(c, req.UserID, req.Name, req.Artists, req.TimeStart, req.TimeEnd)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, status)
}
