package handler

import (
	"net/http"

	"drone-management/internal/domain"
	mw "drone-management/internal/middleware"
	"drone-management/internal/service"

	"github.com/labstack/echo/v5"
)

type DroneHandler struct {
	svc *service.DroneService
}

func NewDroneHandler(svc *service.DroneService) *DroneHandler {
	return &DroneHandler{svc: svc}
}

func (h *DroneHandler) ListJobs(c *echo.Context) error {
	var jt *domain.JobType
	if t := c.QueryParam("type"); t != "" {
		v := domain.JobType(t)
		if !v.Valid() {
			return domain.ErrInvalidInput
		}
		jt = &v
	}
	jobs, err := h.svc.ListOpenJobs(c.Request().Context(), jt)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"jobs": jobs})
}

func (h *DroneHandler) Reserve(c *echo.Context) error {
	jobID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	job, order, err := h.svc.ReserveJob(c.Request().Context(), pid, jobID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"job": job, "order": order})
}

func (h *DroneHandler) Pickup(c *echo.Context) error {
	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	order, err := h.svc.Pickup(c.Request().Context(), pid, orderID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order})
}

func (h *DroneHandler) Delivered(c *echo.Context) error {
	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	order, err := h.svc.MarkDelivered(c.Request().Context(), pid, orderID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order})
}

func (h *DroneHandler) Failed(c *echo.Context) error {
	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	order, err := h.svc.MarkFailed(c.Request().Context(), pid, orderID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order})
}

func (h *DroneHandler) Broken(c *echo.Context) error {
	pid := mw.PrincipalID(c)
	drone, err := h.svc.MarkSelfBroken(c.Request().Context(), pid)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"drone": drone})
}

func (h *DroneHandler) Heartbeat(c *echo.Context) error {
	var req HeartbeatRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidInput
	}
	pid := mw.PrincipalID(c)
	drone, order, err := h.svc.Heartbeat(c.Request().Context(), pid, req.Lat, req.Lng)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, HeartbeatResponse{Drone: drone, Order: order})
}

func (h *DroneHandler) Assigned(c *echo.Context) error {
	pid := mw.PrincipalID(c)
	drone, order, err := h.svc.AssignedOrder(c.Request().Context(), pid)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"drone": drone, "order": order})
}
