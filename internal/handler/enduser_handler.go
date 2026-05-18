package handler

import (
	"net/http"
	"strconv"

	"drone-management/internal/domain"
	mw "drone-management/internal/middleware"
	"drone-management/internal/service"

	"github.com/labstack/echo/v5"
)

type EnduserHandler struct {
	svc *service.OrderService
}

func NewEnduserHandler(svc *service.OrderService) *EnduserHandler {
	return &EnduserHandler{svc: svc}
}

func (h *EnduserHandler) Submit(c *echo.Context) error {
	var req SubmitOrderRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidInput
	}
	pid := mw.PrincipalID(c)
	order, err := h.svc.Submit(c.Request().Context(), service.SubmitOrderInput{
		EnduserPrincipalID: pid,
		OriginLat:          req.Origin.Lat,
		OriginLng:          req.Origin.Lng,
		DestLat:            req.Destination.Lat,
		DestLng:            req.Destination.Lng,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, OrderResponse{Order: order})
}

func (h *EnduserHandler) Get(c *echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	order, events, err := h.svc.GetOwn(c.Request().Context(), pid, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order, Timeline: events})
}

func (h *EnduserHandler) Withdraw(c *echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	pid := mw.PrincipalID(c)
	order, err := h.svc.Withdraw(c.Request().Context(), pid, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OrderResponse{Order: order})
}

func parseUintParam(c *echo.Context, name string) (uint, error) {
	raw := c.Param(name)
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, domain.ErrInvalidInput
	}
	return uint(n), nil
}
