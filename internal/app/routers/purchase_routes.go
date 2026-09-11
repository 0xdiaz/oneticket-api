package routers

import (
	"github.com/0xdiaz/oneticket-api/internal/app/controllers"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/gin-gonic/gin"
)

// RegisterPurchaseRoutes registers checkout under the given group.
//
// The group passed in must already carry AuthMiddleware: buying is not public,
// and attaching the guard here instead would leave the route reachable from
// any other group this function is ever called with.
//
// The wildcard must stay named :id. RegisterEventRoutes already registers
// /events/:id, and Gin panics while building the engine if the same position
// is given a second name -- which would take down every route, not just this
// one.
func RegisterPurchaseRoutes(group *gin.RouterGroup, purchaseService *services.PurchaseService) {
	purchaseController := controllers.NewPurchaseController(purchaseService)
	group.POST("/events/:id/purchase", purchaseController.Purchase)
}
