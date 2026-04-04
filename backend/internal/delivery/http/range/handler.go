package rangehandler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	rangeservice "github.com/AbelHaro/url-shortener/backend/internal/service/range"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *rangeservice.Service
}

func NewHandler(svc *rangeservice.Service) *Handler {
	return &Handler{Service: svc}
}

// Allocate a range of IDs
// @Summary Allocate a range of IDs
// @Description Allocate a range of IDs for URL shortening
// @Tags Range Allocation
// @Accept json
// @Produce json
// @Success 200 {object} dtos.V1AllocateRangeResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /ranges/allocate [post]
func (h *Handler) Allocate(c *gin.Context) {
	var req dtos.V1AllocateRangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid owner_id"})
		return
	}

	rangeAllocated, err := h.Service.AllocateRange(ownerID)

	fmt.Printf("Allocated range for owner %s: %+v\n", ownerID, rangeAllocated)

	if err != nil {
		fmt.Printf("Error allocating range for owner %s: %v\n", ownerID, err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dtos.V1AllocateRangeResponse{
		ID:            rangeAllocated.ID,
		Start:         rangeAllocated.Start,
		Last:          rangeAllocated.Last,
		CurrentOffset: rangeAllocated.CurrentOffset,
	})
}

// Update the offset of a range
// @Summary Update the offset of a range. Given a range of 100_000, if the offset is 1_000, each 1_000 numbers incremented in the range will be marked as used. So if the offset is 1_000, the first 1_000 numbers in the range will be marked as used, this is useful to not loose IDs in case of a failure in the URL shortening service, so the next time the service is restarted, it can continue from the last offset instead of starting from the beginning of the range avoiding to loose IDs.
// @Description Update the offset of a range
// @Tags Range Allocation
// @Accept json
// @Param id path string true "Range ID"
// @Param offset query int true "New offset value"
// @Success 201
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /ranges/{id}/offset [put]
func (h *Handler) UpdateOffset(c *gin.Context) {
	idParam := c.Param("id")
	rangeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid range ID"})
		return
	}

	var req dtos.V1UpdateRangeOffsetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid owner_id"})
		return
	}

	err = h.Service.UpdateRangeOffset(rangeID, ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRangeAllocFailed):
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "failed to allocate range"})
	case errors.Is(err, domain.ErrInvalidRange):
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid range"})
	case errors.Is(err, domain.ErrRangeConsumed):
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "range consumed"})
	case errors.Is(err, domain.ErrRangeNotFound):
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "range not found"})
	case errors.Is(err, domain.ErrInternal):
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "internal server error"})
	default:
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "internal server error"})
	}
}
