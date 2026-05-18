package handler

import (
	"net/http"

	"drone-management/internal/domain"
	mw "drone-management/internal/middleware"
	"drone-management/internal/utils"

	"github.com/labstack/echo/v5"
)

type Handlers struct {
	Auth    *AuthHandler
	Enduser *EnduserHandler
	Drone   *DroneHandler
	Admin   *AdminHandler
}

func Register(e *echo.Echo, signer *utils.JWTSigner, h Handlers) {
	e.HTTPErrorHandler = ErrorHandler

	e.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/auth/token", h.Auth.IssueToken)

	jwt := mw.JWT(signer)

	enduser := e.Group("/orders", jwt, mw.RequireRole(domain.RoleEnduser))
	enduser.POST("", h.Enduser.Submit)
	enduser.GET("/:id", h.Enduser.Get)
	enduser.POST("/:id/withdraw", h.Enduser.Withdraw)

	drone := e.Group("/drone", jwt, mw.RequireRole(domain.RoleDrone))
	drone.GET("/jobs", h.Drone.ListJobs)
	drone.POST("/jobs/:id/reserve", h.Drone.Reserve)
	drone.POST("/orders/:id/pickup", h.Drone.Pickup)
	drone.POST("/orders/:id/delivered", h.Drone.Delivered)
	drone.POST("/orders/:id/failed", h.Drone.Failed)
	drone.POST("/self/broken", h.Drone.Broken)
	drone.POST("/self/heartbeat", h.Drone.Heartbeat)
	drone.GET("/self/order", h.Drone.Assigned)

	admin := e.Group("/admin", jwt, mw.RequireRole(domain.RoleAdmin))
	admin.GET("/orders", h.Admin.ListOrders)
	admin.PATCH("/orders/:id", h.Admin.PatchOrder)
	admin.GET("/drones", h.Admin.ListDrones)
	admin.POST("/drones/:id/broken", h.Admin.MarkBroken)
	admin.POST("/drones/:id/fixed", h.Admin.MarkFixed)
}
