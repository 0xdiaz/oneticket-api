package routers

import (
	"github.com/0xdiaz/oneticket-api/internal/app/controllers"
	"github.com/0xdiaz/oneticket-api/internal/app/services"
	"github.com/gin-gonic/gin"
)

// RegisterEventRoutes registers event browsing routes under the given group
// (e.g. /api/v1). Full paths are groupPrefix/events and groupPrefix/events/:id.
//
// Browsing is public: buying is not, and lands with the checkout flow.
func RegisterEventRoutes(group *gin.RouterGroup, eventService *services.EventService) {
	eventController := controllers.NewEventController(eventService)
	group.GET("/events", eventController.List)
	group.GET("/events/:id", eventController.Get)
}
