package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const (
	HeaderXRequestID = "X-Request-ID"
	CtxRequestIDKey  = "request_id"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(HeaderXRequestID)
			if rid == "" {
				rid = uuid.New().String()
			}
			res.Header().Set(HeaderXRequestID, rid)
			c.Set(CtxRequestIDKey, rid)
			return next(c)
		}
	}
}

func GetRequestID(c *echo.Context) string {
	rid, _ := c.Get(CtxRequestIDKey).(string)
	return rid
}
