package handler

import (
	"net/http"
	"strconv"
	"strings"

	"drone-management/internal/domain"
	mw "drone-management/internal/middleware"
	"drone-management/internal/repo"
	"drone-management/internal/service"

	"github.com/labstack/echo/v5"
)

type AdminHandler struct {
	admin  *service.AdminService
	drones *service.DroneService
}

func NewAdminHandler(a *service.AdminService, d *service.DroneService) *AdminHandler {
	return &AdminHandler{admin: a, drones: d}
}

func (h *AdminHandler) ListOrders(c *echo.Context) error {
	f := repo.ListFilter{}
	if idsRaw := c.QueryParam("ids"); idsRaw != "" {
		for part := range strings.SplitSeq(idsRaw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return domain.ErrInvalidInput
			}
			f.IDs = append(f.IDs, uint(n))
		}
	}
	if statusRaw := c.QueryParam("status"); statusRaw != "" {
		for part := range strings.SplitSeq(statusRaw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			f.Statuses = append(f.Statuses, domain.OrderStatus(part))
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n < 0 {
			return domain.ErrInvalidInput
		}
		f.Limit = n
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		n, err := strconv.Atoi(offsetStr)
		if err != nil || n < 0 {
			return domain.ErrInvalidInput
		}
		f.Offset = n
	}
	orders, err := h.admin.ListOrders(c.Request().Context(), f)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"orders": orders})
}

func (h *AdminHandler) PatchOrder(c *echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req PatchOrderRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidInput
	}
	in := service.PatchOrderInput{}
	if req.Origin != nil {
		in.OriginLat = &req.Origin.Lat
		in.OriginLng = &req.Origin.Lng
	}
	if req.Destination != nil {
		in.DestLat = &req.Destination.Lat
		in.DestLng = &req.Destination.Lng
	}
	actor := mw.PrincipalID(c)
	order, err := h.admin.PatchOrder(c.Request().Context(), id, actor, in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order})
}

func (h *AdminHandler) ListDrones(c *echo.Context) error {
	drones, err := h.admin.ListDrones(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"drones": drones})
}

func (h *AdminHandler) MarkBroken(c *echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	actor := mw.PrincipalID(c)
	drone, err := h.drones.AdminMarkBroken(c.Request().Context(), id, actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"drone": drone})
}

func (h *AdminHandler) MarkFixed(c *echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	actor := mw.PrincipalID(c)
	drone, err := h.drones.AdminMarkFixed(c.Request().Context(), id, actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"drone": drone})
}
