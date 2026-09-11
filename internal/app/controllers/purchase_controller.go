package controllers

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"github.com/0xdiaz/oneticket-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// maxEventID is the largest value events.id can hold: the column is SERIAL,
// which is int4. Anything above this is rejected here rather than handed to
// the driver, so a database encoding error never becomes the response.
const maxEventID = 2147483647

// PurchaseController handles ticket checkout.
//
// Every utils.* call below passes nil for the error argument on purpose.
// utils.HandleErrors copies err.Error() into the response body, so passing a
// wrapped error would ship driver text to the client. Detail goes to the log;
// the client gets a fixed message. This deliberately differs from
// EventController, which does pass err -- do not "harmonise" the two.
type PurchaseController struct {
	service *services.PurchaseService
}

// NewPurchaseController creates a new PurchaseController instance.
func NewPurchaseController(service *services.PurchaseService) *PurchaseController {
	return &PurchaseController{
		service: service,
	}
}

// Purchase buys one ticket of an event for the authenticated user.
//
// POST /api/v1/events/:id/purchase
func (ctrl *PurchaseController) Purchase(c *gin.Context) {
	ctx, start := logger.LogStart(c.Request.Context(), "PurchaseController.Purchase")

	eventID, ok := parseEventID(c)
	if !ok {
		logger.LogFinish(ctx, "PurchaseController.Purchase", nil, start)
		utils.BadRequest(c, nil, "Invalid event id")
		return
	}

	// The buyer is always the bearer of the token, never a value from the
	// request. There is no way to purchase on someone else's behalf.
	userID, ok := currentUserID(c)
	if !ok {
		logger.Warnf("purchase reached the handler without a user id in context")
		logger.LogFinish(ctx, "PurchaseController.Purchase", nil, start)
		utils.Unauthorized(c, nil, "Unauthorized")
		return
	}

	purchase, err := ctrl.service.Purchase(ctx, eventID, userID)
	if err != nil {
		ctrl.respondError(c, ctx, start, err)
		return
	}

	logger.LogFinish(ctx, "PurchaseController.Purchase", nil, start)
	utils.Created(c, purchase, "Ticket purchased successfully")
}

// respondError maps a service sentinel to its status. Both sold-out and
// sale-not-open are 409: the event's state conflicts with what was asked for.
// They are told apart by the message, not by the code.
func (ctrl *PurchaseController) respondError(c *gin.Context, ctx context.Context, start time.Time, err error) {
	logger.LogFinish(ctx, "PurchaseController.Purchase", err, start)

	switch {
	case errors.Is(err, services.ErrEventNotFound):
		utils.NotFound(c, nil, "Event not found")
	case errors.Is(err, services.ErrNoTicketsAvailable):
		utils.Conflict(c, nil, "Event is sold out")
	case errors.Is(err, services.ErrSaleNotOpen):
		utils.Conflict(c, nil, "Sale has not started yet")
	default:
		logger.Errorf("failed to purchase ticket: %v", err)
		utils.InternalServerError(c, nil, "Failed to purchase ticket")
	}
}

// parseEventID reads :id and rejects anything the events table could not hold,
// before the value reaches the database.
func parseEventID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")

	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		logger.Warnf("invalid event id %q: %v", raw, err)
		return 0, false
	}
	if id < 1 || id > maxEventID {
		logger.Warnf("event id %d outside the range of the id column", id)
		return 0, false
	}
	return uint(id), true
}

// currentUserID reads the id AuthMiddleware put in the context.
func currentUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := value.(uint)
	if !ok || id == 0 {
		return 0, false
	}
	return id, true
}
