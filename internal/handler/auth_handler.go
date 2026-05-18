package handler

import (
	"net/http"

	"drone-management/internal/domain"
	"drone-management/internal/service"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) IssueToken(c *echo.Context) error {
	var req TokenRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidInput
	}
	res, err := h.svc.IssueToken(c.Request().Context(), req.Name, req.Role)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, TokenResponse{
		Token:     res.Token,
		Principal: res.Principal,
	})
}
