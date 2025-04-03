package http

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/response"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/utils"
)

type ParsingHandler struct {
	service service.IMusicService
}

func NewParsingHandler(service service.IMusicService) *ParsingHandler {
	return &ParsingHandler{
		service: service,
	}
}

// GetMusicByID godoc
//
//	@Summary	get music details
//	@Tags		music
//	@Produce	json
//	@Param		id	path		string	true	"Music ID"
//	@Success	200	{object}	dto.MusicTrack
//	@Router		/api/v1/music/{id} [get]
func (h *ParsingHandler) GetMusicByID(c *gin.Context) {
	var res dto.MusicTrack

	musicId := c.Param("id")

	music, err := h.service.GetTrackByID(c, musicId)
	if err != nil {
		response.Error(c, http.StatusNotFound, err, "Not found")
		return
	}

	utils.Copy(&res, &music)
	response.JSON(c, http.StatusOK, res)
}

// FindMusic godoc
//
//	@Summary	get music
//	@Tags		music
//	@Produce	json
//	@Param		_	   body		 dto.FindMusicReq	true	"Body"
//	@Success    200    {object}  dto.MusicResultDto
//	@Router     /api/v1/music/find [post]
func (h *ParsingHandler) FindMusic(c *gin.Context) {
	var req dto.FindMusicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err, "Invalid parameters")
		return
	}

	if req.Link == "" {
		response.Error(c, http.StatusBadRequest, errors.New("link is required"), "Invalid parameters")
		return
	}

	parsedURL, err := url.Parse(req.Link)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err, "link parse error")
		return
	}

	music, err := h.service.FindMusic(*parsedURL)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, music)
}
