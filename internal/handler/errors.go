package handler

import (
	"errors"
	"net/http"

	"drone-management/internal/domain"

	"github.com/labstack/echo/v5"
)

type ErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(c *echo.Context, status int, code, message string) error {
	body := ErrorBody{}
	body.Error.Code = code
	body.Error.Message = message
	return c.JSON(status, body)
}

func ErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		_ = writeError(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		_ = writeError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		_ = writeError(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		_ = writeError(c, http.StatusConflict, "invalid_transition", err.Error())
	case errors.Is(err, domain.ErrAlreadyReserved):
		_ = writeError(c, http.StatusConflict, "already_reserved", err.Error())
	case errors.Is(err, domain.ErrConflict):
		_ = writeError(c, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, domain.ErrDroneBusy):
		_ = writeError(c, http.StatusConflict, "drone_busy", err.Error())
	case errors.Is(err, domain.ErrDroneBroken):
		_ = writeError(c, http.StatusConflict, "drone_broken", err.Error())
	case errors.Is(err, domain.ErrUnauthenticated):
		_ = writeError(c, http.StatusUnauthorized, "unauthenticated", err.Error())
	default:
		var he *echo.HTTPError
		if errors.As(err, &he) {
			msg := he.Message
			if msg == "" {
				msg = http.StatusText(he.Code)
			}
			_ = writeError(c, he.Code, statusCodeName(he.Code), msg)
			return
		}
		_ = writeError(c, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

func statusCodeName(code int) string {
	switch code {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusBadRequest:
		return "bad_request"
	default:
		return "error"
	}
}
