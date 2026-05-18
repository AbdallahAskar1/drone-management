package middleware

import (
	"net/http"
	"strings"

	"drone-management/internal/domain"
	"drone-management/internal/utils"

	"github.com/labstack/echo/v5"
)

const (
	CtxClaimsKey      = "claims"
	CtxPrincipalIDKey = "principal_id"
	CtxRoleKey        = "role"
)

func JWT(signer *utils.JWTSigner) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			h := c.Request().Header.Get("Authorization")
			if h == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(h, prefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid Authorization header")
			}
			tokenStr := strings.TrimSpace(h[len(prefix):])
			claims, err := signer.Parse(tokenStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			pid, err := utils.ClaimSubjectUint(claims)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
			}
			c.Set(CtxClaimsKey, claims)
			c.Set(CtxPrincipalIDKey, pid)
			c.Set(CtxRoleKey, claims.Role)
			return next(c)
		}
	}
}

func RequireRole(roles ...domain.Role) echo.MiddlewareFunc {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, ok := c.Get(CtxRoleKey).(domain.Role)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing role")
			}
			if _, ok := allowed[role]; !ok {
				return echo.NewHTTPError(http.StatusForbidden, "forbidden")
			}
			return next(c)
		}
	}
}

func PrincipalID(c *echo.Context) uint {
	v, _ := c.Get(CtxPrincipalIDKey).(uint)
	return v
}

func Role(c *echo.Context) domain.Role {
	v, _ := c.Get(CtxRoleKey).(domain.Role)
	return v
}
