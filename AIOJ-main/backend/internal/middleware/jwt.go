package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/utils"
)

const (
	ctxUserID   = "x-user-id"
	ctxUsername = "x-username"
)

// JWTAuth verifies the Authorization header and stashes the user in the gin
// context. Requests without a valid token are short-circuited with 401.
func JWTAuth(mgr *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Unauthorized(c, "missing authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "invalid authorization format")
			return
		}
		claims, err := mgr.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Next()
	}
}

// CurrentUserID extracts the authenticated user id from the context. It
// returns (0, false) if the middleware was not executed.
func CurrentUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

// CurrentUsername extracts the authenticated username from the context.
func CurrentUsername(c *gin.Context) string {
	v, ok := c.Get(ctxUsername)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
