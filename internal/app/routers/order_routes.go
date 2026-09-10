package routers

import (
	"github.com/0xdiaz/oneticket-api/internal/app/controllers"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/gin-gonic/gin"
)

// RegisterOrderRoutes registers ticket purchase routes under the given group.
//
// The group passed in must already carry the auth middleware: buying a ticket
// requires knowing who is buying.
func RegisterOrderRoutes(group *gin.RouterGroup, orderService *services.OrderService) {
	orderController := controllers.NewOrderController(orderService)
	group.POST("/events/:id/purchase", orderController.Purchase)
}
