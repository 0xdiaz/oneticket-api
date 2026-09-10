package controllers

import (
	"errors"
	"strconv"

	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/0xdiaz/oneticket-api/pkg/logger"
	"github.com/0xdiaz/oneticket-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// OrderController handles ticket purchase endpoints.
type OrderController struct {
	service *services.OrderService
}

// NewOrderController creates a new OrderController instance.
func NewOrderController(service *services.OrderService) *OrderController {
	return &OrderController{service: service}
}

// Purchase sells one ticket of the event to the authenticated user.
//
// POST /api/v1/events/:id/purchase
func (ctrl *OrderController) Purchase(c *gin.Context) {
	ctx, start := logger.LogStart(c.Request.Context(), "OrderController.Purchase")

	// Set by AuthMiddleware. Zero means the guard did not run, which should be
	// impossible on this route — fail closed rather than selling to user 0.
	userID := c.GetUint("user_id")
	if userID == 0 {
		logger.LogFinish(ctx, "OrderController.Purchase", nil, start)
		utils.Unauthorized(c, nil, "Authentication required")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warnf("invalid event id: %v", err)
		logger.LogFinish(ctx, "OrderController.Purchase", err, start)
		utils.BadRequest(c, err, "Invalid event id")
		return
	}

	purchase, err := ctrl.service.PurchaseTicket(ctx, uint(eventID), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEventNotFound):
			logger.LogFinish(ctx, "OrderController.Purchase", err, start)
			utils.NotFound(c, err, "Event not found")
		case errors.Is(err, services.ErrSoldOut):
			logger.LogFinish(ctx, "OrderController.Purchase", err, start)
			utils.Conflict(c, err, "Sold out")
		default:
			logger.Errorf("failed to purchase ticket: %v", err)
			logger.LogFinish(ctx, "OrderController.Purchase", err, start)
			utils.InternalServerError(c, err, "Failed to purchase ticket")
		}
		return
	}

	logger.LogFinish(ctx, "OrderController.Purchase", nil, start)
	utils.Created(c, purchase, "Ticket purchased successfully")
}
