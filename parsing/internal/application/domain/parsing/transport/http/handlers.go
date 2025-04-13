package http

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/response"
)

type ParsingHandler struct {
	service service.IMusicService
}

func NewParsingHandler(service service.IMusicService) *ParsingHandler {
	return &ParsingHandler{
		service: service,
	}
}

// GetTrackByID godoc
//
//	@Summary	get track details
//	@Tags		track
//	@Produce	json
//	@Param		id	path		string	true	"Track ID"
//	@Success	200	{object}	model.Track
//	@Router		/api/v1/track/{id} [get]
func (h *ParsingHandler) GetTrackByID(c *gin.Context) {
	trackID := c.Param("id")

	track, err := h.service.GetTrackByID(c, trackID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err, "Not found")
		return
	}

	response.JSON(c, http.StatusOK, track)
}

// GetAlbumByID godoc
//
//	@Summary	get album details
//	@Tags		album
//	@Produce	json
//	@Param		id	path		string	true	"Album ID"
//	@Success	200	{object}	model.Album
//	@Router		/api/v1/album/{id} [get]
func (h *ParsingHandler) GetAlbumByID(c *gin.Context) {
	albumID := c.Param("id")

	album, err := h.service.GetAlbumByID(c, albumID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err, "Not found")
		return
	}

	response.JSON(c, http.StatusOK, album)
}

// GetAlbumTrackByID godoc
//
//	@Summary	get album's track details
//	@Tags		track
//	@Produce	json
//	@Param		album_id	path		string	true	"Album ID"
//	@Param		track_id	path		string	true	"Track ID"
//	@Success	200	{object}	model.Track
//	@Router		/api/v1/album/{album_id}/track/{track_id} [get]
func (h *ParsingHandler) GetAlbumTrackByID(c *gin.Context) {
	albumID := c.Param("album_id")
	trackID := c.Param("track_id")

	// Call the service to get the track details
	track, err := h.service.GetAlbumTrackByID(c, albumID, trackID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err, "Track not found")
		return
	}

	response.JSON(c, http.StatusOK, track)
}

// FindMusic godoc
//
//	@Summary	get music
//	@Tags		music
//	@Produce	json
//	@Param		_	   body		 dto.FindMusicReq	true	"Body"
//	@Success    200    {object}  dto.MusicResult
//	@Router     /api/v1/music/search [post]
func (h *ParsingHandler) SearchMusic(c *gin.Context) {
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
