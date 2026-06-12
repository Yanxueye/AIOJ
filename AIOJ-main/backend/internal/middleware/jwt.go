package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

const (
	ctxUserID   = "x-user-id"
	ctxUsername = "x-username"
	ctxUserRole = "x-user-role"
)

// JWTAuth verifies the Authorization header (or ?token= query parameter as
// fallback for SSE/EventSource which cannot set custom headers) and stashes
// the user in the gin context.
func JWTAuth(mgr *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		if header := c.GetHeader("Authorization"); header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				utils.Unauthorized(c, "invalid authorization format")
				return
			}
			tokenStr = strings.TrimSpace(parts[1])
		} else {
			// Fallback: accept token from query parameter (for EventSource/SSE)
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			utils.Unauthorized(c, "missing authorization token")
			return
		}

		claims, err := mgr.Parse(tokenStr)
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxUserRole, claims.Role)
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

// CurrentUserRole extracts the authenticated user role from the context.
func CurrentUserRole(c *gin.Context) string {
	v, ok := c.Get(ctxUserRole)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// RequireAdmin only allows admin users through.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUserRole(c) != models.RoleAdmin {
			utils.Forbidden(c, "admin role required")
			return
		}
		c.Next()
	}
}

// RequireAdminDB verifies admin role by querying the database directly,
// preventing stale JWT claims from granting access after role demotion.
func RequireAdminDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := CurrentUserID(c)
		if !ok {
			utils.Unauthorized(c, "not authenticated")
			return
		}
		var user models.User
		if err := db.Select("role").First(&user, uid).Error; err != nil {
			utils.Unauthorized(c, "user not found")
			return
		}
		if user.Role != models.RoleAdmin {
			utils.Forbidden(c, "admin role required")
			return
		}
		// Update context with fresh role
		c.Set(ctxUserRole, user.Role)
		c.Next()
	}
}
