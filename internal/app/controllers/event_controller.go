package controllers

import (
	"errors"
	"strconv"

	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"github.com/0xdiaz/oneticket-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// EventController handles event browsing endpoints.
type EventController struct {
	service *services.EventService
}

// NewEventController creates a new EventController instance.
func NewEventController(service *services.EventService) *EventController {
	return &EventController{
		service: service,
	}
}

// List returns all events with live ticket availability.
//
// GET /api/v1/events
func (ctrl *EventController) List(c *gin.Context) {
	ctx, start := logger.LogStart(c.Request.Context(), "EventController.List")

	events, err := ctrl.service.List(ctx)
	if err != nil {
		logger.Errorf("failed to list events: %v", err)
		logger.LogFinish(ctx, "EventController.List", err, start)
		utils.InternalServerError(c, err, "Failed to list events")
		return
	}

	logger.LogFinish(ctx, "EventController.List", nil, start)
	utils.Ok(c, events, "Events retrieved successfully")
}

// Get returns a single event with live ticket availability.
//
// GET /api/v1/events/:id
func (ctrl *EventController) Get(c *gin.Context) {
	ctx, start := logger.LogStart(c.Request.Context(), "EventController.Get")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warnf("invalid event id: %v", err)
		logger.LogFinish(ctx, "EventController.Get", err, start)
		utils.BadRequest(c, err, "Invalid event id")
		return
	}

	event, err := ctrl.service.Get(ctx, uint(id))
	if err != nil {
		if errors.Is(err, services.ErrEventNotFound) {
			logger.LogFinish(ctx, "EventController.Get", err, start)
			utils.NotFound(c, err, "Event not found")
			return
		}
		logger.Errorf("failed to get event: %v", err)
		logger.LogFinish(ctx, "EventController.Get", err, start)
		utils.InternalServerError(c, err, "Failed to get event")
		return
	}

	logger.LogFinish(ctx, "EventController.Get", nil, start)
	utils.Ok(c, event, "Event retrieved successfully")
}
